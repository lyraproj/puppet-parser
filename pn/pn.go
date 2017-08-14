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

    // Give this PN a name, or change the name if it already has one
    WithName(name string) Entry
  }

  // Entry in hash
  Entry interface {
    PN
    formatEntry(b *Buffer)
    key() string
    value() interface{}
  }

  hash struct {
    entries []Entry
  }

  array struct {
    elements []PN
  }

  stringValue struct {
    val string
  }

  literal struct {
    val interface{}
  }

  namedArray struct {
    name string
    elements []PN
  }

  namedValue struct {
    name string
    val PN
  }

  namedStringValue struct {
    name string
    val string
  }
)

func Array(elements []PN) PN {
  return &array{elements}
}

func Hash(entries []Entry) PN {
  return &hash{entries}
}

func StringValue(val string) PN {
  return &stringValue{val}
}

func Literal(val interface{}) PN {
  return &literal{val}
}

func NamedValue(name string, val PN) Entry {
  return &namedValue{name, val}
}

func NamedArray(name string, elements []PN) Entry {
  return &namedArray{name, elements}
}

func NamedString(name string, val string) Entry {
  return &namedStringValue{name, val}
}

func (pn *array) Format(b *Buffer) {
  b.WriteByte('[')
  formatElements(pn.elements, b)
  b.WriteByte(']')
}

func (pn *array) ToData() interface{} {
  me := make([]interface{}, 0, 1 + len(pn.elements))
  for _, op := range pn.elements {
    me = append(me, op.ToData())
  }
  return me
}

func (pn *array) WithName(name string) Entry {
  return &namedArray{name, pn.elements}
}

func (pn *namedArray) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  if len(pn.elements) > 0 {
    b.WriteByte(' ')
    formatElements(pn.elements, b)
  }
  b.WriteByte(')')
}

func (pn *namedArray) ToData() interface{} {
  me := make([]interface{}, 0, len(pn.elements))
  for _, op := range pn.elements {
    me = append(me, op.ToData())
  }
  return []interface{} { pn.name, me }
}

func (pn *namedArray) formatEntry(b *Buffer) {
  b.WriteByte(':')
  b.WriteString(pn.name)
  b.WriteString(` [`)
  formatElements(pn.elements, b)
  b.WriteByte(']')
}

func (pn *namedArray) WithName(name string) Entry {
  return &namedArray{name, pn.elements}
}

func (pn *namedArray) key() string {
  return pn.name
}

func (pn *namedArray) value() interface{} {
  de := make([]interface{}, len(pn.elements))
  for idx := range pn.elements {
    de[idx] = pn.elements[idx].ToData()
  }
  return de
}

func (pn *namedValue) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  pn.val.Format(b)
  b.WriteByte(')')
}

func (pn *namedValue) formatEntry(b *Buffer) {
  b.WriteByte(':')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  pn.val.Format(b)
}

func (pn *namedValue) ToData() interface{} {
  return []interface{} { pn.name, pn.val.ToData() }
}

func (pn *namedValue) WithName(name string) Entry {
  return &namedValue{name, pn.val}
}

func (pn *namedValue) key() string {
  return pn.name
}

func (pn *namedValue) value() interface{} {
  return pn.val.ToData()
}

func (pn *hash) Format(b *Buffer) {
  b.WriteByte('{')
  formatEntries(pn.entries, b)
  b.WriteByte('}')
}

func (pn *hash) ToData() interface{} {
  me := make(map[string]interface{}, len(pn.entries))
  for _, op := range pn.entries {
    entry, _ := op.(Entry)
    me[entry.key()] = entry.value()
  }
  return me
}

func (pn *hash) WithName(name string) Entry {
  return &namedValue{name, pn}
}

func (pn *stringValue) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.val)
  b.WriteByte(')')
}

func (pn *stringValue) ToData() interface{} {
  return []interface{} { pn.val }
}

func (pn *stringValue) WithName(name string) Entry {
  return &namedStringValue{name, pn.val}
}

func (pn *namedStringValue) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  b.WriteString(pn.val)
  b.WriteByte(')')
}

func (pn *namedStringValue) ToData() interface{} {
  return []interface{} { pn.name, pn.val }
}

func (pn *namedStringValue) formatEntry(b *Buffer) {
  b.WriteByte(':')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  b.WriteString(pn.val)
}

func (pn *namedStringValue) WithName(name string) Entry {
  return &namedStringValue{name, pn.val}
}

func (pn *namedStringValue) key() string {
  return pn.name
}

func (pn *namedStringValue) value() interface{} {
  return pn.val
}

func (pn *literal) Format(b *Buffer) {
  switch pn.val.(type) {
  case string:
    doubleQuote(pn.val.(string), b)
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

func (pn *literal) ToData() interface{} {
  return pn.val
}

func (pn *literal) WithName(name string) Entry {
  return &namedValue{name, pn}
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

func formatEntries(elements []Entry, b *Buffer) {
  top := len(elements)
  if top > 0 {
    elements[0].formatEntry(b)
    for idx := 1; idx < top; idx++ {
      b.WriteByte(' ')
      elements[idx].formatEntry(b)
    }
  }
}

func doubleQuote(str string, b *Buffer) {
  b.WriteByte('"')
  for _, c := range str {
    switch c {
    case '\t':
      b.WriteString(`\t`)
    case '\n':
      b.WriteString(`\n`)
    case '\r':
      b.WriteString(`\r`)
    case '"':
      b.WriteString(`\"`)
    case '\\':
      b.WriteString(`\\`)
    default:
      if c < 0x20 {
        Fprintf(b, `\u{%X}`, c)
      } else {
        b.WriteRune(c)
      }
    }
  }
  b.WriteByte('"')
}

