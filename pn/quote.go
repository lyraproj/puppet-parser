package pn

import (
	"bytes"
)

func DoubleQuote(str string, b *bytes.Buffer) {
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
				Fprintf(b, `\o%3.3o`, c)
			} else {
				b.WriteRune(c)
			}
		}
	}
	b.WriteByte('"')
}
