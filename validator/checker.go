package validator

import (
  . "regexp"
  . "github.com/puppetlabs/go-parser/issue"
  . "github.com/puppetlabs/go-parser/parser"
  . "github.com/puppetlabs/go-parser/literal"
)

var NUMERIC_VAR_NAME_EXPR = MustCompile(`\A(?:0|(?:[1-9][0-9]*))\z`)
var DOUBLE_COLON_EXPR = MustCompile(`::`)

// CLASSREF_EXT matches a class reference the same way as the lexer - i.e. the external source form
// where each part must start with a capital letter A-Z.
var CLASSREF_EXT = MustCompile(`\A(?:::)?[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)


// CLASSREF matches a class reference the way it is represented internally in the
// model (i.e. in lower case).
var CLASSREF_DECL = MustCompile(`\A[a-z][\w]*(?:::[a-z][\w]*)*\z`)

var RESERVED_TYPE_NAMES = map[string]bool {
  `type`: true,
  `any`: true,
  `unit`: true,
  `scalar`: true,
  `boolean`: true,
  `numeric`: true,
  `integer`: true,
  `float`: true,
  `collection`: true,
  `array`: true,
  `hash`: true,
  `tuple`: true,
  `struct`: true,
  `variant`: true,
  `optional`: true,
  `enum`: true,
  `regexp`: true,
  `pattern`: true,
  `runtime`: true,

  `init`: true,
  `object`: true,
  `sensitive`: true,
  `semver`: true,
  `semverrange`: true,
  `string`: true,
  `timestamp`: true,
  `timespan`: true,
  `typeset`: true,
}

var FUTURE_RESERVED_WORDS = map[string]bool {
  `application`: true,
  `produces`: true,
  `consumes`: true,
}

var RESERVED_PARAMETERS = map[string]bool {
  `name`: true,
  `title`: true,
}

type Checker struct {
  AbstractValidator
}

func NewChecker() *Checker {
  checker := &Checker{AbstractValidator{nil, nil, make([]*ReportedIssue, 0, 5), make(map[IssueCode]Severity, 5)}}
  checker.Demote(VALIDATE_IDEM_EXPRESSION_NOT_LAST, SEVERITY_WARNING)
  checker.Demote(VALIDATE_FUTURE_RESERVED_WORD, SEVERITY_DEPRECATION)
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
  case *CollectExpression:
    v.check_CollectExpression(e.(*CollectExpression))
  case *EppExpression:
    v.check_EppExpression(e.(*EppExpression))
  case *FunctionDefinition:
    v.check_FunctionDefinition(e.(*FunctionDefinition))
  case *HostClassDefinition:
    v.check_HostClassDefinition(e.(*HostClassDefinition))
  case *IfExpression:
    v.check_IfExpression(e.(*IfExpression))
  case *KeyedEntry:
    v.check_KeyedEntry(e.(*KeyedEntry))
  case *LambdaExpression:
    v.check_LambdaExpression(e.(*LambdaExpression))
  case *LiteralHash:
    v.check_LiteralHash(e.(*LiteralHash))
  case *ResourceTypeDefinition:
    v.check_ResourceTypeDefinition(e.(*ResourceTypeDefinition))
  case *UnlessExpression:
    v.check_UnlessExpression(e.(*UnlessExpression))

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
  exprOk := false
  switch e.Component().(type) {
  case *QualifiedReference:
    exprOk = true

  case *QualifiedName:
    v.Accept(VALIDATE_ILLEGAL_CLASSREF, e.Component(), e.Component().(*QualifiedName).Name())
    exprOk = true // OK, besides from what was just reported

  case *AccessExpression:
    ae, _ := e.Component().(*AccessExpression)
    if _, ok := ae.Operand().(*QualifiedReference); ok && len(ae.Keys()) == 1 {
      switch ae.Keys()[0].(type) {
      case *LiteralString, *QualifiedName, *QualifiedReference:
        exprOk = true
      }
    }
  }

  if !exprOk {
    v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Component(), A_an(e.Component()), `capability mapping`, A_an(e))
  }

  if !CLASSREF_EXT.MatchString(e.Capability()) {
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

func (v *Checker) check_CollectExpression(e *CollectExpression) {
  if _, ok := e.ResourceType().(*QualifiedReference); !ok {
    v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.ResourceType(), A_an(e), `type name`)
  }
}

func (v *Checker) check_EppExpression(e *EppExpression) {
  p := v.Container()
  if lambda, ok := p.(*LambdaExpression); ok {
    v.checkNoCapture(lambda, lambda.Parameters())
    v.checkParameterNameUniqueness(lambda, lambda.Parameters())
  }
}

func (v *Checker) check_FunctionDefinition(e *FunctionDefinition) {
  v.check_NamedDefinition(e)
  v.checkCaptureLast(e, e.Parameters())
  v.checkReturnType(e, e.ReturnType())
}

func (v *Checker) check_HostClassDefinition(e *HostClassDefinition) {
  v.check_NamedDefinition(e)
  v.checkNoCapture(e, e.Parameters())
  v.checkReservedParams(e, e.Parameters())
  v.checkNoIdemLast(e)
}

func (v *Checker) check_IfExpression(e *IfExpression) {
  v.checkRValue(e.Test())
}

func (v *Checker) check_KeyedEntry(e *KeyedEntry) {
  v.checkRValue(e.Key())
  v.checkRValue(e.Value())
}

func (v *Checker) check_LambdaExpression(e *LambdaExpression) {
  v.checkCaptureLast(e, e.Parameters())
  v.checkReturnType(e, e.ReturnType())
}

func (v *Checker) check_LiteralHash(e *LiteralHash) {
  unique := make(map[interface{}]bool, len(e.Entries()))
  for _, entry := range e.Entries() {
    key := entry.(*KeyedEntry).Key()
    if literalKey, ok := ToLiteral(key); ok {
      if _, ok = unique[literalKey]; ok {
        v.Accept(VALIDATE_DUPLICATE_KEY, entry, key.String())
      } else {
        unique[literalKey] = true
      }
    }
  }
}

func (v *Checker) check_NamedDefinition(e NamedDefinition) {
  v.checkTop(e, v.Container())
  if !CLASSREF_DECL.MatchString(e.Name()) {
    v.Accept(VALIDATE_ILLEGAL_DEFINITION_NAME, e, e.Name(), A_an(e))
  }
  v.checkReservedTypeName(e, e.Name())
  v.checkFutureReservedWord(e, e.Name())
  v.checkParameterNameUniqueness(e, e.Parameters())
}

func (v *Checker) check_ResourceTypeDefinition(e *ResourceTypeDefinition) {
  v.check_NamedDefinition(e)
  v.checkNoCapture(e, e.Parameters())
  v.checkReservedParams(e, e.Parameters())
  v.checkNoIdemLast(e)
}

func (v *Checker) check_UnlessExpression(e *UnlessExpression) {
  v.checkRValue(e.Test())
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

func (v *Checker) checkFutureReservedWord(e Expression, w string) {
  if _, ok := FUTURE_RESERVED_WORDS[w]; ok {
    v.Accept(VALIDATE_FUTURE_RESERVED_WORD, e, w)
  }
}

func (v *Checker) checkNoIdemLast(e NamedDefinition) {
  if violator := v.endsWithIdem(e.Body().(*BlockExpression)); violator != nil && !v.isResourceWithoutTitle(violator) {
    v.Accept(VALIDATE_IDEM_NOT_ALLOWED_LAST, violator, violator.Label(), A_anUc(e))
  }
}

func (v *Checker) endsWithIdem(e *BlockExpression) Expression {
  top := len(e.Statements())
  if top > 0 {
    last := e.Statements()[top-1]
    if v.isIdem(last) {
      return last
    }
  }
  return nil
}

func (v *Checker) checkCaptureLast(container Expression, parameters []Expression) {
  last := len(parameters) - 1
  for idx := 0; idx < last; idx++ {
    if param, ok := parameters[idx].(*Parameter); ok && param.CapturesRest() {
      v.Accept(VALIDATE_CAPTURES_REST_NOT_LAST, param, param.Name())
    }
  }
}

func (v *Checker) checkNoCapture(container Expression, parameters []Expression) {
  for _, parameter := range parameters {
    if param, ok := parameter.(*Parameter); ok && param.CapturesRest() {
      v.Accept(VALIDATE_CAPTURES_REST_NOT_SUPPORTED, param, param.Name(), A_an(container))
    }
  }
}

func (v *Checker) checkParameterNameUniqueness(container Expression, parameters []Expression) {
  unique := make(map[string]bool, 10)
  for _, parameter := range parameters {
    param := parameter.(*Parameter)
    if _, found := unique[param.Name()]; found {
      v.Accept(VALIDATE_DUPLICATE_PARAMETER, parameter, param.Name())
    } else {
      unique[param.Name()] = true
    }
  }
}

func (v *Checker) checkReservedParams(container Expression, parameters []Expression) {
  for _, parameter := range parameters {
    param := parameter.(*Parameter)
    if _, ok := RESERVED_PARAMETERS[param.Name()]; ok {
      v.Accept(VALIDATE_RESERVED_PARAMETER, container, param.Name(), A_an(container))
    }
  }
}

func (v *Checker) checkReservedTypeName(e Expression, w string) {
  if _, ok := RESERVED_TYPE_NAMES[w]; ok {
    v.Accept(VALIDATE_RESERVED_TYPE_NAME, e, w, A_an(e))
  }
}

func (v *Checker) checkReturnType(function Expression, returnType Expression) {
  if returnType != nil {
    v.checkTypeRef(function, returnType)
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

func (v *Checker) checkTop(e Expression, c Expression) {
  switch c.(type) {
  case nil, *HostClassDefinition, *Program:
    return

  case *BlockExpression:
    c = v.ContainerOf(c)
    if _, ok := c.(*Program); !ok {
      switch e.(type) {
      case *FunctionDefinition, *TypeAlias, *TypeDefinition:
        // not ok. These can never be nested in a block
        v.Accept(VALIDATE_NOT_ABSOLUTE_TOP_LEVEL, e, A_anUc(e))
        return
      }
    }
    v.checkTop(e, c)

  default:
    v.Accept(VALIDATE_NOT_TOP_LEVEL, e, A_anUc(e))
  }
}

func (v *Checker) checkTypeRef(function Expression, r Expression) {
  n := r
  if ae, ok := r.(*AccessExpression); ok {
    n = ae.Operand();
  }
  if qr, ok := n.(*QualifiedReference); ok {
    v.checkFutureReservedWord(r, qr.DowncasedName())
  } else {
    v.Accept(VALIDATE_ILLEGAL_EXPRESSION, r, `a type reference`, A_an(function))
  }
}

// Checks if the expression has side effect ('idem' is latin for 'the same', here meaning that the evaluation state
// is known to be unchanged after the expression has been evaluated). The result is not 100% authoritative for
// negative answers since analysis of function behavior is not possible.
func (v *Checker) isIdem(e Expression) bool {
  switch e.(type) {
  case nil, *AccessExpression, *ConcatenatedString, *HeredocExpression, *LiteralList, *LiteralHash, *Nop, *SelectorExpression:
    return true
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
  case *AssignmentExpression, *RelationshipExpression, *RenderExpression, *RenderStringExpression:
    return false
  case BinaryExpression, LiteralValue, UnaryExpression:
    return true
  default:
    return false
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

func (v *Checker) isResourceWithoutTitle(e Expression) bool {
  if be, ok := e.(*BlockExpression); ok {
    statements := be.Statements()
    if len(statements) == 2 {
      _, ok = statements[0].(*QualifiedReference)
      if ok {
        _, ok = statements[1].(*LiteralHash)
        return ok
      }
    }
  }
  return false
}
