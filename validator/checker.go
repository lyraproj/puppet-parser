package validator

import (
  . "regexp"
  . "github.com/puppetlabs/go-parser/issue"
  . "github.com/puppetlabs/go-parser/parser"
)

var NUMERIC_VAR_NAME_EXPR = MustCompile(`\A(?:0|(?:[1-9][0-9]*))\z`)
var DOUBLE_COLON_EXPR = MustCompile(`::`)
var CLASSREF_EXPR = MustCompile(`\A(?:::)?[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)

type Checker struct {
  AbstractValidator
}

func NewChecker() *Checker {
  checker := &Checker{AbstractValidator{nil, nil, make([]*ReportedIssue, 0, 5), make(map[IssueCode]Severity, 5)}}
  checker.Demote(VALIDATE_IDEM_EXPRESSION_NOT_LAST, SEVERITY_WARNING)
  return checker
}

func (v *Checker) Validate(e Expression) {
  switch e.(type) {
  case *AssignmentExpression:
    v.check_AssignmentExpression(e.(*AssignmentExpression))
  case *AttributeOperation:
    v.check_AttributeOperation(e.(*AttributeOperation))
  case *AttributesOperation:
    v.check_AttributesOperation(e.(*AttributesOperation))
  case *BlockExpression:
    v.check_BlockExpression(e.(*BlockExpression))
  case *CallNamedFunctionExpression:
    v.check_CallNamedFunctionExpression(e.(*CallNamedFunctionExpression))
  case *CapabilityMapping:
    v.check_CapabilityMapping(e.(*CapabilityMapping))
  case *CaseExpression:
    v.check_CaseExpression(e.(*CaseExpression))
  case *CaseOption:
    v.check_CaseOption(e.(*CaseOption))

  // Interface switches
  case BinaryExpression:
    v.check_BinaryExpression(e.(BinaryExpression))
  }
}

func (v *Checker) check_AssignmentExpression(e *AssignmentExpression) {
  switch e.Operator() {
  case `=`:
    checkAssign(v, e.Lhs())
  default:
    v.Accept(VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED, e, e.Operator())
  }
}

func (v *Checker) check_AttributeOperation(e *AttributeOperation) {
  if e.Operator() == `+>` {
    p := v.Container()
    switch p.(type) {
    case *CollectExpression, *ResourceOverrideExpression:
      return
    default:
      v.Accept(VALIDATE_ILLEGAL_ATTRIBUTE_APPEND, e, e.Name(), A_an(p))
    }
  }
}

func (v *Checker) check_AttributesOperation(e *AttributesOperation) {
  p := v.Container()
  switch p.(type) {
  case AbstractResource, *CollectExpression, *CapabilityMapping:
    v.Accept(VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT, p, `* =>`, A_an(p))
  }
  v.checkRValue(e.Expr())
}

func (v *Checker) check_BinaryExpression(e BinaryExpression) {
  v.checkRValue(e.Lhs())
  v.checkRValue(e.Rhs())
}

func (v *Checker) check_BlockExpression(e *BlockExpression) {
  last := len(e.Statements()) - 1
  for idx, statement := range e.Statements() {
    if idx != last && v.isIdem(statement) {
      v.Accept(VALIDATE_IDEM_EXPRESSION_NOT_LAST, statement, statement.Label())
      break
    }
  }
}

func (v *Checker) check_CallNamedFunctionExpression(e *CallNamedFunctionExpression) {
  switch e.Functor().(type) {
  case *QualifiedName:
    return
  case *QualifiedReference:
    // Call to type
    return
  case *AccessExpression:
    ae, _ := e.Functor().(*AccessExpression)
    if _, ok := ae.Operand().(*QualifiedReference); ok {
      // Call to parameterized type
      return
    }
  }
  v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Functor(), A_an(e.Functor()), `function name`, A_an(e))
}

func (v *Checker) check_CapabilityMapping(e *CapabilityMapping) {
  var exprOk bool
  switch e.Component().(type) {
  case *QualifiedReference:
    exprOk = true

  case *QualifiedName:
    v.Accept(VALIDATE_ILLEGAL_CLASSREF, e.Component(), e.Component().(*QualifiedName).Name())
    exprOk = true // OK, besides from what was just reported

  case *AccessExpression:
    exprOk = false
    ae, _ := e.Component().(*AccessExpression)
    if _, ok := ae.Operand().(*QualifiedReference); ok && len(ae.Keys()) == 1 {
      switch ae.Keys()[0].(type) {
      case *LiteralString, *QualifiedName, *QualifiedReference:
        exprOk = true
      }
    }

  default:
    // Parser will make sure that this never happens
    exprOk = false
  }

  if !exprOk {
    v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Component(), A_an(e.Component()), `capability mapping`, A_an(e))
  }

  if !CLASSREF_EXPR.MatchString(e.Capability()) {
    v.Accept(VALIDATE_ILLEGAL_CLASSREF, e, e.Capability())
  }
}

func (v *Checker) check_CaseExpression(e *CaseExpression) {
  v.checkRValue(e.Test())
  foundDefault := false
  for _, option := range e.Options() {
    co := option.(*CaseOption)
    for _, value := range co.Values() {
      if _, ok := value.(*LiteralDefault); ok {
        if foundDefault {
          v.Accept(VALIDATE_DUPLICATE_DEFAULT, value, e.Label())
        }
        foundDefault = true
      }
    }
  }
}

func (v *Checker) check_CaseOption(e *CaseOption) {
  for _, value := range e.Values() {
    v.checkRValue(value)
  }
}

// TODO: Add more validations here

// Helper functions
func checkAssign(v *Checker, e Expression) {
  switch e.(type) {
  case *AccessExpression:
    v.Accept(VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX, e)

  case *LiteralList:
    for _, elem := range e.(*LiteralList).Elements() {
      checkAssign(v, elem)
    }

  case *VariableExpression:
    name := e.(*VariableExpression).Name()
    if NUMERIC_VAR_NAME_EXPR.MatchString(name) {
      v.Accept(VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT, e, name)
    }
    if DOUBLE_COLON_EXPR.MatchString(name) {
      v.Accept(VALIDATE_CROSS_SCOPE_ASSIGNMENT, e, name)
    }
  }
}

func (v *Checker) checkRValue(e Expression) {
  switch e.(type) {
  case UnaryExpression:
    v.checkRValue(e.(UnaryExpression).Expr())
  case Definition, *CollectExpression:
    v.Accept(VALIDATE_NOT_RVALUE, e, A_anUc(e))
  }
}

// Checks if the expression has side effect ('idem' is latin for 'the same', here meaning that the evaluation state
// is known to be unchanged after the expression has been evaluated). The result is not 100% authoritative for
// negative answers since analysis of function behavior is not possible.
func (v *Checker) isIdem(e Expression) bool {
  switch e.(type) {
  case *AssignmentExpression, *RelationshipExpression, *RenderExpression, *RenderStringExpression:
    return false
  case *BlockExpression:
    return v.idem_BlockExpression(e.(*BlockExpression))
  case *CaseExpression:
    return v.idem_CaseExpression(e.(*CaseExpression))
  case *CaseOption:
    return v.idem_CaseOption(e.(*CaseOption))
  case *IfExpression:
    return v.idem_IfExpression(e.(*IfExpression))
  case *UnlessExpression:
    return v.idem_IfExpression(&e.(*UnlessExpression).IfExpression)
  case *ParenthesizedExpression:
    return v.isIdem(e.(*ParenthesizedExpression).Expr())
  default:
    return true
  }
}

func (v *Checker) idem_BlockExpression(e *BlockExpression) bool {
  for _, statement := range e.Statements() {
    if !v.isIdem(statement) {
      return false
    }
  }
  return true
}

func (v *Checker) idem_CaseExpression(e *CaseExpression) bool {
  if v.isIdem(e.Test()) {
    for _, option := range e.Options() {
      if !v.isIdem(option) {
        return false
      }
    }
    return true
  }
  return false
}

func (v *Checker) idem_CaseOption(e *CaseOption) bool {
  for _, value := range e.Values() {
    if !v.isIdem(value) {
      return false
    }
  }
  return v.isIdem(e.Then())
}

func (v *Checker) idem_IfExpression(e *IfExpression) bool {
  return v.isIdem(e.Test()) && v.isIdem(e.Then()) && v.isIdem(e.Else())
}
