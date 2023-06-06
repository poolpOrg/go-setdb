/*
 * Copyright (c) 2023 Gilles Chehade <gilles@poolp.org>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package parser

import (
	"fmt"

	"github.com/poolpOrg/go-setdb/query/ast"
	"github.com/poolpOrg/go-setdb/query/lexer"
)

var binopPrecedence = map[lexer.TokenType]int{
	lexer.UNION:                10,
	lexer.INTERSECTION:         10,
	lexer.DIFFERENCE:           10,
	lexer.SYMMETRIC_DIFFERENCE: 10,
}

func getTokenPrecedence(tokenType lexer.TokenType) int {
	if value, exists := binopPrecedence[tokenType]; !exists {
		return -1
	} else {
		return value
	}
}

type ParserError struct {
	token lexer.Token
	msg   string
}

func (e ParserError) Error() string {
	token := e.token
	msg := e.msg
	output := fmt.Sprintf("line %d, column %d: %s, got: %s", token.Position().Line(), token.Position().Column(), msg, token.Type())
	if token.Type().String() != token.Value() {
		output += fmt.Sprintf(" (%s)", token.Value())
	}
	return output
}

func (e ParserError) Line() int {
	return e.token.Position().Line()
}

func (e ParserError) Column() int {
	return e.token.Position().Column()
}

func ParseError(token lexer.Token, format string, args ...any) ParserError {
	output := fmt.Sprintf(format, args...)
	output = fmt.Sprintf("[%d:%d] %s, got: %s", token.Position().Line(), token.Position().Column(), output, token.Type())
	if token.Type().String() != token.Value() {
		output += fmt.Sprintf(" (%s)", token.Value())
	}
	return ParserError{token: token, msg: fmt.Sprintf(format, args...)}
}

type Parser struct {
	lexer *lexer.Lexer

	readAhead *lexer.Token
}

func NewParser(l *lexer.Lexer) *Parser {
	return &Parser{lexer: l, readAhead: nil}
}

func (p *Parser) peekToken() lexer.Token {
	var token lexer.Token

	if p.readAhead != nil {
		token = *p.readAhead
	} else {
		token = p.lexer.Lex()
		p.readAhead = &token
	}
	return token
}

func (p *Parser) readToken() lexer.Token {
	var token lexer.Token

	if p.readAhead != nil {
		token = *p.readAhead
		p.readAhead = nil
	} else {
		token = p.lexer.Lex()
	}
	return token
}

func (p *Parser) Parse() (ast.Node, error) {
	parsedAST, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	token := p.peekToken()
	if token.Type() != lexer.EOF {
		return nil, ParseError(token, "expected EOF")
	}
	return parsedAST, nil
}

/* LITERAL NODES */
func (p *Parser) parseInlineSet() (ast.Node, error) {
	token := p.readToken()
	if token.Type() != lexer.SET_OPEN {
		return nil, ParseError(token, "expected '{'")
	}

	items := make([]ast.Node, 0)
	for {
		token = p.peekToken()
		if token.Type() == lexer.EOF || token.Type() == lexer.SET_CLOSE {
			break
		}
		item, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		items = append(items, item)

		token = p.peekToken()
		if token.Type() != lexer.SET_CLOSE {
			if token.Type() != lexer.COMMA {
				return nil, ParseError(token, "expected ','")
			}
			token = p.readToken()
		}

	}
	if token.Type() != lexer.SET_CLOSE {
		return nil, ParseError(token, "expected '}'")
	}
	p.readToken()

	return &ast.Set{Node: items}, nil
}

func (p *Parser) parseAssign() (ast.Node, error) {
	token := p.readToken()
	if token.Type() != lexer.ASSIGN {
		return nil, ParseError(token, "expected set name")
	}
	return p.parseExpr()
}

func (p *Parser) parseSet() (ast.Node, error) {
	token := p.readToken()
	if token.Type() != lexer.SET {
		return nil, ParseError(token, "expected set name")
	}
	name := token.Value()

	token = p.peekToken()
	if token.Type() == lexer.ASSIGN {
		expr, err := p.parseAssign()
		if err != nil {
			return nil, err
		}
		return &ast.AssignExpr{Name: name, Expr: expr}, nil
	}

	return &ast.Set{Name: name}, nil
}

func (p *Parser) parseItem() (ast.Node, error) {
	token := p.readToken()
	if token.Type() != lexer.ITEM {
		return nil, ParseError(token, "expected item name")
	}

	return &ast.Item{Name: token.Value()}, nil
}

/* Expr NODES */

func (p *Parser) parseExprPrimary() (ast.Node, error) {
	token := p.peekToken()
	if token.Type() == lexer.SET {
		return p.parseSet()
	} else if token.Type() == lexer.ITEM {
		return p.parseItem()
	} else if token.Type() == lexer.SET_OPEN {
		return p.parseInlineSet()
	} else {
		return nil, ParseError(token, "unexpected token %s", token.Type())
	}
}

func (p *Parser) parseExprBinOpRHS(precedence int, LHS ast.Node) (ast.Node, error) {
	for {
		curToken := p.peekToken()
		curPrecedence := getTokenPrecedence(curToken.Type())

		// either not a binop OR has lower precedence than last binop
		if curPrecedence < precedence {
			return LHS, nil
		}

		// consume binop
		binOp := p.readToken()

		RHS, err := p.parseExprPrimary()
		if err != nil {
			return nil, err
		}

		nextToken := p.peekToken()
		nextPrecedence := getTokenPrecedence(nextToken.Type())
		if curPrecedence < nextPrecedence {
			RHS, err = p.parseExprBinOpRHS(curPrecedence+1, RHS)
			if err != nil {
				return nil, err
			}
		}
		// MERGE LHS
		LHS = &ast.BinaryExpr{
			Operator: binOp.Type(),
			LHS:      LHS,
			RHS:      RHS,
		}
	}
}

func (p *Parser) parseExpr() (ast.Node, error) {
	LHS, err := p.parseExprPrimary()
	if err != nil {
		return nil, err
	}

	return p.parseExprBinOpRHS(0, LHS)
}
