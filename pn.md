# The PN (Pops Notation) format

### Objective

The primary objective for this format is to provide a short but precise
human readable format of the parser AST. This format is needed to write
tests that verifies the parsers function.

Another objective is to provide easy transport of the AST using JSON
or similar protocol that are confined to boolean, number, string,
undef, array, and object. The PN conversion to `Data` was specifically
designed with this objective in mind.

### Format
A PN forms a directed acyclig graph of nodes. There are five types of nodes:

* Literal: A boolean, integer, float, string, or undef
* List: An ordered list of nodes
* Map: An unordered map of string to node associations
* Call: A named list of nodes.

### PN represented as Data

A PN can represent itself as the Puppet `Data` type and will then use
the following conversion rules:

PN Node | Data type
--------|----------
`Literal` | `Boolean`, `Integer`, `Float`, `String`, and `Undef`
`List` | `Array[Data]`
`Map` | `Hash[Pattern[/^[A-Za-z_-][0-9A-Za-z_-]*$/, `Data`]
`Call` | `Struct['^',Tuple[String,Data,1]]`

### PN represented as String

The native string representation of a PN is similar to Clojure.

PN Node | Sample string representation
--------|----------
`Literal boolean` | `true`, `false`
`Literal integer` | `834`, `-123`, `0`
`Literal float` | `32.28`, `-1.0`, `33.45e18`
`Literal string` | `"plain"`, `"quote \\""`, `"tab \\t"`, `"return \\r"`, `"newline \\n"`, `"control \\u{14}"`
`Literal undef` | `nil`
`List` | `["a" "b" 32 true]`
`Map` | `{:a 2 :b 3 :c true}`
`Call` | `(myFunc 1 2 "b")`

### PN represented as JSON or YAML

When representing PN as JSON or YAML it must first be converted to `Data`. For JSON, this
means that literals are represented verbatim, lists will be a JSON arrays, maps will be
a JSON objects. The only thing that is a bit special is the `Call` which becomes a JSON
object similar to: `{ "^": [ "myFunc", 1, 2, "b" ] }`


