package validator

import . "github.com/puppetlabs/go-parser/issue"

const (
  VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED = `VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED`
  VALIDATE_CAPTURES_REST_NOT_LAST              = `VALIDATE_CAPTURES_REST_NOT_LAST`
  VALIDATE_CAPTURES_REST_NOT_SUPPORTED         = `VALIDATE_CAPTURES_REST_NOT_SUPPORTED`
  VALIDATE_CROSS_SCOPE_ASSIGNMENT              = `VALIDATE_CROSS_SCOPE_ASSIGNMENT`
  VALIDATE_DUPLICATE_DEFAULT                   = `VALIDATE_DUPLICATE_DEFAULT`
  VALIDATE_DUPLICATE_KEY                       = `VALIDATE_DUPLICATE_KEY`
  VALIDATE_DUPLICATE_PARAMETER                 = `VALIDATE_DUPLICATE_PARAMETER`
  VALIDATE_FUTURE_RESERVED_WORD                = `VALIDATE_FUTURE_RESERVED_WORD`
  VALIDATE_IDEM_EXPRESSION_NOT_LAST            = `VALIDATE_IDEM_EXPRESSION_NOT_LAST`
  VALIDATE_IDEM_NOT_ALLOWED_LAST               = `VALIDATE_IDEM_NOT_ALLOWED_LAST`
  VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX        = `VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX`
  VALIDATE_ILLEGAL_ATTRIBUTE_APPEND            = `VALIDATE_ILLEGAL_ATTRIBUTE_APPEND`
  VALIDATE_ILLEGAL_CLASSREF                    = `VALIDATE_ILLEGAL_CLASSREF`
  VALIDATE_ILLEGAL_DEFINITION_NAME             = `VALIDATE_ILLEGAL_DEFINITION_NAME`
  VALIDATE_ILLEGAL_EXPRESSION                  = `VALIDATE_ILLEGAL_EXPRESSION`
  VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT          = `VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT`
  VALIDATE_NOT_ABSOLUTE_TOP_LEVEL              = `VALIDATE_NOT_ABSOLUTE_TOP_LEVEL`
  VALIDATE_NOT_RVALUE                          = `VALIDATE_NOT_RVALUE`
  VALIDATE_NOT_TOP_LEVEL                       = `VALIDATE_NOT_TOP_LEVEL`
  VALIDATE_RESERVED_PARAMETER                  = `VALIDATE_RESERVED_PARAMETER`
  VALIDATE_RESERVED_TYPE_NAME                  = `VALIDATE_RESERVED_TYPE_NAME`
  VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT     = `VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT`
)

func init() {
  HardIssue(VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED, `The operator '%s' is no longer supported. See http://links.puppet.com/remove-plus-equals`)
  HardIssue(VALIDATE_CAPTURES_REST_NOT_LAST, `Parameter $%s is not last, and has 'captures rest'`)
  HardIssue(VALIDATE_CAPTURES_REST_NOT_SUPPORTED, `Parameter $%s has 'captures rest' - not supported in %s`)
  HardIssue(VALIDATE_CROSS_SCOPE_ASSIGNMENT, `Illegal attempt to assign to '%s'. Cannot assign to variables in other namespaces`)
  HardIssue(VALIDATE_DUPLICATE_DEFAULT, `This %s already has a 'default' entry - this is a duplicate`)
  SoftIssue(VALIDATE_DUPLICATE_KEY, `The key %s is declared more than once`)
  HardIssue(VALIDATE_DUPLICATE_PARAMETER, `The parameter '%s' is declared more than once in the parameter list`)
  SoftIssue(VALIDATE_FUTURE_RESERVED_WORD, `Use of future reserved word: '%s'`)
  SoftIssue(VALIDATE_IDEM_EXPRESSION_NOT_LAST, `This %s has no effect. A value was produced and then forgotten (one or more preceding expressions may have the wrong form)`)
  HardIssue(VALIDATE_IDEM_NOT_ALLOWED_LAST, `This %s has no effect. %s can not end with a value-producing expression without other effect`)
  HardIssue(VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX, `Illegal attempt to assign via [index/key]. Not an assignable reference`)
  HardIssue(VALIDATE_ILLEGAL_ATTRIBUTE_APPEND, `Illegal +> operation on attribute %s. This operator can not be used in %s`)
  HardIssue(VALIDATE_ILLEGAL_CLASSREF, `Illegal type reference. The given name '%s' does not conform to the naming rule`)
  HardIssue(VALIDATE_ILLEGAL_DEFINITION_NAME, `Unacceptable name. The name '%s' is unacceptable as the name of %s`)
  HardIssue(VALIDATE_ILLEGAL_EXPRESSION, `Illegal expression. %s is unacceptable as %s in %s`)
  HardIssue(VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT, `Illegal attempt to assign to the numeric match result variable '$%s'. Numeric variables are not assignable`)
  HardIssue(VALIDATE_NOT_ABSOLUTE_TOP_LEVEL, `%s may only appear at top level`)
  HardIssue(VALIDATE_NOT_TOP_LEVEL, `Classes, definitions, and nodes may only appear at top level or inside other classes`)
  HardIssue(VALIDATE_NOT_RVALUE, `Invalid use of expression. %s does not produce a value`)
  HardIssue(VALIDATE_RESERVED_PARAMETER, `The parameter $%s redefines a built in parameter in %s`)
  HardIssue(VALIDATE_RESERVED_TYPE_NAME, `The name: '%s' is already defined by Puppet and can not be used as the name of %s`)
  HardIssue(VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT, `The operator '%s' in %s is not supported`)
}
