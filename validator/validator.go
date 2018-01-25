package validator

import (
	. "fmt"
	. "github.com/puppetlabs/go-parser/issue"
	. "github.com/puppetlabs/go-parser/parser"
	. "strings"
)

const (
	STRICT_OFF     = Strictness(SEVERITY_IGNORE)
	STRICT_WARNING = Strictness(SEVERITY_WARNING)
	STRICT_ERROR   = Strictness(SEVERITY_ERROR)
)

type (
	Validator interface {
		// Validate the semantics of the given expression
		Validate(e Expression)

		// Return all reported issues (should be called after validation)
		Issues() []*ReportedIssue

		setPathAndSubject(path []Expression, expr Expression)
	}

	// All validators should "inherit" from this struct
	AbstractValidator struct {
		path       []Expression
		subject    Expression
		issues     []*ReportedIssue
		severities map[IssueCode]Severity
	}

	Strictness int
)

func Strict(str string) Strictness {
	switch ToLower(str) {
	case ``, `off`:
		return STRICT_OFF
	case `warning`:
		return STRICT_WARNING
	case `error`:
		return STRICT_ERROR
	default:
		panic(Sprintf(`Invalid Strictness value '%s'`, str))
	}
}

func (s Strictness) String() string {
	switch s {
	case STRICT_OFF:
		return `off`
	case STRICT_WARNING:
		return `warning`
	case STRICT_ERROR:
		return `error`
	default:
		panic(Sprintf(`Invalid Strictness value %d`, s))
	}
}

func (v *AbstractValidator) Demote(code IssueCode, severity Severity) {
	issue := IssueForCode(code)
	severity.AssertValid()
	if !issue.IsDemotable() {
		panic(Sprintf(`Attempt to demote the hard issue '%s' to %s`, code, severity.String()))
	}
	v.severities[code] = severity
}

// Accept an issue during validation
func (v *AbstractValidator) Accept(code IssueCode, e Expression, args H) {
	severity, ok := v.severities[code]
	if !ok {
		severity = SEVERITY_ERROR
	}
	if severity != SEVERITY_IGNORE {
		v.issues = append(v.issues, NewReportedIssue(code, severity, args, e))
	}
}

// Returns the container of the currently validated expression
func (v *AbstractValidator) Container() Expression {
	if v.path != nil && len(v.path) > 0 {
		return v.path[len(v.path)-1]
	}
	return nil
}

// Returns the container of some parent of the currently validated expression
//
// Note: This will return nil for the expression that is currently validated
func (v *AbstractValidator) ContainerOf(e Expression) Expression {
	if e == v.subject {
		return v.Container()
	}
	for last := len(v.path) - 1; last > 0; last-- {
		if e == v.path[last] {
			return v.path[last-1]
		}
	}
	return nil
}

func (v *AbstractValidator) Issues() []*ReportedIssue {
	return v.issues
}

func (v *AbstractValidator) setPathAndSubject(path []Expression, subject Expression) {
	v.path = path
	v.subject = subject
}

// Validate the expression using the Puppet validator
func ValidatePuppet(e Expression, strict Strictness) Validator {
	v := NewChecker(strict)
	Validate(v, e)
	return v
}

// Validate the expression using the Tasks validator
func ValidateTasks(e Expression) Validator {
	v := NewTasksChecker()
	Validate(v, e)
	return v
}

// Iterate over all expressions contained in the given expression (including the expression itself)
// and validate each one.
func Validate(v Validator, e Expression) {
	path := make([]Expression, 0, 16)

	v.setPathAndSubject(path, e)
	v.Validate(e)
	e.AllContents(path, func(path []Expression, expr Expression) {
		v.setPathAndSubject(path, expr)
		v.Validate(expr)
	})
}
