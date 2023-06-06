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

package setdb

import (
	"fmt"

	"strings"

	"log"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/poolpOrg/go-setdb/query/ast"
	"github.com/poolpOrg/go-setdb/query/lexer"
	"github.com/poolpOrg/go-setdb/query/parser"
	"github.com/poolpOrg/go-setdb/sets"
)

type Backend interface {
	List() ([]SetInfo, error)
	Info(string) (SetInfo, error)

	Persist(name string, pattern string, dependencies []string) error
	Pattern(name string) (string, error)

	Close() error
}

var muBackends sync.Mutex
var backends map[string]func(string) Backend = make(map[string]func(string) Backend)

type Database struct {
	backend Backend
	name    string
}

type Set struct {
	items *sets.Set

	patternAST ast.Node
	name       string
	database   *Database

	dependsOn []string
}

func (db *Database) Query(pattern string) (*Set, error) {

	queryParser := parser.NewParser(lexer.NewLexer(strings.NewReader(pattern)))
	queryAST, err := queryParser.Parse()
	if err != nil {
		return nil, err
	}

	name := ""
	if node, ok := queryAST.(*ast.AssignExpr); ok {
		name = node.Name
		queryAST = node.Expr
	}

	dependencies := make([]string, 0)
	setResolver := func(_name string) (*ast.ResolvedSet, error) {
		if name == _name {
			return nil, fmt.Errorf("cyclic reference is forbidden")
		}

		subpattern, err := db.backend.Pattern(_name)
		if err != nil {
			return nil, err
		}
		subqueryParser := parser.NewParser(lexer.NewLexer(strings.NewReader(subpattern)))
		subqueryAST, err := subqueryParser.Parse()
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, _name)
		return ast.NewResolvedSet(name, subqueryAST), nil
	}

	resultset, err := queryAST.Evaluate(setResolver)
	if err != nil {
		return nil, err
	}

	if name != "" {
		err = db.backend.Persist(name, queryAST.ToQuery(), dependencies)
		if err != nil {
			return nil, err
		}
	}

	return &Set{
		items:      resultset,
		name:       name,
		database:   db,
		patternAST: queryAST,
		dependsOn:  dependencies,
	}, err
}

type SetInfo struct {
	Name      string    `json:"name"`
	Uuid      uuid.UUID `json:"uuid"`
	Ctime     time.Time `json:"ctime"`
	Mtime     time.Time `json:"mtime"`
	DependsOn []string  `json:"dependsOn"`
}

func Register(backendName string, backend func(string) Backend) {
	muBackends.Lock()
	defer muBackends.Unlock()
	if _, ok := backends[backendName]; ok {
		log.Fatalf("backend %s registered twice", backendName)
	}
	backends[backendName] = backend
}

func Backends() []string {
	muBackends.Lock()
	defer muBackends.Unlock()

	ret := make([]string, 0)
	for backendName := range backends {
		ret = append(ret, backendName)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}

func Open(backendName string, dbname string) (*Database, error) {
	muBackends.Lock()
	defer muBackends.Unlock()

	if backend, exists := backends[backendName]; !exists {
		return nil, fmt.Errorf("backend %s does not exist", backendName)
	} else {
		database := &Database{}
		database.name = dbname
		database.backend = backend(dbname)
		return database, nil
	}
}

func (db *Database) Close() error {
	return db.backend.Close()
}

func (db *Database) Name() string {
	return db.name
}

func (db *Database) List() ([]SetInfo, error) {
	return db.backend.List()
}

func (db *Database) Info(name string) (SetInfo, error) {
	return db.backend.Info(name)
}

func (s *Set) Pattern() string {
	return s.patternAST.ToQuery()
}

func (s *Set) Items() []string {
	return s.items.ItemsList()
}
