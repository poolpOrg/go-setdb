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

package ast

import (
	"fmt"

	"github.com/poolpOrg/go-setdb/query/lexer"
	"github.com/poolpOrg/go-setdb/sets"
)

type ResolvedSet struct {
	Name    string
	Pattern Node
}

func NewResolvedSet(name string, pattern Node) *ResolvedSet {
	return &ResolvedSet{
		Name:    name,
		Pattern: pattern,
	}
}

type Node interface {
	Evaluate(func(string) (*ResolvedSet, error)) (*sets.Set, error)
	ToQuery() string
}

type AssignExpr struct {
	Name string
	Expr Node
}

func (n AssignExpr) Evaluate(cb func(string) (*ResolvedSet, error)) (*sets.Set, error) {
	return n.Expr.Evaluate(cb)
}

func (n AssignExpr) ToQuery() string {
	return fmt.Sprintf("%s = %s", n.Name, n.Expr.ToQuery())
}

type BinaryExpr struct {
	Operator lexer.TokenType
	LHS      Node
	RHS      Node
}

func (n BinaryExpr) Evaluate(cb func(string) (*ResolvedSet, error)) (*sets.Set, error) {
	var op func(...*sets.Set) *sets.Set
	switch n.Operator {
	case lexer.UNION:
		op = sets.Union
	case lexer.INTERSECTION:
		op = sets.Intersection
	case lexer.DIFFERENCE:
		op = sets.Difference
	case lexer.SYMMETRIC_DIFFERENCE:
		op = sets.SymmetricDifference
	default:
		panic("unknown operation: " + n.Operator.String())
	}

	lhs, err := n.LHS.Evaluate(cb)
	if err != nil {
		return nil, err
	}
	rhs, err := n.RHS.Evaluate(cb)
	if err != nil {
		return nil, err
	}

	return op(lhs, rhs), nil
}

func (n BinaryExpr) ToQuery() string {
	return fmt.Sprintf("%s%s%s", n.LHS.ToQuery(), n.Operator.String(), n.RHS.ToQuery())
}

type Set struct {
	Name string
	Node []Node
}

func (n Set) Evaluate(cb func(string) (*ResolvedSet, error)) (*sets.Set, error) {
	if n.Name != "" {
		resolvedSet, err := cb(n.Name)
		if err != nil {
			return nil, err
		}
		return resolvedSet.Pattern.Evaluate(cb)
	}

	resolvedSets := make([]*sets.Set, 0)
	for _, item := range n.Node {
		results, err := item.Evaluate(cb)
		if err != nil {
			return nil, err
		}
		resolvedSets = append(resolvedSets, results)
	}
	return sets.Union(resolvedSets...), nil
}

func (n Set) ToQuery() string {
	name := n.Name
	if name == "" {
		buf := "{"
		for i, item := range n.Node {
			buf += item.ToQuery()
			if i != len(n.Node)-1 {
				buf += ","
			}
		}
		buf += "}"
		return buf

	} else {
		return name
	}
}

type Item struct {
	Name string
}

func (n Item) Evaluate(cb func(string) (*ResolvedSet, error)) (*sets.Set, error) {
	return sets.NewSet(n.Name), nil
}

func (n Item) ToQuery() string {
	return n.Name
}
