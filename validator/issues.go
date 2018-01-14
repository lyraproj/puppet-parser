package validator

import (
	. "github.com/puppetlabs/go-parser/issue"
	. "github.com/puppetlabs/go-parser/parser"
)

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
	VALIDATE_ILLEGAL_ASSIGNMENT_CONTEXT          = `VALIDATE_ILLEGAL_ASSIGNMENT_CONTEXT`
	VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX        = `VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX`
	VALIDATE_ILLEGAL_ATTRIBUTE_APPEND            = `VALIDATE_ILLEGAL_ATTRIBUTE_APPEND`
	VALIDATE_ILLEGAL_CLASSREF                    = `VALIDATE_ILLEGAL_CLASSREF`
	VALIDATE_ILLEGAL_DEFINITION_NAME             = `VALIDATE_ILLEGAL_DEFINITION_NAME`
	VALIDATE_ILLEGAL_EXPRESSION                  = `VALIDATE_ILLEGAL_EXPRESSION`
	VALIDATE_ILLEGAL_HOSTNAME_CHARS              = `VALIDATE_ILLEGAL_HOSTNAME_CHARS`
	VALIDATE_ILLEGAL_HOSTNAME_INTERPOLATION      = `VALIDATE_ILLEGAL_HOSTNAME_INTERPOLATION`
	VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT          = `VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT`
	VALIDATE_ILLEGAL_NUMERIC_PARAMETER           = `VALIDATE_ILLEGAL_NUMERIC_PARAMETER`
	VALIDATE_ILLEGAL_PARAMETER_NAME              = `VALIDATE_ILLEGAL_PARAMETER_NAME`
	VALIDATE_ILLEGAL_QUERY_EXPRESSION            = `VALIDATE_ILLEGAL_QUERY_EXPRESSION`
	VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING         = `VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING`
	VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING         = `VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING`
	VALIDATE_MULTIPLE_ATTRIBUTES_UNFOLD          = `VALIDATE_MULTIPLE_ATTRIBUTES_UNFOLD`
	VALIDATE_NOT_ABSOLUTE_TOP_LEVEL              = `VALIDATE_NOT_ABSOLUTE_TOP_LEVEL`
	VALIDATE_NOT_RVALUE                          = `VALIDATE_NOT_RVALUE`
	VALIDATE_NOT_TOP_LEVEL                       = `VALIDATE_NOT_TOP_LEVEL`
	VALIDATE_NOT_VIRTUALIZABLE                   = `VALIDATE_NOT_VIRTUALIZABLE`
	VALIDATE_RESERVED_PARAMETER                  = `VALIDATE_RESERVED_PARAMETER`
	VALIDATE_RESERVED_TYPE_NAME                  = `VALIDATE_RESERVED_TYPE_NAME`
	VALIDATE_RESERVED_WORD                       = `VALIDATE_RESERVED_WORD`
	VALIDATE_UNSUPPORTED_EXPRESSION              = `VALIDATE_UNSUPPORTED_EXPRESSION`
	VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT     = `VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT`
)

func init() {
	HardIssue(VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED, `The operator '%{operator}' is no longer supported. See http://links.puppet.com/remove-plus-equals`)

	HardIssue(VALIDATE_CAPTURES_REST_NOT_LAST, `Parameter $%{param} is not last, and has 'captures rest'`)

	HardIssue2(VALIDATE_CAPTURES_REST_NOT_SUPPORTED,
		`Parameter $%{param} has 'captures rest' - not supported in %{container}`,
		HF{`container`: A_an})

	HardIssue(VALIDATE_CROSS_SCOPE_ASSIGNMENT, `Illegal attempt to assign to '%{name}'. Cannot assign to variables in other namespaces`)

	HardIssue2(VALIDATE_DUPLICATE_DEFAULT,
		`This %{container} already has a 'default' entry - this is a duplicate`,
		HF{`container`: Label})

	SoftIssue(VALIDATE_DUPLICATE_KEY, `The key '%{key}' is declared more than once`)

	HardIssue(VALIDATE_DUPLICATE_PARAMETER, `The parameter '%{param}' is declared more than once in the parameter list`)

	SoftIssue(VALIDATE_FUTURE_RESERVED_WORD, `Use of future reserved word: '%{word}'`)

	SoftIssue2(VALIDATE_IDEM_EXPRESSION_NOT_LAST,
		`This %{expression} has no effect. A value was produced and then forgotten (one or more preceding expressions may have the wrong form)`,
			HF{`expression`: Label})

	HardIssue2(VALIDATE_IDEM_NOT_ALLOWED_LAST,
		`This %{expression} has no effect. %{container} can not end with a value-producing expression without other effect`,
		HF{`expression`: Label, `container`: A_anUc })

	HardIssue(VALIDATE_ILLEGAL_ASSIGNMENT_CONTEXT, `Assignment not allowed here`)

	HardIssue(VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX, `Illegal attempt to assign via [index/key]. Not an assignable reference`)

	HardIssue2(VALIDATE_ILLEGAL_ATTRIBUTE_APPEND,
		`Illegal +> operation on attribute %{attr}. This operator can not be used in %{expression}`,
			HF{`expression`: A_an})

	HardIssue(VALIDATE_ILLEGAL_CLASSREF, `Illegal type reference. The given name '%{name}' does not conform to the naming rule`)

	HardIssue2(VALIDATE_ILLEGAL_DEFINITION_NAME,
		`Unacceptable name. The name '%{name}' is unacceptable as the name of %{value}`,
		HF{`value`: A_an})

	HardIssue2(
		VALIDATE_ILLEGAL_EXPRESSION,
		`Illegal expression. %{expression} is unacceptable as %{feature} in %{container}`,
		HF{`expression`: A_anUc, `container`: A_an })

	HardIssue(VALIDATE_ILLEGAL_HOSTNAME_CHARS, `The hostname '%{hostname}' contains illegal characters (only letters, digits, '_', '-', and '.' are allowed)`)

	HardIssue(VALIDATE_ILLEGAL_HOSTNAME_INTERPOLATION, `An interpolated expression is not allowed in a hostname of a node`)

	HardIssue(VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT, `Illegal attempt to assign to the numeric match result variable '$%{var}'. Numeric variables are not assignable`)

	HardIssue(VALIDATE_ILLEGAL_NUMERIC_PARAMETER, `The numeric parameter name '$%{name}' cannot be used (clashes with numeric match result variables)`)

	HardIssue(VALIDATE_ILLEGAL_PARAMETER_NAME, `Illegal parameter name. The given name '%{name}' does not conform to the naming rule /^[a-z_]\w*$/`)

	HardIssue2(VALIDATE_ILLEGAL_QUERY_EXPRESSION,
		`Illegal query expression. %{expression} cannot be used in a query`,
		HF{`expression`: A_anUc})

	HardIssue2(VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING,
		`Illegal type mapping. Expected a Tuple[Regexp,String] on the left side, got %{expression}`,
		HF{`expression`: A_an})

	HardIssue2(VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING,
		`Illegal type mapping. Expected a Type on the left side, got %{expression}`,
		HF{`expression`: A_an})

	HardIssue(VALIDATE_MULTIPLE_ATTRIBUTES_UNFOLD, `Unfolding of attributes from Hash can only be used once per resource body`)

	HardIssue2(VALIDATE_NOT_ABSOLUTE_TOP_LEVEL,
		`%{value} may only appear at top level`,
		HF{`value`: A_anUc})

	HardIssue(VALIDATE_NOT_TOP_LEVEL,`Classes, definitions, and nodes may only appear at top level or inside other classes`)

	HardIssue2(VALIDATE_NOT_RVALUE,
		`Invalid use of expression. %{value} does not produce a value`,
		HF{`value`: A_anUc})

	HardIssue(VALIDATE_NOT_VIRTUALIZABLE, `Resource Defaults/Overrides are not virtualizable`)

	HardIssue2(VALIDATE_RESERVED_PARAMETER,
		`The parameter $%{param} redefines a built in parameter in %{container}`,
		HF{`container`: A_an})

	HardIssue2(VALIDATE_RESERVED_TYPE_NAME,
		`The name: '%{name}' is already defined by Puppet and can not be used as the name of %{expression}`,
		HF{`expression`: A_an})

	HardIssue(VALIDATE_RESERVED_WORD, `Use of reserved word: %{word}, must be quoted if intended to be a String value`)

	HardIssue2(VALIDATE_UNSUPPORTED_EXPRESSION,
		`Expressions of type %{expression} are not supported in this version of Puppet`,
		HF{`expression`: A_an})

	HardIssue2(VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT,
		`The operator '%{operator}' in %{value} is not supported`,
		HF{`value`: A_an})
}
