package validator

import (
	"regexp"
	"strings"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-parser/literal"
	"github.com/lyraproj/puppet-parser/parser"
)

var DoubleColonExpr = regexp.MustCompile(`::`)

// CLASSREF_EXT matches a class reference the same way as the lexer - i.e. the external source form
// where each part must start with a capital letter A-Z.
var ClassrefExt = regexp.MustCompile(`\A(?:::)?[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)

// Same as CLASSREF_EXT but cannot start with '::'
var ClassrefExtDecl = regexp.MustCompile(`\A[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)

// CLASSREF matches a class reference the way it is represented internally in the
// model (i.e. in lower case).
var ClassrefDecl = regexp.MustCompile(`\A[a-z][\w]*(?:::[a-z][\w]*)*\z`)

// ILLEGAL_P3_1_HOSTNAME matches if a hostname contains illegal characters.
// This check does not prevent pathological names like 'a....b', '.....', "---". etc.
var IllegalHostnameChars = regexp.MustCompile(`[^-\w.]`)

// PARAM_NAME matches the name part of a parameter (The $ character is not included)
var ParamName = regexp.MustCompile(`\A[a-z_]\w*\z`)

var StartsWithNumber = regexp.MustCompile(`\A[0-9]`)

var ReservedTypeNames = map[string]bool{
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

var FutureReservedWords = map[string]bool{
	`application`: true,
	`produces`:    true,
	`consumes`:    true,
}

var ReservedParameters = map[string]bool{
	`name`:  true,
	`title`: true,
}

type basicChecker struct {
	AbstractValidator
}

type Checker interface {
	Validator

	checkActivityExpression(e *parser.ActivityExpression)
	checkApplication(e *parser.Application)
	checkAssignmentExpression(e *parser.AssignmentExpression)
	checkAttributeOperation(e *parser.AttributeOperation)
	checkAttributesOperation(e *parser.AttributesOperation)
	checkBinaryExpression(e parser.BinaryExpression)
	checkBlockExpression(e *parser.BlockExpression)
	checkCallNamedFunctionExpression(e *parser.CallNamedFunctionExpression)
	checkCapabilityMapping(e *parser.CapabilityMapping)
	checkCaseExpression(e *parser.CaseExpression)
	checkCaseOption(e *parser.CaseOption)
	checkCollectExpression(e *parser.CollectExpression)
	checkEppExpression(e *parser.EppExpression)
	checkFunctionDefinition(e *parser.FunctionDefinition)
	checkHostClassDefinition(e *parser.HostClassDefinition)
	checkIfExpression(e *parser.IfExpression)
	checkKeyedEntry(e *parser.KeyedEntry)
	checkLambdaExpression(e *parser.LambdaExpression)
	checkLiteralHash(e *parser.LiteralHash)
	checkLiteralList(e *parser.LiteralList)
	checkNamedAccessExpression(e *parser.NamedAccessExpression)
	checkNamedDefinition(e parser.NamedDefinition)
	checkNodeDefinition(e *parser.NodeDefinition)
	checkParameter(e *parser.Parameter)
	checkQueryExpression(e parser.QueryExpression)
	checkRelationshipExpression(e *parser.RelationshipExpression)
	checkReservedWord(e *parser.ReservedWord)
	checkResourceBody(e *parser.ResourceBody)
	checkResourceDefaultsExpression(e *parser.ResourceDefaultsExpression)
	checkResourceExpression(e *parser.ResourceExpression)
	checkResourceOverrideExpression(e *parser.ResourceOverrideExpression)
	checkResourceTypeDefinition(e *parser.ResourceTypeDefinition)
	checkSelectorEntry(e *parser.SelectorEntry)
	checkSelectorExpression(e *parser.SelectorExpression)
	checkSiteDefinition(e *parser.SiteDefinition)
	checkTypeAlias(e *parser.TypeAlias)
	checkTypeMapping(e *parser.TypeMapping)
	checkUnaryExpression(e parser.UnaryExpression)
	checkUnlessExpression(e *parser.UnlessExpression)
}

func NewChecker(strict Strictness) Checker {
	basicChecker := &basicChecker{}
	basicChecker.initialize(strict)
	return basicChecker
}

func Check(v Checker, e parser.Expression) {
	switch e := e.(type) {
	case *parser.ActivityExpression:
		v.checkActivityExpression(e)
	case *parser.Application:
		v.checkApplication(e)
	case *parser.AssignmentExpression:
		v.checkAssignmentExpression(e)
	case *parser.AttributeOperation:
		v.checkAttributeOperation(e)
	case *parser.AttributesOperation:
		v.checkAttributesOperation(e)
	case *parser.BlockExpression:
		v.checkBlockExpression(e)
	case *parser.CallNamedFunctionExpression:
		v.checkCallNamedFunctionExpression(e)
	case *parser.CapabilityMapping:
		v.checkCapabilityMapping(e)
	case *parser.CaseExpression:
		v.checkCaseExpression(e)
	case *parser.CaseOption:
		v.checkCaseOption(e)
	case *parser.CollectExpression:
		v.checkCollectExpression(e)
	case *parser.EppExpression:
		v.checkEppExpression(e)
	case *parser.FunctionDefinition:
		v.checkFunctionDefinition(e)
	case *parser.HostClassDefinition:
		v.checkHostClassDefinition(e)
	case *parser.IfExpression:
		v.checkIfExpression(e)
	case *parser.KeyedEntry:
		v.checkKeyedEntry(e)
	case *parser.LambdaExpression:
		v.checkLambdaExpression(e)
	case *parser.LiteralHash:
		v.checkLiteralHash(e)
	case *parser.LiteralList:
		v.checkLiteralList(e)
	case *parser.NamedAccessExpression:
		v.checkNamedAccessExpression(e)
	case *parser.NodeDefinition:
		v.checkNodeDefinition(e)
	case *parser.Parameter:
		v.checkParameter(e)
	case *parser.RelationshipExpression:
		v.checkRelationshipExpression(e)
	case *parser.ReservedWord:
		v.checkReservedWord(e)
	case *parser.ResourceBody:
		v.checkResourceBody(e)
	case *parser.ResourceDefaultsExpression:
		v.checkResourceDefaultsExpression(e)
	case *parser.ResourceExpression:
		v.checkResourceExpression(e)
	case *parser.ResourceOverrideExpression:
		v.checkResourceOverrideExpression(e)
	case *parser.ResourceTypeDefinition:
		v.checkResourceTypeDefinition(e)
	case *parser.SelectorEntry:
		v.checkSelectorEntry(e)
	case *parser.SelectorExpression:
		v.checkSelectorExpression(e)
	case *parser.SiteDefinition:
		v.checkSiteDefinition(e)
	case *parser.TypeAlias:
		v.checkTypeAlias(e)
	case *parser.TypeMapping:
		v.checkTypeMapping(e)
	case *parser.UnlessExpression:
		v.checkUnlessExpression(e)

	// Interface switches
	case parser.BinaryExpression:
		v.checkBinaryExpression(e)
	case parser.QueryExpression:
		v.checkQueryExpression(e)
	case parser.UnaryExpression:
		v.checkUnaryExpression(e)
	}
}

func (v *basicChecker) Validate(e parser.Expression) {
	Check(v, e)
}

func (v *basicChecker) initialize(strict Strictness) {
	v.severities = make(map[issue.Code]issue.Severity, 5)
	v.Demote(ValidateFutureReservedWord, issue.SeverityDeprecation)
	v.Demote(ValidateDuplicateKey, issue.Severity(strict))
	v.Demote(ValidateIdemExpressionNotLast, issue.Severity(strict))
}

func (v *basicChecker) illegalWorkflowOperation(e parser.Expression) {
	v.Accept(ValidateWorkflowOperationNotSupported, e, issue.H{`operation`: e})
}

func (v *basicChecker) checkActivityExpression(e *parser.ActivityExpression) {
	v.illegalWorkflowOperation(e)
}

func (v *basicChecker) checkAssignmentExpression(e *parser.AssignmentExpression) {
	switch e.Operator() {
	case `=`:
		v.checkAssign(e.Lhs())
	default:
		v.Accept(ValidateAppendsDeletesNoLongerSupported, e, issue.H{`operator`: e.Operator()})
	}
}

func (v *basicChecker) checkApplication(e *parser.Application) {
}

func (v *basicChecker) checkAttributeOperation(e *parser.AttributeOperation) {
	if e.Operator() == `+>` {
		p := v.Container()
		switch p.(type) {
		case *parser.CollectExpression, *parser.ResourceOverrideExpression:
			return
		default:
			v.Accept(ValidateIllegalAttributeAppend, e, issue.H{`attr`: e.Name(), `expression`: p})
		}
	}
}

func (v *basicChecker) checkAttributesOperation(e *parser.AttributesOperation) {
	p := v.Container()
	switch p.(type) {
	case parser.AbstractResource, *parser.CollectExpression, *parser.CapabilityMapping:
		v.Accept(ValidateUnsupportedOperatorInContext, p, issue.H{`operator`: `* =>`, `value`: p})
	}
	v.checkRValue(e.Expr())
}

func (v *basicChecker) checkBinaryExpression(e parser.BinaryExpression) {
	v.checkRValue(e.Lhs())
	v.checkRValue(e.Rhs())
}

func (v *basicChecker) checkBlockExpression(e *parser.BlockExpression) {
	last := len(e.Statements()) - 1
	for idx, statement := range e.Statements() {
		if idx != last && v.isIdem(statement) {
			v.Accept(ValidateIdemExpressionNotLast, statement, issue.H{`expression`: statement})
			break
		}
	}
}

func (v *basicChecker) checkCallNamedFunctionExpression(e *parser.CallNamedFunctionExpression) {
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
	v.Accept(ValidateIllegalExpression, e.Functor(),
		issue.H{`expression`: e.Functor(), `feature`: `function name`, `container`: e})
}

func (v *basicChecker) checkCapabilityMapping(e *parser.CapabilityMapping) {
	exprOk := false
	switch e.Component().(type) {
	case *parser.QualifiedReference:
		exprOk = true

	case *parser.QualifiedName:
		v.Accept(ValidateIllegalClassref, e.Component(), issue.H{`name`: e.Component().(*parser.QualifiedName).Name()})
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
		v.Accept(ValidateIllegalExpression, e.Component(),
			issue.H{`expression`: e.Component(), `feature`: `capability mapping`, `container`: e})
	}

	if !ClassrefExt.MatchString(e.Capability()) {
		v.Accept(ValidateIllegalClassref, e, issue.H{`name`: e.Capability()})
	}
}

func (v *basicChecker) checkCaseExpression(e *parser.CaseExpression) {
	v.checkRValue(e.Test())
	foundDefault := false
	for _, option := range e.Options() {
		co := option.(*parser.CaseOption)
		for _, value := range co.Values() {
			if _, ok := value.(*parser.LiteralDefault); ok {
				if foundDefault {
					v.Accept(ValidateDuplicateDefault, value, issue.H{`container`: e})
				}
				foundDefault = true
			}
		}
	}
}

func (v *basicChecker) checkCaseOption(e *parser.CaseOption) {
	for _, value := range e.Values() {
		v.checkRValue(value)
	}
}

func (v *basicChecker) checkCollectExpression(e *parser.CollectExpression) {
	if _, ok := e.ResourceType().(*parser.QualifiedReference); !ok {
		v.Accept(ValidateIllegalExpression, e.ResourceType(),
			issue.H{`expression`: e.ResourceType(), `feature`: `type name`, `container`: e})
	}
}

func (v *basicChecker) checkEppExpression(e *parser.EppExpression) {
	p := v.Container()
	if lambda, ok := p.(*parser.LambdaExpression); ok {
		v.checkNoCapture(lambda, lambda.Parameters())
		v.checkParameterNameUniqueness(lambda, lambda.Parameters())
	}
}

func (v *basicChecker) checkFunctionDefinition(e *parser.FunctionDefinition) {
	v.checkNamedDefinition(e)
	v.checkCaptureLast(e, e.Parameters())
	v.checkReturnType(e, e.ReturnType())
}

func (v *basicChecker) checkHostClassDefinition(e *parser.HostClassDefinition) {
	v.checkNamedDefinition(e)
	v.checkNoCapture(e, e.Parameters())
	v.checkReservedParams(e, e.Parameters())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) checkIfExpression(e *parser.IfExpression) {
	v.checkRValue(e.Test())
}

func (v *basicChecker) checkKeyedEntry(e *parser.KeyedEntry) {
	v.checkRValue(e.Key())
	v.checkRValue(e.Value())
}

func (v *basicChecker) checkLambdaExpression(e *parser.LambdaExpression) {
	v.checkCaptureLast(e, e.Parameters())
	v.checkReturnType(e, e.ReturnType())
}

func (v *basicChecker) checkLiteralHash(e *parser.LiteralHash) {
	unique := make(map[interface{}]bool, len(e.Entries()))
	for _, entry := range e.Entries() {
		key := entry.(*parser.KeyedEntry).Key()
		if literalKey, ok := literal.ToLiteral(key); ok {
			switch key.(type) {
			case *parser.LiteralList, *parser.LiteralHash:
				literalKey = key.ToPN().String()
			}
			if _, ok = unique[literalKey]; ok {
				v.Accept(ValidateDuplicateKey, entry, issue.H{`key`: key.String()})
			} else {
				unique[literalKey] = true
			}
		}
	}
}

func (v *basicChecker) checkLiteralList(e *parser.LiteralList) {
	for _, element := range e.Elements() {
		v.checkRValue(element)
	}
}

func (v *basicChecker) checkNamedAccessExpression(e *parser.NamedAccessExpression) {
	if _, ok := e.Rhs().(*parser.QualifiedName); !ok {
		v.Accept(ValidateIllegalExpression, e.Rhs(),
			issue.H{`expression`: e.Rhs(), `feature`: `method name`, `container`: v.Container()})
	}
}

func (v *basicChecker) checkNamedDefinition(e parser.NamedDefinition) {
	v.checkTop(e, v.Container())
	if !ClassrefDecl.MatchString(e.Name()) {
		v.Accept(ValidateIllegalDefinitionName, e, issue.H{`name`: e.Name(), `value`: e})
	}
	v.checkReservedTypeName(e, e.Name())
	v.checkFutureReservedWord(e, e.Name())
	v.checkParameterNameUniqueness(e, e.Parameters())
}

func (v *basicChecker) checkNodeDefinition(e *parser.NodeDefinition) {
	v.checkHostname(e, e.HostMatches())
	v.checkTop(e, v.Container())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) checkParameter(e *parser.Parameter) {
	if StartsWithNumber.MatchString(e.Name()) {
		v.Accept(ValidateIllegalNumericParameter, e, issue.H{`name`: e.Name()})
	} else if !ParamName.MatchString(e.Name()) {
		v.Accept(ValidateIllegalParameterName, e, issue.H{`name`: e.Name()})
	}
	if e.Value() != nil {
		v.checkIllegalAssignment(e.Value())
	}
}

func (v *basicChecker) checkQueryExpression(e parser.QueryExpression) {
	if e.Expr() != nil {
		v.checkQuery(e.Expr())
	}
}

func (v *basicChecker) checkRelationshipExpression(e *parser.RelationshipExpression) {
	v.checkRelation(e.Lhs())
	v.checkRelation(e.Rhs())
}

func (v *basicChecker) checkReservedWord(e *parser.ReservedWord) {
	if e.Future() {
		v.Accept(ValidateFutureReservedWord, e, issue.H{`word`: e.Name()})
	} else {
		v.Accept(ValidateReservedWord, e, issue.H{`word`: e.Name()})
	}
}

func (v *basicChecker) checkResourceBody(e *parser.ResourceBody) {
	seenUnfolding := false
	for _, ao := range e.Operations() {
		if _, ok := ao.(*parser.AttributesOperation); ok {
			if seenUnfolding {
				v.Accept(ValidateMultipleAttributesUnfold, ao, issue.NoArgs)
			} else {
				seenUnfolding = true
			}
		}
	}
}

func (v *basicChecker) checkResourceDefaultsExpression(e *parser.ResourceDefaultsExpression) {
	if e.Form() != parser.REGULAR {
		v.Accept(ValidateNotVirtualizable, e, issue.NoArgs)
	}
}

func (v *basicChecker) checkResourceExpression(e *parser.ResourceExpression) {
	// # The expression for type name cannot be statically checked - this is instead done at runtime
	// to enable better error message of the result of the expression rather than the static instruction.
	// (This can be revised as there are static constructs that are illegal, but require updating many
	// tests that expect the detailed reporting).
	if e.Form() != parser.REGULAR {
		if typeName, ok := e.TypeName().(*parser.QualifiedName); ok && typeName.Name() == `class` {
			v.Accept(ValidateNotVirtualizable, e, issue.NoArgs)
		}
	}
}

func (v *basicChecker) checkResourceOverrideExpression(e *parser.ResourceOverrideExpression) {
	if e.Form() != parser.REGULAR {
		v.Accept(ValidateNotVirtualizable, e, issue.NoArgs)
	}
}

func (v *basicChecker) checkResourceTypeDefinition(e *parser.ResourceTypeDefinition) {
	v.checkNamedDefinition(e)
	v.checkNoCapture(e, e.Parameters())
	v.checkReservedParams(e, e.Parameters())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) checkSelectorEntry(e *parser.SelectorEntry) {
	v.checkRValue(e.Matching())
}

func (v *basicChecker) checkSelectorExpression(e *parser.SelectorExpression) {
	v.checkRValue(e.Lhs())
	seenDefault := false
	for _, entry := range e.Selectors() {
		se := entry.(*parser.SelectorEntry)
		if _, ok := se.Matching().(*parser.LiteralDefault); ok {
			if seenDefault {
				v.Accept(ValidateDuplicateDefault, se, issue.H{`container`: e})
			} else {
				seenDefault = true
			}
		}
	}
}

func (v *basicChecker) checkSiteDefinition(e *parser.SiteDefinition) {
}

func (v *basicChecker) checkTypeAlias(e *parser.TypeAlias) {
	v.checkTop(e, v.Container())
	if !ClassrefExtDecl.MatchString(e.Name()) {
		v.Accept(ValidateIllegalDefinitionName, e, issue.H{`name`: e.Name(), `value`: e})
	}
	v.checkReservedTypeName(e, e.Name())
	v.checkTypeRef(e, e.Type())
}

func (v *basicChecker) checkTypeMapping(e *parser.TypeMapping) {
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
		v.Accept(ValidateUnsupportedExpression, e, issue.H{`expression`: e})
	} else {
		rhs := e.Mapping()
		if isPatternWithReplacement(rhs) {
			if lhsType == 1 {
				v.Accept(ValidateIllegalSingleTypeMapping, e, issue.H{`expression`: e})
			}
		} else if isTypeRef(rhs) {
			if lhsType == 2 {
				v.Accept(ValidateIllegalRegexpTypeMapping, e, issue.H{`expression`: e})
			}
		} else {
			if lhsType == 1 {
				v.Accept(ValidateIllegalSingleTypeMapping, e, issue.H{`expression`: e})
			} else {
				v.Accept(ValidateIllegalRegexpTypeMapping, e, issue.H{`expression`: e})
			}
		}
	}
}

func (v *basicChecker) checkUnaryExpression(e parser.UnaryExpression) {
	v.checkRValue(e.Expr())
}

func (v *basicChecker) checkUnlessExpression(e *parser.UnlessExpression) {
	v.checkRValue(e.Test())
}

// TODO: Add more validations here

// Helper functions
func (v *basicChecker) checkAssign(e parser.Expression) {
	switch e := e.(type) {
	case *parser.AccessExpression:
		v.Accept(ValidateIllegalAssignmentViaIndex, e, issue.NoArgs)

	case *parser.LiteralList:
		for _, elem := range e.Elements() {
			v.checkAssign(elem)
		}

	case *parser.VariableExpression:
		if name, ok := e.Name(); ok {
			if DoubleColonExpr.MatchString(name) {
				v.Accept(ValidateCrossScopeAssignment, e, issue.H{`name`: name})
			}
		} else {
			idx, _ := e.Index()
			v.Accept(ValidateIllegalNumericAssignment, e, issue.H{`var`: idx})
		}
	}
}

func (v *basicChecker) checkCaptureLast(container parser.Expression, parameters []parser.Expression) {
	last := len(parameters) - 1
	for idx := 0; idx < last; idx++ {
		if param, ok := parameters[idx].(*parser.Parameter); ok && param.CapturesRest() {
			v.Accept(ValidateCapturesRestNotLast, param, issue.H{`param`: param.Name()})
		}
	}
}

func (v *basicChecker) checkFutureReservedWord(e parser.Expression, w string) {
	if _, ok := FutureReservedWords[w]; ok {
		v.Accept(ValidateFutureReservedWord, e, issue.H{`word`: w})
	}
}

func (v *basicChecker) checkHostname(e parser.Expression, hostMatches []parser.Expression) {
	for _, hostMatch := range hostMatches {
		// Parser syntax prevents a hostMatch from being something other
		// than a ConcatenatedString or LiteralString. It converts numbers and identifiers
		// to LiteralString.
		switch hostMatch := hostMatch.(type) {
		case *parser.ConcatenatedString:
			if lit, ok := literal.ToLiteral(hostMatch); ok {
				v.checkHostnameString(hostMatch, lit.(string))
			} else {
				v.Accept(ValidateIllegalHostnameInterpolation, hostMatch, issue.NoArgs)
			}
		case *parser.LiteralString:
			v.checkHostnameString(hostMatch, hostMatch.StringValue())
		}
	}
}

func (v *basicChecker) checkHostnameString(e parser.Expression, str string) {
	if IllegalHostnameChars.MatchString(str) {
		v.Accept(ValidateIllegalHostnameChars, e, issue.H{`hostname`: str})
	}
}

func (v *basicChecker) checkIllegalAssignment(e parser.Expression) {
	if _, ok := e.(*parser.AssignmentExpression); ok {
		v.Accept(ValidateIllegalAssignmentContext, e, issue.NoArgs)
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
			v.Accept(ValidateCapturesRestNotSupported, param,
				issue.H{`name`: param.Name(), `container`: container})
		}
	}
}

func (v *basicChecker) checkNoIdemLast(e parser.Definition, body parser.Expression) {
	if violator := v.endsWithIdem(body.(*parser.BlockExpression)); violator != nil {
		v.Accept(ValidateIdemNotAllowedLast, violator,
			issue.H{`expression`: violator, `container`: e})
	}
}

func (v *basicChecker) checkParameterNameUniqueness(container parser.Expression, parameters []parser.Expression) {
	unique := make(map[string]bool, 10)
	for _, parameter := range parameters {
		param := parameter.(*parser.Parameter)
		if _, found := unique[param.Name()]; found {
			v.Accept(ValidateDuplicateParameter, parameter, issue.H{`param`: param.Name()})
		} else {
			unique[param.Name()] = true
		}
	}
}

func (v *basicChecker) checkQuery(e parser.Expression) {
	switch e := e.(type) {
	case *parser.ComparisonExpression:
		switch e.Operator() {
		case `==`, `!=`:
			// OK
		default:
			v.Accept(ValidateIllegalQueryExpression, e, issue.H{`expression`: e})
		}
	case *parser.ParenthesizedExpression:
		v.checkQuery(e.Expr())
	case *parser.VariableExpression, *parser.QualifiedName, *parser.LiteralInteger, *parser.LiteralFloat, *parser.LiteralString, *parser.LiteralBoolean:
		// OK
	case parser.BooleanExpression:
		v.checkQuery(e.Lhs())
		v.checkQuery(e.Rhs())
	default:
		v.Accept(ValidateIllegalQueryExpression, e, issue.H{`expression`: e})
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
		if _, ok := ReservedParameters[param.Name()]; ok {
			v.Accept(ValidateReservedParameter, container, issue.H{`param`: param.Name(), `container`: container})
		}
	}
}

func (v *basicChecker) checkReservedTypeName(e parser.Expression, w string) {
	if _, ok := ReservedTypeNames[strings.ToLower(w)]; ok {
		v.Accept(ValidateReservedTypeName, e, issue.H{`name`: w, `expression`: e})
	}
}

func (v *basicChecker) checkReturnType(function parser.Expression, returnType parser.Expression) {
	if returnType != nil {
		v.checkTypeRef(function, returnType)
	}
}

func (v *basicChecker) checkRValue(e parser.Expression) {
	switch e := e.(type) {
	case parser.UnaryExpression:
		v.checkRValue(e.Expr())
	case parser.Definition, *parser.CollectExpression:
		v.Accept(ValidateNotRvalue, e, issue.H{`value`: e})
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
				v.Accept(ValidateNotAbsoluteTopLevel, e, issue.H{`value`: e})
				return
			}
		}
		v.checkTop(e, c)

	default:
		v.Accept(ValidateNotTopLevel, e, issue.NoArgs)
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
		v.Accept(ValidateIllegalExpression, r,
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
	switch e := e.(type) {
	case *parser.AccessExpression, *parser.ConcatenatedString, *parser.HeredocExpression, *parser.LiteralList, *parser.LiteralHash, *parser.Nop, *parser.SelectorExpression:
		return true
	case *parser.BlockExpression:
		return v.idemBlockExpression(e)
	case *parser.CaseExpression:
		return v.idemCaseExpression(e)
	case *parser.CaseOption:
		return v.idemCaseOption(e)
	case *parser.IfExpression:
		return v.idemIfExpression(e)
	case *parser.UnlessExpression:
		return v.idemIfExpression(&e.IfExpression)
	case *parser.ParenthesizedExpression:
		return v.isIdem(e.Expr())
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
		els := v.Elements()
		if _, ok := els[0].(*parser.RegexpExpression); ok {
			_, ok := els[1].(*parser.LiteralString)
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

func (v *basicChecker) idemBlockExpression(e *parser.BlockExpression) bool {
	for _, statement := range e.Statements() {
		if !v.isIdem(statement) {
			return false
		}
	}
	return true
}

func (v *basicChecker) idemCaseExpression(e *parser.CaseExpression) bool {
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

func (v *basicChecker) idemCaseOption(e *parser.CaseOption) bool {
	for _, value := range e.Values() {
		if !v.isIdem(value) {
			return false
		}
	}
	return v.isIdem(e.Then())
}

func (v *basicChecker) idemIfExpression(e *parser.IfExpression) bool {
	return v.isIdem(e.Test()) && v.isIdem(e.Then()) && v.isIdem(e.Else())
}
