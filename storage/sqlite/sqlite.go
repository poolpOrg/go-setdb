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

package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/poolpOrg/go-setdb"
)

type backend struct {
	conn   *sql.DB
	dbname string
}

func init() {
	setdb.Register("sqlite", newBackend)
}

func newBackend(name string) setdb.Backend {
	conn, err := sql.Open("sqlite3", "/tmp/"+name+".db")
	if err != nil {
		panic(err)
	}

	const createTableSets string = `
			CREATE TABLE IF NOT EXISTS sets (
				id INTEGER NOT NULL PRIMARY KEY,
				ctime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				mtime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				name char(255) UNIQUE NOT NULL,
				uuid  char(36) UNIQUE NOT NULL DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
				dependsOn TEXT NOT NULL,
				pattern TEXT DEFAULT ''
			);
			`
	_, err = conn.Exec(createTableSets)
	if err != nil {
		panic(err)
	}

	return &backend{
		conn:   conn,
		dbname: name,
	}
}

func (bck *backend) Close() error {
	return bck.conn.Close()
}

func (bck *backend) Info(name string) (setdb.SetInfo, error) {

	stmt, err := bck.conn.Prepare(`SELECT name, uuid, ctime, mtime, dependsOn FROM sets WHERE name=?`)
	if err != nil {
		return setdb.SetInfo{}, err
	}
	defer stmt.Close()

	res, err := stmt.Query(name)
	if err != nil {
		return setdb.SetInfo{}, err
	}
	defer res.Close()

	if res.Next() {
		var name string
		var uid uuid.UUID
		var ctime time.Time
		var mtime time.Time
		var dependsOnSerialized []byte

		err = res.Scan(&name, &uid, &ctime, &mtime, &dependsOnSerialized)
		if err != nil {
			fmt.Println(err)
			return setdb.SetInfo{}, err
		}

		var dependsOn []string
		err = json.Unmarshal(dependsOnSerialized, &dependsOn)
		if err != nil {
			fmt.Println(err)
			return setdb.SetInfo{}, err
		}

		return setdb.SetInfo{
			Name:      name,
			Uuid:      uid,
			Ctime:     ctime,
			Mtime:     mtime,
			DependsOn: dependsOn,
		}, nil
	}

	return setdb.SetInfo{}, err
}

func (bck *backend) List() ([]setdb.SetInfo, error) {

	res, err := bck.conn.Query(`SELECT name, uuid, ctime, mtime, dependsOn FROM sets`)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	resultSet := make([]setdb.SetInfo, 0)
	for res.Next() {
		var name string
		var uid uuid.UUID
		var ctime time.Time
		var mtime time.Time
		var dependsOnSerialized []byte

		err = res.Scan(&name, &uid, &ctime, &mtime, &dependsOnSerialized)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		var dependsOn []string
		err = json.Unmarshal(dependsOnSerialized, &dependsOn)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		resultSet = append(resultSet, setdb.SetInfo{
			Name:      name,
			Uuid:      uid,
			Ctime:     ctime,
			Mtime:     mtime,
			DependsOn: dependsOn,
		})

	}

	return resultSet, nil
}

func (bck *backend) Persist(name string, pattern string, dependencies []string) error {
	stmt, err := bck.conn.Prepare(`INSERT OR REPLACE INTO sets (mtime, name, pattern, dependsOn) VALUES(CURRENT_TIMESTAMP, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	deps, err := json.Marshal(dependencies)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(name, pattern, deps)

	return err
}

func (bck *backend) Pattern(name string) (string, error) {
	stmt, err := bck.conn.Prepare(`SELECT pattern FROM sets WHERE name=?`)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	res, err := stmt.Query(name)
	if err != nil {
		return "", err
	}
	defer res.Close()

	if !res.Next() {
		return "", fmt.Errorf("set %s does not exist", name)
	}
	var template string
	err = res.Scan(&template)
	if err != nil {
		return "", err
	}

	return template, nil
}
