package parser

import (
  . "strings"
  . "sort"
  . "unicode/utf8"
  . "github.com/puppetlabs/go-parser/issue"
  . "github.com/puppetlabs/go-parser/pn"
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
  PathVisitor func(path [] Expression, e Expression)

  Expression interface {
    // Return the location in source for this expression.
    Location

    // Let the given visitor iterate all contents. The iteration starts with this
    // expression and will then traverse, depth first, into all contained expressions
    AllContents(path []Expression, visitor PathVisitor)

    // Returns a very brief description of this expression suitable to use in error messages
    Label() string

    // Returns the string that represents the parsed text that resulted in this expression
    String() string

    // Returns false for all expressions except the Noop expression
    IsNop() bool

    // Represent the expression using polish notation
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
    // Literals, Arrays, and Hashes do not have operands although they too, implement the PN
    // interface
    //
    ToPN() PN

    byteLength() int

    byteOffset() int

    updateOffsetAndLength(offset int, length int)
  }

  AbstractResource interface {
    Expression
    Form() string
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
    QueryExpression
  }

  FunctionDefinition struct {
    namedDefinition
    returnType Expression
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
    definitions []Expression
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

  QueryExpression struct {
    positioned
    expr Expression
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
    QueryExpression
  }

  // Abstract types
  abstractResource struct {
    positioned
    form string
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

  unaryExpression struct {
    positioned
    expr Expression
  }
)

func (e *Locator) String() string {
  return e.string
}

func (e *Locator) File() string {
  return e.file
}

// Return the line in the source for the given byte offset
func (e *Locator) LineForOffset(offset int) int {
  return SearchInts(e.getLineIndex(), offset+1)
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
  line := SearchInts(li, offset+1)
  lineStart := li[line-1]
  if offset == lineStart {
    return 0
  }
  return RuneCountInString(e.string[lineStart:offset])
}

func (e *positioned) String() string {
  return e.locator.String()[e.offset:e.offset+e.length]
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

func (e *positioned) byteLength() int {
  return e.length
}

func (e *positioned) byteOffset() int {
  return e.offset
}

func (e *positioned) updateOffsetAndLength(offset int, length int) {
  e.offset = offset
  e.length = length
}

func deepVisit(e Expression, path []Expression, visitor PathVisitor, children ...interface{}) {
  visitor(path, e)
  if len(children) == 0 {
    return
  }
  path = append(path, e)
  for _, child := range children {
    if expr, ok := child.(Expression); ok {
      expr.AllContents(path, visitor)
    } else if exprs, ok := child.([]Expression); ok {
      for _, expr := range exprs {
        expr.AllContents(path, visitor)
      }
    }
  }
}

func (e *abstractResource) Form() string {
  return e.form
}

func (e *AccessExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.operand, e.keys)
}

func (e *AccessExpression) Operand() Expression {
  return e.operand
}

func (e *AccessExpression) Keys() []Expression {
  return e.keys
}

func (e *AccessExpression) ToPN() PN { return NamedArray(`[]`, append(pnArgs(e.Operand()), pnElems(e.Keys())...)) }

func (e *AndExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *AndExpression) ToBooleanExpression() BooleanExpression {
  return e
}

func (e *AndExpression) ToPN() PN { return e.binaryOp(`and`) }

func (e *Application) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *Application) ToDefinition() Definition {
  return e
}

func (e *Application) ToPN() PN { return e.pnNamedDefinition(`application`) }

func (e *ArithmeticExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ArithmeticExpression) ToPN() PN { return e.binaryOp(e.Operator()) }

func (e *ArithmeticExpression) Operator() string {
  return e.operator
}

func (e *AssignmentExpression) Operator() string {
  return e.operator
}

func (e *AssignmentExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *AssignmentExpression) ToPN() PN { return e.binaryOp(e.Operator()) }

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

func (e *AttributeOperation) ToPN() PN { return NamedArray(e.Operator(), []PN{StringValue(e.Name()), e.Value().ToPN()}) }

func (e *AttributesOperation) Expr() Expression {
  return e.expr
}

func (e *AttributesOperation) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *AttributesOperation) ToPN() PN { return NamedArray(`=>`, []PN{StringValue(`*`), e.Expr().ToPN()}) }

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

func (e *BlockExpression) ToPN() PN { return NamedArray(`block`, pnElems(e.Statements())) }

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

func (e *CallFunctionExpression) ToPN() PN {
  s := `invoke_lambda`
  if e.RvalRequired() {
    s = `call_lambda`
  }
  entries := []Entry{NamedValue(`functor`, e.Functor().ToPN()), NamedArray(`args`, pnElems(e.Arguments()))}
  if e.Lambda() != nil {
    entries = append(entries, e.Lambda().ToPN().WithName(`block`))
  }
  return NamedValue(s, Hash(entries))
}

func (e *CallMethodExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallMethodExpression) ToPN() PN {
  s := `invoke_method`
  if e.RvalRequired() {
    s = `call_method`
  }
  entries := []Entry{NamedValue(`functor`, e.Functor().ToPN()), NamedArray(`args`, pnElems(e.Arguments()))}
  if e.Lambda() != nil {
    entries = append(entries, e.Lambda().ToPN().WithName(`block`))
  }
  return NamedValue(s, Hash(entries))
}

func (e *CallNamedFunctionExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
}

func (e *CallNamedFunctionExpression) ToPN() PN {
  s := `invoke`
  if e.RvalRequired() {
    s = `call`
  }
  entries := []Entry{NamedValue(`functor`, e.Functor().ToPN()), NamedArray(`args`, pnElems(e.Arguments()))}
  if e.Lambda() != nil {
    entries = append(entries, e.Lambda().ToPN().WithName(`block`))
  }
  return NamedValue(s, Hash(entries))
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

func (e *CapabilityMapping) ToDefinition() Definition {
  return e
}

func (e *CapabilityMapping) ToPN() PN { return NamedArray(e.Kind(), []PN{e.Component().ToPN(), NamedArray(e.Capability(), pnElems(e.Mappings()))}) }

func (e *CaseExpression) Test() Expression {
  return e.test
}

func (e *CaseExpression) Options() []Expression {
  return e.options
}

func (e *CaseExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.test, e.options)
}

func (e *CaseExpression) ToPN() PN { return NamedArray(`case`, pnElems(e.Options())) }

func (e *CaseOption) Values() []Expression {
  return e.values
}

func (e *CaseOption) Then() Expression {
  return e.then
}

func (e *CaseOption) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.values, e.then)
}

func (e *CaseOption) ToPN() PN {
  return Hash([]Entry{
    NamedArray(`when`, pnElems(e.Values())),
    NamedValue(`then`, e.Then().ToPN())})
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

func (e *CollectExpression) ToPN() PN {
  entries := make([]Entry, 0, 3)
  entries = append(entries, NamedValue(`type`, e.ResourceType().ToPN()), NamedValue(`query`, e.Query().ToPN()))
  if len(e.Operations()) > 0 {
    entries = append(entries, NamedArray(`ops`, pnElems(e.Operations())))
  }
  return NamedValue(`collect`, Hash(entries))
}

func (e *ComparisonExpression) Operator() string {
  return e.operator
}

func (e *ComparisonExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ComparisonExpression) ToPN() PN { return e.binaryOp(e.Operator()) }

func (e *ConcatenatedString) Segments() []Expression {
  return e.segments
}

func (e *ConcatenatedString) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.segments)
}

func (e *ConcatenatedString) ToPN() PN { return NamedArray(`concat`, pnElems(e.Segments())) }

func (e *EppExpression) ParametersSpecified() bool {
  return e.parametersSpecified
}

func (e *EppExpression) Body() Expression {
  return e.body
}

func (e *EppExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.body)
}

func (e *EppExpression) ToPN() PN {
  return e.Body().ToPN().WithName(`epp`)
}

func (e *ExportedQuery) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *ExportedQuery) ToPN() PN {
  if e.Expr().IsNop() {
    return StringValue(`<<| |>>`)
  }
  return NamedValue(`<<| |>>`, e.Expr().ToPN())
}

func (e *FunctionDefinition) ReturnType() Expression {
  return e.returnType
}

func (e *FunctionDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *FunctionDefinition) ToDefinition() Definition {
  return e
}

func (e *FunctionDefinition) ToPN() PN {
  entries := make([]Entry, 0, 4)
  entries = append(entries, NamedString(`name`, e.Name()))
  if len(e.Parameters()) > 0 {
    entries = append(entries, NamedArray(`params`, pnElems(e.Parameters())))
  }
  if e.ReturnType() != nil {
    entries = append(entries, NamedValue(`returns`, e.ReturnType().ToPN()))
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().WithName(`body`))
  }
  return NamedValue(`function`, Hash(entries))
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

func (e *HeredocExpression) ToPN() PN {
  entries := make([]Entry, 0, 2)
  if e.Syntax() != `` {
    entries = append(entries, NamedString(`syntax`, e.Syntax()))
  }
  entries = append(entries, NamedValue(`text`, e.Text().ToPN()))
  return NamedValue(`heredoc`, Hash(entries))
}

func (e *HostClassDefinition) ParentClass() string {
  return e.parentClass
}

func (e *HostClassDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *HostClassDefinition) ToDefinition() Definition {
  return e
}

func (e *HostClassDefinition) ToPN() PN {
  entries := make([]Entry, 0, 3)
  entries = append(entries, NamedString(`name`, e.Name()))
  if e.ParentClass() != `` {
    entries = append(entries, NamedString(`parent`, e.ParentClass()))
  }
  if len(e.Parameters()) > 0 {
    entries = append(entries, NamedArray(`params`, pnElems(e.Parameters())))
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().WithName(`body`))
  }
  return NamedValue(`class`, Hash(entries))
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

func (e *IfExpression) ToPN() PN { return e.pnIf(`if`) }

func (e *InExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *InExpression) ToPN() PN { return e.binaryOp(`in`) }

func (e *KeyedEntry) Key() Expression {
  return e.key
}

func (e *KeyedEntry) Value() Expression {
  return e.value
}

func (e *KeyedEntry) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.key, e.value)
}

func (e *KeyedEntry) ToPN() PN { return NamedArray(`=>`, pnArgs(e.Key(), e.Value())) }

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

func (e *LambdaExpression) ToPN() PN {
  entries := make([]Entry, 0, 4)
  if len(e.Parameters()) > 0 {
    entries = append(entries, NamedArray(`params`, pnElems(e.Parameters())))
  }
  if e.ReturnType() != nil {
    entries = append(entries, NamedValue(`returns`, e.ReturnType().ToPN()))
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().WithName(`body`))
  }
  return NamedValue(`lambda`, Hash(entries))
}

func (e *LiteralBoolean) Bool() bool {
  return e.value
}

func (e *LiteralBoolean) ToPN() PN { return Literal(e.Value()) }

func (e *LiteralBoolean) Value() interface{} {
  return e.value
}

func (e *LiteralBoolean) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *LiteralBoolean) ToLiteralValue() LiteralValue {
  return e
}

func (e *LiteralDefault) Value() interface{} {
  return DEFAULT_INSTANCE
}

func (e *LiteralDefault) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *LiteralDefault) ToLiteralValue() LiteralValue {
  return e
}

func (e *LiteralDefault) ToPN() PN { return StringValue(`default`) }

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
  visitor(path, e)
}

func (e *LiteralFloat) ToLiteralValue() LiteralValue {
  return e
}

func (e *LiteralFloat) ToPN() PN { return Literal(e.Value()) }

func (e *LiteralHash) Entries() []Expression {
  return e.entries
}

func (e *LiteralHash) Get(key string) Expression {
  for _, entry := range e.entries {
    ex, _ := entry.(*KeyedEntry)
    if str, ok := ex.Key().(*LiteralString); ok && key == str.String() {
      return ex.Value()
    }
  }
  return nil
}

func (e *LiteralHash) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.entries)
}

func (e *LiteralHash) ToPN() PN { return NamedArray(`hash`, pnElems(e.Entries())) }

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
  visitor(path, e)
}

func (e *LiteralInteger) ToLiteralValue() LiteralValue {
  return e
}

func (e *LiteralInteger) ToPN() PN { return Literal(e.Value()) }

func (e *LiteralList) Elements() []Expression {
  return e.elements
}

func (e *LiteralList) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.elements)
}

func (e *LiteralList) ToPN() PN { return NamedArray(`array`, pnElems(e.Elements())) }

func (e *LiteralString) String() string {
  return e.value
}

func (e *LiteralString) Value() interface{} {
  return e.value
}

func (e *LiteralString) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *LiteralString) ToLiteralValue() LiteralValue {
  return e
}

func (e *LiteralString) ToPN() PN { return Literal(e.Value()) }

func (e *LiteralUndef) Value() interface{} {
  return nil
}

func (e *LiteralUndef) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *LiteralUndef) ToLiteralValue() LiteralValue {
  return e
}

func (e *LiteralUndef) ToPN() PN { return StringValue(`undef`) }

func (e *MatchExpression) Operator() string {
  return e.operator
}

func (e *MatchExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *MatchExpression) ToPN() PN { return e.binaryOp(e.Operator()) }

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

func (e *NamedAccessExpression) ToPN() PN { return e.binaryOp(`.`) }

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

func (e *NodeDefinition) ToDefinition() Definition {
  return e
}

func (e *NodeDefinition) ToPN() PN {
  entries := make([]Entry, 0, 4)
  entries = append(entries, NamedArray(`matches`, pnElems(e.HostMatches())))
  if e.Parent() != nil {
    entries = append(entries, NamedValue(`parent`, e.Parent().ToPN()))
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().WithName(`body`))
  }
  return NamedValue(`node`, Hash(entries))
}

func (e *Nop) IsNop() bool { return true }

func (e *Nop) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *Nop) ToPN() PN { return StringValue(`nop`) }

func (e *NotExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *NotExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *NotExpression) ToPN() PN { return NamedValue(`!`, e.Expr().ToPN()) }

func (e *OrExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *OrExpression) ToBooleanExpression() BooleanExpression {
  return e
}

func (e *OrExpression) ToPN() PN { return e.binaryOp(`or`) }

func (e *Parameter) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.typeExpr, e.value)
}

func (e *Parameter) CapturesRest() bool {
  return e.capturesRest
}

func (e *Parameter) Name() string {
  return e.name
}

func (e *Parameter) ToPN() PN {
  entries := make([]Entry, 0, 3)
  entries = append(entries, NamedString(`name`, e.Name()))
  if e.Type() != nil {
    entries = append(entries, NamedValue(`type`, e.Type().ToPN()))
  }
  if e.CapturesRest() {
    entries = append(entries, NamedValue(`splat`, Literal(true)))
  }
  if e.Value() != nil {
    entries = append(entries, NamedValue(`value`, e.Value().ToPN()))
  }
  return Hash(entries)
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

func (e *ParenthesizedExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *ParenthesizedExpression) ToPN() PN { return NamedValue(`()`, e.Expr().ToPN()) }

func (e *Program) Definitions() []Expression {
  return e.definitions
}

func (e *Program) Body() Expression {
  return e.body
}

func (e *Program) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.body)
}

func (e *Program) ToPN() PN { return e.Body().ToPN() }

func (e *qRefDefinition) Name() string {
  return e.name
}

func (e *QualifiedName) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *QualifiedName) Name() string {
  return e.name
}

func (e *QualifiedName) ToPN() PN { return NamedString(`qn`, e.Name()) }

func (e *QualifiedName) Value() interface{} {
  return e.name
}

func (e *QualifiedName) ToLiteralValue() LiteralValue {
  return e
}

func (e *QualifiedReference) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *QualifiedReference) DowncasedName() string {
  if e.downcasedName == "" {
    e.downcasedName = ToLower(e.name)
  }
  return e.downcasedName
}

func (e *QualifiedReference) Name() string {
  return e.name
}

func (e *QualifiedReference) ToPN() PN { return NamedString(`qr`, e.Name()) }

func (e *QueryExpression) Expr() Expression {
  return e.expr
}

func (e *QueryExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *QueryExpression) ToPN() PN { return NamedValue(`query`, e.Expr().ToPN()) }

func (e *RegexpExpression) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *RegexpExpression) Value() interface{} {
  return e.value
}

func (e *RegexpExpression) ToLiteralValue() LiteralValue {
  return e
}

func (e *RegexpExpression) ToPN() PN { return NamedValue(`regexp`, Literal(e.Value())) }

func (e *RelationshipExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *RelationshipExpression) Operator() string {
  return e.operator
}

func (e *RelationshipExpression) ToPN() PN { return e.binaryOp(e.Operator()) }

func (e *RenderExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *RenderExpression) ToPN() PN { return NamedValue(`render`, e.Expr().ToPN()) }

func (e *RenderExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *RenderStringExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor)
}

func (e *RenderStringExpression) ToPN() PN { return NamedValue(`render-s`, Literal(e.Value())) }

func (e *ReservedWord) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *ReservedWord) ToPN() PN { return NamedString(`reserved`, e.Name()) }

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

func (e *ResourceBody) ToPN() PN {
  return Hash([]Entry{
    NamedValue(`title`, e.Title().ToPN()),
    NamedArray(`ops`, pnElems(e.Operations()))})
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

func (e *ResourceDefaultsExpression) ToPN() PN {
  entries := make([]Entry, 0, 3)
  if e.Form() != `regular` {
    entries = append(entries, NamedString(`form`, e.Form()))
  }
  entries = append(entries, NamedValue(`type`, e.TypeRef().ToPN()))
  entries = append(entries, NamedArray(`ops`, pnElems(e.Operations())))
  return NamedValue(`resource-defaults`, Hash(entries))
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

func (e *ResourceExpression) ToPN() PN {
  entries := make([]Entry, 0, 3)
  if e.Form() != `regular` {
    entries = append(entries, NamedString(`form`, e.Form()))
  }
  entries = append(entries, NamedValue(`type`, e.TypeName().ToPN()))
  entries = append(entries, NamedArray(`bodies`, pnElems(e.Bodies())))
  return NamedValue(`resource`, Hash(entries))
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

func (e *ResourceOverrideExpression) ToPN() PN {
  entries := make([]Entry, 0, 3)
  if e.Form() != `regular` {
    entries = append(entries, NamedString(`form`, e.Form()))
  }
  entries = append(entries, NamedValue(`resources`, e.Resources().ToPN()))
  entries = append(entries, NamedArray(`ops`, pnElems(e.Operations())))
  return NamedValue(`resource-override`, Hash(entries))
}

func (e *ResourceTypeDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *ResourceTypeDefinition) ToDefinition() Definition {
  return e
}

func (e *ResourceTypeDefinition) ToPN() PN { return e.pnNamedDefinition(`define`) }

func (e *SelectorEntry) Matching() Expression {
  return e.matching
}

func (e *SelectorEntry) Value() Expression {
  return e.value
}

func (e *SelectorEntry) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.matching, e.value)
}

func (e *SelectorEntry) ToPN() PN { return NamedArray(`=>`, pnArgs(e.Matching(), e.Value())) }

func (e *SelectorExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.selectors)
}

func (e *SelectorExpression) Lhs() Expression {
  return e.lhs
}

func (e *SelectorExpression) Selectors() []Expression {
  return e.selectors
}

func (e *SelectorExpression) ToPN() PN { return NamedArray(`?`, []PN{e.Lhs().ToPN(), Array(pnElems(e.Selectors()))}) }

func (e *SiteDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.body)
}

func (e *SiteDefinition) Body() Expression {
  return e.body
}

func (e *SiteDefinition) ToDefinition() Definition {
  return e
}

func (e *SiteDefinition) ToPN() PN {
  return NamedValue(`site`, e.Body().ToPN())
}

func (e *TextExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *TextExpression) ToPN() PN { return NamedValue(`str`, e.Expr().ToPN()) }

func (e *TextExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *TypeAlias) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.typeExpr)
}

func (e *TypeAlias) ToDefinition() Definition {
  return e
}

func (e *TypeAlias) ToPN() PN { return NamedArray(`type-alias`, []PN{StringValue(e.Name()), e.Type().ToPN()}) }

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

func (e *TypeDefinition) ToDefinition() Definition {
  return e
}

func (e *TypeDefinition) ToPN() PN { return NamedArray(`type-definition`, []PN{StringValue(e.Name()), StringValue(e.Parent()), e.Body().ToPN()}) }

func (e *TypeMapping) Type() Expression {
  return e.typeExpr
}

func (e *TypeMapping) Mapping() Expression {
  return e.mappingExpr
}

func (e *TypeMapping) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.typeExpr, e.mappingExpr)
}

func (e *TypeMapping) ToDefinition() Definition {
  return e
}

func (e *TypeMapping) ToPN() PN { return NamedArray(`type-mapping`, pnArgs(e.Type(), e.Mapping())) }

func (e *unaryExpression) Expr() Expression {
  return e.expr
}

func (e *UnaryMinusExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *UnaryMinusExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *UnaryMinusExpression) ToPN() PN { return NamedValue(`-`, e.Expr().ToPN()) }

func (e *UnfoldExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *UnfoldExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *UnfoldExpression) ToPN() PN { return NamedValue(`unfold`, e.Expr().ToPN()) }

func (e *UnlessExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.test, e.then, e.elseExpr)
}

func (e *UnlessExpression) ToPN() PN { return e.pnIf(`unless`) }

func (e *VariableExpression) Name() string {
  return e.expr.(*QualifiedName).name
}

func (e *VariableExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *VariableExpression) ToPN() PN { return NamedString(`$`, e.Expr().(*QualifiedName).Name()) }

func (e *VariableExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *VirtualQuery) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *VirtualQuery) ToPN() PN {
  if e.Expr().IsNop() {
    return StringValue(`<| |>`)
  }
  return NamedValue(`<| |>`, e.Expr().ToPN())
}

func (e *IfExpression) pnIf(name string) PN {
  entries := make([]Entry, 0, 3)
  entries = append(entries, NamedValue(`test`, e.Test().ToPN()))
  if !e.Then().IsNop() {
    entries = append(entries, e.Then().ToPN().WithName(`then`))
  }
  if !e.Else().IsNop() {
    entries = append(entries, e.Else().ToPN().WithName(`else`))
  }
  return NamedValue(name, Hash(entries))
}

func (e *namedDefinition) pnNamedDefinition(name string) PN {
  return NamedValue(name, Hash([]Entry{
    NamedString(`name`, e.Name()),
    NamedArray(`params`, pnElems(e.Parameters())),
    e.Body().ToPN().WithName(`body`)}))
}

func (e *binaryExpression) binaryOp(op string) PN { return NamedArray(op, pnArgs(e.Lhs(), e.Rhs())) }

func pnElems(elements []Expression) []PN {
  result := make([]PN, len(elements))
  for idx, element := range elements {
    result[idx] = element.ToPN()
  }
  return result
}

func pnArgs(elements ...Expression) []PN {
  result := make([]PN, len(elements))
  for idx, element := range elements {
    result[idx] = element.ToPN()
  }
  return result
}
