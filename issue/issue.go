package issue

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/puppetlabs/go-parser/pn"
)

// this would be an enum in most other languages
const (
	SEVERITY_IGNORE      = Severity(1)
	SEVERITY_DEPRECATION = Severity(2)
	SEVERITY_WARNING     = Severity(3)
	SEVERITY_ERROR       = Severity(4)
)

var NO_ARGS = H{}

type (
	Code string

	Severity int

	ArgFormat func(value interface{}) string

	H map[string]interface{}

	HF map[string]ArgFormat

	Location interface {
		File() string

		Line() int

		// Position on line
		Pos() int
	}

	Issue struct {
		code          Code
		messageFormat string
		argFormats    HF
		demotable     bool
	}

	Reported struct {
		issueCode Code
		severity  Severity
		args      H
		location  Location
	}

	location struct {
		file string
		line int
		pos  int
	}
)

func NewLocation(file string, line, pos int) Location {
	return &location{file, line, pos}
}

func (l *location) File() string {
	return l.file
}

func (l *location) Line() int {
	return l.line
}

func (l *location) Pos() int {
	return l.pos
}

var issues = map[Code]*Issue{}

func Hard(code Code, messageFormat string) *Issue {
	return addIssue(code, messageFormat, false, nil)
}

func Hard2(code Code, messageFormat string, argFormats HF) *Issue {
	return addIssue(code, messageFormat, false, argFormats)
}

func SoftIssue(code Code, messageFormat string) *Issue {
	return addIssue(code, messageFormat, true, nil)
}

func SoftIssue2(code Code, messageFormat string, argFormats HF) *Issue {
	return addIssue(code, messageFormat, true, argFormats)
}

func (issue *Issue) Code() Code {
	return issue.code
}

func (issue *Issue) IsDemotable() bool {
	return issue.demotable
}

func (issue *Issue) MessageFormat() string {
	return issue.messageFormat
}

func (severity Severity) String() string {
	switch severity {
	case SEVERITY_IGNORE:
		return `ignore`
	case SEVERITY_DEPRECATION:
		return `warning`
	case SEVERITY_WARNING:
		return `warning`
	case SEVERITY_ERROR:
		return `error`
	default:
		panic(fmt.Sprintf(`Illegal severity level: %d`, severity))
	}
}

func (severity Severity) AssertValid() {
	switch severity {
	case SEVERITY_IGNORE, SEVERITY_DEPRECATION, SEVERITY_WARNING, SEVERITY_ERROR:
		return
	default:
		panic(fmt.Sprintf(`Illegal severity level: %d`, severity))
	}
}

func addIssue(code Code, messageFormat string, demotable bool, argFormats HF) *Issue {
	dsc := &Issue{code, messageFormat, argFormats, demotable}
	issues[code] = dsc
	return dsc
}

// Returns the Issue for a Code. Will panic if the given code does not represent
// an existing issue
func IssueForCode(code Code) *Issue {
	if dsc, ok := issues[code]; ok {
		return dsc
	}
	panic(fmt.Sprintf("internal error: no issue found for issue code '%s'", code))
}

func IssueForCode2(code Code) (dsc *Issue, ok bool) {
	dsc, ok = issues[code]
	return
}

func NewReported(code Code, severity Severity, args H, location Location) *Reported {
	return &Reported{code, severity, args, location}
}

func (ri *Reported) Argument(str string) interface{} {
	return ri.args[str]
}

func (ri *Reported) OffsetByLocation(location Location) *Reported {
	loc := ri.location
	if loc == nil {
		loc = location
	} else {
		loc = NewLocation(location.File(), location.Line()+loc.Line(), location.Pos())
	}
	return &Reported{ri.issueCode, ri.severity, ri.args, loc}
}

func (ri *Reported) Error() (str string) {
	issue := IssueForCode(ri.issueCode)
	var args H
	af := issue.argFormats
	if af != nil {
		args = make(H, len(ri.args))
		for k, v := range ri.args {
			if a, ok := af[k]; ok {
				v = a(v)
			}
			args[k] = v
		}
	} else {
		args = ri.args
	}
	return appendLocation(MapSprintf(issue.messageFormat, args), ri.location)
}

func (ri *Reported) String() (str string) {
	return ri.Error()
}

func (ri *Reported) Code() Code {
	return ri.issueCode
}

func (ri *Reported) Severity() Severity {
	return ri.severity
}

// Represent the Reported using Puppet Extended S-Expresssion Notation (PN)
func (ri *Reported) ToPN() pn.PN {
	return pn.Map([]pn.Entry{
		pn.Literal(ri.issueCode).WithName(`code`),
		pn.Literal(ri.severity.String()).WithName(`severity`),
		pn.Literal(ri.Error()).WithName(`message`)})
}

func appendLocation(str string, location Location) string {
	if location == nil {
		return str
	}
	b := bytes.NewBufferString(str)
	line := location.Line()
	pos := location.Pos()
	if file := location.File(); file != `` {
		if line > 0 {
			b.WriteString(` at `)
			b.WriteString(file)
			b.WriteByte(':')
			fmt.Fprintf(b, `%d`, line)
			if pos > 0 {
				b.WriteByte(':')
				fmt.Fprintf(b, `%d`, pos)
			}
		} else {
			b.WriteString(` in `)
			b.WriteString(file)
		}
	} else if line > 0 {
		b.WriteString(` at line `)
		fmt.Fprintf(b, `%d`, line)
		if pos > 0 {
			b.WriteByte(':')
			fmt.Fprintf(b, `%d`, pos)
		}
	}
	return b.String()
}

type stringReader struct {
	i    int
	text string
}

func (r *stringReader) next() rune {
	if r.i >= len(r.text) {
		return 0
	}
	c := rune(r.text[r.i])
	if c < utf8.RuneSelf {
		r.i++
		return c
	}
	c, size := utf8.DecodeRuneInString(r.text[r.i:])
	if c == utf8.RuneError {
		panic(`Invalid unicode in string`)
	}
	r.i += size
	return c
}

func MapSprintf(formatString string, args H) string {
	b := bytes.NewBufferString(``)
	MapFprintf(b, formatString, args)
	return b.String()
}

func MapFprintf(writer io.Writer, formatString string, args H) {
	posFormatString, argCount, expectedArgs := extractNamesAndLocations(formatString)
	posArgs := make([]interface{}, argCount)
	for k, v := range expectedArgs {
		if arg, ok := args[k]; ok {
			for _, pos := range v {
				posArgs[pos] = arg
			}
		} else {
			panic(fmt.Sprintf(`missing argument matching key {%s} in format string %s`, k, formatString))
		}
	}
	fmt.Fprintf(writer, posFormatString, posArgs...)
}

func extractNamesAndLocations(formatString string) (string, int, map[string][]int) {
	b := bytes.NewBufferString(``)
	rdr := stringReader{0, formatString}
	locations := make(map[string][]int, 8)
	c := rdr.next()
	location := 0
	for c != 0 {
		b.WriteRune(c)
		if c != '%' {
			c = rdr.next()
			continue
		}
		c = rdr.next()
		if c != '{' && c != '<' {
			if c != '%' {
				panic(fmt.Sprintf(`keyed formats cannot be combined with other %% formats at position %d in string '%s'`,
					rdr.i, formatString))
			}
			b.WriteRune(c)
			c = rdr.next()
			continue
		}
		ec := '}'
		bc := c
		if bc == '<' {
			ec = '>'
		}
		s := rdr.i
		c = rdr.next()
		for c != 0 && c != ec {
			c = rdr.next()
		}
		if c == 0 {
			panic(fmt.Sprintf(`unterminated %%%c at position %d in string '%s'`, bc, s-2, formatString))
		}
		e := rdr.i - 1
		if s == e {
			panic(fmt.Sprintf(`empty %%%c%c at position %d in string '%s'`, bc, ec, s-2, formatString))
		}
		key := formatString[s:e]
		if ps, ok := locations[key]; ok {
			locations[key] = append(ps, location)
		} else {
			locations[key] = []int{location}
		}
		location++
		if bc == '{' {
			// %{} constructs uses default format specifier whereas %<> uses whatever was specified
			b.WriteByte('v')
		}
		c = rdr.next()
	}
	return b.String(), location, locations
}
