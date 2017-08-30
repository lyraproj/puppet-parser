package pspec

import . `github.com/puppetlabs/go-parser/issue`

const (
  SPEC_EXPRESSION_NOT_PARAMETER_TO = `SPEC_EXPRESSION_NOT_PARAMETER_TO`
  SPEC_ILLEGAL_ARGUMENT_TYPE = `SPEC_ILLEGAL_ARGUMENT_TYPE`
  SPEC_MISSING_ARGUMENT = `SPEC_MISSING_ARGUMENT`
  SPEC_NOT_TOP_EXPRESSION = `SPEC_NOT_TOP_EXPRESSION`
  SPEC_ILLEGAL_CALL_RECEIVER = `SPEC_ILLEGAL_CALL_RECEIVER`
  SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS = `SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS`
  SPEC_UNKNOWN_IDENTIFIER = `SPEC_UNKNOWN_IDENTIFIER`
)

func init() {
  HardIssue(SPEC_EXPRESSION_NOT_PARAMETER_TO, `%v can only be a parameter to %v or assigned to a variable`)
  HardIssue(SPEC_ILLEGAL_ARGUMENT_TYPE, `Illegal argument type. Function %v, parameter %d expected %v, got %v`)
  HardIssue(SPEC_MISSING_ARGUMENT, `Missing required value. Function %v, parameter %d requires a value of type %v`)
  HardIssue(SPEC_NOT_TOP_EXPRESSION, `%v is only legal at top level`)
  HardIssue(SPEC_ILLEGAL_CALL_RECEIVER, `Illegal call receiver`)
  HardIssue(SPEC_ILLEGAL_NUMBER_OF_ARGUMENTS, `Illegal number of arguments. Function %v expects %v, got %d`)
  HardIssue(SPEC_UNKNOWN_IDENTIFIER, `unknown identifier %v`)
}
