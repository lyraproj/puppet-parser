package parser

import (
  "fmt"
  "bytes"
)

const (
  // Lexer issues
  LEX_DOUBLE_COLON_NOT_FOLLOWED_BY_NAME = `DOUBLE_COLON_NOT_FOLLOWED_BY_NAME`
  LEX_DIGIT_EXPECTED                    = `LEX_DIGIT_EXPECTED`
  LEX_HEREDOC_EMPTY_TAG                 = `LEX_HEREDOC_EMPTY_TAG`
  LEX_HEREDOC_ILLEGAL_ESCAPE            = `LEX_HEREDOC_ILLEGAL_ESCAPE`
  LEX_HEREDOC_MULTIPLE_ESCAPE           = `LEX_HEREDOC_MULTIPLE_ESCAPE`
  LEX_HEREDOC_MULTIPLE_SYNTAX           = `LEX_HEREDOC_MULTIPLE_SYNTAX`
  LEX_HEREDOC_MULTIPLE_TAG              = `LEX_HEREDOC_MULTIPLE_TAG`
  LEX_HEREDOC_DECL_UNTERMINATED         = `LEX_HEREDOC_DECL_UNTERMINATED`
  LEX_HEREDOC_UNTERMINATED              = `LEX_HEREDOC_UNTERMINATED`
  LEX_HEXDIGIT_EXPECTED                 = `LEX_HEXDIGIT_EXPECTED`
  LEX_INVALID_NAME                      = `LEX_INVALID_NAME`
  LEX_INVALID_OPERATOR                  = `LEX_INVALID_OPERATOR`
  LEX_INVALID_TYPE_NAME                 = `LEX_INVALID_TYPE_NAME`
  LEX_INVALID_VARIABLE_NAME             = `LEX_INVALID_VARIABLE_NAME`
  LEX_MALFORMED_INTERPOLATION           = `LEX_MALFORMED_INTERPOLATION`
  LEX_MALFORMED_UNICODE_ESCAPE          = `LEX_MALFORMED_UNICODE_ESCAPE`
  LEX_OCTALDIGIT_EXPECTED               = `LEX_OCTALDIGIT_EXPECTED`
  LEX_UNBALANCED_EPP_COMMENT            = `LEX_UNBALANCED_EPP_COMMENT`
  LEX_UNEXPECTED_TOKEN                  = `LEX_UNEXPECTED_TOKEN`
  LEX_UNTERMINATED_COMMENT              = `LEX_UNTERMINATED_COMMENT`
  LEX_UNTERMINATED_STRING               = `LEX_UNTERMINATED_STRING`

  PARSE_CLASS_NOT_VALID_HERE              = `PARSE_CLASS_NOT_VALID_HERE`
  PARSE_ELSIF_IN_UNLESS                   = `PARSE_ELSIF_IN_UNLESS`
  PARSE_EXPECTED_ATTRIBUTE_NAME           = `PARSE_EXPECTED_ATTRIBUTE_NAME`
  PARSE_EXPECTED_CLASS_NAME               = `PARSE_EXPECTED_CLASS_NAME`
  PARSE_EXPECTED_FARROW_AFTER_KEY         = `PARSE_EXPECTED_FARROW_AFTER_KEY`
  PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT = `PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT`
  PARSE_EXPECTED_NAME_AFTER_FUNCTION      = `PARSE_EXPECTED_NAME_AFTER_FUNCTION`
  PARSE_EXPECTED_HOSTNAME                 = `PARSE_EXPECTED_HOSTNAME`
  PARSE_EXPECTED_TITLE                    = `PARSE_EXPECTED_TITLE`
  PARSE_EXPECTED_TOKEN                    = `PARSE_EXPECTED_TOKEN`
  PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE     = `PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE`
  PARSE_EXPECTED_VARIABLE                 = `PARSE_EXPECTED_VARIABLE`
  PARSE_ILLEGAL_EPP_PARAMETERS            = `PARSE_ILLEGAL_EPP_PARAMETERS`
  PARSE_INVALID_RESOURCE                  = `PARSE_INVALID_RESOURCE`
  PARSE_INVALID_ATTRIBUTE                 = `PARSE_INVALID_ATTRIBUTE`
  PARSE_INHERITS_MUST_BE_TYPE_NAME        = `PARSE_INHERITS_MUST_BE_TYPE_NAME`
  PARSE_RESOURCE_WITHOUT_TITLE            = `PARSE_RESOURCE_WITHOUT_TITLE`
  PARSE_QUOTED_NOT_VALID_NAME             = `PARSE_QUOTED_NOT_VALID_NAME`

  VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED = `VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED`
  VALIDATE_CROSS_SCOPE_ASSIGNMENT              = `VALIDATE_CROSS_SCOPE_ASSIGNMENT`
  VALIDATE_IDEM_EXPRESSION_NOT_LAST            = `VALIDATE_IDEM_EXPRESSION_NOT_LAST`
  VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX        = `VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX`
  VALIDATE_ILLEGAL_ATTRIBUTE_APPEND            = `VALIDATE_ILLEGAL_ATTRIBUTE_APPEND`
  VALIDATE_ILLEGAL_EXPRESSION                  = `VALIDATE_ILLEGAL_EXPRESSION`
  VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT          = `VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT`
  VALIDATE_NOT_RVALUE                          = `VALIDATE_NOT_RVALUE`
)

func init() {
  newIssue(LEX_DOUBLE_COLON_NOT_FOLLOWED_BY_NAME, `:: not followed by name segment`, false)
  newIssue(LEX_DIGIT_EXPECTED, `digit expected`, false)
  newIssue(LEX_HEXDIGIT_EXPECTED, `hexadecimal digit expected`, false)
  newIssue(LEX_HEREDOC_DECL_UNTERMINATED, `unterminated @(`, false)
  newIssue(LEX_HEREDOC_EMPTY_TAG, `empty heredoc tag`, false)
  newIssue(LEX_HEREDOC_ILLEGAL_ESCAPE, `illegal heredoc escape '%s'`, false)
  newIssue(LEX_HEREDOC_MULTIPLE_ESCAPE, `more than one declaration of escape flags in heredoc`, false)
  newIssue(LEX_HEREDOC_MULTIPLE_SYNTAX, `more than one syntax declaration in heredoc`, false)
  newIssue(LEX_HEREDOC_MULTIPLE_TAG, `more than one tag declaration in heredoc`, false)
  newIssue(LEX_HEREDOC_UNTERMINATED, `unterminated heredoc`, false)
  newIssue(LEX_INVALID_OPERATOR, `invalid operator '%s'`, false)
  newIssue(LEX_INVALID_NAME, `invalid name`, false)
  newIssue(LEX_INVALID_TYPE_NAME, `invalid type name`, false)
  newIssue(LEX_INVALID_VARIABLE_NAME, `invalid variable name`, false)
  newIssue(LEX_MALFORMED_INTERPOLATION, `malformed interpolation expression`, false)
  newIssue(LEX_MALFORMED_UNICODE_ESCAPE, `malformed unicode escape sequence`, false)
  newIssue(LEX_OCTALDIGIT_EXPECTED, `octal digit expected`, false)
  newIssue(LEX_UNBALANCED_EPP_COMMENT, `unbalanced epp comment`, false)
  newIssue(LEX_UNEXPECTED_TOKEN, `unexpected token '%s'`, false)
  newIssue(LEX_UNTERMINATED_COMMENT, `unterminated /* */ comment`, false)
  newIssue(LEX_UNTERMINATED_STRING, `unterminated %s quoted string`, false)

  newIssue(PARSE_CLASS_NOT_VALID_HERE, `'class' keyword not allowed at this location`, false)
  newIssue(PARSE_ELSIF_IN_UNLESS, `elsif not supported in unless expression`, false)
  newIssue(PARSE_EXPECTED_ATTRIBUTE_NAME, `expected attribute name`, false)
  newIssue(PARSE_EXPECTED_CLASS_NAME, `expected name of class`, false)
  newIssue(PARSE_EXPECTED_FARROW_AFTER_KEY, `expected '=>' to follow hash key`, false)
  newIssue(PARSE_EXPECTED_HOSTNAME, `hostname expected`, false)
  newIssue(PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT, `expected name or number to follow '.'`, false)
  newIssue(PARSE_EXPECTED_NAME_AFTER_FUNCTION, `expected a name to follow keyword 'function'`, false)
  newIssue(PARSE_EXPECTED_TITLE, `resource title expected`, false)
  newIssue(PARSE_EXPECTED_TOKEN, `expected token '%s'`, false)
  newIssue(PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE, `expected type name to follow 'type'`, false)
  newIssue(PARSE_EXPECTED_VARIABLE, `expected variable declaration`, false)
  newIssue(PARSE_ILLEGAL_EPP_PARAMETERS, `Ambiguous EPP parameter expression. Probably missing '<%-' before parameters to remove leading whitespace`, false)
  newIssue(PARSE_INVALID_ATTRIBUTE, `invalid attribute operation`, false)
  newIssue(PARSE_INVALID_RESOURCE, `invalid resource expression`, false)
  newIssue(PARSE_INHERITS_MUST_BE_TYPE_NAME, `expected type name to follow 'inherits'`, false)
  newIssue(PARSE_RESOURCE_WITHOUT_TITLE, `This expression is invalid. Did you try declaring a '%s' resource without a title?`, false)
  newIssue(PARSE_QUOTED_NOT_VALID_NAME, `a quoted string is not valid as a name at this location`, false)

  newIssue(VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED, `The operator '%s' is no longer supported. See http://links.puppet.com/remove-plus-equals`, false)
  newIssue(VALIDATE_IDEM_EXPRESSION_NOT_LAST, `This %s has no effect. A value was produced and then forgotten (one or more preceding expressions may have the wrong form)`, false)
  newIssue(VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX, `Illegal attempt to assign via [index/key]. Not an assignable reference`, false)
  newIssue(VALIDATE_ILLEGAL_ATTRIBUTE_APPEND, `Illegal +> operation on attribute %s. This operator can not be used in %s`, false)
  newIssue(VALIDATE_ILLEGAL_EXPRESSION, `Illegal expression. %s is unacceptable as %s in %s`, false)
  newIssue(VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT, `Illegal attempt to assign to the numeric match result variable '$%s'. Numeric variables are not assignable`, false)
  newIssue(VALIDATE_CROSS_SCOPE_ASSIGNMENT, `Illegal attempt to assign to '%s'. Cannot assign to variables in other namespaces`, false)
  newIssue(VALIDATE_NOT_RVALUE, `Invalid use of expression. %s does not produce a value`, false)
}

type (
  Issue struct {
    code          string
    messageFormat string
    demotable     bool
  }

  ReportedIssue struct {
    issueCode string
    args      []interface{}
    location  Location
  }
)

var Issues = map[string]*Issue{
}

func newIssue(code string, messageFormat string, demotable bool) {
  Issues[code] = &Issue{code, messageFormat, demotable}
}

func NewReportedIssue(issueCode string, args []interface{}, location Location) *ReportedIssue {
  return &ReportedIssue{issueCode, args, location}
}

func (e *ReportedIssue) Error() (str string) {
  if issue, ok := Issues[e.issueCode]; ok {
    return appendLocation(fmt.Sprintf(issue.messageFormat, e.args...), e.location)
  }
  return fmt.Sprintf("internal error: no issue found for issue code '%s'", e.issueCode)
}

func (e *ReportedIssue) String() (str string) {
  return e.Error()
}

func (e *ReportedIssue) Code() string {
  return e.issueCode
}

func appendLocation(str string, location Location) string {
  if location == nil {
    return str
  }
  b := bytes.NewBufferString(str)
  line := location.Line()
  pos := location.Pos()
  if file := location.File(); file != `` {
    if line > 0 {
      b.WriteString(` at `)
      b.WriteString(file)
      b.WriteByte(':')
      fmt.Fprintf(b, `%d`, line)
      if pos > 0 {
        b.WriteByte(':')
        fmt.Fprintf(b, `%d`, pos)
      }
    } else {
      b.WriteString(` in `)
      b.WriteString(file)
    }
  } else if line > 0 {
    b.WriteString(` at line `)
    fmt.Fprintf(b, `%d`, line)
    if pos > 0 {
      b.WriteByte(':')
      fmt.Fprintf(b, `%d`, pos)
    }
  }
  return b.String()
}
