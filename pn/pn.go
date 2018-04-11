package pn

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"github.com/puppetlabs/go-issues/issue"
)

type (
	// PN - Puppet Extended S-Expresssion Notation
	//
	// A PN forms a directed acyclig graph of nodes. There are four types of nodes:
	//
	// * Literal: A boolean, integer, float, string, or undef
	//
	// * List: An ordered list of nodes
  //
	// * Map: An ordered map of string to node associations
	//
	// * Call: A named list of nodes.
	PN interface {
		// Format produces a compact, Clojure like syntax. Suitable for tests.
		Format(b *bytes.Buffer)

		// ToData produces an object that where all values are of primitive type or
		// slices or maps. This format is suitable for output as JSON or YAML
		//
		ToData() interface{}

		// AsCall turns this PN into an argument in a call or change the name if the PN Is a call already
		AsCall(name string) PN

		// AsParameters returns the PN as a parameter list
		AsParameters() []PN

		// WithName creates a key/value pair from the given name and this PN
		WithName(name string) Entry

		// String returns the Format output as a string
		String() string
	}

	pnError struct {
		message string
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

var keyPattern = regexp.MustCompile(`^[A-Za-z_-][0-9A-Za-z_-]*$`)

// Represent the Reported using Puppet Extended S-Expresssion Notation (PN)
func ReportedToPN(ri issue.Reported) PN {
	return Map([]Entry{
		Literal(ri.Code()).WithName(`code`),
		Literal(ri.Severity().String()).WithName(`severity`),
		Literal(ri.Error()).WithName(`message`)})
}

func (e *pnError) Error() string {
	return e.message
}

func List(elements []PN) PN {
	return &listPN{elements}
}

func Map(entries []Entry) PN {
	for _, e := range entries {
		if !keyPattern.MatchString(e.Key()) {
			panic(pnError{fmt.Sprintf("key '%s' does not conform to pattern %s",
				e.Key(), keyPattern.String())})
		}
	}
	return &mapPN{entries}
}

func Literal(val interface{}) PN {
	return &literalPN{val}
}

func Call(name string, elements ...PN) PN {
	return &callPN{listPN{elements}, name}
}

func ToString(pn PN) string {
	b := bytes.NewBufferString(``)
	pn.Format(b)
	return b.String()
}

func (pn *listPN) AsCall(name string) PN {
	return Call(name, pn.elements...)
}

func (pn *listPN) AsParameters() []PN {
	return pn.elements
}

func (pn *listPN) Format(b *bytes.Buffer) {
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

func (pn *callPN) Format(b *bytes.Buffer) {
	b.WriteByte('(')
	b.WriteString(pn.name)
	if len(pn.elements) > 0 {
		b.WriteByte(' ')
		formatElements(pn.elements, b)
	}
	b.WriteByte(')')
}

func (pn *callPN) ToData() interface{} {
	top := len(pn.elements)
	args := make([]interface{}, 0, top+1)
	args = append(args, pn.name)
	if top > 0 {
		params := pn.listPN.ToData()
		args = append(args, params.([]interface{})...)
	}
	return map[string]interface{}{`^`: args}
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
	return Call(name, pn)
}

func (pn *mapPN) AsParameters() []PN {
	return []PN{pn}
}

func (pn *mapPN) Format(b *bytes.Buffer) {
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

func formatEntry(entry Entry, b *bytes.Buffer) {
	b.WriteByte(':')
	b.WriteString(entry.Key())
	b.WriteByte(' ')
	entry.Value().Format(b)
}

func (pn *mapPN) ToData() interface{} {
	top := len(pn.entries) * 2
	args := make([]interface{}, 0, top)
	for _, entry := range pn.entries {
		args = append(args, entry.Key(), entry.Value().ToData())
	}
	return map[string]interface{}{`#`: args}
}

func (pn *mapPN) String() string {
	return ToString(pn)
}

func (pn *mapPN) WithName(name string) Entry {
	return &mapEntry{name, pn}
}

func (pn *literalPN) AsCall(name string) PN {
	return Call(name, pn)
}

func (pn *literalPN) AsParameters() []PN {
	return []PN{pn}
}

// Strip zeroes between last significant digit and end or exponent. The
// zero following the decimal point is considered significant.
var STRIP_TRAILING_ZEROES = regexp.MustCompile("\\A(.*(?:\\.0|[1-9]))0+(e[+-]?\\d+)?\\z")

func (pn *literalPN) Format(b *bytes.Buffer) {
	switch pn.val.(type) {
	case nil:
		b.WriteString(`nil`)
	case string:
		DoubleQuote(pn.val.(string), b)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		fmt.Fprintf(b, `%d`, pn.val)
	case float32, float64:
		str := fmt.Sprintf(`%.16g`, pn.val)
		// We want 16 digit precision that overflows into scientific notation and no trailing zeroes
		if strings.IndexByte(str, '.') < 0 && strings.IndexByte(str, 'e') < 0 {
			// %g sometimes yields an integer number without decimals or scientific
			// notation. Scientific notation must then be used to retain type information
			str = fmt.Sprintf(`%.16e`, pn.val)
		}

		if groups := STRIP_TRAILING_ZEROES.FindStringSubmatch(str); groups != nil {
			b.WriteString(groups[1])
			b.WriteString(groups[2])
		} else {
			b.WriteString(str)
		}
	case bool:
		fmt.Fprintf(b, `%t`, pn.val)
	default:
		fmt.Fprintf(b, `%v`, pn.val)
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

func formatElements(elements []PN, b *bytes.Buffer) {
	top := len(elements)
	if top > 0 {
		elements[0].Format(b)
		for idx := 1; idx < top; idx++ {
			b.WriteByte(' ')
			elements[idx].Format(b)
		}
	}
}
