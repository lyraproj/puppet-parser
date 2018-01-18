package issue

import (
	. "bytes"
	. "fmt"
	. "github.com/puppetlabs/go-parser/pn"
	"unicode/utf8"
	"io"
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
	IssueCode string

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
		code          IssueCode
		messageFormat string
		argFormats    HF
		demotable     bool
	}

	ReportedIssue struct {
		issueCode IssueCode
		severity  Severity
		args      H
		location  Location
	}
)

var issues = map[IssueCode]*Issue{}

func HardIssue(code IssueCode, messageFormat string) *Issue {
	return addIssue(code, messageFormat, false, nil)
}

func HardIssue2(code IssueCode, messageFormat string, argFormats HF) *Issue {
	return addIssue(code, messageFormat, false, argFormats)
}

func SoftIssue(code IssueCode, messageFormat string) *Issue {
	return addIssue(code, messageFormat, true, nil)
}

func SoftIssue2(code IssueCode, messageFormat string, argFormats HF) *Issue {
	return addIssue(code, messageFormat, true, argFormats)
}

func (issue *Issue) Code() IssueCode {
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
		panic(Sprintf(`Illegal severity level: %d`, severity))
	}
}

func (severity Severity) AssertValid() {
	switch severity {
	case SEVERITY_IGNORE, SEVERITY_DEPRECATION, SEVERITY_WARNING, SEVERITY_ERROR:
		return
	default:
		panic(Sprintf(`Illegal severity level: %d`, severity))
	}
}

func addIssue(code IssueCode, messageFormat string, demotable bool, argFormats HF) *Issue {
	dsc := &Issue{code, messageFormat, argFormats,demotable}
	issues[code] = dsc
	return dsc
}

// Returns the Issue for an IssueCode. Will panic if the given code does not represent
// an existing issue
func IssueForCode(code IssueCode) *Issue {
	if dsc, ok := issues[code]; ok {
		return dsc
	}
	panic(Sprintf("internal error: no issue found for issue code '%s'", code))
}

func IssueForCode2(code IssueCode) (dsc *Issue, ok bool) {
	dsc, ok = issues[code]
	return
}

func NewReportedIssue(code IssueCode, severity Severity, args H, location Location) *ReportedIssue {
	return &ReportedIssue{code, severity, args, location}
}

func (e *ReportedIssue) Argument(str string) interface{} {
	return e.args[str]
}

func (e *ReportedIssue) Error() (str string) {
	issue := IssueForCode(e.issueCode)
	var args H
	af := issue.argFormats
	if af != nil {
		args = make(H, len(e.args))
		for k, v := range e.args {
			if a, ok := af[k]; ok {
				v = a(v)
			}
			args[k] = v
		}
	} else {
		args = e.args
	}
	return appendLocation(MapSprintf(issue.messageFormat, args), e.location)
}

func (e *ReportedIssue) String() (str string) {
	return e.Error()
}

func (e *ReportedIssue) Code() IssueCode {
	return e.issueCode
}

func (e *ReportedIssue) Severity() Severity {
	return e.severity
}

// Represent the reported using polish notation
func (e *ReportedIssue) ToPN() PN {
	return MapPN([]Entry{
		LiteralPN(e.issueCode).WithName(`code`),
		LiteralPN(e.severity.String()).WithName(`severity`),
		LiteralPN(e.Error()).WithName(`message`)})
}

func appendLocation(str string, location Location) string {
	if location == nil {
		return str
	}
	b := NewBufferString(str)
	line := location.Line()
	pos := location.Pos()
	if file := location.File(); file != `` {
		if line > 0 {
			b.WriteString(` at `)
			b.WriteString(file)
			b.WriteByte(':')
			Fprintf(b, `%d`, line)
			if pos > 0 {
				b.WriteByte(':')
				Fprintf(b, `%d`, pos)
			}
		} else {
			b.WriteString(` in `)
			b.WriteString(file)
		}
	} else if line > 0 {
		b.WriteString(` at line `)
		Fprintf(b, `%d`, line)
		if pos > 0 {
			b.WriteByte(':')
			Fprintf(b, `%d`, pos)
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
	b := NewBufferString(``)
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
			panic(Sprintf(`missing argument matching key {%s} in format string %s`, k, formatString))
		}
	}
	Fprintf(writer, posFormatString, posArgs...)
}

func extractNamesAndLocations(formatString string) (string, int, map[string][]int) {
	b := NewBufferString(``)
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
				panic(Sprintf(`keyed formats cannot be combined with other % formats at position %d in string '%s'`,
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
			panic(Sprintf(`unterminated %%%c at position %d in string '%s'`, bc, s - 2, formatString))
		}
		e := rdr.i - 1
		if s == e {
			panic(Sprintf(`empty %%%c%c at position %d in string '%s'`, bc, ec, s - 2, formatString))
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