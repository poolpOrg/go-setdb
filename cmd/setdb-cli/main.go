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
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/poolpOrg/go-setdb"
	_ "github.com/poolpOrg/go-setdb/storage/sqlite"
)

type Query struct {
	Expression string `json:"expression"`
}

func main() {
	var databaseName string
	var serverURL string
	var useStdin bool

	flag.StringVar(&serverURL, "server", "", "server URL")
	flag.StringVar(&databaseName, "database", "default", "database name")
	flag.Parse()

	if flag.NArg() < 1 {
		useStdin = true
	} else {
		useStdin = false
	}

	if serverURL == "" {
		db, err := setdb.Open("sqlite", databaseName)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		if !useStdin {
			expression := flag.Arg(0)
			set, err := db.Query(expression)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			} else {
				fmt.Println(set.Items())
			}
		} else {
			fmt.Printf("setdb> ")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				if scanner.Text() == "quit" {
					break
				}
				set, err := db.Query(scanner.Text())
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
				} else {
					fmt.Println(set.Items())
				}
				fmt.Printf("setdb> ")
			}

			if scanner.Err() != nil {
				panic(err)
			}
		}

	} else {
		var q Query

		if !useStdin {
			expression := flag.Arg(0)
			q.Expression = expression

			serializedQuery, err := json.Marshal(&q)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			}

			res, err := http.Post(fmt.Sprintf("%s/database/%s", serverURL, databaseName), "application/json", bytes.NewReader(serializedQuery))
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			}
			//fmt.Println(res.StatusCode)
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERR: %s\n", err)

			}

			var result []string
			err = json.Unmarshal(body, &result)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			} else {
				fmt.Println(result)
			}
		} else {
			fmt.Printf("setdb> ")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				if scanner.Text() == "quit" {
					break
				}

				expression := scanner.Text()
				q.Expression = expression
				serializedQuery, err := json.Marshal(&q)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
				}

				res, err := http.Post(fmt.Sprintf("%s/database/%s", serverURL, databaseName), "application/json", bytes.NewReader(serializedQuery))
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
				}
				//fmt.Println(res.StatusCode)
				body, err := ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %s\n", err)

				}
				var result []string
				err = json.Unmarshal(body, &result)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
				} else {
					fmt.Println(result)
				}
				fmt.Printf("setdb> ")
			}

			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "ERR: %s\n", err)
			}
		}

	}
}
