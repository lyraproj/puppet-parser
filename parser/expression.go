package parser

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-parser/pn"
)

// The AST Model. Designed to match the AST model used by the Puppet
// ruby parser.
//
// See: https://github.com/puppetlabs/puppet/blob/master/lib/puppet/pops/model/ast.pp
//
// This model uses a name starting with a lowercase letter for the structs that represent
// abstract types to avoid them being exported. Each such struct then has a corresponding
// interface which is exported. Structs representing concrete types in the AST model use
// names starting with an uppercase letter. This scheme allows all names of interfaces
// and structs in the model to participate in type switches.
//
// TODO: Ideally, this model should be generated from the TypeSet described in the ast.pp
type (
	PathVisitor func(path []Expression, e Expression)

	Expression interface {
		issue.Labeled

		// Return the location in source for this expression.
		issue.Location

		// Let the given visitor recursively iterate all contained expressions, depth first.
		AllContents(path []Expression, visitor PathVisitor)

		// Let the given visitor iterate all contained expressions.
		Contents(path []Expression, visitor PathVisitor)

		// Returns the string that represents the parsed text that resulted in this expression
		String() string

		// Returns false for all expressions except the Noop expression
		IsNop() bool

		// Represent the expression using Puppet Extended S-Expresssion Notation (PN)
		//
		// An expression is in one of two forms:
		//
		// (op <space separated list of zero to many operands>)
		//
		// or:
		//
		// (op <hash>)
		//
		// in which case the <hash> will contain named arguments for the given op. The first form
		// is always used for ops that take zero or one operand and, in some cases, such as 'concat'
		// when it is impossible to name the operands.
		//
		// Literals, ListPNs, and Hashes do not have operands although they too, implement the PN
		// interface
		//
		ToPN() pn.PN

		ByteLength() int

		ByteOffset() int

		Locator() *Locator

		updateOffsetAndLength(offset int, length int)
	}

	ResourceForm string

	AbstractResource interface {
		Expression
		Form() ResourceForm
	}

	Definition interface {
		Expression

		// Marker method, to ensure unique interface
		ToDefinition() Definition
	}

	BinaryExpression interface {
		Expression

		Lhs() Expression
		Rhs() Expression
	}

	BooleanExpression interface {
		BinaryExpression

		// Marker method, to ensure unique interface
		ToBooleanExpression() BooleanExpression
	}

	CallExpression interface {
		Expression
		Functor() Expression
		Arguments() []Expression
		Lambda() Expression
	}

	NamedDefinition interface {
		Definition
		Name() string
		Parameters() []Expression
		Body() Expression
	}

	QueryExpression interface {
		Expression

		Expr() Expression

		// Marker method, to ensure unique interface
		ToQueryExpression() QueryExpression
	}

	UnaryExpression interface {
		Expression

		Expr() Expression

		// Marker method, to ensure unique interface
		ToUnaryExpression() UnaryExpression
	}

	NameExpression interface {
		Name() string
	}

	LiteralValue interface {
		Expression

		Value() interface{}

		// Marker method, to ensure unique interface
		ToLiteralValue() LiteralValue
	}

	LiteralNumber interface {
		LiteralValue
		Float() float64
		Int() int64
	}

	AccessExpression struct {
		positioned
		operand Expression
		keys    []Expression
	}

	AndExpression struct {
		binaryExpression
	}

	ArithmeticExpression struct {
		binaryExpression
		operator string
	}

	Application struct {
		namedDefinition
	}

	AssignmentExpression struct {
		binaryExpression
		operator string
	}

	AttributeOperation struct {
		positioned
		operator string
		name     string
		value    Expression
	}

	AttributesOperation struct {
		positioned
		expr Expression
	}

	BlockExpression struct {
		positioned
		statements []Expression
	}

	CallFunctionExpression struct {
		callExpression
	}

	CallMethodExpression struct {
		callExpression
	}

	CallNamedFunctionExpression struct {
		callExpression
	}

	CapabilityMapping struct {
		positioned
		kind       string
		capability string
		component  Expression
		mappings   []Expression
	}

	CaseExpression struct {
		positioned
		test    Expression
		options []Expression
	}

	CaseOption struct {
		positioned
		values []Expression
		then   Expression
	}

	CollectExpression struct {
		positioned
		resourceType Expression
		query        Expression
		operations   []Expression
	}

	ComparisonExpression struct {
		binaryExpression
		operator string
	}

	ConcatenatedString struct {
		positioned
		segments []Expression
	}

	EppExpression struct {
		positioned
		parametersSpecified bool
		body                Expression
	}

	ExportedQuery struct {
		queryExpression
	}

	FunctionDefinition struct {
		namedDefinition
		returnType Expression
	}

	PlanDefinition struct {
		FunctionDefinition
	}

	HeredocExpression struct {
		positioned
		syntax string
		text   Expression
	}

	HostClassDefinition struct {
		namedDefinition
		parentClass string
	}

	IfExpression struct {
		positioned
		test     Expression
		then     Expression
		elseExpr Expression
	}

	InExpression struct {
		binaryExpression
	}

	KeyedEntry struct {
		positioned
		key   Expression
		value Expression
	}

	LambdaExpression struct {
		positioned
		parameters []Expression
		body       Expression
		returnType Expression
	}

	LiteralBoolean struct {
		positioned
		value bool
	}

	LiteralDefault struct {
		positioned
	}

	LiteralFloat struct {
		positioned
		value float64
	}

	LiteralHash struct {
		positioned
		entries []Expression
	}

	LiteralInteger struct {
		positioned
		radix int
		value int64
	}

	LiteralList struct {
		positioned
		elements []Expression
	}

	LiteralString struct {
		positioned
		value string
	}

	Locator struct {
		string    string
		file      string
		lineIndex []int
	}

	MatchExpression struct {
		binaryExpression
		operator string
	}

	NamedAccessExpression struct {
		binaryExpression
	}

	NodeDefinition struct {
		positioned
		parent      Expression
		hostMatches []Expression
		body        Expression
	}

	Nop struct {
		positioned
	}

	NotExpression struct {
		unaryExpression
	}

	OrExpression struct {
		binaryExpression
	}

	Parameter struct {
		positioned
		name         string
		value        Expression
		typeExpr     Expression
		capturesRest bool
	}

	ParenthesizedExpression struct {
		unaryExpression
	}

	Program struct {
		positioned
		body        Expression
		definitions []Definition
	}

	qRefDefinition struct {
		positioned
		name string
	}

	QualifiedName struct {
		positioned
		name string
	}

	QualifiedReference struct {
		QualifiedName
		downcasedName string
	}

	RegexpExpression struct {
		positioned
		value string
	}

	RelationshipExpression struct {
		binaryExpression
		operator string
	}

	RenderExpression struct {
		unaryExpression
	}

	RenderStringExpression struct {
		LiteralString
	}

	ReservedWord struct {
		positioned
		word   string
		future bool
	}

	ResourceBody struct {
		positioned
		title      Expression
		operations []Expression
	}

	ResourceDefaultsExpression struct {
		abstractResource
		typeRef    Expression
		operations []Expression
	}

	ResourceExpression struct {
		abstractResource
		typeName Expression
		bodies   []Expression
	}

	ResourceOverrideExpression struct {
		abstractResource
		resources  Expression
		operations []Expression
	}

	ResourceTypeDefinition struct {
		namedDefinition
	}

	SelectorEntry struct {
		positioned
		matching Expression
		value    Expression
	}

	SelectorExpression struct {
		positioned
		lhs       Expression
		selectors []Expression
	}

	SiteDefinition struct {
		positioned
		body Expression
	}

	TextExpression struct {
		unaryExpression
	}

	TypeAlias struct {
		qRefDefinition
		typeExpr Expression
	}

	TypeDefinition struct {
		qRefDefinition
		parent string
		body   Expression
	}

	TypeMapping struct {
		positioned
		typeExpr    Expression
		mappingExpr Expression
	}

	UnaryMinusExpression struct {
		unaryExpression
	}

	UnfoldExpression struct {
		unaryExpression
	}

	LiteralUndef struct {
		positioned
	}

	UnlessExpression struct {
		IfExpression
	}

	VariableExpression struct {
		unaryExpression
	}

	VirtualQuery struct {
		queryExpression
	}

	// Abstract types
	abstractResource struct {
		positioned
		form ResourceForm
	}

	binaryExpression struct {
		positioned
		lhs Expression
		rhs Expression
	}

	callExpression struct {
		positioned
		rvalRequired bool
		functor      Expression
		arguments    []Expression
		lambda       Expression
	}

	namedDefinition struct {
		positioned
		name       string
		parameters []Expression
		body       Expression
	}

	positioned struct {
		locator *Locator
		offset  int
		length  int
	}

	queryExpression struct {
		positioned
		expr Expression
	}

	unaryExpression struct {
		positioned
		expr Expression
	}
)

const(
	VIRTUAL = ResourceForm(`virtual`)
	EXPORTED = ResourceForm(`exported`)
	REGULAR = ResourceForm(`regular`)
)

func NewLocator(file, content string) *Locator {
	return &Locator{string: content, file: file}
}

func (e *Locator) String() string {
	return e.string
}

func (e *Locator) File() string {
	return e.file
}

// Return the line in the source for the given byte offset
func (e *Locator) LineForOffset(offset int) int {
	return sort.SearchInts(e.getLineIndex(), offset+1)
}

// Return the position on a line in the source for the given byte offset
func (e *Locator) PosOnLine(offset int) int {
	return e.offsetOnLine(offset) + 1
}

func (e *Locator) getLineIndex() []int {
	if e.lineIndex == nil {
		li := append(make([]int, 0, 32), 0)
		rdr := NewStringReader(e.string)
		for c, _ := rdr.Next(); c != 0; c, _ = rdr.Next() {
			if c == '\n' {
				li = append(li, rdr.Pos())
			}
		}
		e.lineIndex = li
	}
	return e.lineIndex
}

func (e *Locator) offsetOnLine(offset int) int {
	li := e.getLineIndex()
	line := sort.SearchInts(li, offset+1)
	lineStart := li[line-1]
	if offset == lineStart {
		return 0
	}
	if offset > len(e.string) {
		offset = len(e.string)
	}
	return utf8.RuneCountInString(e.string[lineStart:offset])
}

func (e *positioned) String() string {
	return e.locator.String()[e.offset : e.offset+e.length]
}

func (e *positioned) File() string {
	return e.locator.File()
}

func (e *positioned) Line() int {
	return e.locator.LineForOffset(e.offset)
}

func (e *positioned) Pos() int {
	return e.locator.PosOnLine(e.offset)
}

func (e *positioned) IsNop() bool { return false }

func (e *positioned) ByteLength() int {
	return e.length
}

func (e *positioned) ByteOffset() int {
	return e.offset
}

func (e *positioned) Location() issue.Location {
	return e
}

func (e *positioned) Locator() *Locator {
	return e.locator
}

func (e *positioned) updateOffsetAndLength(offset int, length int) {
	e.offset = offset
	e.length = length
}

func deepVisit(e Expression, path []Expression, visitor PathVisitor, children ...interface{}) {
	if len(children) == 0 {
		return
	}
	childPath := append(path, e)
	for _, child := range children {
		if expr, ok := child.(Expression); ok {
			visitor(childPath, expr)
			expr.AllContents(childPath, visitor)
		} else if exprs, ok := child.([]Expression); ok {
			for _, expr := range exprs {
				visitor(childPath, expr)
				expr.AllContents(childPath, visitor)
			}
		}
	}
}

func shallowVisit(e Expression, path []Expression, visitor PathVisitor, children ...interface{}) {
	if len(children) == 0 {
		return
	}
	childPath := append(path, e)
	for _, child := range children {
		if expr, ok := child.(Expression); ok {
			visitor(childPath, expr)
		} else if exprs, ok := child.([]Expression); ok {
			for _, expr := range exprs {
				visitor(childPath, expr)
			}
		}
	}
}

func (e *abstractResource) Form() ResourceForm {
	return e.form
}

func (e *AccessExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.operand, e.keys)
}

func (e *AccessExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.operand, e.keys)
}

func (e *AccessExpression) Operand() Expression {
	return e.operand
}

func (e *AccessExpression) Keys() []Expression {
	return e.keys
}

func (e *AccessExpression) ToPN() pn.PN {
	return pn.List(append(pnMapArgs(e.Operand()), pnMap(e.Keys())...)).AsCall(`access`)
}

func (e *AndExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *AndExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *AndExpression) ToBooleanExpression() BooleanExpression {
	return e
}

func (e *AndExpression) ToPN() pn.PN { return e.binaryOp(`and`) }

func (e *Application) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *Application) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.parameters, e.body)
}

func (e *Application) ToDefinition() Definition {
	return e
}

func (e *Application) ToPN() pn.PN { return e.definitionPN(`application`, ``, nil) }

func (e *ArithmeticExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ArithmeticExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ArithmeticExpression) ToPN() pn.PN { return e.binaryOp(e.Operator()) }

func (e *ArithmeticExpression) Operator() string {
	return e.operator
}

func (e *AssignmentExpression) Operator() string {
	return e.operator
}

func (e *AssignmentExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *AssignmentExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *AssignmentExpression) ToPN() pn.PN { return e.binaryOp(e.Operator()) }

func (e *AttributeOperation) Operator() string {
	return e.operator
}

func (e *AttributeOperation) Name() string {
	return e.name
}

func (e *AttributeOperation) Value() Expression {
	return e.value
}

func (e *AttributeOperation) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.value)
}

func (e *AttributeOperation) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.value)
}

func (e *AttributeOperation) ToPN() pn.PN {
	return pn.Call(e.Operator(), pn.Literal(e.Name()), e.Value().ToPN())
}

func (e *AttributesOperation) Expr() Expression {
	return e.expr
}

func (e *AttributesOperation) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *AttributesOperation) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *AttributesOperation) ToPN() pn.PN { return pn.Call(`splat-hash`, e.Expr().ToPN()) }

func (e *binaryExpression) Lhs() Expression {
	return e.lhs
}

func (e *binaryExpression) Rhs() Expression {
	return e.rhs
}

func (e *BlockExpression) Statements() []Expression {
	return e.statements
}

func (e *BlockExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.statements)
}

func (e *BlockExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.statements)
}

func (e *BlockExpression) ToPN() pn.PN { return pnList(e.Statements()).AsCall(`block`) }

func (e *callExpression) RvalRequired() bool {
	return e.rvalRequired
}

func (e *callExpression) Functor() Expression {
	return e.functor
}

func (e *callExpression) Arguments() []Expression {
	return e.arguments
}

func (e *callExpression) Lambda() Expression {
	return e.lambda
}

func (e *CallFunctionExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallFunctionExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallFunctionExpression) ToPN() pn.PN {
	s := `invoke-lambda`
	if e.RvalRequired() {
		s = `call-lambda`
	}
	entries := []pn.Entry{e.Functor().ToPN().WithName(`functor`), pnList(e.Arguments()).WithName(`args`)}
	if e.Lambda() != nil {
		entries = append(entries, e.Lambda().ToPN().WithName(`block`))
	}
	return pn.Map(entries).AsCall(s)
}

func (e *CallMethodExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallMethodExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallMethodExpression) ToPN() pn.PN {
	s := `invoke-method`
	if e.RvalRequired() {
		s = `call-method`
	}
	entries := []pn.Entry{e.Functor().ToPN().WithName(`functor`), pnList(e.Arguments()).WithName(`args`)}
	if e.Lambda() != nil {
		entries = append(entries, e.Lambda().ToPN().WithName(`block`))
	}
	return pn.Map(entries).AsCall(s)
}

func (e *CallNamedFunctionExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallNamedFunctionExpression) WithFunctor(functor Expression) *CallNamedFunctionExpression {
	cr := &CallNamedFunctionExpression{}
	*cr = *e
	cr.functor = functor
	return cr
}

func (e *CallNamedFunctionExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallNamedFunctionExpression) ToPN() pn.PN {
	s := `invoke`
	if e.RvalRequired() {
		s = `call`
	}
	entries := []pn.Entry{e.Functor().ToPN().WithName(`functor`), pnList(e.Arguments()).WithName(`args`)}
	if e.Lambda() != nil {
		entries = append(entries, e.Lambda().ToPN().WithName(`block`))
	}
	return pn.Map(entries).AsCall(s)
}

func (e *CapabilityMapping) Kind() string {
	return e.kind
}

func (e *CapabilityMapping) Capability() string {
	return e.capability
}

func (e *CapabilityMapping) Component() Expression {
	return e.component
}

func (e *CapabilityMapping) Mappings() []Expression {
	return e.mappings
}

func (e *CapabilityMapping) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.component, e.mappings)
}

func (e *CapabilityMapping) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.component, e.mappings)
}

func (e *CapabilityMapping) ToDefinition() Definition {
	return e
}

func (e *CapabilityMapping) ToPN() pn.PN {
	return pn.Call(e.Kind(), e.Component().ToPN(), pn.List(append([]pn.PN{pn.Literal(e.Capability())}, pnMap(e.Mappings())...)))
}

func (e *CaseExpression) Test() Expression {
	return e.test
}

func (e *CaseExpression) Options() []Expression {
	return e.options
}

func (e *CaseExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.test, e.options)
}

func (e *CaseExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.test, e.options)
}

func (e *CaseExpression) ToPN() pn.PN { return pn.Call(`case`, e.Test().ToPN(), pnList(e.Options())) }

func (e *CaseOption) Values() []Expression {
	return e.values
}

func (e *CaseOption) Then() Expression {
	return e.then
}

func (e *CaseOption) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.values, e.then)
}

func (e *CaseOption) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.values, e.then)
}

func (e *CaseOption) ToPN() pn.PN {
	return pn.Map([]pn.Entry{pnList(e.Values()).WithName(`when`), pnBlockAsEntry(`then`, e.Then())})
}

func (e *CollectExpression) ResourceType() Expression {
	return e.resourceType
}

func (e *CollectExpression) Query() Expression {
	return e.query
}

func (e *CollectExpression) Operations() []Expression {
	return e.operations
}

func (e *CollectExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.resourceType, e.query, e.operations)
}

func (e *CollectExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.resourceType, e.query, e.operations)
}

func (e *CollectExpression) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 3)
	entries = append(entries, e.ResourceType().ToPN().WithName(`type`), e.Query().ToPN().WithName(`query`))
	if len(e.Operations()) > 0 {
		entries = append(entries, pnList(e.Operations()).WithName(`ops`))
	}
	return pn.Map(entries).AsCall(`collect`)
}

func (e *ComparisonExpression) Operator() string {
	return e.operator
}

func (e *ComparisonExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ComparisonExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ComparisonExpression) ToPN() pn.PN { return e.binaryOp(e.Operator()) }

func (e *ConcatenatedString) Segments() []Expression {
	return e.segments
}

func (e *ConcatenatedString) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.segments)
}

func (e *ConcatenatedString) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.segments)
}

func (e *ConcatenatedString) ToPN() pn.PN { return pnList(e.Segments()).AsCall(`concat`) }

func (e *EppExpression) ParametersSpecified() bool {
	return e.parametersSpecified
}

func (e *EppExpression) Body() Expression {
	return e.body
}

func (e *EppExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.body)
}

func (e *EppExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.body)
}

func (e *EppExpression) ToPN() pn.PN {
	return e.Body().ToPN().AsCall(`epp`)
}

func (e *ExportedQuery) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *ExportedQuery) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *ExportedQuery) Expr() Expression {
	return e.expr
}

func (e *ExportedQuery) ToQueryExpression() QueryExpression {
	return e
}

func (e *ExportedQuery) ToPN() pn.PN {
	if e.Expr().IsNop() {
		return pn.Call(`exported-query`)
	}
	return pn.Call(`exported-query`, e.Expr().ToPN())
}

func (e *FunctionDefinition) ReturnType() Expression {
	return e.returnType
}

func (e *FunctionDefinition) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *FunctionDefinition) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.parameters, e.body)
}

func (e *FunctionDefinition) ToDefinition() Definition {
	return e
}

func (e *FunctionDefinition) ToPN() pn.PN {
	return e.definitionPN(`function`, ``, e.returnType)
}

func (e *HeredocExpression) Syntax() string {
	return e.syntax
}

func (e *HeredocExpression) Text() Expression {
	return e.text
}

func (e *HeredocExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.text)
}

func (e *HeredocExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.text)
}

func (e *HeredocExpression) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 2)
	if e.Syntax() != `` {
		entries = append(entries, pn.Literal(e.Syntax()).WithName(`syntax`))
	}
	entries = append(entries, e.Text().ToPN().WithName(`text`))
	return pn.Map(entries).AsCall(`heredoc`)
}

func (e *HostClassDefinition) ParentClass() string {
	return e.parentClass
}

func (e *HostClassDefinition) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *HostClassDefinition) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.parameters, e.body)
}

func (e *HostClassDefinition) ToDefinition() Definition {
	return e
}

func (e *HostClassDefinition) ToPN() pn.PN {
	return e.definitionPN(`class`, e.parentClass, nil)
}

func (e *IfExpression) Test() Expression {
	return e.test
}

func (e *IfExpression) Then() Expression {
	return e.then
}

func (e *IfExpression) Else() Expression {
	return e.elseExpr
}

func (e *IfExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.test, e.then, e.elseExpr)
}

func (e *IfExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.test, e.then, e.elseExpr)
}

func (e *IfExpression) ToPN() pn.PN { return e.pnIf(`if`) }

func (e *InExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *InExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *InExpression) ToPN() pn.PN { return e.binaryOp(`in`) }

func (e *KeyedEntry) Key() Expression {
	return e.key
}

func (e *KeyedEntry) Value() Expression {
	return e.value
}

func (e *KeyedEntry) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.key, e.value)
}

func (e *KeyedEntry) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.key, e.value)
}

func (e *KeyedEntry) ToPN() pn.PN { return pn.Call(`=>`, e.Key().ToPN(), e.Value().ToPN()) }

func (e *LambdaExpression) Body() Expression {
	return e.body
}

func (e *LambdaExpression) Parameters() []Expression {
	return e.parameters
}

func (e *LambdaExpression) ReturnType() Expression {
	return e.returnType
}

func (e *LambdaExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.parameters, e.body, e.returnType)
}

func (e *LambdaExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.parameters, e.body, e.returnType)
}

func (e *LambdaExpression) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 3)
	if len(e.Parameters()) > 0 {
		entries = append(entries, parametersEntry(e.Parameters()))
	}
	if e.ReturnType() != nil {
		entries = append(entries, e.ReturnType().ToPN().WithName(`returns`))
	}
	if e.Body() != nil {
		entries = append(entries, pnBlockAsEntry(`body`, e.Body()))
	}
	return pn.Map(entries).AsCall(`lambda`)
}

func (e *LiteralBoolean) Bool() bool {
	return e.value
}

func (e *LiteralBoolean) ToPN() pn.PN { return pn.Literal(e.Value()) }

func (e *LiteralBoolean) Value() interface{} {
	return e.value
}

func (e *LiteralBoolean) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralBoolean) Contents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralBoolean) ToLiteralValue() LiteralValue {
	return e
}

func (e *LiteralDefault) Value() interface{} {
	return DEFAULT_INSTANCE
}

func (e *LiteralDefault) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralDefault) Contents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralDefault) ToLiteralValue() LiteralValue {
	return e
}

func (e *LiteralDefault) ToPN() pn.PN { return pn.Call(`default`) }

func (e *LiteralFloat) Float() float64 {
	return e.value
}

func (e *LiteralFloat) Value() interface{} {
	return e.value
}

func (e *LiteralFloat) Int() int64 {
	return int64(e.value)
}

func (e *LiteralFloat) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralFloat) Contents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralFloat) ToLiteralValue() LiteralValue {
	return e
}

func (e *LiteralFloat) ToPN() pn.PN { return pn.Literal(e.Value()) }

func (e *LiteralHash) Entries() []Expression {
	return e.entries
}

func (e *LiteralHash) Get(key string) Expression {
	for _, entry := range e.entries {
		ex, _ := entry.(*KeyedEntry)
		ek := ex.Key()
		switch ek.(type) {
		case *LiteralString:
			if ek.(*LiteralString).value == key {
				return ex.Value()
			}
		case *QualifiedName:
			if ek.(*QualifiedName).name == key {
				return ex.Value()
			}
		case *QualifiedReference:
			if ek.(*QualifiedReference).name == key {
				return ex.Value()
			}
		}
	}
	return nil
}

func (e *LiteralHash) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.entries)
}

func (e *LiteralHash) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.entries)
}

func (e *LiteralHash) ToPN() pn.PN { return pnList(e.Entries()).AsCall(`hash`) }

func (e *LiteralInteger) Float() float64 {
	return float64(e.value)
}

func (e *LiteralInteger) Value() interface{} {
	return e.value
}

func (e *LiteralInteger) Int() int64 {
	return e.value
}

func (e *LiteralInteger) Radix() int {
	return e.radix
}

func (e *LiteralInteger) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralInteger) Contents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralInteger) ToLiteralValue() LiteralValue {
	return e
}

func (e *LiteralInteger) ToPN() pn.PN {
	if e.radix == 10 {
		return pn.Literal(e.Value())
	}
	return pn.Map([]pn.Entry{
		pn.Literal(e.radix).WithName(`radix`), pn.Literal(e.value).WithName(`value`)}).AsCall(`int`)
}

func (e *LiteralList) Elements() []Expression {
	return e.elements
}

func (e *LiteralList) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.elements)
}

func (e *LiteralList) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.elements)
}

func (e *LiteralList) ToPN() pn.PN { return pnList(e.Elements()).AsCall(`array`) }

func (e *LiteralString) StringValue() string {
	return e.value
}

func (e *LiteralString) Value() interface{} {
	return e.value
}

func (e *LiteralString) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralString) Contents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralString) ToLiteralValue() LiteralValue {
	return e
}

func (e *LiteralString) ToPN() pn.PN { return pn.Literal(e.Value()) }

func (e *LiteralUndef) Value() interface{} {
	return nil
}

func (e *LiteralUndef) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralUndef) Contents(path []Expression, visitor PathVisitor) {
}

func (e *LiteralUndef) ToLiteralValue() LiteralValue {
	return e
}

func (e *LiteralUndef) ToPN() pn.PN { return pn.Literal(nil) }

func (e *MatchExpression) Operator() string {
	return e.operator
}

func (e *MatchExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *MatchExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *MatchExpression) ToPN() pn.PN { return e.binaryOp(e.Operator()) }

func (e *namedDefinition) Name() string {
	return e.name
}

func (e *namedDefinition) Parameters() []Expression {
	return e.parameters
}

func (e *namedDefinition) Body() Expression {
	return e.body
}

func (e *NamedAccessExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *NamedAccessExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *NamedAccessExpression) ToPN() pn.PN { return e.binaryOp(`.`) }

func (e *NodeDefinition) Body() Expression {
	return e.body
}

func (e *NodeDefinition) HostMatches() []Expression {
	return e.hostMatches
}

func (e *NodeDefinition) Parent() Expression {
	return e.parent
}

func (e *NodeDefinition) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.parent, e.hostMatches, e.body)
}

func (e *NodeDefinition) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.parent, e.hostMatches, e.body)
}

func (e *NodeDefinition) ToDefinition() Definition {
	return e
}

func (e *NodeDefinition) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 4)
	entries = append(entries, pnList(e.HostMatches()).WithName(`matches`))
	if e.Parent() != nil {
		entries = append(entries, e.Parent().ToPN().WithName(`parent`))
	}
	if e.Body() != nil {
		entries = append(entries, pnBlockAsEntry(`body`, e.Body()))
	}
	return pn.Map(entries).AsCall(`node`)
}

func (e *Nop) IsNop() bool { return true }

func (e *Nop) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *Nop) Contents(path []Expression, visitor PathVisitor) {
}

func (e *Nop) ToPN() pn.PN { return pn.Call(`nop`) }

func (e *NotExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *NotExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *NotExpression) ToUnaryExpression() UnaryExpression {
	return e
}

func (e *NotExpression) ToPN() pn.PN { return pn.Call(`!`, e.Expr().ToPN()) }

func (e *OrExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *OrExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *OrExpression) ToBooleanExpression() BooleanExpression {
	return e
}

func (e *OrExpression) ToPN() pn.PN { return e.binaryOp(`or`) }

func (e *Parameter) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.typeExpr, e.value)
}

func (e *Parameter) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.typeExpr, e.value)
}

func (e *Parameter) CapturesRest() bool {
	return e.capturesRest
}

func (e *Parameter) Name() string {
	return e.name
}

func (e *Parameter) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 3)
	entries = append(entries, pn.Literal(e.Name()).WithName(`name`))
	if e.Type() != nil {
		entries = append(entries, e.Type().ToPN().WithName(`type`))
	}
	if e.CapturesRest() {
		entries = append(entries, pn.Literal(true).WithName(`splat`))
	}
	if e.Value() != nil {
		entries = append(entries, e.Value().ToPN().WithName(`value`))
	}
	return pn.Map(entries).AsCall(`param`)
}

func (e *Parameter) Type() Expression {
	return e.typeExpr
}

func (e *Parameter) Value() Expression {
	return e.value
}

func (e *ParenthesizedExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *ParenthesizedExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *ParenthesizedExpression) ToUnaryExpression() UnaryExpression {
	return e
}

func (e *ParenthesizedExpression) ToPN() pn.PN { return pn.Call(`paren`, e.Expr().ToPN()) }

func (e *PlanDefinition) ToPN() pn.PN {
	return e.definitionPN(`plan`, ``, e.returnType)
}

func (e *Program) Definitions() []Definition {
	return e.definitions
}

func (e *Program) Body() Expression {
	return e.body
}

func (e *Program) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.body)
}

func (e *Program) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.body)
}

func (e *Program) ToPN() pn.PN { return e.Body().ToPN() }

func (e *qRefDefinition) Name() string {
	return e.name
}

func (e *QualifiedName) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *QualifiedName) Contents(path []Expression, visitor PathVisitor) {
}

func (e *QualifiedName) Name() string {
	return e.name
}

func (e *QualifiedName) ToPN() pn.PN { return pn.Literal(e.Name()).AsCall(`qn`) }

func (e *QualifiedName) Value() interface{} {
	return e.name
}

func (e *QualifiedName) ToLiteralValue() LiteralValue {
	return e
}

func (e *QualifiedReference) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *QualifiedReference) Contents(path []Expression, visitor PathVisitor) {
}

func (e *QualifiedReference) DowncasedName() string {
	return e.downcasedName
}

func (e *QualifiedReference) Name() string {
	return e.name
}

func (e *QualifiedReference) WithName(name string) *QualifiedReference {
	rn := &QualifiedReference{}
	*rn = *e
	rn.name = name
	rn.downcasedName = strings.ToLower(name)
	return rn
}

func (e *QualifiedReference) ToPN() pn.PN { return pn.Literal(e.Name()).AsCall(`qr`) }

func (e *RegexpExpression) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *RegexpExpression) Contents(path []Expression, visitor PathVisitor) {
}

func (e *RegexpExpression) Value() interface{} {
	return e.value
}

func (e *RegexpExpression) PatternString() string {
	return e.value
}

func (e *RegexpExpression) ToLiteralValue() LiteralValue {
	return e
}

func (e *RegexpExpression) ToPN() pn.PN { return pn.Literal(e.Value()).AsCall(`regexp`) }

func (e *RelationshipExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *RelationshipExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *RelationshipExpression) Operator() string {
	return e.operator
}

func (e *RelationshipExpression) ToPN() pn.PN { return e.binaryOp(e.Operator()) }

func (e *RenderExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *RenderExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *RenderExpression) ToPN() pn.PN { return pn.Call(`render`, e.Expr().ToPN()) }

func (e *RenderExpression) ToUnaryExpression() UnaryExpression {
	return e
}

func (e *RenderStringExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor)
}

func (e *RenderStringExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor)
}

func (e *RenderStringExpression) ToPN() pn.PN { return pn.Literal(e.Value()).AsCall(`render-s`) }

func (e *ReservedWord) AllContents(path []Expression, visitor PathVisitor) {
}

func (e *ReservedWord) Contents(path []Expression, visitor PathVisitor) {
}

func (e *ReservedWord) ToPN() pn.PN { return pn.Literal(e.Name()).AsCall(`reserved`) }

func (e *ReservedWord) Name() string {
	return e.word
}

func (e *ReservedWord) Future() bool {
	return e.future
}

func (e *ReservedWord) Value() interface{} {
	return e.word
}

func (e *ReservedWord) ToLiteralValue() LiteralValue {
	return e
}

func (e *ResourceBody) Title() Expression {
	return e.title
}

func (e *ResourceBody) Operations() []Expression {
	return e.operations
}

func (e *ResourceBody) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.title, e.operations)
}

func (e *ResourceBody) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.title, e.operations)
}

func (e *ResourceBody) ToPN() pn.PN {
	return pn.Map([]pn.Entry{
		e.Title().ToPN().WithName(`title`),
		pnList(e.Operations()).WithName(`ops`)}).AsCall(`resource-body`)
}

func (e *ResourceDefaultsExpression) TypeRef() Expression {
	return e.typeRef
}

func (e *ResourceDefaultsExpression) Operations() []Expression {
	return e.operations
}

func (e *ResourceDefaultsExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.typeRef, e.operations)
}

func (e *ResourceDefaultsExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.typeRef, e.operations)
}

func (e *ResourceDefaultsExpression) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 3)
	entries = append(entries, e.TypeRef().ToPN().WithName(`type`), pnList(e.Operations()).WithName(`ops`))
	if e.Form() != REGULAR {
		entries = append(entries, pn.Literal(string(e.Form())).WithName(`form`))
	}
	return pn.Map(entries).AsCall(`resource-defaults`)
}

func (e *ResourceExpression) TypeName() Expression {
	return e.typeName
}

func (e *ResourceExpression) Bodies() []Expression {
	return e.bodies
}

func (e *ResourceExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.typeName, e.bodies)
}

func (e *ResourceExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.typeName, e.bodies)
}

func (e *ResourceExpression) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 3)
	entries = append(entries, e.TypeName().ToPN().WithName(`type`))
	bodies := make([]pn.PN, 0, len(e.Bodies()))
	for _, body := range e.bodies {
		bodies = append(bodies, body.ToPN().AsParameters()...)
	}
	entries = append(entries, pn.List(bodies).WithName(`bodies`))
	if e.Form() != REGULAR {
		entries = append(entries, pn.Literal(string(e.Form())).WithName(`form`))
	}
	return pn.Map(entries).AsCall(`resource`)
}

func (e *ResourceOverrideExpression) Resources() Expression {
	return e.resources
}

func (e *ResourceOverrideExpression) Operations() []Expression {
	return e.operations
}

func (e *ResourceOverrideExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.resources, e.operations)
}

func (e *ResourceOverrideExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.resources, e.operations)
}

func (e *ResourceOverrideExpression) ToPN() pn.PN {
	entries := make([]pn.Entry, 0, 3)
	entries = append(entries, e.Resources().ToPN().WithName(`resources`), pnList(e.Operations()).WithName(`ops`))
	if e.Form() != REGULAR {
		entries = append(entries, pn.Literal(string(e.Form())).WithName(`form`))
	}
	return pn.Map(entries).AsCall(`resource-override`)
}

func (e *ResourceTypeDefinition) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *ResourceTypeDefinition) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.parameters, e.body)
}

func (e *ResourceTypeDefinition) ToDefinition() Definition {
	return e
}

func (e *ResourceTypeDefinition) ToPN() pn.PN { return e.definitionPN(`define`, ``, nil) }

func (e *SelectorEntry) Matching() Expression {
	return e.matching
}

func (e *SelectorEntry) Value() Expression {
	return e.value
}

func (e *SelectorEntry) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.matching, e.value)
}

func (e *SelectorEntry) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.matching, e.value)
}

func (e *SelectorEntry) ToPN() pn.PN { return pn.Call(`=>`, e.Matching().ToPN(), e.Value().ToPN()) }

func (e *SelectorExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.lhs, e.selectors)
}

func (e *SelectorExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.lhs, e.selectors)
}

func (e *SelectorExpression) Lhs() Expression {
	return e.lhs
}

func (e *SelectorExpression) Selectors() []Expression {
	return e.selectors
}

func (e *SelectorExpression) ToPN() pn.PN {
	return pn.Call(`?`, e.Lhs().ToPN(), pnList(e.Selectors()))
}

func (e *SiteDefinition) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.body)
}

func (e *SiteDefinition) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.body)
}

func (e *SiteDefinition) Body() Expression {
	return e.body
}

func (e *SiteDefinition) ToDefinition() Definition {
	return e
}

func (e *SiteDefinition) ToPN() pn.PN {
	return e.Body().ToPN().AsCall(`site`)
}

func (e *TextExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *TextExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *TextExpression) ToPN() pn.PN { return pn.Call(`str`, e.Expr().ToPN()) }

func (e *TextExpression) ToUnaryExpression() UnaryExpression {
	return e
}

func (e *TypeAlias) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.typeExpr)
}

func (e *TypeAlias) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.typeExpr)
}

func (e *TypeAlias) ToDefinition() Definition {
	return e
}

func (e *TypeAlias) ToPN() pn.PN {
	return pn.Call(`type-alias`, pn.Literal(e.Name()), e.Type().ToPN())
}

func (e *TypeAlias) Type() Expression {
	return e.typeExpr
}

func (e *TypeDefinition) Parent() string {
	return e.parent
}

func (e *TypeDefinition) Body() Expression {
	return e.body
}

func (e *TypeDefinition) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.body)
}

func (e *TypeDefinition) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.body)
}

func (e *TypeDefinition) ToDefinition() Definition {
	return e
}

func (e *TypeDefinition) ToPN() pn.PN {
	return pn.Call(`type-definition`, pn.Literal(e.Name()), pn.Literal(e.Parent()), e.Body().ToPN())
}

func (e *TypeMapping) Type() Expression {
	return e.typeExpr
}

func (e *TypeMapping) Mapping() Expression {
	return e.mappingExpr
}

func (e *TypeMapping) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.typeExpr, e.mappingExpr)
}

func (e *TypeMapping) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.typeExpr, e.mappingExpr)
}

func (e *TypeMapping) ToDefinition() Definition {
	return e
}

func (e *TypeMapping) ToPN() pn.PN {
	return pn.Call(`type-mapping`, e.Type().ToPN(), e.Mapping().ToPN())
}

func (e *unaryExpression) Expr() Expression {
	return e.expr
}

func (e *UnaryMinusExpression) Contents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *UnaryMinusExpression) AllContents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *UnaryMinusExpression) ToUnaryExpression() UnaryExpression {
	return e
}

func (e *UnaryMinusExpression) ToPN() pn.PN { return pn.Call(`-`, e.Expr().ToPN()) }

func (e *UnfoldExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *UnfoldExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *UnfoldExpression) ToUnaryExpression() UnaryExpression {
	return e
}

func (e *UnfoldExpression) ToPN() pn.PN { return pn.Call(`unfold`, e.Expr().ToPN()) }

func (e *UnlessExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.test, e.then, e.elseExpr)
}

func (e *UnlessExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.test, e.then, e.elseExpr)
}

func (e *UnlessExpression) ToPN() pn.PN { return e.pnIf(`unless`) }

func (e *VariableExpression) Index() (index int64, ok bool) {
	var ix *LiteralInteger
	if ix, ok = e.expr.(*LiteralInteger); ok {
		index = ix.value
	}
	return
}

func (e *VariableExpression) Name() (name string, ok bool) {
	var qn *QualifiedName
	if qn, ok = e.expr.(*QualifiedName); ok {
		name = qn.name
	}
	return
}

func (e *VariableExpression) NameOrIndex() interface{} {
	if qn, ok := e.expr.(*QualifiedName); ok {
		return qn.name
	}
	return e.expr.(*LiteralInteger).value
}

func (e *VariableExpression) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *VariableExpression) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *VariableExpression) ToPN() pn.PN { return pn.Call(`var`, pn.Literal(e.NameOrIndex())) }

func (e *VariableExpression) ToUnaryExpression() UnaryExpression {
	return e
}

func (e *VirtualQuery) AllContents(path []Expression, visitor PathVisitor) {
	deepVisit(e, path, visitor, e.expr)
}

func (e *VirtualQuery) Contents(path []Expression, visitor PathVisitor) {
	shallowVisit(e, path, visitor, e.expr)
}

func (e *VirtualQuery) Expr() Expression {
	return e.expr
}

func (e *VirtualQuery) ToPN() pn.PN {
	if e.Expr().IsNop() {
		return pn.Call(`virtual-query`)
	}
	return pn.Call(`virtual-query`, e.Expr().ToPN())
}

func (e *VirtualQuery) ToQueryExpression() QueryExpression {
	return e
}

func (e *IfExpression) pnIf(name string) pn.PN {
	entries := make([]pn.Entry, 0, 3)
	entries = append(entries, e.Test().ToPN().WithName(`test`))
	if !e.Then().IsNop() {
		entries = append(entries, pnBlockAsEntry(`then`, e.Then()))
	}
	if !e.Else().IsNop() {
		entries = append(entries, pnBlockAsEntry(`else`, e.Else()))
	}
	return pn.Map(entries).AsCall(name)
}

func (e *namedDefinition) definitionPN(typeName string, parent string, returnType Expression) pn.PN {
	entries := make([]pn.Entry, 0, 3)
	entries = append(entries, pn.Literal(e.Name()).WithName(`name`))
	if parent != `` {
		entries = append(entries, pn.Literal(parent).WithName(`parent`))
	}
	if len(e.Parameters()) > 0 {
		entries = append(entries, parametersEntry(e.Parameters()))
	}
	if e.Body() != nil {
		entries = append(entries, pnBlockAsEntry(`body`, e.Body()))
	}
	if returnType != nil {
		entries = append(entries, returnType.ToPN().WithName(`returns`))
	}
	return pn.Map(entries).AsCall(typeName)
}

func parametersEntry(parameters []Expression) pn.Entry {
	params := make([]pn.Entry, len(parameters))
	for idx, param := range parameters {
		p, _ := param.(*Parameter)
		entries := make([]pn.Entry, 0, 3)
		if p.Type() != nil {
			entries = append(entries, p.Type().ToPN().WithName(`type`))
		}
		if p.CapturesRest() {
			entries = append(entries, pn.Literal(true).WithName(`splat`))
		}
		if p.Value() != nil {
			entries = append(entries, p.Value().ToPN().WithName(`value`))
		}
		params[idx] = pn.Map(entries).WithName(p.Name())
	}
	return pn.Map(params).WithName(`params`)
}

func (e *binaryExpression) binaryOp(op string) pn.PN {
	return pn.Call(op, e.Lhs().ToPN(), e.Rhs().ToPN())
}

func pnList(elements []Expression) pn.PN {
	return pn.List(pnMap(elements))
}

func pnMap(elements []Expression) []pn.PN {
	return pnMapArgs(elements...)
}

func pnMapArgs(elements ...Expression) []pn.PN {
	result := make([]pn.PN, len(elements))
	for idx, element := range elements {
		result[idx] = element.ToPN()
	}
	return result
}

func pnBlockAsEntry(name string, expr Expression) pn.Entry {
	if block, ok := expr.(*BlockExpression); ok {
		return pnList(block.Statements()).WithName(name)
	}
	return pn.List(pnMapArgs(expr)).WithName(name)
}
