package validator

import (
	"regexp"
	"strings"

	"github.com/puppetlabs/go-parser/issue"
	"github.com/puppetlabs/go-parser/literal"
	"github.com/puppetlabs/go-parser/parser"
)

var DOUBLE_COLON_EXPR = regexp.MustCompile(`::`)

// CLASSREF_EXT matches a class reference the same way as the lexer - i.e. the external source form
// where each part must start with a capital letter A-Z.
var CLASSREF_EXT = regexp.MustCompile(`\A(?:::)?[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)

// Same as CLASSREF_EXT but cannot start with '::'
var CLASSREF_EXT_DECL = regexp.MustCompile(`\A[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)

// CLASSREF matches a class reference the way it is represented internally in the
// model (i.e. in lower case).
var CLASSREF_DECL = regexp.MustCompile(`\A[a-z][\w]*(?:::[a-z][\w]*)*\z`)

// ILLEGAL_P3_1_HOSTNAME matches if a hostname contains illegal characters.
// This check does not prevent pathological names like 'a....b', '.....', "---". etc.
var ILLEGAL_HOSTNAME_CHARS = regexp.MustCompile(`[^-\w.]`)

// PARAM_NAME matches the name part of a parameter (The $ character is not included)
var PARAM_NAME = regexp.MustCompile(`\A[a-z_]\w*\z`)

var STARTS_WITH_NUMBER = regexp.MustCompile(`\A[0-9]`)

var RESERVED_TYPE_NAMES = map[string]bool{
	`type`:       true,
	`any`:        true,
	`unit`:       true,
	`scalar`:     true,
	`boolean`:    true,
	`numeric`:    true,
	`integer`:    true,
	`float`:      true,
	`collection`: true,
	`array`:      true,
	`hash`:       true,
	`tuple`:      true,
	`struct`:     true,
	`variant`:    true,
	`optional`:   true,
	`enum`:       true,
	`regexp`:     true,
	`pattern`:    true,
	`runtime`:    true,

	`init`:        true,
	`object`:      true,
	`sensitive`:   true,
	`semver`:      true,
	`semverrange`: true,
	`string`:      true,
	`timestamp`:   true,
	`timespan`:    true,
	`typeset`:     true,
}

var FUTURE_RESERVED_WORDS = map[string]bool{
	`application`: true,
	`produces`:    true,
	`consumes`:    true,
}

var RESERVED_PARAMETERS = map[string]bool{
	`name`:  true,
	`title`: true,
}

type basicChecker struct {
	AbstractValidator
}

type Checker interface {
	Validator

	check_Application(e *parser.Application)
	check_AssignmentExpression(e *parser.AssignmentExpression)
	check_AttributeOperation(e *parser.AttributeOperation)
	check_AttributesOperation(e *parser.AttributesOperation)
	check_BinaryExpression(e parser.BinaryExpression)
	check_BlockExpression(e *parser.BlockExpression)
	check_CallNamedFunctionExpression(e *parser.CallNamedFunctionExpression)
	check_CapabilityMapping(e *parser.CapabilityMapping)
	check_CaseExpression(e *parser.CaseExpression)
	check_CaseOption(e *parser.CaseOption)
	check_CollectExpression(e *parser.CollectExpression)
	check_EppExpression(e *parser.EppExpression)
	check_FunctionDefinition(e *parser.FunctionDefinition)
	check_HostClassDefinition(e *parser.HostClassDefinition)
	check_IfExpression(e *parser.IfExpression)
	check_KeyedEntry(e *parser.KeyedEntry)
	check_LambdaExpression(e *parser.LambdaExpression)
	check_LiteralHash(e *parser.LiteralHash)
	check_LiteralList(e *parser.LiteralList)
	check_NamedAccessExpression(e *parser.NamedAccessExpression)
	check_NamedDefinition(e parser.NamedDefinition)
	check_NodeDefinition(e *parser.NodeDefinition)
	check_Parameter(e *parser.Parameter)
	check_QueryExpression(e parser.QueryExpression)
	check_RelationshipExpression(e *parser.RelationshipExpression)
	check_ReservedWord(e *parser.ReservedWord)
	check_ResourceBody(e *parser.ResourceBody)
	check_ResourceDefaultsExpression(e *parser.ResourceDefaultsExpression)
	check_ResourceExpression(e *parser.ResourceExpression)
	check_ResourceOverrideExpression(e *parser.ResourceOverrideExpression)
	check_ResourceTypeDefinition(e *parser.ResourceTypeDefinition)
	check_SelectorEntry(e *parser.SelectorEntry)
	check_SelectorExpression(e *parser.SelectorExpression)
	check_SiteDefinition(e *parser.SiteDefinition)
	check_TypeAlias(e *parser.TypeAlias)
	check_TypeMapping(e *parser.TypeMapping)
	check_UnaryExpression(e parser.UnaryExpression)
	check_UnlessExpression(e *parser.UnlessExpression)
}

func NewChecker(strict Strictness) Checker {
	basicChecker := &basicChecker{}
	basicChecker.initialize(strict)
	return basicChecker
}

func Check(v Checker, e parser.Expression) {
	switch e.(type) {
	case *parser.Application:
		v.check_Application(e.(*parser.Application))
	case *parser.AssignmentExpression:
		v.check_AssignmentExpression(e.(*parser.AssignmentExpression))
	case *parser.AttributeOperation:
		v.check_AttributeOperation(e.(*parser.AttributeOperation))
	case *parser.AttributesOperation:
		v.check_AttributesOperation(e.(*parser.AttributesOperation))
	case *parser.BlockExpression:
		v.check_BlockExpression(e.(*parser.BlockExpression))
	case *parser.CallNamedFunctionExpression:
		v.check_CallNamedFunctionExpression(e.(*parser.CallNamedFunctionExpression))
	case *parser.CapabilityMapping:
		v.check_CapabilityMapping(e.(*parser.CapabilityMapping))
	case *parser.CaseExpression:
		v.check_CaseExpression(e.(*parser.CaseExpression))
	case *parser.CaseOption:
		v.check_CaseOption(e.(*parser.CaseOption))
	case *parser.CollectExpression:
		v.check_CollectExpression(e.(*parser.CollectExpression))
	case *parser.EppExpression:
		v.check_EppExpression(e.(*parser.EppExpression))
	case *parser.FunctionDefinition:
		v.check_FunctionDefinition(e.(*parser.FunctionDefinition))
	case *parser.HostClassDefinition:
		v.check_HostClassDefinition(e.(*parser.HostClassDefinition))
	case *parser.IfExpression:
		v.check_IfExpression(e.(*parser.IfExpression))
	case *parser.KeyedEntry:
		v.check_KeyedEntry(e.(*parser.KeyedEntry))
	case *parser.LambdaExpression:
		v.check_LambdaExpression(e.(*parser.LambdaExpression))
	case *parser.LiteralHash:
		v.check_LiteralHash(e.(*parser.LiteralHash))
	case *parser.LiteralList:
		v.check_LiteralList(e.(*parser.LiteralList))
	case *parser.NamedAccessExpression:
		v.check_NamedAccessExpression(e.(*parser.NamedAccessExpression))
	case *parser.NodeDefinition:
		v.check_NodeDefinition(e.(*parser.NodeDefinition))
	case *parser.Parameter:
		v.check_Parameter(e.(*parser.Parameter))
	case *parser.RelationshipExpression:
		v.check_RelationshipExpression(e.(*parser.RelationshipExpression))
	case *parser.ReservedWord:
		v.check_ReservedWord(e.(*parser.ReservedWord))
	case *parser.ResourceBody:
		v.check_ResourceBody(e.(*parser.ResourceBody))
	case *parser.ResourceDefaultsExpression:
		v.check_ResourceDefaultsExpression(e.(*parser.ResourceDefaultsExpression))
	case *parser.ResourceExpression:
		v.check_ResourceExpression(e.(*parser.ResourceExpression))
	case *parser.ResourceOverrideExpression:
		v.check_ResourceOverrideExpression(e.(*parser.ResourceOverrideExpression))
	case *parser.ResourceTypeDefinition:
		v.check_ResourceTypeDefinition(e.(*parser.ResourceTypeDefinition))
	case *parser.SelectorEntry:
		v.check_SelectorEntry(e.(*parser.SelectorEntry))
	case *parser.SelectorExpression:
		v.check_SelectorExpression(e.(*parser.SelectorExpression))
	case *parser.SiteDefinition:
		v.check_SiteDefinition(e.(*parser.SiteDefinition))
	case *parser.TypeAlias:
		v.check_TypeAlias(e.(*parser.TypeAlias))
	case *parser.TypeMapping:
		v.check_TypeMapping(e.(*parser.TypeMapping))
	case *parser.UnlessExpression:
		v.check_UnlessExpression(e.(*parser.UnlessExpression))

	// Interface switches
	case parser.BinaryExpression:
		v.check_BinaryExpression(e.(parser.BinaryExpression))
	case parser.QueryExpression:
		v.check_QueryExpression(e.(parser.QueryExpression))
	case parser.UnaryExpression:
		v.check_UnaryExpression(e.(parser.UnaryExpression))
	}
}

func (v *basicChecker) Validate(e parser.Expression) {
	Check(v, e)
}

func (v *basicChecker) initialize(strict Strictness) {
	v.issues = make([]*issue.Reported, 0, 5)
	v.severities = make(map[issue.Code]issue.Severity, 5)
	v.Demote(VALIDATE_FUTURE_RESERVED_WORD, issue.SEVERITY_DEPRECATION)
	v.Demote(VALIDATE_DUPLICATE_KEY, issue.Severity(strict))
	v.Demote(VALIDATE_IDEM_EXPRESSION_NOT_LAST, issue.Severity(strict))
}

func (v *basicChecker) check_AssignmentExpression(e *parser.AssignmentExpression) {
	switch e.Operator() {
	case `=`:
		v.checkAssign(e.Lhs())
	default:
		v.Accept(VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED, e, issue.H{`operator`: e.Operator()})
	}
}

func (v *basicChecker) check_Application(e *parser.Application) {
}

func (v *basicChecker) check_AttributeOperation(e *parser.AttributeOperation) {
	if e.Operator() == `+>` {
		p := v.Container()
		switch p.(type) {
		case *parser.CollectExpression, *parser.ResourceOverrideExpression:
			return
		default:
			v.Accept(VALIDATE_ILLEGAL_ATTRIBUTE_APPEND, e, issue.H{`attr`: e.Name(), `expression`: p})
		}
	}
}

func (v *basicChecker) check_AttributesOperation(e *parser.AttributesOperation) {
	p := v.Container()
	switch p.(type) {
	case parser.AbstractResource, *parser.CollectExpression, *parser.CapabilityMapping:
		v.Accept(VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT, p, issue.H{`operator`: `* =>`, `value`: p})
	}
	v.checkRValue(e.Expr())
}

func (v *basicChecker) check_BinaryExpression(e parser.BinaryExpression) {
	v.checkRValue(e.Lhs())
	v.checkRValue(e.Rhs())
}

func (v *basicChecker) check_BlockExpression(e *parser.BlockExpression) {
	last := len(e.Statements()) - 1
	for idx, statement := range e.Statements() {
		if idx != last && v.isIdem(statement) {
			v.Accept(VALIDATE_IDEM_EXPRESSION_NOT_LAST, statement, issue.H{`expression`: statement})
			break
		}
	}
}

func (v *basicChecker) check_CallNamedFunctionExpression(e *parser.CallNamedFunctionExpression) {
	switch e.Functor().(type) {
	case *parser.QualifiedName:
		return
	case *parser.QualifiedReference:
		// Call to type
		return
	case *parser.AccessExpression:
		ae, _ := e.Functor().(*parser.AccessExpression)
		if _, ok := ae.Operand().(*parser.QualifiedReference); ok {
			// Call to parameterized type
			return
		}
	}
	v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Functor(),
		issue.H{`expression`: e.Functor(), `feature`: `function name`, `container`: e})
}

func (v *basicChecker) check_CapabilityMapping(e *parser.CapabilityMapping) {
	exprOk := false
	switch e.Component().(type) {
	case *parser.QualifiedReference:
		exprOk = true

	case *parser.QualifiedName:
		v.Accept(VALIDATE_ILLEGAL_CLASSREF, e.Component(), issue.H{`name`: e.Component().(*parser.QualifiedName).Name()})
		exprOk = true // OK, besides from what was just reported

	case *parser.AccessExpression:
		ae, _ := e.Component().(*parser.AccessExpression)
		if _, ok := ae.Operand().(*parser.QualifiedReference); ok && len(ae.Keys()) == 1 {
			switch ae.Keys()[0].(type) {
			case *parser.LiteralString, *parser.QualifiedName, *parser.QualifiedReference:
				exprOk = true
			}
		}
	}

	if !exprOk {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Component(),
			issue.H{`expression`: e.Component(), `feature`: `capability mapping`, `container`: e})
	}

	if !CLASSREF_EXT.MatchString(e.Capability()) {
		v.Accept(VALIDATE_ILLEGAL_CLASSREF, e, issue.H{`name`: e.Capability()})
	}
}

func (v *basicChecker) check_CaseExpression(e *parser.CaseExpression) {
	v.checkRValue(e.Test())
	foundDefault := false
	for _, option := range e.Options() {
		co := option.(*parser.CaseOption)
		for _, value := range co.Values() {
			if _, ok := value.(*parser.LiteralDefault); ok {
				if foundDefault {
					v.Accept(VALIDATE_DUPLICATE_DEFAULT, value, issue.H{`container`: e})
				}
				foundDefault = true
			}
		}
	}
}

func (v *basicChecker) check_CaseOption(e *parser.CaseOption) {
	for _, value := range e.Values() {
		v.checkRValue(value)
	}
}

func (v *basicChecker) check_CollectExpression(e *parser.CollectExpression) {
	if _, ok := e.ResourceType().(*parser.QualifiedReference); !ok {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.ResourceType(),
			issue.H{`expression`: e.ResourceType(), `feature`: `type name`, `container`: e})
	}
}

func (v *basicChecker) check_EppExpression(e *parser.EppExpression) {
	p := v.Container()
	if lambda, ok := p.(*parser.LambdaExpression); ok {
		v.checkNoCapture(lambda, lambda.Parameters())
		v.checkParameterNameUniqueness(lambda, lambda.Parameters())
	}
}

func (v *basicChecker) check_FunctionDefinition(e *parser.FunctionDefinition) {
	v.check_NamedDefinition(e)
	v.checkCaptureLast(e, e.Parameters())
	v.checkReturnType(e, e.ReturnType())
}

func (v *basicChecker) check_HostClassDefinition(e *parser.HostClassDefinition) {
	v.check_NamedDefinition(e)
	v.checkNoCapture(e, e.Parameters())
	v.checkReservedParams(e, e.Parameters())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) check_IfExpression(e *parser.IfExpression) {
	v.checkRValue(e.Test())
}

func (v *basicChecker) check_KeyedEntry(e *parser.KeyedEntry) {
	v.checkRValue(e.Key())
	v.checkRValue(e.Value())
}

func (v *basicChecker) check_LambdaExpression(e *parser.LambdaExpression) {
	v.checkCaptureLast(e, e.Parameters())
	v.checkReturnType(e, e.ReturnType())
}

func (v *basicChecker) check_LiteralHash(e *parser.LiteralHash) {
	unique := make(map[interface{}]bool, len(e.Entries()))
	for _, entry := range e.Entries() {
		key := entry.(*parser.KeyedEntry).Key()
		if literalKey, ok := literal.ToLiteral(key); ok {
			switch key.(type) {
			case *parser.LiteralList, *parser.LiteralHash:
				literalKey = key.ToPN().String()
			}
			if _, ok = unique[literalKey]; ok {
				v.Accept(VALIDATE_DUPLICATE_KEY, entry, issue.H{`key`: key.String()})
			} else {
				unique[literalKey] = true
			}
		}
	}
}

func (v *basicChecker) check_LiteralList(e *parser.LiteralList) {
	for _, element := range e.Elements() {
		v.checkRValue(element)
	}
}

func (v *basicChecker) check_NamedAccessExpression(e *parser.NamedAccessExpression) {
	if _, ok := e.Rhs().(*parser.QualifiedName); !ok {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Rhs(),
			issue.H{`expression`: e.Rhs(), `feature`: `method name`, `container`: v.Container()})
	}
}

func (v *basicChecker) check_NamedDefinition(e parser.NamedDefinition) {
	v.checkTop(e, v.Container())
	if !CLASSREF_DECL.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_DEFINITION_NAME, e, issue.H{`name`: e.Name(), `value`: e})
	}
	v.checkReservedTypeName(e, e.Name())
	v.checkFutureReservedWord(e, e.Name())
	v.checkParameterNameUniqueness(e, e.Parameters())
}

func (v *basicChecker) check_NodeDefinition(e *parser.NodeDefinition) {
	v.checkHostname(e, e.HostMatches())
	v.checkTop(e, v.Container())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) check_Parameter(e *parser.Parameter) {
	if STARTS_WITH_NUMBER.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_NUMERIC_PARAMETER, e, issue.H{`name`: e.Name()})
	} else if !PARAM_NAME.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_PARAMETER_NAME, e, issue.H{`name`: e.Name()})
	}
	if e.Value() != nil {
		v.checkIllegalAssignment(e.Value())
	}
}

func (v *basicChecker) check_QueryExpression(e parser.QueryExpression) {
	if e.Expr() != nil {
		v.checkQuery(e.Expr())
	}
}

func (v *basicChecker) check_RelationshipExpression(e *parser.RelationshipExpression) {
	v.checkRelation(e.Lhs())
	v.checkRelation(e.Rhs())
}

func (v *basicChecker) check_ReservedWord(e *parser.ReservedWord) {
	if e.Future() {
		v.Accept(VALIDATE_FUTURE_RESERVED_WORD, e, issue.H{`word`: e.Name()})
	} else {
		v.Accept(VALIDATE_RESERVED_WORD, e, issue.H{`word`: e.Name()})
	}
}

func (v *basicChecker) check_ResourceBody(e *parser.ResourceBody) {
	seenUnfolding := false
	for _, ao := range e.Operations() {
		if _, ok := ao.(*parser.AttributesOperation); ok {
			if seenUnfolding {
				v.Accept(VALIDATE_MULTIPLE_ATTRIBUTES_UNFOLD, ao, issue.NO_ARGS)
			} else {
				seenUnfolding = true
			}
		}
	}
}

func (v *basicChecker) check_ResourceDefaultsExpression(e *parser.ResourceDefaultsExpression) {
	if e.Form() != parser.REGULAR {
		v.Accept(VALIDATE_NOT_VIRTUALIZABLE, e, issue.NO_ARGS)
	}
}

func (v *basicChecker) check_ResourceExpression(e *parser.ResourceExpression) {
	// # The expression for type name cannot be statically checked - this is instead done at runtime
	// to enable better error message of the result of the expression rather than the static instruction.
	// (This can be revised as there are static constructs that are illegal, but require updating many
	// tests that expect the detailed reporting).
	if e.Form() != parser.REGULAR {
		if typeName, ok := e.TypeName().(*parser.QualifiedName); ok && typeName.Name() == `class` {
			v.Accept(VALIDATE_NOT_VIRTUALIZABLE, e, issue.NO_ARGS)
		}
	}
}

func (v *basicChecker) check_ResourceOverrideExpression(e *parser.ResourceOverrideExpression) {
	if e.Form() != parser.REGULAR {
		v.Accept(VALIDATE_NOT_VIRTUALIZABLE, e, issue.NO_ARGS)
	}
}

func (v *basicChecker) check_ResourceTypeDefinition(e *parser.ResourceTypeDefinition) {
	v.check_NamedDefinition(e)
	v.checkNoCapture(e, e.Parameters())
	v.checkReservedParams(e, e.Parameters())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) check_SelectorEntry(e *parser.SelectorEntry) {
	v.checkRValue(e.Matching())
}

func (v *basicChecker) check_SelectorExpression(e *parser.SelectorExpression) {
	v.checkRValue(e.Lhs())
	seenDefault := false
	for _, entry := range e.Selectors() {
		se := entry.(*parser.SelectorEntry)
		if _, ok := se.Matching().(*parser.LiteralDefault); ok {
			if seenDefault {
				v.Accept(VALIDATE_DUPLICATE_DEFAULT, se, issue.H{`container`: e})
			} else {
				seenDefault = true
			}
		}
	}
}

func (v *basicChecker) check_SiteDefinition(e *parser.SiteDefinition) {
}

func (v *basicChecker) check_TypeAlias(e *parser.TypeAlias) {
	v.checkTop(e, v.Container())
	if !CLASSREF_EXT_DECL.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_DEFINITION_NAME, e, issue.H{`name`: e.Name(), `value`: e})
	}
	v.checkReservedTypeName(e, e.Name())
	v.checkTypeRef(e, e.Type())
}

func (v *basicChecker) check_TypeMapping(e *parser.TypeMapping) {
	v.checkTop(e, v.Container())
	lhs := e.Type()
	lhsType := 0 // Not Runtime
	if ae, ok := lhs.(*parser.AccessExpression); ok {
		if left, ok := ae.Operand().(*parser.QualifiedReference); ok && left.Name() == `Runtime` {
			lhsType = 1 // Runtime
			keys := ae.Keys()

			// Must be a literal string or pattern replacement
			if len(keys) == 2 && isPatternWithReplacement(keys[1]) {
				lhsType = 2
			}
		}
	}
	if lhsType == 0 {
		v.Accept(VALIDATE_UNSUPPORTED_EXPRESSION, e, issue.H{`expression`: e})
	} else {
		rhs := e.Mapping()
		if isPatternWithReplacement(rhs) {
			if lhsType == 1 {
				v.Accept(VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING, e, issue.H{`expression`: e})
			}
		} else if isTypeRef(rhs) {
			if lhsType == 2 {
				v.Accept(VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING, e, issue.H{`expression`: e})
			}
		} else {
			if lhsType == 1 {
				v.Accept(VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING, e, issue.H{`expression`: e})
			} else {
				v.Accept(VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING, e, issue.H{`expression`: e})
			}
		}
	}
}

func (v *basicChecker) check_UnaryExpression(e parser.UnaryExpression) {
	v.checkRValue(e.Expr())
}

func (v *basicChecker) check_UnlessExpression(e *parser.UnlessExpression) {
	v.checkRValue(e.Test())
}

// TODO: Add more validations here

// Helper functions
func (v *basicChecker) checkAssign(e parser.Expression) {
	switch e.(type) {
	case *parser.AccessExpression:
		v.Accept(VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX, e, issue.NO_ARGS)

	case *parser.LiteralList:
		for _, elem := range e.(*parser.LiteralList).Elements() {
			v.checkAssign(elem)
		}

	case *parser.VariableExpression:
		ve := e.(*parser.VariableExpression)
		if name, ok := ve.Name(); ok {
			if DOUBLE_COLON_EXPR.MatchString(name) {
				v.Accept(VALIDATE_CROSS_SCOPE_ASSIGNMENT, e, issue.H{`name`: name})
			}
		} else {
			idx, _ := ve.Index()
			v.Accept(VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT, e, issue.H{`var`: idx})
		}
	}
}

func (v *basicChecker) checkCaptureLast(container parser.Expression, parameters []parser.Expression) {
	last := len(parameters) - 1
	for idx := 0; idx < last; idx++ {
		if param, ok := parameters[idx].(*parser.Parameter); ok && param.CapturesRest() {
			v.Accept(VALIDATE_CAPTURES_REST_NOT_LAST, param, issue.H{`param`: param.Name()})
		}
	}
}

func (v *basicChecker) checkFutureReservedWord(e parser.Expression, w string) {
	if _, ok := FUTURE_RESERVED_WORDS[w]; ok {
		v.Accept(VALIDATE_FUTURE_RESERVED_WORD, e, issue.H{`word`: w})
	}
}

func (v *basicChecker) checkHostname(e parser.Expression, hostMatches []parser.Expression) {
	for _, hostMatch := range hostMatches {
		// Parser syntax prevents a hostMatch from being something other
		// than a ConcatenatedString or LiteralString. It converts numbers and identifiers
		// to LiteralString.
		switch hostMatch.(type) {
		case *parser.ConcatenatedString:
			if lit, ok := literal.ToLiteral(hostMatch); ok {
				v.checkHostnameString(hostMatch, lit.(string))
			} else {
				v.Accept(VALIDATE_ILLEGAL_HOSTNAME_INTERPOLATION, hostMatch, issue.NO_ARGS)
			}
		case *parser.LiteralString:
			v.checkHostnameString(hostMatch, hostMatch.(*parser.LiteralString).StringValue())
		}
	}
}

func (v *basicChecker) checkHostnameString(e parser.Expression, str string) {
	if ILLEGAL_HOSTNAME_CHARS.MatchString(str) {
		v.Accept(VALIDATE_ILLEGAL_HOSTNAME_CHARS, e, issue.H{`hostname`: str})
	}
}

func (v *basicChecker) checkIllegalAssignment(e parser.Expression) {
	if _, ok := e.(*parser.AssignmentExpression); ok {
		v.Accept(VALIDATE_ILLEGAL_ASSIGNMENT_CONTEXT, e, issue.NO_ARGS)
	} else {
		if _, ok := e.(*parser.LambdaExpression); !ok {
			e.Contents(v.path, func(path []parser.Expression, child parser.Expression) {
				v.checkIllegalAssignment(child)
			})
		}
	}
}

func (v *basicChecker) checkNoCapture(container parser.Expression, parameters []parser.Expression) {
	for _, parameter := range parameters {
		if param, ok := parameter.(*parser.Parameter); ok && param.CapturesRest() {
			v.Accept(VALIDATE_CAPTURES_REST_NOT_SUPPORTED, param,
				issue.H{`name`: param.Name(), `container`: container})
		}
	}
}

func (v *basicChecker) checkNoIdemLast(e parser.Definition, body parser.Expression) {
	if violator := v.endsWithIdem(body.(*parser.BlockExpression)); violator != nil {
		v.Accept(VALIDATE_IDEM_NOT_ALLOWED_LAST, violator,
			issue.H{`expression`: violator, `container`: e})
	}
}

func (v *basicChecker) checkParameterNameUniqueness(container parser.Expression, parameters []parser.Expression) {
	unique := make(map[string]bool, 10)
	for _, parameter := range parameters {
		param := parameter.(*parser.Parameter)
		if _, found := unique[param.Name()]; found {
			v.Accept(VALIDATE_DUPLICATE_PARAMETER, parameter, issue.H{`param`: param.Name()})
		} else {
			unique[param.Name()] = true
		}
	}
}

func (v *basicChecker) checkQuery(e parser.Expression) {
	switch e.(type) {
	case *parser.ComparisonExpression:
		switch e.(*parser.ComparisonExpression).Operator() {
		case `==`, `!=`:
			// OK
		default:
			v.Accept(VALIDATE_ILLEGAL_QUERY_EXPRESSION, e, issue.H{`expression`: e})
		}
	case *parser.ParenthesizedExpression:
		v.checkQuery(e.(*parser.ParenthesizedExpression).Expr())
	case *parser.VariableExpression, *parser.QualifiedName, *parser.LiteralInteger, *parser.LiteralFloat, *parser.LiteralString, *parser.LiteralBoolean:
		// OK
	case parser.BooleanExpression:
		be := e.(parser.BooleanExpression)
		v.checkQuery(be.Lhs())
		v.checkQuery(be.Rhs())
	default:
		v.Accept(VALIDATE_ILLEGAL_QUERY_EXPRESSION, e, issue.H{`expression`: e})
	}
}

func (v *basicChecker) checkRelation(e parser.Expression) {
	switch e.(type) {
	case *parser.CollectExpression, *parser.RelationshipExpression:
		// OK
	default:
		v.checkRValue(e)
	}
}

func (v *basicChecker) checkReservedParams(container parser.Expression, parameters []parser.Expression) {
	for _, parameter := range parameters {
		param := parameter.(*parser.Parameter)
		if _, ok := RESERVED_PARAMETERS[param.Name()]; ok {
			v.Accept(VALIDATE_RESERVED_PARAMETER, container, issue.H{`param`: param.Name(), `container`: container})
		}
	}
}

func (v *basicChecker) checkReservedTypeName(e parser.Expression, w string) {
	if _, ok := RESERVED_TYPE_NAMES[strings.ToLower(w)]; ok {
		v.Accept(VALIDATE_RESERVED_TYPE_NAME, e, issue.H{`name`: w, `expression`: e})
	}
}

func (v *basicChecker) checkReturnType(function parser.Expression, returnType parser.Expression) {
	if returnType != nil {
		v.checkTypeRef(function, returnType)
	}
}

func (v *basicChecker) checkRValue(e parser.Expression) {
	switch e.(type) {
	case parser.UnaryExpression:
		v.checkRValue(e.(parser.UnaryExpression).Expr())
	case parser.Definition, *parser.CollectExpression:
		v.Accept(VALIDATE_NOT_RVALUE, e, issue.H{`value`: e})
	}
}

func (v *basicChecker) checkTop(e parser.Expression, c parser.Expression) {
	if c == nil {
		return
	}
	switch c.(type) {
	case *parser.HostClassDefinition, *parser.Program:
		return

	case *parser.BlockExpression:
		c = v.ContainerOf(c)
		if _, ok := c.(*parser.Program); !ok {
			switch e.(type) {
			case *parser.FunctionDefinition, *parser.TypeAlias, *parser.TypeDefinition, *parser.TypeMapping:
				// not ok. These can never be nested in a block
				v.Accept(VALIDATE_NOT_ABSOLUTE_TOP_LEVEL, e, issue.H{`value`: e})
				return
			}
		}
		v.checkTop(e, c)

	default:
		v.Accept(VALIDATE_NOT_TOP_LEVEL, e, issue.NO_ARGS)
	}
}

func (v *basicChecker) checkTypeRef(function parser.Expression, r parser.Expression) {
	n := r
	if ae, ok := r.(*parser.AccessExpression); ok {
		n = ae.Operand()
	}
	if qr, ok := n.(*parser.QualifiedReference); ok {
		v.checkFutureReservedWord(r, qr.DowncasedName())
	} else {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, r,
			issue.H{`expression`: r, `feature`: `a type reference`, `container`: function})
	}
}

func (v *basicChecker) endsWithIdem(e *parser.BlockExpression) parser.Expression {
	top := len(e.Statements())
	if top > 0 {
		last := e.Statements()[top-1]
		if v.isIdem(last) {
			return last
		}
	}
	return nil
}

// Checks if the expression has side effect ('idem' is latin for 'the same', here meaning that the evaluation state
// is known to be unchanged after the expression has been evaluated). The result is not 100% authoritative for
// negative answers since analysis of function behavior is not possible.
func (v *basicChecker) isIdem(e parser.Expression) bool {
	if e == nil {
		return true
	}
	switch e.(type) {
	case *parser.AccessExpression, *parser.ConcatenatedString, *parser.HeredocExpression, *parser.LiteralList, *parser.LiteralHash, *parser.Nop, *parser.SelectorExpression:
		return true
	case *parser.BlockExpression:
		return v.idem_BlockExpression(e.(*parser.BlockExpression))
	case *parser.CaseExpression:
		return v.idem_CaseExpression(e.(*parser.CaseExpression))
	case *parser.CaseOption:
		return v.idem_CaseOption(e.(*parser.CaseOption))
	case *parser.IfExpression:
		return v.idem_IfExpression(e.(*parser.IfExpression))
	case *parser.UnlessExpression:
		return v.idem_IfExpression(&e.(*parser.UnlessExpression).IfExpression)
	case *parser.ParenthesizedExpression:
		return v.isIdem(e.(*parser.ParenthesizedExpression).Expr())
	case *parser.AssignmentExpression, *parser.RelationshipExpression,
		*parser.RenderExpression, *parser.RenderStringExpression,
		*parser.MatchExpression:
		return false
	case parser.BinaryExpression, parser.LiteralValue, parser.UnaryExpression:
		return true
	default:
		return false
	}
}

func isPatternWithReplacement(e parser.Expression) bool {
	if v, ok := e.(*parser.LiteralList); ok && len(v.Elements()) == 2 {
		elems := v.Elements()
		if _, ok := elems[0].(*parser.RegexpExpression); ok {
			_, ok := elems[1].(*parser.LiteralString)
			return ok
		}
	}
	return false
}

func isTypeRef(e parser.Expression) bool {
	n := e
	if ae, ok := e.(*parser.AccessExpression); ok {
		n = ae.Operand()
	}
	_, ok := n.(*parser.QualifiedReference)
	return ok
}

func (v *basicChecker) idem_BlockExpression(e *parser.BlockExpression) bool {
	for _, statement := range e.Statements() {
		if !v.isIdem(statement) {
			return false
		}
	}
	return true
}

func (v *basicChecker) idem_CaseExpression(e *parser.CaseExpression) bool {
	if v.isIdem(e.Test()) {
		for _, option := range e.Options() {
			if !v.isIdem(option) {
				return false
			}
		}
		return true
	}
	return false
}

func (v *basicChecker) idem_CaseOption(e *parser.CaseOption) bool {
	for _, value := range e.Values() {
		if !v.isIdem(value) {
			return false
		}
	}
	return v.isIdem(e.Then())
}

func (v *basicChecker) idem_IfExpression(e *parser.IfExpression) bool {
	return v.isIdem(e.Test()) && v.isIdem(e.Then()) && v.isIdem(e.Else())
}
