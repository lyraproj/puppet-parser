package pspec

import (
  . "github.com/puppetlabs/go-parser/parser"
  . "github.com/puppetlabs/go-parser/issue"
  . "bytes"
  . "fmt"
  "github.com/puppetlabs/go-parser/internal/testutils"
  "github.com/puppetlabs/go-parser/validator"
)

type (
  Assertions interface {
    AssertEquals(a interface{}, b interface{})

    Fail(message string)
  }

  Executable func(assertions Assertions)

  SpecEvaluator struct {
    path []Expression
  }

  SpecFunction func (s *SpecEvaluator, semantic Expression, args []interface{}) interface{}

  Input interface {
    CreateOkTests(expected Result) []Executable

    CreateIssuesTests(expected Result) []Executable
  }

  Node interface {
    Description() string
    CreateTest() Test
  }

  Result interface {
    CreateTest(actual interface{}) Executable
  }

  Test interface {
    Name() string
  }

  Example struct {
    description string
    given *Given
    result Result
  }

  Examples struct {
    description string
    children []Node
  }

  Given struct {
    inputs []Input
  }

  ParseResult struct {
    expected string
  }

  Source struct {
    parser ExpressionParser
    sources []string
  }

  TestExecutable struct {
    name string
    test Executable
  }

  TestGroup struct {
    name string
    tests []Test
  }

  ValidationResult struct {
    issue *Issue
    severity Severity
  }

  ValidatesWith struct {
    expectedIssues []*ValidationResult
  }
)

func NewSpecEvaluator() *SpecEvaluator {
  return &SpecEvaluator{}
}

var functions = map[string]SpecFunction {
  `Error`: errorIssue,
  `Example`: example,
  `Examples`: examples,
  `Given`: given,
  `Parses_to`: parsesTo,
  `Source`: source,
  `Unindent`: unindent,
  `Validates_with`: validatesWith,
  `Warning`: warningIssue,
}

func errorIssue(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  s.assertVariableOrParameterTo(`Error`, `Validates_with`)
  if len(args) != 1 {
    panic(s.specError(SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS, semantic, `Error`, `1`, len(args)))
  }
  if issue, ok := args[0].(*Issue); ok {
    return &ValidationResult{issue,SEVERITY_ERROR}
  }
  panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Error`, 1, `Issue`, Sprintf(`%T`, args[0])))
}

func example(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  if !s.atTop() {
    s.assertVariableOrParameterTo(`Example`, `Examples`)
  }
  if len(args) != 3 {
    panic(s.specError(SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS, semantic, `Example`, '3', len(args)))
  }
  var (
    example Example
    ok bool
  )
  if example.description, ok = args[0].(string); !ok {
    panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Example`, 1, `string`, Sprintf(`%T`, args[0])))
  }
  if example.given, ok = args[1].(*Given); !ok {
    panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Example`, 2, `Given`, Sprintf(`%T`, args[1])))
  }
  if example.result, ok = args[2].(Result); !ok {
    panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Example`, 3, `Result`, Sprintf(`%T`, args[2])))
  }
  return &example
}

func examples(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  if !s.atTop() {
    s.assertVariableOrParameterTo(`Examples`, `Examples`)
  }
  var examples Examples
  ok := false
  top := len(args)
  if top == 0 {
    panic(s.specError(SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS, semantic, `Example`, `at least one`, len(args)))
  }
  if examples.description, ok = args[0].(string); !ok {
    panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Examples`, 1, `string`, Sprintf(`%T`, args[0])))
  }
  examples.children = make([]Node, top - 1)
  for idx := 1; idx < top; idx++ {
    var node Node
    if node, ok = args[idx].(Node); !ok {
      panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Examples`, idx + 1, `Examples or Example`, Sprintf(`%T`, args[idx])))
    }
    examples.children[idx - 1] = node
  }
  return &examples
}

func given(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  s.assertVariableOrParameterTo(`Given`, `Example`)

  var (
    str string
    ok bool
    input Input
  )

  top := len(args)
  inputs := make([]Input, top)
  for idx := 0; idx < top; idx++ {
    arg := args[idx]
    if str, ok = arg.(string); ok {
      inputs[idx] = &Source{CreateParser(), []string{str}}
    } else if input, ok = arg.(Input); ok {
      inputs[idx] = input
    } else {
      panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Given`, idx + 1, `Input`, Sprintf(`%T`, args[idx])))
    }
  }
  return &Given{inputs}
}

func parsesTo(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  s.assertVariableOrParameterTo(`Parses_to`, `Example`)
  if len(args) != 1 {
    panic(s.specError(SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS, semantic, `Parses_to`, `1`, len(args)))
  }
  if str, ok := args[0].(string); ok {
    return &ParseResult{str}
  }
  panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Parses_to`, 1, `string`, Sprintf(`%T`, args[0])))
}

func source(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  s.assertVariableOrParameterTo(`Source`, `Given`)

  var (
    str string
    ok bool
  )
  top := len(args)
  sources := make([]string, top)
  for idx := 0; idx < top; idx++ {
    if str, ok = args[idx].(string); !ok {
      panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Source`, idx+1, `string`, Sprintf(`%T`, args[idx])))
    }
    sources[idx] = str
  }
  return &Source{CreateParser(), sources}
}

func unindent(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  if len(args) != 1 {
    panic(s.specError(SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS, semantic, `Unindent`, `1`, len(args)))
  }
  if str, ok := args[0].(string); ok {
    return testutils.Unindent(str)
  }
  panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Unindent`, 1, `string`, Sprintf(`%T`, args[0])))
}

func validatesWith(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  s.assertVariableOrParameterTo(`Validates_with`, `Example`)
  var (
    result *ValidationResult
    ok bool
  )
  top := len(args)
  results := make([]*ValidationResult, top)
  for idx := 0; idx < top; idx++ {
    if result, ok = args[idx].(*ValidationResult); !ok {
      panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `ValidationResult`, idx+1, `string`, Sprintf(`%T`, args[idx])))
    }
    results[idx] = result
  }
  return &ValidatesWith{results}
}

func warningIssue(s *SpecEvaluator, semantic Expression, args []interface{}) interface{} {
  s.assertVariableOrParameterTo(`Error`, `Validates_with`)
  if len(args) != 1 {
    panic(s.specError(SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS, semantic, `Warning`, `1`, len(args)))
  }
  if issue, ok := args[0].(*Issue); ok {
    return &ValidationResult{issue,SEVERITY_WARNING}
  }
  panic(s.specError(SPEC_ILLEGAL_ARGUMENT_TYPE, semantic, `Warning`, 1, `Issue`, Sprintf(`%T`, args[0])))
}

func (s *SpecEvaluator) atTop() bool {
  if len(s.path) == 3 {
    if _, ok := s.path[0].(*Program); ok {
      _, ok = s.path[1].(*BlockExpression)
      return ok
    }
  }
  return false
}

func (s *SpecEvaluator) assertVariableOrParameterTo(exprName string, callName string) {
  pathSize := len(s.path)
  if pathSize > 1 {
    container := s.path[pathSize-2]
    if callNamed, ok := container.(*CallNamedFunctionExpression); ok {
      var functor *QualifiedReference
      if functor, ok = callNamed.Functor().(*QualifiedReference); ok && callName == functor.String() {
        return
      }
    } else {
      var asgn *AssignmentExpression
      if asgn, ok = container.(*AssignmentExpression); ok {
        if _, ok = asgn.Lhs().(*VariableExpression); ok {
          return
        }
      }
    }
  }
  panic(s.specError(SPEC_EXPRESSION_NOT_PARAMETER_TO, s.path[pathSize-1], exprName, callName))
}

func (s *SpecEvaluator) specError(issueCode IssueCode, semantic Expression, args ...interface{}) *ReportedIssue {
  return NewReportedIssue(issueCode, SEVERITY_ERROR, args, semantic)
}

func (s *SpecEvaluator) CreateTests(expression Expression) []Test {
  s.path = make([]Expression, 0, 5)
  tests := make([]Test, 0, 64)
  if nodes, ok := s.eval(expression).([]interface{}); ok {
    var node Node
    for _, v := range nodes {
      if node, ok = v.(Node); ok {
        tests = append(tests, node.CreateTest())
      }
    }
  }
  return tests
}

func (s *SpecEvaluator) eval(expression Expression) interface{} {
  switch expression.(type) {
  case *BlockExpression:
    return s.eval_BlockExpression(expression.(*BlockExpression))
  case *CallNamedFunctionExpression:
    return s.eval_CallNamedFunctionExpression(expression.(*CallNamedFunctionExpression))
  case *Program:
    return s.eval_Program(expression.(*Program))
  case *QualifiedReference:
    return s.eval_QualifiedReference(expression.(*QualifiedReference))
  case *LiteralString:
    return s.eval_LiteralString(expression.(*LiteralString))
  default:
    panic(s.specError(validator.VALIDATE_ILLEGAL_EXPRESSION, expression, A_anUc(expression), `entry`, `a pspec`))
  }
}

func (s *SpecEvaluator) eval_BlockExpression(block *BlockExpression) interface{} {
  path := s.path
  s.path = append(path, block)

  statements := block.Statements()
  top := len(statements)
  result := make([]interface{}, top)
  for idx, statement := range statements {
    result[idx] = s.eval(statement)
  }

  s.path = path
  return result
}

func (s *SpecEvaluator) eval_CallNamedFunctionExpression(call *CallNamedFunctionExpression) interface{} {
  path := s.path
  s.path = append(path, call)
  var (
    fn SpecFunction
    ok bool
  )
  if fn, ok = s.eval(call.Functor()).(SpecFunction); !ok {
    panic(s.specError(SPEC_ILLEGAL_CALL_RECEIVER, call.Functor()))
  }
  args := make([]interface{}, len(call.Arguments()))
  for idx, arg := range call.Arguments() {
    args[idx] = s.eval(arg)
  }
  result := fn(s, call, args)
  s.path = path
  return result
}

func (s *SpecEvaluator) eval_LiteralString(lit *LiteralString) interface{} {
  return lit.StringValue()
}

func (s *SpecEvaluator) eval_Program(program *Program) interface{} {
  path := s.path
  s.path = append(path, program)
  result := s.eval(program.Body())
  s.path = path
  return result
}

func (s *SpecEvaluator) eval_QualifiedReference(qr *QualifiedReference) interface{} {
  fn, ok := functions[qr.Name()]
  if ok {
    return fn
  }
  defer func() {
    if r := recover(); r != nil {
      panic(s.specError(SPEC_UNKNOWN_IDENTIFIER, qr, qr.Name()))
    }
  }()
  return IssueForCode(IssueCode(qr.Name()))
}

func (e *Example) CreateTest() Test {
  tests := make([]Executable, 0, 8)
  if _, ok := e.result.(*ValidatesWith); ok {
    for _, input := range e.given.inputs {
      tests = append(tests, input.CreateIssuesTests(e.result)...)
    }
  } else {
    for _, input := range e.given.inputs {
      tests = append(tests, input.CreateOkTests(e.result)...)
    }
  }
  test := func(assertions Assertions) {
    for _, test := range tests {
      test(assertions)
    }
  }
  return &TestExecutable{e.description, test}
}

func (e *Example) Description() string {
  return e.description
}

func (e *Examples) CreateTest() Test {
  tests := make([]Test, len(e.children))
  for idx, child := range e.children {
    tests[idx] = child.CreateTest()
  }
  return &TestGroup{e.description, tests}
}

func (e *Examples) Description() string {
  return e.description
}

func (p *ParseResult) CreateTest(actual interface{}) Executable {
  actualPN := actual.(Expression).ToPN()
  expectedPN := ParsePN(``, p.expected)
  return func(assertions Assertions) {
    assertions.AssertEquals(expectedPN.String(), actualPN.String())
  }
}

func (i *Source)  CreateOkTests(expected Result) []Executable {
  result := make([]Executable, len(i.sources))
  for idx, source := range i.sources {
    result[idx] = i.createOkTest(source, expected)
  }
  return result
}

func (i *Source)  CreateIssuesTests(expected Result) []Executable {
  result := make([]Executable, len(i.sources))
  for idx, source := range i.sources {
    result[idx] = i.createIssuesTest(source, expected)
  }
  return result
}

func (i *Source)  createOkTest(source string, expected Result) Executable {
  actual, err := i.parser.Parse(``, source, false, true)
  if err != nil {
    return func(assertions Assertions) {
      assertions.Fail(err.Error())
    }
  }
  return expected.CreateTest(actual)
}

func (i *Source)  createIssuesTest(source string, expected Result) Executable {
  _, err := i.parser.Parse(``, source, false, true)
  if issue, ok := err.(*ReportedIssue); ok {
    return expected.CreateTest([]*ReportedIssue { issue })
  } else if err != nil {
    return func(assertions Assertions) {
      assertions.Fail(err.Error())
    }
  }
  return expected.CreateTest([]*ReportedIssue {})
}

func (v *TestExecutable) Name() string {
  return v.name
}

func (v *TestExecutable) Executable() Executable {
  return v.test
}

func (v *TestGroup) Name() string {
  return v.name
}

func (v *TestGroup) Tests() []Test {
  return v.tests
}

func (v *ValidatesWith) CreateTest(actual interface{}) Executable {
  issues := actual.([]*ReportedIssue)
  return func(assertions Assertions) {
    bld := NewBufferString(``)
    nextExpected: for _, expected := range v.expectedIssues {
      for _, issue := range issues {
        if expected.issue.Code() == issue.Code() {
          continue nextExpected
        }
      }
      Fprint(bld, `Expected %s %s but it was not produced`, expected.severity.String(), expected.issue.Code())
    }

    nextIssue: for _, issue := range issues {
      for _, expected := range v.expectedIssues {
        if expected.issue.Code() == issue.Code() {
          continue nextIssue
        }
        Fprint(bld, `Unexpected %s %s`, expected.severity.String(), expected.issue.Code())
      }
    }
    if bld.Len() > 0 {
      assertions.Fail(bld.String())
    }
  }
}