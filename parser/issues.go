package parser

import "github.com/lyraproj/issue/issue"

const (
	// Lexer issues
	lexDoubleColonNotFollowedByName = `LEX_DOUBLE_COLON_NOT_FOLLOWED_BY_NAME`
	lexDigitExpected                = `LEX_DIGIT_EXPECTED`
	lexHeredocEmptyTag              = `LEX_HEREDOC_EMPTY_TAG`
	lexHeredocIllegalEscape         = `LEX_HEREDOC_ILLEGAL_ESCAPE`
	lexHeredocMultipleEscape        = `LEX_HEREDOC_MULTIPLE_ESCAPE`
	lexHeredocMultipleSyntax        = `LEX_HEREDOC_MULTIPLE_SYNTAX`
	lexHeredocMultipleTag           = `LEX_HEREDOC_MULTIPLE_TAG`
	lexHeredocDeclUnterminated      = `LEX_HEREDOC_DECL_UNTERMINATED`
	lexHeredocUnterminated          = `LEX_HEREDOC_UNTERMINATED`
	lexHexdigitExpected             = `LEX_HEXDIGIT_EXPECTED`
	lexInvalidName                  = `LEX_INVALID_NAME`
	lexInvalidOperator              = `LEX_INVALID_OPERATOR`
	lexInvalidTypeName              = `LEX_INVALID_TYPE_NAME`
	lexInvalidVariableName          = `LEX_INVALID_VARIABLE_NAME`
	lexMalformedHexEscape           = `LEX_MALFORMED_HEX_ESCAPE`
	lexMalformedInterpolation       = `LEX_MALFORMED_INTERPOLATION`
	lexMalformedUnicodeEscape       = `LEX_MALFORMED_UNICODE_ESCAPE`
	lexOctaldigitExpected           = `LEX_OCTALDIGIT_EXPECTED`
	lexUnbalancedEppComment         = `LEX_UNBALANCED_EPP_COMMENT`
	lexUnexpectedToken              = `LEX_UNEXPECTED_TOKEN`
	lexUnterminatedComment          = `LEX_UNTERMINATED_COMMENT`
	lexUnterminatedString           = `LEX_UNTERMINATED_STRING`

	parseClassNotValidHere            = `PARSE_CLASS_NOT_VALID_HERE`
	parseElsifInUnless                = `PARSE_ELSIF_IN_UNLESS`
	parseExpectedStepName             = `PARSE_EXPECTED_STEP_NAME`
	parseExpectedStepStyle            = `PARSE_EXPECTED_STEP_STYLE`
	parseExpectedAttributeName        = `PARSE_EXPECTED_ATTRIBUTE_NAME`
	parseExpectedStepOperation        = `PARSE_EXPECTED_STEP_OPERATION`
	parseExpectedIteratorStyle        = `PARSE_EXPECTED_ITERATOR_STYLE`
	parseExpectedClassName            = `PARSE_EXPECTED_CLASS_NAME`
	parseExpectedFarrowAfterKey       = `PARSE_EXPECTED_FARROW_AFTER_KEY`
	parseExpectedNameOrNumberAfterDot = `PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT`
	parseExpectedNameAfterFunction    = `PARSE_EXPECTED_NAME_AFTER_FUNCTION`
	parseExpectedNameAfterPlan        = `PARSE_EXPECTED_NAME_AFTER_PLAN`
	parseExpectedHostname             = `PARSE_EXPECTED_HOSTNAME`
	parseExpectedTitle                = `PARSE_EXPECTED_TITLE`
	parseExpectedToken                = `PARSE_EXPECTED_TOKEN`
	parseExpectedOneOfTokens          = `PARSE_EXPECTED_ONE_OF_TOKENS`
	parseExpectedTypeName             = `PARSE_EXPECTED_TYPE_NAME`
	parseExpectedTypeNameAfterType    = `PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE`
	parseExpectedVariable             = `PARSE_EXPECTED_VARIABLE`
	parseExtraneousComma              = `PARSE_EXTRANEOUS_COMMA`
	parseIllegalEppParameters         = `PARSE_ILLEGAL_EPP_PARAMETERS`
	parseInvalidStepAttribute         = `PARSE_INVALID_STEP_ATTRIBUTE`
	parseInvalidAttribute             = `PARSE_INVALID_ATTRIBUTE`
	parseInvalidResource              = `PARSE_INVALID_RESOURCE`
	parseInheritsMustBeTypeName       = `PARSE_INHERITS_MUST_BE_TYPE_NAME`
	parseResourceWithoutTitle         = `PARSE_RESOURCE_WITHOUT_TITLE`
	parseQuotedNotValidName           = `PARSE_QUOTED_NOT_VALID_NAME`
)

func init() {
	issue.Hard(lexDoubleColonNotFollowedByName, `:: not followed by name segment`)
	issue.Hard(lexDigitExpected, `digit expected`)
	issue.Hard(lexHeredocDeclUnterminated, `unterminated @(`)
	issue.Hard(lexHeredocEmptyTag, `empty heredoc tag`)
	issue.Hard(lexHeredocIllegalEscape, `illegal heredoc escape '%{flag}'`)
	issue.Hard(lexHeredocMultipleEscape, `more than one declaration of escape flags in heredoc`)
	issue.Hard(lexHeredocMultipleSyntax, `more than one syntax declaration in heredoc`)
	issue.Hard(lexHeredocMultipleTag, `more than one tag declaration in heredoc`)
	issue.Hard(lexHeredocUnterminated, `unterminated heredoc`)
	issue.Hard(lexHexdigitExpected, `hexadecimal digit expected`)
	issue.Hard(lexInvalidName, `invalid name`)
	issue.Hard(lexInvalidOperator, `invalid operator '%{op}'`)
	issue.Hard(lexInvalidTypeName, `invalid type name`)
	issue.Hard(lexInvalidVariableName, `invalid variable name`)
	issue.Hard(lexMalformedHexEscape, `malformed hexadecimal escape sequence`)
	issue.Hard(lexMalformedInterpolation, `malformed interpolation expression`)
	issue.Hard(lexMalformedUnicodeEscape, `malformed unicode escape sequence`)
	issue.Hard(lexOctaldigitExpected, `octal digit expected`)
	issue.Hard(lexUnbalancedEppComment, `unbalanced epp comment`)
	issue.Hard(lexUnexpectedToken, `unexpected token '%{token}'`)
	issue.Hard(lexUnterminatedComment, `unterminated /* */ comment`)
	issue.Hard(lexUnterminatedString, `unterminated %{string_type} quoted string`)

	issue.Hard(parseClassNotValidHere, `'class' keyword not allowed at this location`)
	issue.Hard(parseElsifInUnless, `elsif not supported in unless expression`)
	issue.Hard(parseExpectedStepName, `expected %{step} name`)
	issue.Hard(parseExpectedStepOperation, `expected one of 'delete', 'read', or 'upsert'. Got '%{operation}'`)
	issue.Hard(parseExpectedIteratorStyle, `expected one of 'each', 'range', or 'times'. Got '%{style}`)
	issue.Hard(parseExpectedStepStyle, `expected one of 'action', 'resource', or 'workflow'`)
	issue.Hard(parseExpectedAttributeName, `expected attribute name`)
	issue.Hard(parseExpectedClassName, `expected name of class`)
	issue.Hard(parseExpectedFarrowAfterKey, `expected '=>' to follow hash key`)
	issue.Hard(parseExpectedHostname, `hostname expected`)
	issue.Hard(parseExpectedNameOrNumberAfterDot, `expected name or number to follow '.'`)
	issue.Hard(parseExpectedNameAfterFunction, `expected a name to follow keyword 'function'`)
	issue.Hard(parseExpectedNameAfterPlan, `expected a name to follow keyword 'plan'`)
	issue.Hard(parseExpectedOneOfTokens, `expected one of %{expected}, got '%{actual}'`)
	issue.Hard(parseExpectedTitle, `resource title expected`)
	issue.Hard(parseExpectedToken, `expected token '%{expected}', got '%{actual}'`)
	issue.Hard(parseExpectedTypeName, `expected type name`)
	issue.Hard(parseExpectedTypeNameAfterType, `expected type name to follow 'type'`)
	issue.Hard(parseExpectedVariable, `expected variable declaration`)
	issue.Hard(parseExtraneousComma, `Extraneous comma between statements`)
	issue.Hard(parseIllegalEppParameters, `Ambiguous EPP parameter expression. Probably missing '<%%-' before parameters to remove leading whitespace`)
	issue.Hard(parseInvalidStepAttribute, `Attribute '%{name}' is not valid in a '%{style}' definition`)
	issue.Hard(parseInvalidAttribute, `invalid attribute operation`)
	issue.Hard(parseInvalidResource, `invalid resource expression`)
	issue.Hard(parseInheritsMustBeTypeName, `expected type name to follow 'inherits'`)
	issue.Hard(parseResourceWithoutTitle, `This expression is invalid. Did you try declaring a '%{name}' resource without a title?`)
	issue.Hard(parseQuotedNotValidName, `a quoted string is not valid as a name at this location`)
}
