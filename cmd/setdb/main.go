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

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/poolpOrg/go-setdb"
	_ "github.com/poolpOrg/go-setdb/storage/sqlite"
)

var globalDatabasesMutex = sync.Mutex{}
var database = make(map[string]*setdb.Database)
var databaseMutex = make(map[string]*sync.Mutex)

func openDatabase(name string) (*setdb.Database, error) {
	globalDatabasesMutex.Lock()
	defer globalDatabasesMutex.Unlock()

	if conn, exists := database[name]; exists {
		databaseMutex[name].Lock()
		return conn, nil
	} else {
		conn, err := setdb.Open("sqlite", name)
		if err != nil {
			return nil, err
		}
		database[name] = conn
		databaseMutex[name] = &sync.Mutex{}
		databaseMutex[name].Lock()
		return conn, nil
	}
}

func closeDatabase(db *setdb.Database) error {
	if _, exists := database[db.Name()]; !exists {
		return fmt.Errorf("database connection for %s does not exist", db.Name())
	} else {
		databaseMutex[db.Name()].Unlock()
		return nil
	}
}

func getDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["dbname"]

	db, err := openDatabase(dbname)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer closeDatabase(db)

	sets, err := db.List()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	json.NewEncoder(w).Encode(&sets)
}

type Query struct {
	Expression string `json:"expression"`
}

func postDatabaseQueryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbname := vars["dbname"]

	var q Query
	err := json.NewDecoder(r.Body).Decode(&q)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	db, err := openDatabase(dbname)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	defer closeDatabase(db)

	set, err := db.Query(q.Expression)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	items := set.Items()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(items)
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/database/{dbname}", getDatabaseHandler).Methods("GET")
	r.HandleFunc("/database/{dbname}", postDatabaseQueryHandler).Methods("POST")

	http.ListenAndServe("0.0.0.0:3031", r)
}
