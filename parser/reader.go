package parser

import (
  . "unicode/utf8"
  . "fmt"
)

type StringReader interface {
  // Returns the the current rune and its position in the parsed string and advances the position. Returns 0, 0
  // when EOS is reached
  Next() (c rune, start int)

  // Returns the the current rune and its size in the parsed string. The position does not change
  Peek() (c rune, size int)

  Advance(size int)

  Pos() (int)

  SetPos(int)

  // Returns the string that is backing the reader
  String() (string)

  // Returns the substring starting at start and up to, but not including, the current position
  From(start int) (string)
}

type ParseError struct {
  rootCause error
  message string
  offset int
}

type stringReader struct {
  i int
  Text string
}

func (e *ParseError) Error() string {
  return Sprintf(`%s at offset %d`, e.message, e.offset)
}

func NewStringReader(s string) StringReader {
  return &stringReader{i: 0, Text: s}
}

func (r *stringReader) parseError(message string) *ParseError {
  return &ParseError{ message: message, offset: r.i }
}

func (r *stringReader)invalidUnicode() *ParseError {
  return r.parseError("invalid unicode character")
}

func (r *stringReader)Next() (c rune, start int) {
  start = r.i
  if r.i >= len(r.Text) {
    return
  }
  c = rune(r.Text[r.i])
  if c < RuneSelf {
    r.i++
    return
  }
  c, size := DecodeRuneInString(r.Text[r.i:])
  if c == RuneError {
    panic(r.invalidUnicode())
  }
  r.i += size
  return
}

func (r *stringReader)Peek() (c rune, size int) {
  if r.i >= len(r.Text) {
    return
  }
  c = rune(r.Text[r.i])
  if c < RuneSelf {
    size = 1
    return
  }
  c, size = DecodeRuneInString(r.Text[r.i:])
  if c == RuneError {
    panic(r.invalidUnicode())
  }
  return c, size
}

func (r *stringReader)Advance(size int) {
  r.i += size
}

func (r *stringReader)Pos() (int) {
  return r.i
}

func (r *stringReader)SetPos(pos int) {
  r.i = pos
}

func (r *stringReader)String() (string) {
  return r.Text
}

// Returns a substring of the contained string that starts at the given position and ends at
// the current position
func (r *stringReader)From(start int) (string) {
  return r.Text[start:r.i]
}

