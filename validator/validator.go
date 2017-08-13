package validator

import (
  . "github.com/puppetlabs/go-parser/parser"
  . "fmt"
)

type Validator interface {
  // Validate the semantics of the given expression
  Validate(e Expression)

  // Return all reported issues (should be called after validation)
  Issues() []*ReportedIssue

  setPathAndSubject(path []Expression, expr Expression)
}

// All validators should "inherit" from this struct
type AbstractValidator struct {
  path []Expression
  subject Expression
  issues []*ReportedIssue
  severities map[IssueCode]Severity
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
func (v *AbstractValidator) Accept(code IssueCode, e Expression, args...interface{}) {
  severity, ok := v.severities[code]
  if !ok {
    severity = SEVERITY_ERROR
  }
  v.issues = append(v.issues, NewReportedIssue(code, severity, args, e))
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
    return v.Container();
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

// Validate the expression using the Checker validator
func ValidatePuppet(e Expression) Validator {
  v := NewChecker()
  Validate(v, e)
  return v
}

// Iterate over all expressions contained in the given expression (including the expression itself)
// and validate each one.
func Validate(v Validator, e Expression) {
  path := make([]Expression, 0, 16)

  e.AllContents(path, func(path []Expression, expr Expression) {
    v.setPathAndSubject(path, expr)
    v.Validate(expr)
  })
}
