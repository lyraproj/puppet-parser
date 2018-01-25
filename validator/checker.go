package validator

import (
	. "github.com/puppetlabs/go-parser/issue"
	. "github.com/puppetlabs/go-parser/literal"
	. "github.com/puppetlabs/go-parser/parser"
	. "regexp"
	"strings"
)

var DOUBLE_COLON_EXPR = MustCompile(`::`)

// CLASSREF_EXT matches a class reference the same way as the lexer - i.e. the external source form
// where each part must start with a capital letter A-Z.
var CLASSREF_EXT = MustCompile(`\A(?:::)?[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)

// Same as CLASSREF_EXT but cannot start with '::'
var CLASSREF_EXT_DECL = MustCompile(`\A[A-Z][\w]*(?:::[A-Z][\w]*)*\z`)

// CLASSREF matches a class reference the way it is represented internally in the
// model (i.e. in lower case).
var CLASSREF_DECL = MustCompile(`\A[a-z][\w]*(?:::[a-z][\w]*)*\z`)

// ILLEGAL_P3_1_HOSTNAME matches if a hostname contains illegal characters.
// This check does not prevent pathological names like 'a....b', '.....', "---". etc.
var ILLEGAL_HOSTNAME_CHARS = MustCompile(`[^-\w.]`)

// PARAM_NAME matches the name part of a parameter (The $ character is not included)
var PARAM_NAME = MustCompile(`\A[a-z_]\w*\z`)

var STARTS_WITH_NUMBER = MustCompile(`\A[0-9]`)

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

	check_Application(e *Application)
	check_AssignmentExpression(e *AssignmentExpression)
	check_AttributeOperation(e *AttributeOperation)
	check_AttributesOperation(e *AttributesOperation)
	check_BinaryExpression(e BinaryExpression)
	check_BlockExpression(e *BlockExpression)
	check_CallNamedFunctionExpression(e *CallNamedFunctionExpression)
	check_CapabilityMapping(e *CapabilityMapping)
	check_CaseExpression(e *CaseExpression)
	check_CaseOption(e *CaseOption)
	check_CollectExpression(e *CollectExpression)
	check_EppExpression(e *EppExpression)
	check_FunctionDefinition(e *FunctionDefinition)
	check_HostClassDefinition(e *HostClassDefinition)
	check_IfExpression(e *IfExpression)
	check_KeyedEntry(e *KeyedEntry)
	check_LambdaExpression(e *LambdaExpression)
	check_LiteralHash(e *LiteralHash)
	check_LiteralList(e *LiteralList)
	check_NamedAccessExpression(e *NamedAccessExpression)
	check_NamedDefinition(e NamedDefinition)
	check_NodeDefinition(e *NodeDefinition)
	check_Parameter(e *Parameter)
	check_QueryExpression(e QueryExpression)
	check_RelationshipExpression(e *RelationshipExpression)
	check_ReservedWord(e *ReservedWord)
	check_ResourceBody(e *ResourceBody)
	check_ResourceDefaultsExpression(e *ResourceDefaultsExpression)
	check_ResourceExpression(e *ResourceExpression)
	check_ResourceOverrideExpression(e *ResourceOverrideExpression)
	check_ResourceTypeDefinition(e *ResourceTypeDefinition)
	check_SelectorEntry(e *SelectorEntry)
	check_SelectorExpression(e *SelectorExpression)
	check_SiteDefinition(e *SiteDefinition)
	check_TypeAlias(e *TypeAlias)
	check_TypeMapping(e *TypeMapping)
	check_UnaryExpression(e UnaryExpression)
	check_UnlessExpression(e *UnlessExpression)
}

func NewChecker(strict Strictness) Checker {
	basicChecker := &basicChecker{}
	basicChecker.initialize(strict)
	return basicChecker
}

func Check(v Checker, e Expression) {
	switch e.(type) {
	case *Application:
		v.check_Application(e.(*Application))
	case *AssignmentExpression:
		v.check_AssignmentExpression(e.(*AssignmentExpression))
	case *AttributeOperation:
		v.check_AttributeOperation(e.(*AttributeOperation))
	case *AttributesOperation:
		v.check_AttributesOperation(e.(*AttributesOperation))
	case *BlockExpression:
		v.check_BlockExpression(e.(*BlockExpression))
	case *CallNamedFunctionExpression:
		v.check_CallNamedFunctionExpression(e.(*CallNamedFunctionExpression))
	case *CapabilityMapping:
		v.check_CapabilityMapping(e.(*CapabilityMapping))
	case *CaseExpression:
		v.check_CaseExpression(e.(*CaseExpression))
	case *CaseOption:
		v.check_CaseOption(e.(*CaseOption))
	case *CollectExpression:
		v.check_CollectExpression(e.(*CollectExpression))
	case *EppExpression:
		v.check_EppExpression(e.(*EppExpression))
	case *FunctionDefinition:
		v.check_FunctionDefinition(e.(*FunctionDefinition))
	case *HostClassDefinition:
		v.check_HostClassDefinition(e.(*HostClassDefinition))
	case *IfExpression:
		v.check_IfExpression(e.(*IfExpression))
	case *KeyedEntry:
		v.check_KeyedEntry(e.(*KeyedEntry))
	case *LambdaExpression:
		v.check_LambdaExpression(e.(*LambdaExpression))
	case *LiteralHash:
		v.check_LiteralHash(e.(*LiteralHash))
	case *LiteralList:
		v.check_LiteralList(e.(*LiteralList))
	case *NamedAccessExpression:
		v.check_NamedAccessExpression(e.(*NamedAccessExpression))
	case *NodeDefinition:
		v.check_NodeDefinition(e.(*NodeDefinition))
	case *Parameter:
		v.check_Parameter(e.(*Parameter))
	case *RelationshipExpression:
		v.check_RelationshipExpression(e.(*RelationshipExpression))
	case *ReservedWord:
		v.check_ReservedWord(e.(*ReservedWord))
	case *ResourceBody:
		v.check_ResourceBody(e.(*ResourceBody))
	case *ResourceDefaultsExpression:
		v.check_ResourceDefaultsExpression(e.(*ResourceDefaultsExpression))
	case *ResourceExpression:
		v.check_ResourceExpression(e.(*ResourceExpression))
	case *ResourceOverrideExpression:
		v.check_ResourceOverrideExpression(e.(*ResourceOverrideExpression))
	case *ResourceTypeDefinition:
		v.check_ResourceTypeDefinition(e.(*ResourceTypeDefinition))
	case *SelectorEntry:
		v.check_SelectorEntry(e.(*SelectorEntry))
	case *SelectorExpression:
		v.check_SelectorExpression(e.(*SelectorExpression))
	case *SiteDefinition:
		v.check_SiteDefinition(e.(*SiteDefinition))
	case *TypeAlias:
		v.check_TypeAlias(e.(*TypeAlias))
	case *TypeMapping:
		v.check_TypeMapping(e.(*TypeMapping))
	case *UnlessExpression:
		v.check_UnlessExpression(e.(*UnlessExpression))

	// Interface switches
	case BinaryExpression:
		v.check_BinaryExpression(e.(BinaryExpression))
	case QueryExpression:
		v.check_QueryExpression(e.(QueryExpression))
	case UnaryExpression:
		v.check_UnaryExpression(e.(UnaryExpression))
	}
}

func (v *basicChecker) Validate(e Expression) {
	Check(v, e)
}

func (v *basicChecker) initialize(strict Strictness) {
	v.issues = make([]*ReportedIssue, 0, 5)
	v.severities =  make(map[IssueCode]Severity, 5)
	v.Demote(VALIDATE_FUTURE_RESERVED_WORD, SEVERITY_DEPRECATION)
	v.Demote(VALIDATE_DUPLICATE_KEY, Severity(strict))
	v.Demote(VALIDATE_IDEM_EXPRESSION_NOT_LAST, Severity(strict))
}

func (v *basicChecker) check_AssignmentExpression(e *AssignmentExpression) {
	switch e.Operator() {
	case `=`:
		v.checkAssign(e.Lhs())
	default:
		v.Accept(VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED, e, H{`operator`: e.Operator()})
	}
}

func (v *basicChecker) check_Application(e *Application) {
}

func (v *basicChecker) check_AttributeOperation(e *AttributeOperation) {
	if e.Operator() == `+>` {
		p := v.Container()
		switch p.(type) {
		case *CollectExpression, *ResourceOverrideExpression:
			return
		default:
			v.Accept(VALIDATE_ILLEGAL_ATTRIBUTE_APPEND, e, H{`attr`: e.Name(), `expression`: p})
		}
	}
}

func (v *basicChecker) check_AttributesOperation(e *AttributesOperation) {
	p := v.Container()
	switch p.(type) {
	case AbstractResource, *CollectExpression, *CapabilityMapping:
		v.Accept(VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT, p, H{`operator`: `* =>`, `value`: p})
	}
	v.checkRValue(e.Expr())
}

func (v *basicChecker) check_BinaryExpression(e BinaryExpression) {
	v.checkRValue(e.Lhs())
	v.checkRValue(e.Rhs())
}

func (v *basicChecker) check_BlockExpression(e *BlockExpression) {
	last := len(e.Statements()) - 1
	for idx, statement := range e.Statements() {
		if idx != last && v.isIdem(statement) {
			v.Accept(VALIDATE_IDEM_EXPRESSION_NOT_LAST, statement, H{`expression`: statement})
			break
		}
	}
}

func (v *basicChecker) check_CallNamedFunctionExpression(e *CallNamedFunctionExpression) {
	switch e.Functor().(type) {
	case *QualifiedName:
		return
	case *QualifiedReference:
		// Call to type
		return
	case *AccessExpression:
		ae, _ := e.Functor().(*AccessExpression)
		if _, ok := ae.Operand().(*QualifiedReference); ok {
			// Call to parameterized type
			return
		}
	}
	v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Functor(),
		H{ `expression`: e.Functor(), `feature`: `function name`, `container`: e })
}

func (v *basicChecker) check_CapabilityMapping(e *CapabilityMapping) {
	exprOk := false
	switch e.Component().(type) {
	case *QualifiedReference:
		exprOk = true

	case *QualifiedName:
		v.Accept(VALIDATE_ILLEGAL_CLASSREF, e.Component(), H{`name`: e.Component().(*QualifiedName).Name()})
		exprOk = true // OK, besides from what was just reported

	case *AccessExpression:
		ae, _ := e.Component().(*AccessExpression)
		if _, ok := ae.Operand().(*QualifiedReference); ok && len(ae.Keys()) == 1 {
			switch ae.Keys()[0].(type) {
			case *LiteralString, *QualifiedName, *QualifiedReference:
				exprOk = true
			}
		}
	}

	if !exprOk {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Component(),
			H{`expression`: e.Component(), `feature`: `capability mapping`, `container`: e})
	}

	if !CLASSREF_EXT.MatchString(e.Capability()) {
		v.Accept(VALIDATE_ILLEGAL_CLASSREF, e, H{`name`: e.Capability()})
	}
}

func (v *basicChecker) check_CaseExpression(e *CaseExpression) {
	v.checkRValue(e.Test())
	foundDefault := false
	for _, option := range e.Options() {
		co := option.(*CaseOption)
		for _, value := range co.Values() {
			if _, ok := value.(*LiteralDefault); ok {
				if foundDefault {
					v.Accept(VALIDATE_DUPLICATE_DEFAULT, value, H{`container`: e})
				}
				foundDefault = true
			}
		}
	}
}

func (v *basicChecker) check_CaseOption(e *CaseOption) {
	for _, value := range e.Values() {
		v.checkRValue(value)
	}
}

func (v *basicChecker) check_CollectExpression(e *CollectExpression) {
	if _, ok := e.ResourceType().(*QualifiedReference); !ok {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.ResourceType(),
			H{`expression`: e.ResourceType(), `feature`: `type name`, `container`: e})
	}
}

func (v *basicChecker) check_EppExpression(e *EppExpression) {
	p := v.Container()
	if lambda, ok := p.(*LambdaExpression); ok {
		v.checkNoCapture(lambda, lambda.Parameters())
		v.checkParameterNameUniqueness(lambda, lambda.Parameters())
	}
}

func (v *basicChecker) check_FunctionDefinition(e *FunctionDefinition) {
	v.check_NamedDefinition(e)
	v.checkCaptureLast(e, e.Parameters())
	v.checkReturnType(e, e.ReturnType())
}

func (v *basicChecker) check_HostClassDefinition(e *HostClassDefinition) {
	v.check_NamedDefinition(e)
	v.checkNoCapture(e, e.Parameters())
	v.checkReservedParams(e, e.Parameters())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) check_IfExpression(e *IfExpression) {
	v.checkRValue(e.Test())
}

func (v *basicChecker) check_KeyedEntry(e *KeyedEntry) {
	v.checkRValue(e.Key())
	v.checkRValue(e.Value())
}

func (v *basicChecker) check_LambdaExpression(e *LambdaExpression) {
	v.checkCaptureLast(e, e.Parameters())
	v.checkReturnType(e, e.ReturnType())
}

func (v *basicChecker) check_LiteralHash(e *LiteralHash) {
	unique := make(map[interface{}]bool, len(e.Entries()))
	for _, entry := range e.Entries() {
		key := entry.(*KeyedEntry).Key()
		if literalKey, ok := ToLiteral(key); ok {
			if _, ok = unique[literalKey]; ok {
				v.Accept(VALIDATE_DUPLICATE_KEY, entry, H{`key`: key.String()})
			} else {
				unique[literalKey] = true
			}
		}
	}
}

func (v *basicChecker) check_LiteralList(e *LiteralList) {
	for _, element := range e.Elements() {
		v.checkRValue(element)
	}
}

func (v *basicChecker) check_NamedAccessExpression(e *NamedAccessExpression) {
	if _, ok := e.Rhs().(*QualifiedName); !ok {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, e.Rhs(),
			H{`expression`: e.Rhs(), `feature`: `method name`, `container`: v.Container()})
	}
}

func (v *basicChecker) check_NamedDefinition(e NamedDefinition) {
	v.checkTop(e, v.Container())
	if !CLASSREF_DECL.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_DEFINITION_NAME, e, H{`name`: e.Name(), `value`: e})
	}
	v.checkReservedTypeName(e, e.Name())
	v.checkFutureReservedWord(e, e.Name())
	v.checkParameterNameUniqueness(e, e.Parameters())
}

func (v *basicChecker) check_NodeDefinition(e *NodeDefinition) {
	v.checkHostname(e, e.HostMatches())
	v.checkTop(e, v.Container())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) check_Parameter(e *Parameter) {
	if STARTS_WITH_NUMBER.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_NUMERIC_PARAMETER, e, H{`name`: e.Name()})
	} else if !PARAM_NAME.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_PARAMETER_NAME, e, H{`name`: e.Name()})
	}
	if e.Value() != nil {
		v.checkIllegalAssignment(e.Value())
	}
}

func (v *basicChecker) check_QueryExpression(e QueryExpression) {
	if e.Expr() != nil {
		v.checkQuery(e.Expr())
	}
}

func (v *basicChecker) check_RelationshipExpression(e *RelationshipExpression) {
	v.checkRelation(e.Lhs())
	v.checkRelation(e.Rhs())
}

func (v *basicChecker) check_ReservedWord(e *ReservedWord) {
	if e.Future() {
		v.Accept(VALIDATE_FUTURE_RESERVED_WORD, e, H{`word`: e.Name()})
	} else {
		v.Accept(VALIDATE_RESERVED_WORD, e, H{`word`: e.Name()})
	}
}

func (v *basicChecker) check_ResourceBody(e *ResourceBody) {
	seenUnfolding := false
	for _, ao := range e.Operations() {
		if _, ok := ao.(*AttributesOperation); ok {
			if seenUnfolding {
				v.Accept(VALIDATE_MULTIPLE_ATTRIBUTES_UNFOLD, ao, NO_ARGS)
			} else {
				seenUnfolding = true
			}
		}
	}
}

func (v *basicChecker) check_ResourceDefaultsExpression(e *ResourceDefaultsExpression) {
	if e.Form() != `regular` {
		v.Accept(VALIDATE_NOT_VIRTUALIZABLE, e, NO_ARGS)
	}
}

func (v *basicChecker) check_ResourceExpression(e *ResourceExpression) {
	// # The expression for type name cannot be statically checked - this is instead done at runtime
	// to enable better error message of the result of the expression rather than the static instruction.
	// (This can be revised as there are static constructs that are illegal, but require updating many
	// tests that expect the detailed reporting).
	if e.Form() != `regular` {
		if typeName, ok := e.TypeName().(*QualifiedName); ok && typeName.Name() == `class` {
			v.Accept(VALIDATE_NOT_VIRTUALIZABLE, e, NO_ARGS)
		}
	}
}

func (v *basicChecker) check_ResourceOverrideExpression(e *ResourceOverrideExpression) {
	if e.Form() != `regular` {
		v.Accept(VALIDATE_NOT_VIRTUALIZABLE, e, NO_ARGS)
	}
}

func (v *basicChecker) check_ResourceTypeDefinition(e *ResourceTypeDefinition) {
	v.check_NamedDefinition(e)
	v.checkNoCapture(e, e.Parameters())
	v.checkReservedParams(e, e.Parameters())
	v.checkNoIdemLast(e, e.Body())
}

func (v *basicChecker) check_SelectorEntry(e *SelectorEntry) {
	v.checkRValue(e.Matching())
}

func (v *basicChecker) check_SelectorExpression(e *SelectorExpression) {
	v.checkRValue(e.Lhs())
	seenDefault := false
	for _, entry := range e.Selectors() {
		se := entry.(*SelectorEntry)
		if _, ok := se.Matching().(*LiteralDefault); ok {
			if seenDefault {
				v.Accept(VALIDATE_DUPLICATE_DEFAULT, se, H{`container`: e})
			} else {
				seenDefault = true
			}
		}
	}
}

func (v *basicChecker) check_SiteDefinition(e *SiteDefinition) {
}

func (v *basicChecker) check_TypeAlias(e *TypeAlias) {
	v.checkTop(e, v.Container())
	if !CLASSREF_EXT_DECL.MatchString(e.Name()) {
		v.Accept(VALIDATE_ILLEGAL_DEFINITION_NAME, e, H{`name`: e.Name(), `value`: e})
	}
	v.checkReservedTypeName(e, e.Name())
	v.checkTypeRef(e, e.Type())
}

func (v *basicChecker) check_TypeMapping(e *TypeMapping) {
	v.checkTop(e, v.Container())
	lhs := e.Type()
	lhsType := 0 // Not Runtime
	if ae, ok := lhs.(*AccessExpression); ok {
		if left, ok := ae.Operand().(*QualifiedReference); ok && left.Name() == `Runtime` {
			lhsType = 1 // Runtime
			keys := ae.Keys()

			// Must be a literal string or pattern replacement
			if len(keys) == 2 && isPatternWithReplacement(keys[1]) {
				lhsType = 2
			}
		}
	}
	if lhsType == 0 {
		v.Accept(VALIDATE_UNSUPPORTED_EXPRESSION, e, H{`expression`: e})
	} else {
		rhs := e.Mapping()
		if isPatternWithReplacement(rhs) {
			if lhsType == 1 {
				v.Accept(VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING, e, H{`expression`: e})
			}
		} else if isTypeRef(rhs) {
			if lhsType == 2 {
				v.Accept(VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING, e, H{`expression`: e})
			}
		} else {
			if lhsType == 1 {
				v.Accept(VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING, e, H{`expression`: e})
			} else {
				v.Accept(VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING, e, H{`expression`: e})
			}
		}
	}
}

func (v *basicChecker) check_UnaryExpression(e UnaryExpression) {
	v.checkRValue(e.Expr())
}

func (v *basicChecker) check_UnlessExpression(e *UnlessExpression) {
	v.checkRValue(e.Test())
}

// TODO: Add more validations here

// Helper functions
func (v *basicChecker) checkAssign(e Expression) {
	switch e.(type) {
	case *AccessExpression:
		v.Accept(VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX, e, NO_ARGS)

	case *LiteralList:
		for _, elem := range e.(*LiteralList).Elements() {
			v.checkAssign(elem)
		}

	case *VariableExpression:
		ve := e.(*VariableExpression)
		if name, ok := ve.Name(); ok {
			if DOUBLE_COLON_EXPR.MatchString(name) {
				v.Accept(VALIDATE_CROSS_SCOPE_ASSIGNMENT, e, H{`name`: name})
			}
		} else {
			idx, _ := ve.Index()
			v.Accept(VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT, e, H{`var`: idx})
		}
	}
}

func (v *basicChecker) checkCaptureLast(container Expression, parameters []Expression) {
	last := len(parameters) - 1
	for idx := 0; idx < last; idx++ {
		if param, ok := parameters[idx].(*Parameter); ok && param.CapturesRest() {
			v.Accept(VALIDATE_CAPTURES_REST_NOT_LAST, param, H{`param`: param.Name()})
		}
	}
}

func (v *basicChecker) checkFutureReservedWord(e Expression, w string) {
	if _, ok := FUTURE_RESERVED_WORDS[w]; ok {
		v.Accept(VALIDATE_FUTURE_RESERVED_WORD, e, H{`word`: w})
	}
}

func (v *basicChecker) checkHostname(e Expression, hostMatches []Expression) {
	for _, hostMatch := range hostMatches {
		// Parser syntax prevents a hostMatch from being something other
		// than a ConcatenatedString or LiteralString. It converts numbers and identifiers
		// to LiteralString.
		switch hostMatch.(type) {
		case *ConcatenatedString:
			if lit, ok := ToLiteral(hostMatch); ok {
				v.checkHostnameString(hostMatch, lit.(string))
			} else {
				v.Accept(VALIDATE_ILLEGAL_HOSTNAME_INTERPOLATION, hostMatch, NO_ARGS)
			}
		case *LiteralString:
			v.checkHostnameString(hostMatch, hostMatch.(*LiteralString).StringValue())
		}
	}
}

func (v *basicChecker) checkHostnameString(e Expression, str string) {
	if ILLEGAL_HOSTNAME_CHARS.MatchString(str) {
		v.Accept(VALIDATE_ILLEGAL_HOSTNAME_CHARS, e, H{`hostname`: str})
	}
}

func (v *basicChecker) checkIllegalAssignment(e Expression) {
	if _, ok := e.(*AssignmentExpression); ok {
		v.Accept(VALIDATE_ILLEGAL_ASSIGNMENT_CONTEXT, e, NO_ARGS)
	} else {
		if _, ok := e.(*LambdaExpression); !ok {
			e.Contents(v.path, func(path []Expression, child Expression) {
				v.checkIllegalAssignment(child)
			})
		}
	}
}

func (v *basicChecker) checkNoCapture(container Expression, parameters []Expression) {
	for _, parameter := range parameters {
		if param, ok := parameter.(*Parameter); ok && param.CapturesRest() {
			v.Accept(VALIDATE_CAPTURES_REST_NOT_SUPPORTED, param,
				H{`name`: param.Name(), `container`: container})
		}
	}
}

func (v *basicChecker) checkNoIdemLast(e Definition, body Expression) {
	if violator := v.endsWithIdem(body.(*BlockExpression)); violator != nil {
		v.Accept(VALIDATE_IDEM_NOT_ALLOWED_LAST, violator,
			H{`expression`: violator, `container`: e})
	}
}

func (v *basicChecker) checkParameterNameUniqueness(container Expression, parameters []Expression) {
	unique := make(map[string]bool, 10)
	for _, parameter := range parameters {
		param := parameter.(*Parameter)
		if _, found := unique[param.Name()]; found {
			v.Accept(VALIDATE_DUPLICATE_PARAMETER, parameter, H{`param`: param.Name()})
		} else {
			unique[param.Name()] = true
		}
	}
}

func (v *basicChecker) checkQuery(e Expression) {
	switch e.(type) {
	case *ComparisonExpression:
		switch e.(*ComparisonExpression).Operator() {
		case `==`, `!=`:
			// OK
		default:
			v.Accept(VALIDATE_ILLEGAL_QUERY_EXPRESSION, e, H{`expression`: e})
		}
	case *ParenthesizedExpression:
		v.checkQuery(e.(*ParenthesizedExpression).Expr())
	case *VariableExpression, *QualifiedName, *LiteralInteger, *LiteralFloat, *LiteralString, *LiteralBoolean:
		// OK
	case BooleanExpression:
		be := e.(BooleanExpression)
		v.checkQuery(be.Lhs())
		v.checkQuery(be.Rhs())
	default:
		v.Accept(VALIDATE_ILLEGAL_QUERY_EXPRESSION, e, H{`expression`: e})
	}
}

func (v *basicChecker) checkRelation(e Expression) {
	switch e.(type) {
	case *CollectExpression, *RelationshipExpression:
		// OK
	default:
		v.checkRValue(e)
	}
}

func (v *basicChecker) checkReservedParams(container Expression, parameters []Expression) {
	for _, parameter := range parameters {
		param := parameter.(*Parameter)
		if _, ok := RESERVED_PARAMETERS[param.Name()]; ok {
			v.Accept(VALIDATE_RESERVED_PARAMETER, container, H{`param`: param.Name(), `container`: container})
		}
	}
}

func (v *basicChecker) checkReservedTypeName(e Expression, w string) {
	if _, ok := RESERVED_TYPE_NAMES[strings.ToLower(w)]; ok {
		v.Accept(VALIDATE_RESERVED_TYPE_NAME, e, H{`name`: w, `expression`: e})
	}
}

func (v *basicChecker) checkReturnType(function Expression, returnType Expression) {
	if returnType != nil {
		v.checkTypeRef(function, returnType)
	}
}

func (v *basicChecker) checkRValue(e Expression) {
	switch e.(type) {
	case UnaryExpression:
		v.checkRValue(e.(UnaryExpression).Expr())
	case Definition, *CollectExpression:
		v.Accept(VALIDATE_NOT_RVALUE, e, H{`value`: e})
	}
}

func (v *basicChecker) checkTop(e Expression, c Expression) {
	switch c.(type) {
	case nil, *HostClassDefinition, *Program:
		return

	case *BlockExpression:
		c = v.ContainerOf(c)
		if _, ok := c.(*Program); !ok {
			switch e.(type) {
			case *FunctionDefinition, *TypeAlias, *TypeDefinition, *TypeMapping:
				// not ok. These can never be nested in a block
				v.Accept(VALIDATE_NOT_ABSOLUTE_TOP_LEVEL, e, H{`value`: e})
				return
			}
		}
		v.checkTop(e, c)

	default:
		v.Accept(VALIDATE_NOT_TOP_LEVEL, e, NO_ARGS)
	}
}

func (v *basicChecker) checkTypeRef(function Expression, r Expression) {
	n := r
	if ae, ok := r.(*AccessExpression); ok {
		n = ae.Operand()
	}
	if qr, ok := n.(*QualifiedReference); ok {
		v.checkFutureReservedWord(r, qr.DowncasedName())
	} else {
		v.Accept(VALIDATE_ILLEGAL_EXPRESSION, r,
			H{`expression`: r, `feature`: `a type reference`, `container`: function})
	}
}

func (v *basicChecker) endsWithIdem(e *BlockExpression) Expression {
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
func (v *basicChecker) isIdem(e Expression) bool {
	switch e.(type) {
	case nil, *AccessExpression, *ConcatenatedString, *HeredocExpression, *LiteralList, *LiteralHash, *Nop, *SelectorExpression:
		return true
	case *BlockExpression:
		return v.idem_BlockExpression(e.(*BlockExpression))
	case *CaseExpression:
		return v.idem_CaseExpression(e.(*CaseExpression))
	case *CaseOption:
		return v.idem_CaseOption(e.(*CaseOption))
	case *IfExpression:
		return v.idem_IfExpression(e.(*IfExpression))
	case *UnlessExpression:
		return v.idem_IfExpression(&e.(*UnlessExpression).IfExpression)
	case *ParenthesizedExpression:
		return v.isIdem(e.(*ParenthesizedExpression).Expr())
	case *AssignmentExpression, *RelationshipExpression, *RenderExpression, *RenderStringExpression:
		return false
	case BinaryExpression, LiteralValue, UnaryExpression:
		return true
	default:
		return false
	}
}

func isPatternWithReplacement(e Expression) bool {
	if v, ok := e.(*LiteralList); ok && len(v.Elements()) == 2 {
		elems := v.Elements()
		if _, ok := elems[0].(*RegexpExpression); ok {
			_, ok := elems[1].(*LiteralString)
			return ok
		}
	}
	return false
}

func isTypeRef(e Expression) bool {
	n := e
	if ae, ok := e.(*AccessExpression); ok {
		n = ae.Operand()
	}
	_, ok := n.(*QualifiedReference)
	return ok
}

func (v *basicChecker) idem_BlockExpression(e *BlockExpression) bool {
	for _, statement := range e.Statements() {
		if !v.isIdem(statement) {
			return false
		}
	}
	return true
}

func (v *basicChecker) idem_CaseExpression(e *CaseExpression) bool {
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

func (v *basicChecker) idem_CaseOption(e *CaseOption) bool {
	for _, value := range e.Values() {
		if !v.isIdem(value) {
			return false
		}
	}
	return v.isIdem(e.Then())
}

func (v *basicChecker) idem_IfExpression(e *IfExpression) bool {
	return v.isIdem(e.Test()) && v.isIdem(e.Then()) && v.isIdem(e.Else())
}
