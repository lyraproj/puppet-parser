package issue

import (
	. "bytes"
	. "fmt"
	. "github.com/puppetlabs/go-parser/pn"
)

// this would be an enum in most other languages
const (
	SEVERITY_IGNORE      = Severity(1)
	SEVERITY_DEPRECATION = Severity(2)
	SEVERITY_WARNING     = Severity(3)
	SEVERITY_ERROR       = Severity(4)
)

type (
	IssueCode string

	Severity int

	Location interface {
		File() string

		Line() int

		// Position on line
		Pos() int
	}

	Issue struct {
		code          IssueCode
		messageFormat string
		demotable     bool
	}

	ReportedIssue struct {
		issueCode IssueCode
		severity  Severity
		args      []interface{}
		location  Location
	}
)

var issues = map[IssueCode]*Issue{}

func HardIssue(code IssueCode, messageFormat string) *Issue {
	return addIssue(code, messageFormat, false)
}

func SoftIssue(code IssueCode, messageFormat string) *Issue {
	return addIssue(code, messageFormat, true)
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

func addIssue(code IssueCode, messageFormat string, demotable bool) *Issue {
	dsc := &Issue{code, messageFormat, demotable}
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

func NewReportedIssue(code IssueCode, severity Severity, args []interface{}, location Location) *ReportedIssue {
	return &ReportedIssue{code, severity, args, location}
}

func (e *ReportedIssue) Error() (str string) {
	return appendLocation(Sprintf(IssueForCode(e.issueCode).messageFormat, e.args...), e.location)
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
