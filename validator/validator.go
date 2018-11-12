package validator

import (
	"fmt"
	"strings"

	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-parser/parser"
)

const (
	STRICT_OFF     = Strictness(issue.SEVERITY_IGNORE)
	STRICT_WARNING = Strictness(issue.SEVERITY_WARNING)
	STRICT_ERROR   = Strictness(issue.SEVERITY_ERROR)
)

type (
	Validator interface {
		Clear()

		// Validate the semantics of the given expression
		Validate(e parser.Expression)

		// Return all reported issues (should be called after validation)
		Issues() []issue.Reported

		setPathAndSubject(path []parser.Expression, expr parser.Expression)
	}

	ParserValidator interface {
		// Parse parses and validates the given source and returns the resulting expression together
		// with an optional Result. The result will be nil if no warnings or errors were generated.
		Parse(filename string, source string) (parser.Expression, issue.Result)
	}

	// All validators should "inherit" from this struct
	AbstractValidator struct {
		path       []parser.Expression
		subject    parser.Expression
		issues     []issue.Reported
		severities map[issue.Code]issue.Severity
	}

	Strictness int

	parserValidator struct {
		parser parser.ExpressionParser
		validator Validator
	}
)

func Strict(str string) Strictness {
	switch strings.ToLower(str) {
	case ``, `off`:
		return STRICT_OFF
	case `warning`:
		return STRICT_WARNING
	case `error`:
		return STRICT_ERROR
	default:
		panic(fmt.Sprintf(`Invalid Strictness value '%s'`, str))
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
		panic(fmt.Sprintf(`Invalid Strictness value %d`, s))
	}
}

func (v *AbstractValidator) Demote(code issue.Code, severity issue.Severity) {
	issue := issue.IssueForCode(code)
	severity.AssertValid()
	if !issue.IsDemotable() {
		panic(fmt.Sprintf(`Attempt to demote the hard issue '%s' to %s`, code, severity.String()))
	}
	v.severities[code] = severity
}

// Accept an issue during validation
func (v *AbstractValidator) Accept(code issue.Code, e parser.Expression, args issue.H) {
	severity, ok := v.severities[code]
	if !ok {
		severity = issue.SEVERITY_ERROR
	}
	if severity != issue.SEVERITY_IGNORE {
		v.issues = append(v.issues, issue.NewReported(code, severity, args, e))
	}
}

// Returns the container of the currently validated expression
func (v *AbstractValidator) Container() parser.Expression {
	if v.path != nil && len(v.path) > 0 {
		return v.path[len(v.path)-1]
	}
	return nil
}

// Returns the container of some parent of the currently validated expression
//
// Note: This will return nil for the expression that is currently validated
func (v *AbstractValidator) ContainerOf(e parser.Expression) parser.Expression {
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

func (v *AbstractValidator) Issues() []issue.Reported {
	return v.issues
}

func (v *AbstractValidator) Clear() {
	v.issues = make([]issue.Reported, 0, 5)
}

func (v *AbstractValidator) setPathAndSubject(path []parser.Expression, subject parser.Expression) {
	v.path = path
	v.subject = subject
}

// Validate the expression using the Puppet validator
func ValidatePuppet(e parser.Expression, strict Strictness) Validator {
	v := NewChecker(strict)
	Validate(v, e)
	return v
}

// Validate the expression using the Tasks validator
func ValidateTasks(e parser.Expression) Validator {
	v := NewTasksChecker()
	Validate(v, e)
	return v
}

// Validate the expression using the Workflow validator
func ValidateWorkflow(e parser.Expression) Validator {
	v := NewWorkflowChecker()
	Validate(v, e)
	return v
}

// Iterate over all expressions contained in the given expression (including the expression itself)
// and validate each one.
func Validate(v Validator, e parser.Expression) {
	path := make([]parser.Expression, 0, 16)

	v.Clear()
	v.setPathAndSubject(path, e)
	v.Validate(e)
	e.AllContents(path, func(path []parser.Expression, expr parser.Expression) {
		v.setPathAndSubject(path, expr)
		v.Validate(expr)
	})
}

func NewParserValidator(parser parser.ExpressionParser, validator Validator) ParserValidator {
	return &parserValidator{parser, validator}
}

func (pv *parserValidator) Parse(filename string, source string) (parser.Expression, issue.Result) {
	expr, err := pv.parser.Parse(filename, source, false)
	if err != nil {
		if i, ok := err.(issue.Reported); ok {
			return nil, issue.NewResult([]issue.Reported{i})
		}
		panic(err.Error())
	}
	Validate(pv.validator, expr)
	issues := pv.validator.Issues()
	if len(issues) == 0 {
		return expr, nil
	}
	return expr, issue.NewResult(issues)
}
