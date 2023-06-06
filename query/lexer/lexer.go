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

package lexer

import (
	"bufio"
	"io"
	"unicode"
)

type TokenType int

const (
	EOF = iota
	ILLEGAL

	COMMA

	SET
	ITEM

	ASSIGN

	UNION                // |
	INTERSECTION         // &
	DIFFERENCE           // -
	SYMMETRIC_DIFFERENCE // ^

	SET_OPEN
	SET_CLOSE
)

var tokens = []string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	SET:     "SET",
	ITEM:    "ITEM",

	COMMA: ",",

	ASSIGN: "=",

	// Infix ops

	UNION:                "|",
	INTERSECTION:         "&",
	DIFFERENCE:           "-",
	SYMMETRIC_DIFFERENCE: "^",

	SET_OPEN:  "{",
	SET_CLOSE: "}",
}

func (t TokenType) String() string {
	return tokens[t]
}

type Position struct {
	line   int
	column int
}

func (p *Position) Line() int {
	return p.line
}

func (p *Position) Column() int {
	return p.column
}

type Token struct {
	tokenType     TokenType
	tokenPosition *Position
	tokenValue    string
}

func (t *Token) Type() TokenType {
	return t.tokenType
}

func (t *Token) Position() *Position {
	return t.tokenPosition
}

func (t *Token) Value() string {
	return t.tokenValue
}

func tokenFromLexer(tokenType TokenType, tokenPosition Position, tokenValue string) Token {
	return Token{
		tokenType:     tokenType,
		tokenPosition: &tokenPosition,
		tokenValue:    tokenValue,
	}
}

type Lexer struct {
	reader *bufio.Reader
	pos    Position
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewReader(reader),
		pos:    Position{line: 1, column: 0},
	}
}

func (l *Lexer) Lex() Token {

	// keep looping until we return a token
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return tokenFromLexer(EOF, l.pos, "")
			}

			// at this point there isn't much we can do, and the compiler
			// should just return the raw error to the user
			panic(err)
		}
		l.pos.column++

		switch r {
		case '\n':
			l.resetPosition()

		case ',':
			return tokenFromLexer(COMMA, l.pos, ",")

		case '|':
			return tokenFromLexer(UNION, l.pos, "|")
		case '&':
			return tokenFromLexer(INTERSECTION, l.pos, "&")
		case '-':
			return tokenFromLexer(DIFFERENCE, l.pos, "-")
		case '^':
			return tokenFromLexer(SYMMETRIC_DIFFERENCE, l.pos, "^")

		case '{':
			return tokenFromLexer(SET_OPEN, l.pos, "(")
		case '}':
			return tokenFromLexer(SET_CLOSE, l.pos, ")")

		case '=':
			return tokenFromLexer(ASSIGN, l.pos, "=")

		case '\'':
			startPos := l.pos
			l.backup()
			lit := l.lexItem()
			return tokenFromLexer(ITEM, startPos, lit)

		default:
			if unicode.IsSpace(r) {
				continue // nothing to do here, just move on
			} else if unicode.IsLetter(r) {
				// backup and let lexIdent rescan the beginning of the ident
				startPos := l.pos
				l.backup()
				lit := l.lexIdent()
				return tokenFromLexer(SET, startPos, lit)
			} else if unicode.IsDigit(r) {
				// backup and let lexIdent rescan the beginning of the ident
				startPos := l.pos
				l.backup()
				lit := l.lexIdent()
				return tokenFromLexer(ITEM, startPos, lit)
			} else {
				return tokenFromLexer(ILLEGAL, l.pos, string(r))
			}
		}
	}
}

func (l *Lexer) resetPosition() {
	l.pos.line++
	l.pos.column = 0
}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}

	l.pos.column--
}

func (l *Lexer) lexIdent() string {
	var lit string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// at the end of the identifier
				return lit
			}
		}

		l.pos.column++
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ':' {
			lit = lit + string(r)
		} else {
			// scanned something not in the identifier
			l.backup()
			return lit
		}
	}
}

func (l *Lexer) lexItem() string {
	var lit string
	r, _, err := l.reader.ReadRune()
	if err != nil || (r != '"' && r != '\'') {
		// not a string
		return ""
	}
	lit = lit + string(r)

	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// at the end of the string
				return lit
			}
		}

		l.pos.column++
		if r != '"' && r != '\'' {
			lit = lit + string(r)
		} else {
			lit = lit + string(r)
			// scanned the end quote of the string
			return lit
		}
	}
}
