package parser

import (
  "regexp"
)

var NUMERIC_VAR_NAME_EXPR = regexp.MustCompile(`\A(?:0|(?:[1-9][0-9]*))\z`)
var DOUBLE_COLON_EXPR = regexp.MustCompile(`::`)


func (e *AssignmentExpression) Validate(v *Validator) {
  switch e.Operator() {
  case `=`:
    checkAssign(v, e.Lhs())
  default:
    v.accept(VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED, e, e.Operator())
  }
}

func (e *AttributeOperation) Validate(v *Validator) {
  if e.Operator() == `+>` {
    p := v.container()
    switch p.(type) {
    case *CollectExpression, *ResourceOverrideExpression:
      return
    default:
      v.accept(VALIDATE_ILLEGAL_ATTRIBUTE_APPEND, e, e.Name(), A_an(p))
    }
  }
}

func (e *binaryExpression) Validate(v *Validator) {
  checkRValue(v, e.Lhs())
  checkRValue(v, e.Rhs())
}

func (e *BlockExpression) Validate(v *Validator) {
  last := len(e.statements) - 1
  for idx, statement := range e.statements {
    if idx != last && isIdem(v, statement) {
      v.accept(VALIDATE_IDEM_EXPRESSION_NOT_LAST, statement, statement.Label())
      break
    }
  }
}

func (e *CallNamedFunctionExpression) Validate(v *Validator) {
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

  default:
    v.accept(VALIDATE_ILLEGAL_EXPRESSION, e.Functor(), A_an(e.Functor()), `function name`, A_an(e))
  }
}

// TODO: Add more validations here

// Helper functions
func checkAssign(v *Validator, e Expression) {
  switch e.(type) {
  case *AccessExpression:
    v.accept(VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX, e)

  case *LiteralList:
    for _, elem := range e.(*LiteralList).elements {
      checkAssign(v, elem)
    }

  case *VariableExpression:
    name := e.(*VariableExpression).Name()
    if NUMERIC_VAR_NAME_EXPR.MatchString(name) {
      v.accept(VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT, e, name)
    }
    if DOUBLE_COLON_EXPR.MatchString(name) {
      v.accept(VALIDATE_CROSS_SCOPE_ASSIGNMENT, e, name)
    }
  }
}

func checkRValue(v *Validator, e Expression) {
  switch e.(type) {
  case Unary:
    checkRValue(v, e.(Unary).Expr())
  case Definition, *CollectExpression:
      v.accept(VALIDATE_NOT_RVALUE, e, A_anUc(e))
  }
}

// Checks if the expression has side effect ('idem' is latin for 'the same', here meaning that the evaluation state
// is known to be unchanged after the expression has been evaluated). The result is not 100% authoritative for
// negative answers since analysis of function behavior is not possible.
func isIdem(v *Validator, e Expression) bool {
  return e.(idem).isIdem(v)
}

type idem interface {
  isIdem(v *Validator) bool
}

func (e *positioned)isIdem(v *Validator) bool {
  return false
}

func (e *AccessExpression)isIdem(v *Validator) bool {
  return true
}

func (e *AssignmentExpression)isIdem(v *Validator) bool {
  return false
}

func (e *binaryExpression)isIdem(v *Validator) bool {
  return true
}

func (e *BlockExpression)isIdem(v *Validator) bool {
  for _, statement := range e.statements {
    if !statement.(idem).isIdem(v) {
      return false
    }
  }
  return true
}

func (e *CaseExpression)isIdem(v *Validator) bool {
  if e.test.(idem).isIdem(v) {
    for _, option := range e.options {
      if !option.(idem).isIdem(v) {
        return false
      }
    }
    return true
  }
  return false
}

func (e *CaseOption)isIdem(v *Validator) bool {
  for _, value := range e.values {
    if !value.(idem).isIdem(v) {
      return false
    }
  }
  return e.Then().(idem).isIdem(v)
}

func (e *ConcatenatedString)isIdem(v *Validator) bool {
  return true
}

func (e *HeredocExpression)isIdem(v *Validator) bool {
  return true
}

func (e *IfExpression)isIdem(v *Validator) bool {
  return e.Test().(idem).isIdem(v) && e.Then().(idem).isIdem(v) && e.Else().(idem).isIdem(v)
}

func (e *literalExpression)isIdem(v *Validator) bool {
  return true
}

func (e *LiteralHash)isIdem(v *Validator) bool {
  return true
}

func (e *LiteralList)isIdem(v *Validator) bool {
  return true
}

func (e *Nop)isIdem(v *Validator) bool {
  return true
}

func (e *ParenthesizedExpression)isIdem(v *Validator) bool {
  return e.expr.(idem).isIdem(v)
}

func (e *RelationshipExpression)isIdem(v *Validator) bool {
  return false
}

func (e *RenderExpression)isIdem(v *Validator) bool {
  return false
}

func (e *RenderStringExpression)isIdem(v *Validator) bool {
  return false
}

func (e *SelectorExpression)isIdem(v *Validator) bool {
  return true
}

func (e *unaryExpression)isIdem(v *Validator) bool {
  return true
}
