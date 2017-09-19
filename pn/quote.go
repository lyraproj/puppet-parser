package pn

import (
	. "bytes"
	. "fmt"
)

func DoubleQuote(str string, b *Buffer) {
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
