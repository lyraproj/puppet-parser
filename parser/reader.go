package parser

import (
	"fmt"
	"unicode/utf8"
)

type StringReader interface {
	// Returns the the current rune and its position in the parsed string and advances the position. Returns 0, 0
	// when EOS is reached
	Next() (c rune, start int)

	// Returns the the current rune and its size in the parsed string. The position does not change
	Peek() (c rune, size int)

	Advance(size int)

	Pos() int

	SetPos(int)

	// Returns the string that is backing the reader
	Text() string

	// Returns the substring starting at start and up to, but not including, the current position
	From(start int) string
}

type ParseError struct {
	rootCause error
	message   string
	offset    int
}

type stringReader struct {
	i    int
	text string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf(`%s at offset %d`, e.message, e.offset)
}

func NewStringReader(s string) StringReader {
	return &stringReader{i: 0, text: s}
}

func (r *stringReader) parseError(message string) *ParseError {
	return &ParseError{message: message, offset: r.i}
}

func (r *stringReader) invalidUnicode() *ParseError {
	return r.parseError("invalid unicode character")
}

func (r *stringReader) Next() (c rune, start int) {
	start = r.i
	if r.i >= len(r.text) {
		return
	}
	c = rune(r.text[r.i])
	if c < utf8.RuneSelf {
		r.i++
		return
	}
	c, size := utf8.DecodeRuneInString(r.text[r.i:])
	if c == utf8.RuneError {
		panic(r.invalidUnicode())
	}
	r.i += size
	return
}

func (r *stringReader) Peek() (c rune, size int) {
	if r.i >= len(r.text) {
		return
	}
	c = rune(r.text[r.i])
	if c < utf8.RuneSelf {
		size = 1
		return
	}
	c, size = utf8.DecodeRuneInString(r.text[r.i:])
	if c == utf8.RuneError {
		panic(r.invalidUnicode())
	}
	return c, size
}

func (r *stringReader) Advance(size int) {
	r.i += size
}

func (r *stringReader) Pos() int {
	return r.i
}

func (r *stringReader) SetPos(pos int) {
	r.i = pos
}

func (r *stringReader) Text() string {
	return r.text
}

// Returns a substring of the contained string that starts at the given position and ends at
// the current position
func (r *stringReader) From(start int) string {
	return r.text[start:r.i]
}
