package validator

import (
  "testing"
  . "github.com/puppetlabs/go-parser/parser"
  . "github.com/puppetlabs/go-parser/testutils"
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

func TestBinaryOpValidation(t *testing.T) {
  expectIssues(t, `notice(function foo() {} < 3)`, VALIDATE_NOT_RVALUE)
  expectNoIssues(t, `notice(true == !false)`)
}

func TestBlockValidation(t *testing.T) {
  expectIssues(t,
    Unindent(`
      ['a', 'b']
      $x = 3
      `),
    VALIDATE_IDEM_EXPRESSION_NOT_LAST)
}

func expectNoIssues(t *testing.T, str string) {
  expectIssues(t, str)
}

func expectIssues(t *testing.T, str string, expectedIssueCodes...string) {
  issues := parseAndValidate(t, str)
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

func parseAndValidate(t *testing.T, str string) []*ReportedIssue {
  if expr := parse(t, str); expr != nil {
    v := ValidatePuppet(expr)
    return v.Issues()
  }
  return nil
}

func parse(t *testing.T, str string) *BlockExpression {
  expr, err := Parse(``, str, false)
  if err != nil {
    t.Errorf(err.Error())
    return nil
  }
  block, ok := expr.(*BlockExpression)
  if !ok {
    t.Errorf("'%s' did not parse to a block", str)
    return nil
  }
  return block
}
