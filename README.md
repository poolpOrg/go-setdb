# go-setdb

WIP: this is a work in progress, do not use or you'll feel sorry when you lose data

## Should I use it ?

nope.


## What is SetDB ?

SetDB is a database to manage sets,
as in set theory,
and which provides a DSL to query the database for specific set operations.


## What's the license for SetDB ?

SetDB is published under the ISC license,
do what you want with it but keep the copyright in place and don't complain if code blows up.

```
Copyright (c) 2023 Gilles Chehade <gilles@poolp.org>

Permission to use, copy, modify, and distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
```

## But doesn't solution X, Y, Z already do that ?

Yes,
you can technically use several solutions ranging from full-blown SQL databases to data structure servers,
however they are not necessarily all very practical for the use-cases that I have.
Also,
I like writing code so sometimes I do it just because.


## How does it work ?

SetDB manipulates two kind of sets:
persistent sets and transient sets,
the former being persisted across queries and the latter existing solely as a result set.

It provides a very basic query language,
currently only supporting operations that return result sets (union, intersection, difference, symmetric difference).
The query language allows the creation of new sets but isn't complete yet and doesn't cover operations not returning sets (subset of, superset of, ...).


```sh
$ setdb-cli
setdb> x
ERR: set x does not exist
setdb> x = {}
[]
setdb> x
[]
setdb> {1, 2, 3} & {3}
[3]
setdb> {1, 2, 3} | {4}
[3 4 2 1]
setdb> {1, 2, 3} - {1}
[2 3]
setdb> {1, 2, 3} ^ {1}
[2 3]
setdb> x = {1, 2, 3} & {2, 3} | 4 ^ 2
[3 4]
setdb> x
[3 4]
setdb>
```

Sets are handled as patterns, allowing the inclusion of other sets and dynamic resolving:
```sh
setdb> y = {1, 2, 3}
[2 3 1]
setdb> x = y
[2 1 3]
setdb> y = {1, 2, 3, 4}
[1 2 3 4]
setdb> x
[1 2 3 4]
setdb> z = {1, 2, 5 }
[1 2 5]
setdb> x = y & z
[2 1]
setdb> x = {x | 1}
ERR: cyclic reference is forbidden
setdb> a = {1}
[1]
setdb> b = a
[1]
setdb> c = b
[1]
setdb> a = c
ERR: cyclic reference is forbidden
setdb>
```

They are not typed and can contain integers and strings at this point,
I'll decide later if I want to be more strict but I don't see a value.
```sh
setdb> fruits = {'grape', 'orange', 'strawberry'}
['grape' 'orange' 'strawberry']
setdb> vegetables = {'spinash', 'onions'}
['spinash' 'onions']
setdb> healthy = {fruits | vegetables}
['grape' 'orange' 'spinash' 'onions' 'strawberry']
setdb> gross = {'onions'}
['onions']
setdb> healthy
['onions' 'orange' 'strawberry' 'grape' 'spinash']
setdb> healthy - gross
['orange' 'strawberry' 'spinash' 'grape']
setdb>
```

## What's missing ?

- code cleanup
- do a pass to decide on final syntax for the DSL
- implement set dereference so a set can contain the content of another set, not the other set itself (ie: `x = {*y}`)
- implement various caching strategies (some were implemented but temporarily removed)
- disk and memory optimizations have been discussed, they are just not implemented yet



## Code

For now, the code is a moving target, so unless you know what you're doing, don't import it.

Otherwise, look atht the example implementations in cmd/,
one implements a server and the other a command line tool that also ships a client.


## Special thanks
This project was worked on partly during my spare time and partly during my work time,
at [VeepeeTech](https://github.com/veepee-oss),
with many insights from [@aromeyer](https://github.com/aromeyer) who had to endure multiple duck-debugging sessions.


