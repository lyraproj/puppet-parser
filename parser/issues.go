package parser

import "github.com/puppetlabs/go-issues/issue"

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
	PARSE_EXPECTED_ACTIVITY_NAME            = `PARSE_EXPECTED_ACTIVITY_NAME`
	PARSE_EXPECTED_ACTIVITY_STYLE           = `PARSE_EXPECTED_ACTIVITY_STYLE`
	PARSE_EXPECTED_ATTRIBUTE_NAME           = `PARSE_EXPECTED_ATTRIBUTE_NAME`
	PARSE_EXPECTED_ACTIVITY_OPERATION       = `PARSE_EXPECTED_ACTIVITY_OPERATION`
	PARSE_EXPECTED_ITERATOR_STYLE           = `PARSE_EXPECTED_ITERATOR_STYLE`
	PARSE_EXPECTED_CLASS_NAME               = `PARSE_EXPECTED_CLASS_NAME`
	PARSE_EXPECTED_FARROW_AFTER_KEY         = `PARSE_EXPECTED_FARROW_AFTER_KEY`
	PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT = `PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT`
	PARSE_EXPECTED_NAME_AFTER_FUNCTION      = `PARSE_EXPECTED_NAME_AFTER_FUNCTION`
	PARSE_EXPECTED_NAME_AFTER_PLAN          = `PARSE_EXPECTED_NAME_AFTER_PLAN`
	PARSE_EXPECTED_HOSTNAME                 = `PARSE_EXPECTED_HOSTNAME`
	PARSE_EXPECTED_TITLE                    = `PARSE_EXPECTED_TITLE`
	PARSE_EXPECTED_TOKEN                    = `PARSE_EXPECTED_TOKEN`
	PARSE_EXPECTED_ONE_OF_TOKENS            = `PARSE_EXPECTED_ONE_OF_TOKENS`
	PARSE_EXPECTED_TYPE_NAME                = `PARSE_EXPECTED_TYPE_NAME`
	PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE     = `PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE`
	PARSE_EXPECTED_VARIABLE                 = `PARSE_EXPECTED_VARIABLE`
	PARSE_EXTRANEOUS_COMMA                  = `PARSE_EXTRANEOUS_COMMA`
	PARSE_ILLEGAL_EPP_PARAMETERS            = `PARSE_ILLEGAL_EPP_PARAMETERS`
	PARSE_INVALID_ACTIVITY_ATTRIBUTE        = `PARSE_INVALID_ACTIVITY_ATTRIBUTE`
	PARSE_INVALID_ATTRIBUTE                 = `PARSE_INVALID_ATTRIBUTE`
	PARSE_INVALID_RESOURCE                  = `PARSE_INVALID_RESOURCE`
	PARSE_INHERITS_MUST_BE_TYPE_NAME        = `PARSE_INHERITS_MUST_BE_TYPE_NAME`
	PARSE_RESOURCE_WITHOUT_TITLE            = `PARSE_RESOURCE_WITHOUT_TITLE`
	PARSE_QUOTED_NOT_VALID_NAME             = `PARSE_QUOTED_NOT_VALID_NAME`
)

func init() {
	issue.Hard(LEX_DOUBLE_COLON_NOT_FOLLOWED_BY_NAME, `:: not followed by name segment`)
	issue.Hard(LEX_DIGIT_EXPECTED, `digit expected`)
	issue.Hard(LEX_HEREDOC_DECL_UNTERMINATED, `unterminated @(`)
	issue.Hard(LEX_HEREDOC_EMPTY_TAG, `empty heredoc tag`)
	issue.Hard(LEX_HEREDOC_ILLEGAL_ESCAPE, `illegal heredoc escape '%{flag}'`)
	issue.Hard(LEX_HEREDOC_MULTIPLE_ESCAPE, `more than one declaration of escape flags in heredoc`)
	issue.Hard(LEX_HEREDOC_MULTIPLE_SYNTAX, `more than one syntax declaration in heredoc`)
	issue.Hard(LEX_HEREDOC_MULTIPLE_TAG, `more than one tag declaration in heredoc`)
	issue.Hard(LEX_HEREDOC_UNTERMINATED, `unterminated heredoc`)
	issue.Hard(LEX_HEXDIGIT_EXPECTED, `hexadecimal digit expected`)
	issue.Hard(LEX_INVALID_NAME, `invalid name`)
	issue.Hard(LEX_INVALID_OPERATOR, `invalid operator '%{op}'`)
	issue.Hard(LEX_INVALID_TYPE_NAME, `invalid type name`)
	issue.Hard(LEX_INVALID_VARIABLE_NAME, `invalid variable name`)
	issue.Hard(LEX_MALFORMED_HEX_ESCAPE, `malformed hexadecimal escape sequence`)
	issue.Hard(LEX_MALFORMED_INTERPOLATION, `malformed interpolation expression`)
	issue.Hard(LEX_MALFORMED_UNICODE_ESCAPE, `malformed unicode escape sequence`)
	issue.Hard(LEX_OCTALDIGIT_EXPECTED, `octal digit expected`)
	issue.Hard(LEX_UNBALANCED_EPP_COMMENT, `unbalanced epp comment`)
	issue.Hard(LEX_UNEXPECTED_TOKEN, `unexpected token '%{token}'`)
	issue.Hard(LEX_UNTERMINATED_COMMENT, `unterminated /* */ comment`)
	issue.Hard(LEX_UNTERMINATED_STRING, `unterminated %{string_type} quoted string`)

	issue.Hard(PARSE_CLASS_NOT_VALID_HERE, `'class' keyword not allowed at this location`)
	issue.Hard(PARSE_ELSIF_IN_UNLESS, `elsif not supported in unless expression`)
	issue.Hard(PARSE_EXPECTED_ACTIVITY_NAME, `expected %{activity} name`)
	issue.Hard(PARSE_EXPECTED_ACTIVITY_OPERATION, `expected one of 'delete', 'read', or 'upsert'. Got '%{operation}'`)
	issue.Hard(PARSE_EXPECTED_ITERATOR_STYLE, `expected one of 'each', 'range', or 'times'. Got '%{style}`)
	issue.Hard(PARSE_EXPECTED_ACTIVITY_STYLE, `expected one of 'action', 'resource', or 'workflow'`)
	issue.Hard(PARSE_EXPECTED_ATTRIBUTE_NAME, `expected attribute name`)
	issue.Hard(PARSE_EXPECTED_CLASS_NAME, `expected name of class`)
	issue.Hard(PARSE_EXPECTED_FARROW_AFTER_KEY, `expected '=>' to follow hash key`)
	issue.Hard(PARSE_EXPECTED_HOSTNAME, `hostname expected`)
	issue.Hard(PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT, `expected name or number to follow '.'`)
	issue.Hard(PARSE_EXPECTED_NAME_AFTER_FUNCTION, `expected a name to follow keyword 'function'`)
	issue.Hard(PARSE_EXPECTED_NAME_AFTER_PLAN, `expected a name to follow keyword 'plan'`)
	issue.Hard(PARSE_EXPECTED_ONE_OF_TOKENS, `expected one of %{expected}, got '%{actual}'`)
	issue.Hard(PARSE_EXPECTED_TITLE, `resource title expected`)
	issue.Hard(PARSE_EXPECTED_TOKEN, `expected token '%{expected}', got '%{actual}'`)
	issue.Hard(PARSE_EXPECTED_TYPE_NAME, `expected type name`)
	issue.Hard(PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE, `expected type name to follow 'type'`)
	issue.Hard(PARSE_EXPECTED_VARIABLE, `expected variable declaration`)
	issue.Hard(PARSE_EXTRANEOUS_COMMA, `Extraneous comma between statements`)
	issue.Hard(PARSE_ILLEGAL_EPP_PARAMETERS, `Ambiguous EPP parameter expression. Probably missing '<%%-' before parameters to remove leading whitespace`)
	issue.Hard(PARSE_INVALID_ACTIVITY_ATTRIBUTE, `Attribute '%{name}' is not valid in a '%{style}' definition`)
	issue.Hard(PARSE_INVALID_ATTRIBUTE, `invalid attribute operation`)
	issue.Hard(PARSE_INVALID_RESOURCE, `invalid resource expression`)
	issue.Hard(PARSE_INHERITS_MUST_BE_TYPE_NAME, `expected type name to follow 'inherits'`)
	issue.Hard(PARSE_RESOURCE_WITHOUT_TITLE, `This expression is invalid. Did you try declaring a '%{name}' resource without a title?`)
	issue.Hard(PARSE_QUOTED_NOT_VALID_NAME, `a quoted string is not valid as a name at this location`)
}
