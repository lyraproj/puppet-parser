package validator

import (
	"github.com/lyraproj/issue/issue"
)

const (
	ValidateAppendsDeletesNoLongerSupported = `VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED`
	ValidateCapturesRestNotLast             = `VALIDATE_CAPTURES_REST_NOT_LAST`
	ValidateCapturesRestNotSupported        = `VALIDATE_CAPTURES_REST_NOT_SUPPORTED`
	ValidateCatalogOperationNotSupported    = `VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED`
	ValidateCrossScopeAssignment            = `VALIDATE_CROSS_SCOPE_ASSIGNMENT`
	ValidateDuplicateDefault                = `VALIDATE_DUPLICATE_DEFAULT`
	ValidateDuplicateKey                    = `VALIDATE_DUPLICATE_KEY`
	ValidateDuplicateParameter              = `VALIDATE_DUPLICATE_PARAMETER`
	ValidateFutureReservedWord              = `VALIDATE_FUTURE_RESERVED_WORD`
	ValidateIdemExpressionNotLast           = `VALIDATE_IDEM_EXPRESSION_NOT_LAST`
	ValidateIdemNotAllowedLast              = `VALIDATE_IDEM_NOT_ALLOWED_LAST`
	ValidateIllegalAssignmentContext        = `VALIDATE_ILLEGAL_ASSIGNMENT_CONTEXT`
	ValidateIllegalAssignmentViaIndex       = `VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX`
	ValidateIllegalAttributeAppend          = `VALIDATE_ILLEGAL_ATTRIBUTE_APPEND`
	ValidateIllegalClassref                 = `VALIDATE_ILLEGAL_CLASSREF`
	ValidateIllegalDefinitionName           = `VALIDATE_ILLEGAL_DEFINITION_NAME`
	ValidateIllegalExpression               = `VALIDATE_ILLEGAL_EXPRESSION`
	ValidateIllegalHostnameChars            = `VALIDATE_ILLEGAL_HOSTNAME_CHARS`
	ValidateIllegalHostnameInterpolation    = `VALIDATE_ILLEGAL_HOSTNAME_INTERPOLATION`
	ValidateIllegalNumericAssignment        = `VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT`
	ValidateIllegalNumericParameter         = `VALIDATE_ILLEGAL_NUMERIC_PARAMETER`
	ValidateIllegalParameterName            = `VALIDATE_ILLEGAL_PARAMETER_NAME`
	ValidateIllegalQueryExpression          = `VALIDATE_ILLEGAL_QUERY_EXPRESSION`
	ValidateIllegalRegexpTypeMapping        = `VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING`
	ValidateIllegalSingleTypeMapping        = `VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING`
	ValidateInvalidStepStyle                = `VALIDATE_INVALID_STEP_STYLE`
	ValidateMultipleAttributesUnfold        = `VALIDATE_MULTIPLE_ATTRIBUTES_UNFOLD`
	ValidateNotAbsoluteTopLevel             = `VALIDATE_NOT_ABSOLUTE_TOP_LEVEL`
	ValidateNotRvalue                       = `VALIDATE_NOT_RVALUE`
	ValidateNotTopLevel                     = `VALIDATE_NOT_TOP_LEVEL`
	ValidateNotVirtualizable                = `VALIDATE_NOT_VIRTUALIZABLE`
	ValidateReservedParameter               = `VALIDATE_RESERVED_PARAMETER`
	ValidateReservedTypeName                = `VALIDATE_RESERVED_TYPE_NAME`
	ValidateReservedWord                    = `VALIDATE_RESERVED_WORD`
	ValidateUnsupportedExpression           = `VALIDATE_UNSUPPORTED_EXPRESSION`
	ValidateUnsupportedOperatorInContext    = `VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT`
	ValidateWorkflowOperationNotSupported   = `VALIDATE_WORKFLOW_OPERATION_NOT_SUPPORTED`
)

func init() {
	issue.Hard(ValidateAppendsDeletesNoLongerSupported, `The operator '%{operator}' is no longer supported. See http://links.puppet.com/remove-plus-equals`)

	issue.Hard(ValidateCapturesRestNotLast, `Parameter $%{param} is not last, and has 'captures rest'`)

	issue.Hard2(ValidateCapturesRestNotSupported,
		`Parameter $%{param} has 'captures rest' - not supported in %{container}`,
		issue.HF{`container`: issue.AnOrA})

	issue.Hard(ValidateCatalogOperationNotSupported, `The catalog operation '%{operation}' is only available when compiling a catalog`)

	issue.Hard(ValidateCrossScopeAssignment, `Illegal attempt to assign to '%{name}'. Cannot assign to variables in other namespaces`)

	issue.Hard2(ValidateDuplicateDefault,
		`This %{container} already has a 'default' entry - this is a duplicate`,
		issue.HF{`container`: issue.Label})

	issue.Soft(ValidateDuplicateKey, `The key '%{key}' is declared more than once`)

	issue.Hard(ValidateDuplicateParameter, `The parameter '%{param}' is declared more than once in the parameter list`)

	issue.Soft(ValidateFutureReservedWord, `Use of future reserved word: '%{word}'`)

	issue.Soft2(ValidateIdemExpressionNotLast,
		`This %{expression} has no effect. A value was produced and then forgotten (one or more preceding expressions may have the wrong form)`,
		issue.HF{`expression`: issue.Label})

	issue.Hard2(ValidateIdemNotAllowedLast,
		`This %{expression} has no effect. %{container} can not end with a value-producing expression without other effect`,
		issue.HF{`expression`: issue.Label, `container`: issue.UcAnOrA})

	issue.Hard(ValidateIllegalAssignmentContext, `Assignment not allowed here`)

	issue.Hard(ValidateIllegalAssignmentViaIndex, `Illegal attempt to assign via [index/key]. Not an assignable reference`)

	issue.Hard2(ValidateIllegalAttributeAppend,
		`Illegal +> operation on attribute %{attr}. This operator can not be used in %{expression}`,
		issue.HF{`expression`: issue.AnOrA})

	issue.Hard(ValidateIllegalClassref, `Illegal type reference. The given name '%{name}' does not conform to the naming rule`)

	issue.Hard2(ValidateIllegalDefinitionName,
		`Unacceptable name. The name '%{name}' is unacceptable as the name of %{value}`,
		issue.HF{`value`: issue.AnOrA})

	issue.Hard2(
		ValidateIllegalExpression,
		`Illegal expression. %{expression} is unacceptable as %{feature} in %{container}`,
		issue.HF{`expression`: issue.UcAnOrA, `container`: issue.AnOrA})

	issue.Hard(ValidateIllegalHostnameChars, `The hostname '%{hostname}' contains illegal characters (only letters, digits, '_', '-', and '.' are allowed)`)

	issue.Hard(ValidateIllegalHostnameInterpolation, `An interpolated expression is not allowed in a hostname of a node`)

	issue.Hard(ValidateIllegalNumericAssignment, `Illegal attempt to assign to the numeric match result variable '$%{var}'. Numeric variables are not assignable`)

	issue.Hard(ValidateIllegalNumericParameter, `The numeric parameter name '$%{name}' cannot be used (clashes with numeric match result variables)`)

	issue.Hard(ValidateIllegalParameterName, `Illegal parameter name. The given name '%{name}' does not conform to the naming rule /^[a-z_]\w*$/`)

	issue.Hard2(ValidateIllegalQueryExpression,
		`Illegal query expression. %{expression} cannot be used in a query`,
		issue.HF{`expression`: issue.UcAnOrA})

	issue.Hard2(ValidateIllegalRegexpTypeMapping,
		`Illegal type mapping. Expected a Tuple[Regexp,String] on the left side, got %{expression}`,
		issue.HF{`expression`: issue.AnOrA})

	issue.Hard2(ValidateIllegalSingleTypeMapping,
		`Illegal type mapping. Expected a Type on the left side, got %{expression}`,
		issue.HF{`expression`: issue.AnOrA})

	issue.Hard(ValidateInvalidStepStyle, `Expected one of 'for', 'function', 'guard', 'resource', or 'workflow'. Got '%{style}'`)

	issue.Hard(ValidateMultipleAttributesUnfold, `Unfolding of attributes from Hash can only be used once per resource body`)

	issue.Hard2(ValidateNotAbsoluteTopLevel,
		`%{value} may only appear at top level`,
		issue.HF{`value`: issue.UcAnOrA})

	issue.Hard(ValidateNotTopLevel, `Classes, definitions, and nodes may only appear at top level or inside other classes`)

	issue.Hard2(ValidateNotRvalue,
		`Invalid use of expression. %{value} does not produce a value`,
		issue.HF{`value`: issue.UcAnOrA})

	issue.Hard(ValidateNotVirtualizable, `Resource Defaults/Overrides are not virtualizable`)

	issue.Hard2(ValidateReservedParameter,
		`The parameter $%{param} redefines a built in parameter in %{container}`,
		issue.HF{`container`: issue.AnOrA})

	issue.Hard2(ValidateReservedTypeName,
		`The name: '%{name}' is already defined by Puppet and can not be used as the name of %{expression}`,
		issue.HF{`expression`: issue.AnOrA})

	issue.Hard(ValidateReservedWord, `Use of reserved word: %{word}, must be quoted if intended to be a String value`)

	issue.Hard2(ValidateUnsupportedExpression,
		`Expressions of type %{expression} are not supported in this version of Puppet`,
		issue.HF{`expression`: issue.AnOrA})

	issue.Hard2(ValidateUnsupportedOperatorInContext,
		`The operator '%{operator}' in %{value} is not supported`,
		issue.HF{`value`: issue.AnOrA})

	issue.Hard(ValidateWorkflowOperationNotSupported, `The workflow operation '%{operation}' is only available when compiling workflows`)
}
