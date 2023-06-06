# go-setdb

WIP: this is a work in progress, do not use or you'll feel sorry when you lose data


## What is SetDB ?

SetDB is a database to manage sets,
as in set theory,
and which provides a DSL to query the database for specific set operations.


## How does it work ?

SetDB manipulates two kind of sets:
persistent sets and transient sets,
the former being persisted across queries and the latter existing solely as a result set.

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

## TBD





