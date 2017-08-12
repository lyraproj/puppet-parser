package validator

import (
  . "github.com/puppetlabs/go-parser/parser"
)

type Validator interface {
  Validate(e Expression)

  Issues() []*ReportedIssue

  setPathAndSubject(path []Expression, expr Expression)
}

type AbstractValidator struct {
  path []Expression
  subject Expression
  issues []*ReportedIssue
}

func (v *AbstractValidator) Accept(issueCode string, e Expression, args...interface{}) {
  v.issues = append(v.issues, NewReportedIssue(issueCode, args, e))
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

func Validate(v Validator, e Expression) {
  path := make([]Expression, 0, 16)

  e.AllContents(path, func(path []Expression, expr Expression) {
    v.setPathAndSubject(path, expr)
    v.Validate(expr)
  })
}
