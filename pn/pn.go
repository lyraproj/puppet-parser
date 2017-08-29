package pn

import (
  . "fmt"
  . "bytes"
)

type (
  // Roughly Polish notation of an expression, i.e. an operand in string form
  // and Operator()s. Suitable for simple presentation formats such as Clojure or JSON/YAML
  // capable of representing boolean, integer, float, string, hash, and array without
  // type loss.
  //
  PN interface {
    // Produces a compact, LISP like syntax. Suitable for tests.
    Format(b *Buffer)

    // Produces an object that where all values are of primitive type or
    // slices or maps. This format is suitable for output as JSON or YAML
    //
    ToData() interface{}

    // Turn this PN into an argument in a call or change the name if the PN Is a call already
    AsCall(name string) PN

    // Return the PN as a parameter list
    AsParameters() []PN

    // Create a key/value pair from the given name and this PN
    WithName(name string) Entry
  }

  // Entry in hash
  Entry interface {
    Key() string
    Value() PN
  }

  mapEntry struct {
    key string
    value PN
  }

  listPN struct {
    elements []PN
  }

  mapPN struct {
    entries []Entry
  }

  literalPN struct {
    val interface{}
  }

  callNamedPN struct {
    mapPN
    name string
  }

  callPN struct {
    listPN
    name string
  }
)

func ListPN(elements []PN) PN {
  return &listPN{elements}
}

func MapPN(entries []Entry) PN {
  return &mapPN{entries}
}

func LiteralPN(val interface{}) PN {
  return &literalPN{val}
}

func CallPN(name string, elements ...PN) PN {
  return &callPN{listPN{elements}, name}
}

func CallNamedPN(name string, entries ...Entry) PN {
  return &callNamedPN{mapPN{entries}, name}
}

func (pn *listPN) AsCall(name string) PN {
  return CallPN(name, pn.elements...)
}

func (pn *listPN) AsParameters() []PN {
  return pn.elements
}

func (pn *listPN) Format(b *Buffer) {
  b.WriteByte('[')
  formatElements(pn.elements, b)
  b.WriteByte(']')
}

func (pn *listPN) ToData() interface{} {
  me := make([]interface{}, len(pn.elements))
  for idx, op := range pn.elements {
    me[idx] = op.ToData()
  }
  return me
}

func (pn *listPN) WithName(name string) Entry {
  return &mapEntry{name, pn}
}

func (pn *callPN) AsCall(name string) PN {
  return &callPN{listPN{pn.elements}, name}
}

func (pn *callPN) AsParameters() []PN {
  return pn.elements
}

func (pn *callPN) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  if len(pn.elements) > 0 {
    b.WriteByte(' ')
    formatElements(pn.elements, b)
  }
  b.WriteByte(')')
}

func (pn *callPN) ToData() interface{} {
  return []interface{} { pn.name, pn.listPN.ToData() }
}

func (pn *callPN) WithName(name string) Entry {
  return &mapEntry{name, pn}
}

func (pn *callNamedPN) AsCall(name string) PN {
  return &callNamedPN{mapPN{pn.entries}, name}
}

func (pn *callNamedPN) AsParameters() []PN {
  return pn.mapPN.AsParameters()
}

func (pn *callNamedPN) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  pn.mapPN.Format(b)
  b.WriteByte(')')
}

func (pn *callNamedPN) ToData() interface{} {
  return []interface{} { pn.name, pn.mapPN.ToData() }
}

func (pn *callNamedPN) WithName(name string) Entry {
  return &mapEntry{name, pn}
}

func (e *mapEntry) Key() string {
  return e.key
}

func (e *mapEntry) Value() PN {
  return e.value
}

func (pn *mapPN) AsCall(name string) PN {
  return &callNamedPN{mapPN{pn.entries}, name}
}

func (pn *mapPN) AsParameters() []PN {
  return []PN { pn }
}

func (pn *mapPN) Format(b *Buffer) {
  b.WriteByte('{')
  if top := len(pn.entries); top > 0 {
    formatEntry(pn.entries[0], b)
    for idx := 1; idx < top; idx++ {
      b.WriteByte(' ')
      formatEntry(pn.entries[idx], b)
    }
  }
  b.WriteByte('}')
}

func formatEntry(entry Entry, b *Buffer) {
  b.WriteByte(':')
  b.WriteString(entry.Key())
  b.WriteByte(' ')
  entry.Value().Format(b)
}

func (pn *mapPN) ToData() interface{} {
  me := make(map[string]interface{}, len(pn.entries))
  for _, entry := range pn.entries {
    me[entry.Key()] = entry.Value().ToData()
  }
  return me
}

func (pn *mapPN) WithName(name string) Entry {
  return &mapEntry{name, pn}
}

func (pn *literalPN) AsCall(name string) PN {
  return CallPN(name, pn)
}

func (pn *literalPN) AsParameters() []PN {
  return []PN { pn }
}

func (pn *literalPN) Format(b *Buffer) {
  switch pn.val.(type) {
  case string:
    DoubleQuote(pn.val.(string), b)
  case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
    Fprintf(b, `%d`, pn.val)
  case float32, float64:
    Fprintf(b, `%g`, pn.val)
  case bool:
    Fprintf(b, `%t`, pn.val)
  default:
    Fprintf(b, `%v`, pn.val)
  }
}

func (pn *literalPN) ToData() interface{} {
  return pn.val
}

func (pn *literalPN) WithName(name string) Entry {
  return &mapEntry{name, pn}
}

func formatElements(elements []PN, b *Buffer) {
  top := len(elements)
  if top > 0 {
    elements[0].Format(b)
    for idx := 1; idx < top; idx++ {
      b.WriteByte(' ')
      elements[idx].Format(b)
    }
  }
}
