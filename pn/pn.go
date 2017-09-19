package pn

import (
	. "bytes"
	. "fmt"
	"regexp"
	"strings"
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

		// Returns the Format output as a string
		String() string
	}

	// Entry in hash
	Entry interface {
		Key() string
		Value() PN
	}

	mapEntry struct {
		key   string
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

func ToString(pn PN) string {
	b := NewBufferString(``)
	pn.Format(b)
	return b.String()
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

func (pn *listPN) String() string {
	return ToString(pn)
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
	hash := make(map[string]interface{}, 2)
	top := len(pn.elements)
	args := make([]interface{}, 0, top+1)
	args = append(args, pn.name)
	if top > 0 {
		params := pn.listPN.ToData()
		args = append(args, params.([]interface{})...)
	}
	hash[`^`] = args
	return hash
}

func (pn *callPN) String() string {
	return ToString(pn)
}

func (pn *callPN) WithName(name string) Entry {
	return &mapEntry{name, pn}
}

func (e *mapEntry) Key() string {
	return e.key
}

func (e *mapEntry) Value() PN {
	return e.value
}

func (pn *mapPN) AsCall(name string) PN {
	return CallPN(name, pn)
}

func (pn *mapPN) AsParameters() []PN {
	return []PN{pn}
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

func (pn *mapPN) String() string {
	return ToString(pn)
}

func (pn *mapPN) WithName(name string) Entry {
	return &mapEntry{name, pn}
}

func (pn *literalPN) AsCall(name string) PN {
	return CallPN(name, pn)
}

func (pn *literalPN) AsParameters() []PN {
	return []PN{pn}
}

// Strip zeroes between last significant digit and end or exponent. The
// zero following the decimal point is considered significant.
var STRIP_TRAILING_ZEROES = regexp.MustCompile("\\A(.*(?:\\.0|[1-9]))0+(e[+-]?\\d+)?\\z")

func (pn *literalPN) Format(b *Buffer) {
	switch pn.val.(type) {
	case nil:
		b.WriteString(`null`)
	case string:
		DoubleQuote(pn.val.(string), b)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		Fprintf(b, `%d`, pn.val)
	case float32, float64:
		str := Sprintf(`%.16g`, pn.val)
		// We want 16 digit precision that overflows into scientific notation and no trailing zeroes
		if strings.IndexByte(str, '.') < 0 && strings.IndexByte(str, 'e') < 0 {
			// %g sometimes yields an integer number without decimals or scientific
			// notation. Scientific notation must then be used to retain type information
			str = Sprintf(`%.16e`, pn.val)
		}

		if groups := STRIP_TRAILING_ZEROES.FindStringSubmatch(str); groups != nil {
			b.WriteString(groups[1])
			b.WriteString(groups[2])
		} else {
			b.WriteString(str)
		}
	case bool:
		Fprintf(b, `%t`, pn.val)
	default:
		Fprintf(b, `%v`, pn.val)
	}
}

func (pn *literalPN) ToData() interface{} {
	return pn.val
}

func (pn *literalPN) String() string {
	return ToString(pn)
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
