package validator

import (
  "testing"
  . "github.com/puppetlabs/go-parser/issue"
  . "github.com/puppetlabs/go-parser/parser"
  . "github.com/puppetlabs/go-parser/internal/testutils"
)

func TestVariableAssignValidation(t *testing.T) {
  expectNoIssues(t, `$x = 'y'`)
}

func TestNumericVariableAssignValidation(t *testing.T) {
  expectIssues(t, `$1 = 'y'`, VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT)
}

func TestMultipleVariableAssign(t *testing.T) {
  expectNoIssues(t, `[$a, $b] = 'y'`)
  expectIssues(t, `[$a, $1] = 'y'`, VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT)
  expectIssues(t, `[$a, $b['h']] = 'y'`, VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX)
  expectIssues(t, `[$a, $b::z] = 'y'`, VALIDATE_CROSS_SCOPE_ASSIGNMENT)
}

func TestAccessAssignValidation(t *testing.T) {
  expectIssues(t, `$x['h'] = 'y'`, VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX)
}

func TestAppendsDeletesValidation(t *testing.T) {
  expectIssues(t, `$x += 'y'`, VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED)
  expectIssues(t, `$x -= 'y'`, VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED)
}

func TestNamespaceAssignValidation(t *testing.T) {
  expectIssues(t, `$x::z = 'y'`, VALIDATE_CROSS_SCOPE_ASSIGNMENT)
}

func TestAttributeAppendValidation(t *testing.T) {
  expectNoIssues(t, `Service[apache] { require +> File['apache.pem'] }`)

  expectIssues(t, `service { apache: require +> File['apache.pem'] }`, VALIDATE_ILLEGAL_ATTRIBUTE_APPEND)
}

func TestAttributesOpValidation(t *testing.T) {
  expectNoIssues(t,
    Unindent(`
      file { '/tmp/foo':
        ensure => file,
        * => $file_ownership
      }`))

  expectIssues(t,
    Unindent(`
      File <| mode == '0644' |> {
        * => $file_ownership
      }`),
    VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT)

  expectIssues(t,
    Unindent(`
      File {
        ensure => file,
        * => $file_ownership
      }`),
    VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT)

  expectIssues(t,
    Unindent(`
      File['/tmp/foo'] {
        ensure => file,
        * => $file_ownership
      }`),
    VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT)

  expectIssues(t,
    Unindent(`
      file { '/tmp/foo':
        ensure => file,
        * => function foo() {}
      }`),
    VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
}

func TestCallNamedFunctionValidation(t *testing.T) {
  expectNoIssues(t,
    Unindent(`
      include apache
      `))

  expectNoIssues(t,
    Unindent(`
      $x = String(123, 16)
      `))

  expectNoIssues(t,
    Unindent(`
      $x = Enum['a', 'b']('a')
      `))

  expectIssues(t,
    Unindent(`
      $x = enum['a', 'b']('a')
      `),
    VALIDATE_ILLEGAL_EXPRESSION)
}

func TestBinaryOpValidation(t *testing.T) {
  expectIssues(t, `notice(function foo() {} < 3)`, VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
  expectNoIssues(t, `notice(true == !false)`)
}

func TestBlockValidation(t *testing.T) {
  expectIssues(t,
    Unindent(`
      ['a', 'b']
      $x = 3
      `),
    VALIDATE_IDEM_EXPRESSION_NOT_LAST)

  expectIssues(t,
    Unindent(`
      case $z {
      2: { true }
      3: { false }
      default: { false }
      }
      $x = 3
      `),
    VALIDATE_IDEM_EXPRESSION_NOT_LAST)

  expectNoIssues(t,
    Unindent(`
      case $z {
      2: { true }
      3: { false }
      default: { $v = 1 }
      }
      $x = 3
      `))

  expectNoIssues(t,
    Unindent(`
      case ($z = 2) {
      2: { true }
      3: { false }
      default: { false }
      }
      $x = 3
      `))

  expectNoIssues(t,
    Unindent(`
      case $z {
      ($y = 2): { true }
      3: { false }
      default: { false }
      }
      $x = 3
      `))

  expectIssues(t,
    Unindent(`
      if $z { 3 } else { 4 }
      $x = 3
      `),
    VALIDATE_IDEM_EXPRESSION_NOT_LAST)

  expectNoIssues(t,
    Unindent(`
      if $z { $v = 3 } else { $v = 4 }
      $x = 3
      `))

  expectIssues(t,
    Unindent(`
      unless $z { 3 }
      $x = 3
      `),
    VALIDATE_IDEM_EXPRESSION_NOT_LAST)

  expectNoIssues(t,
    Unindent(`
      unless $z { $v = 3 }
      $x = 3
      `))

  expectIssues(t,
    Unindent(`
      (3)
      $x = 3
      `),
    VALIDATE_IDEM_EXPRESSION_NOT_LAST)

  expectNoIssues(t,
    Unindent(`
      ($v = 3)
      $x = 3
      `))
}

func TestCapabilityMappingValidation(t *testing.T) {
  expectNoIssues(t,
    Unindent(`
      Something produces Foo {}
      `))

  expectNoIssues(t,
    Unindent(`
      Something[A] produces Foo {}
      `))

  expectIssues(t,
    Unindent(`
      something produces Foo {}
      `),
    VALIDATE_ILLEGAL_CLASSREF)

  expectIssues(t,
    Unindent(`
      Something produces foo {}
      `),
    VALIDATE_ILLEGAL_CLASSREF)

  expectIssues(t,
    Unindent(`
      Something['A', 'B'] produces Foo {}
      `),
    VALIDATE_ILLEGAL_EXPRESSION)
}


func TestCaseExpressionValidation(t *testing.T) {
  expectNoIssues(t,
    Unindent(`
      case $x {
        'a': { true }
        default: { 'false' }
      }`))

  expectIssues(t,
    Unindent(`
      case $x {
        'a': { true }
        default: { 'false' }
        default: { 'true' }
      }`),
    VALIDATE_DUPLICATE_DEFAULT)

  expectIssues(t,
    Unindent(`
      case $x {
        function foo() {}: { true }
        default: { false }
      }`),
    VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
}

func TestCollectValidation(t *testing.T) {
  expectNoIssues(t,
    Unindent(`
      User <| groups == 'admin' |>`))

  expectIssues(t,
    Unindent(`
      user <| groups == 'admin' |>`),
    VALIDATE_ILLEGAL_EXPRESSION)
}

func TestEppValidation(t *testing.T) {
  expectNoIssuesEPP(t,
    Unindent(`
      <%-| $a, $b |-%>
      This is $a <%= $a %>`))

  expectIssuesEPP(t,
    Unindent(`
      <%-| $a, $b, $a |-%>
      This is $a <%= $a %>`),
    VALIDATE_DUPLICATE_PARAMETER)

  expectIssuesEPP(t,
    Unindent(`
      <%-| $a, *$b |-%>
      This is $a <%= $a %>`),
    VALIDATE_CAPTURES_REST_NOT_SUPPORTED)
}

func TestFunctionDefinitionValidation(t *testing.T) {
  expectNoIssues(t,
    Unindent(`
      function foo() {}`))

  expectIssues(t,
    Unindent(`
      function foo() {
        function bar() {}
      }`),
    VALIDATE_NOT_ABSOLUTE_TOP_LEVEL)

  expectIssues(t,
    Unindent(`
      function foo() >> Application {
      }`),
    VALIDATE_FUTURE_RESERVED_WORD)

  expectIssues(t,
    Unindent(`
      function foo() >> Application[X] {
      }`),
    VALIDATE_FUTURE_RESERVED_WORD)

  expectIssues(t,
    Unindent(`
      function variant() {
      }`),
    VALIDATE_RESERVED_TYPE_NAME)

  expectIssues(t,
    Unindent(`
      function Foo() {
      }`),
    VALIDATE_ILLEGAL_DEFINITION_NAME)
}

func expectNoIssues(t *testing.T, str string) {
  expectIssuesX(t, str, false)
}

func expectNoIssuesEPP(t *testing.T, str string) {
  expectIssuesX(t, str, true)
}

func expectIssues(t *testing.T, str string, expectedIssueCodes...IssueCode) {
  expectIssuesX(t, str, false, expectedIssueCodes...)
}

func expectIssuesEPP(t *testing.T, str string, expectedIssueCodes...IssueCode) {
  expectIssuesX(t, str, true, expectedIssueCodes...)
}

func expectIssuesX(t *testing.T, str string, eppMode bool, expectedIssueCodes...IssueCode) {
  issues := parseAndValidate(t, str, eppMode)
  if issues == nil {
    return
  }
  nextCode: for _, expectedIssueCode := range expectedIssueCodes {
    for _, issue := range issues {
      if expectedIssueCode == issue.Code() {
        continue nextCode
      }
    }
    t.Errorf(`Expected issue '%s' but it was not produced`, expectedIssueCode)
  }

  nextIssue: for _, issue := range issues {
    for _, expectedIssueCode := range expectedIssueCodes {
      if expectedIssueCode == issue.Code() {
        continue nextIssue
      }
    }
    t.Errorf(`Unexpected issue %s: '%s'`, issue.Code(), issue.String())
  }
}

func parseAndValidate(t *testing.T, str string, eppMode bool) []*ReportedIssue {
  if expr := parse(t, str, eppMode); expr != nil {
    v := ValidatePuppet(expr)
    return v.Issues()
  }
  return nil
}

func parse(t *testing.T, str string, eppMode bool) *Program {
  expr, err := Parse(``, str, eppMode)
  if err != nil {
    t.Errorf(err.Error())
    return nil
  }
  block, ok := expr.(*Program)
  if !ok {
    t.Errorf("'%s' did not parse to a program", str)
    return nil
  }
  return block
}
