package parser

import . "github.com/puppetlabs/go-parser/issue"

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
	LEX_MALFORMED_HEX_ESCAPE              = `LEX_MALFORMED_HEX_ESCAPE`
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
	PARSE_EXPECTED_ONE_OF_TOKENS            = `PARSE_EXPECTED_ONE_OF_TOKENS`
	PARSE_EXPECTED_TYPE_NAME                = `PARSE_EXPECTED_TYPE_NAME`
	PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE     = `PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE`
	PARSE_EXPECTED_VARIABLE                 = `PARSE_EXPECTED_VARIABLE`
	PARSE_ILLEGAL_EPP_PARAMETERS            = `PARSE_ILLEGAL_EPP_PARAMETERS`
	PARSE_INVALID_RESOURCE                  = `PARSE_INVALID_RESOURCE`
	PARSE_INVALID_ATTRIBUTE                 = `PARSE_INVALID_ATTRIBUTE`
	PARSE_INHERITS_MUST_BE_TYPE_NAME        = `PARSE_INHERITS_MUST_BE_TYPE_NAME`
	PARSE_RESOURCE_WITHOUT_TITLE            = `PARSE_RESOURCE_WITHOUT_TITLE`
	PARSE_QUOTED_NOT_VALID_NAME             = `PARSE_QUOTED_NOT_VALID_NAME`
)

func init() {
	HardIssue(LEX_DOUBLE_COLON_NOT_FOLLOWED_BY_NAME, `:: not followed by name segment`)
	HardIssue(LEX_DIGIT_EXPECTED, `digit expected`)
	HardIssue(LEX_HEREDOC_DECL_UNTERMINATED, `unterminated @(`)
	HardIssue(LEX_HEREDOC_EMPTY_TAG, `empty heredoc tag`)
	HardIssue(LEX_HEREDOC_ILLEGAL_ESCAPE, `illegal heredoc escape '%v'`)
	HardIssue(LEX_HEREDOC_MULTIPLE_ESCAPE, `more than one declaration of escape flags in heredoc`)
	HardIssue(LEX_HEREDOC_MULTIPLE_SYNTAX, `more than one syntax declaration in heredoc`)
	HardIssue(LEX_HEREDOC_MULTIPLE_TAG, `more than one tag declaration in heredoc`)
	HardIssue(LEX_HEREDOC_UNTERMINATED, `unterminated heredoc`)
	HardIssue(LEX_HEXDIGIT_EXPECTED, `hexadecimal digit expected`)
	HardIssue(LEX_INVALID_NAME, `invalid name`)
	HardIssue(LEX_INVALID_OPERATOR, `invalid operator '%v'`)
	HardIssue(LEX_INVALID_TYPE_NAME, `invalid type name`)
	HardIssue(LEX_INVALID_VARIABLE_NAME, `invalid variable name`)
	HardIssue(LEX_MALFORMED_HEX_ESCAPE, `malformed hexadecimal escape sequence`)
	HardIssue(LEX_MALFORMED_INTERPOLATION, `malformed interpolation expression`)
	HardIssue(LEX_MALFORMED_UNICODE_ESCAPE, `malformed unicode escape sequence`)
	HardIssue(LEX_OCTALDIGIT_EXPECTED, `octal digit expected`)
	HardIssue(LEX_UNBALANCED_EPP_COMMENT, `unbalanced epp comment`)
	HardIssue(LEX_UNEXPECTED_TOKEN, `unexpected token '%v'`)
	HardIssue(LEX_UNTERMINATED_COMMENT, `unterminated /* */ comment`)
	HardIssue(LEX_UNTERMINATED_STRING, `unterminated %v quoted string`)

	HardIssue(PARSE_CLASS_NOT_VALID_HERE, `'class' keyword not allowed at this location`)
	HardIssue(PARSE_ELSIF_IN_UNLESS, `elsif not supported in unless expression`)
	HardIssue(PARSE_EXPECTED_ATTRIBUTE_NAME, `expected attribute name`)
	HardIssue(PARSE_EXPECTED_CLASS_NAME, `expected name of class`)
	HardIssue(PARSE_EXPECTED_FARROW_AFTER_KEY, `expected '=>' to follow hash key`)
	HardIssue(PARSE_EXPECTED_HOSTNAME, `hostname expected`)
	HardIssue(PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT, `expected name or number to follow '.'`)
	HardIssue(PARSE_EXPECTED_NAME_AFTER_FUNCTION, `expected a name to follow keyword 'function'`)
	HardIssue(PARSE_EXPECTED_ONE_OF_TOKENS, `expected one of %v, got '%v'`)
	HardIssue(PARSE_EXPECTED_TITLE, `resource title expected`)
	HardIssue(PARSE_EXPECTED_TOKEN, `expected token '%v', got '%v'`)
	HardIssue(PARSE_EXPECTED_TYPE_NAME, `expected type name`)
	HardIssue(PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE, `expected type name to follow 'type'`)
	HardIssue(PARSE_EXPECTED_VARIABLE, `expected variable declaration`)
	HardIssue(PARSE_ILLEGAL_EPP_PARAMETERS, `Ambiguous EPP parameter expression. Probably missing '<%%-' before parameters to remove leading whitespace`)
	HardIssue(PARSE_INVALID_ATTRIBUTE, `invalid attribute operation`)
	HardIssue(PARSE_INVALID_RESOURCE, `invalid resource expression`)
	HardIssue(PARSE_INHERITS_MUST_BE_TYPE_NAME, `expected type name to follow 'inherits'`)
	HardIssue(PARSE_RESOURCE_WITHOUT_TITLE, `This expression is invalid. Did you try declaring a '%v' resource without a title?`)
	HardIssue(PARSE_QUOTED_NOT_VALID_NAME, `a quoted string is not valid as a name at this location`)
}
