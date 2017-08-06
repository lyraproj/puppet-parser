package parser

type Validator struct {
  path []Expression
  subject Expression
  issues []*ReportedIssue
}

type Validation interface {
  Expression
  Validate(v *Validator)
}

func NewValidator() *Validator {
  return &Validator{nil, nil, make([]*ReportedIssue, 0, 5)}
}

// Returns the container of the currently validated expression
func (v *Validator) container() Expression {
  if v.path != nil && len(v.path) > 0 {
    return v.path[len(v.path)-1]
  }
  return nil
}

// Returns the container of some parent of the currently validated expression
//
// Note: This will return nil for the expression that is currently validated
func (v *Validator) containerOf(e Expression) Expression {
  if e == v.subject {
    return v.container();
  }
  for last := len(v.path) - 1; last > 0; last-- {
    if e == v.path[last] {
      return v.path[last-1]
    }
  }
  return nil
}

func (v *Validator) Issues() []*ReportedIssue {
  return v.issues
}

func (v *Validator) Validate(e Expression) {
  path := make([]Expression, 0, 16)

  e.AllContents(path, func(path []Expression, expr Expression) {
    if vs, ok := expr.(Validation); ok {
      v.path = path
      v.subject = expr
      vs.Validate(v)
    }
  })
}

func (v *Validator) accept(issueCode string, e Expression, args...interface{}) {
  v.issues = append(v.issues, &ReportedIssue{issueCode, args, e})
}
