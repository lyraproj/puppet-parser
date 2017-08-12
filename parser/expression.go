package parser

import (
  . "strings"
  . "sort"
  . "unicode/utf8"
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

  Location interface {
    File() string

    Line() int

    // Position on line
    Pos() int
  }

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
    ToPN() PN

    byteLength() int

    byteOffset() int

    updateOffsetAndLength(offset int, length int)
  }

  Definition interface {
    // Marker method, to ensure unique interface
    ToDefinition() Definition
  }

  BinaryExpression interface {
    Lhs() Expression
    Rhs() Expression
  }

  BooleanExpression interface {
    BinaryExpression

    // Marker method, to ensure unique interface
    ToBooleanExpression() BooleanExpression
  }

  UnaryExpression interface {
    Expr() Expression

    // Marker method, to ensure unique interface
    ToUnaryExpression() UnaryExpression
  }

  NameExpression interface {
    Expression
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
    keys []Expression
  }

  AndExpression struct {
    booleanExpression
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
    name string
    value Expression
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
    definitionExpression
    kind string
    capability string
    component Expression
    mappings []Expression
  }

  CaseExpression struct {
    positioned
    test Expression
    options []Expression
  }

  CaseOption struct {
    positioned
    values []Expression
    then Expression
  }

  CollectExpression struct {
    positioned
    resourceType Expression
    query Expression
    operations []Expression
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
    body Expression
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
    text Expression
  }

  HostClassDefinition struct {
    namedDefinition
    parentClass string
  }

  IfExpression struct {
    positioned
    test Expression
    then Expression
    elseExpr Expression
  }

  InExpression struct {
    binaryExpression
  }

  KeyedEntry struct {
    positioned
    key Expression
    value Expression
  }

  LambdaExpression struct {
    positioned
    parameters []Expression
    body Expression
    returnType Expression
  }

  LiteralBoolean struct {
    literalExpression
    value bool
  }

  LiteralDefault struct {
    literalExpression
  }

  LiteralFloat struct {
    literalExpression
    value float64
  }

  LiteralHash struct {
    positioned
    entries []Expression
  }

  LiteralInteger struct {
    literalExpression
    radix int
    value int64
  }

  LiteralList struct {
    positioned
    elements []Expression
  }

  LiteralString struct {
    literalExpression
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
    definitionExpression
    parent Expression
    hostMatches []Expression
    body Expression
  }

  Nop struct {
    positioned
  }

  NotExpression struct {
    unaryExpression
  }

  OrExpression struct {
    booleanExpression
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
    body Expression
    definitions []Expression
  }

  QRefDefinition struct {
    definitionExpression
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
    literalExpression
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
    literalExpression
    word string
    future bool
  }

  ResourceBody struct {
    positioned
    title Expression
    operations []Expression
  }

  ResourceDefaultsExpression struct {
    AbstractResource
    typeRef Expression
    operations []Expression
  }

  ResourceExpression struct {
    AbstractResource
    typeName Expression
    bodies []Expression
  }

  ResourceOverrideExpression struct {
    AbstractResource
    resources Expression
    operations []Expression
  }

  ResourceTypeDefinition struct {
    namedDefinition
  }

  SelectorEntry struct {
    positioned
    matching Expression
    value Expression
  }

  SelectorExpression struct {
    positioned
    lhs Expression
    selectors []Expression
  }

  SiteDefinition struct {
    definitionExpression
    body Expression
  }

  TextExpression struct {
    unaryExpression
  }

  TypeAlias struct {
    QRefDefinition
    typeExpr Expression
  }

  TypeDefinition struct {
    QRefDefinition
    parent string
    body Expression
  }

  TypeMapping struct {
    definitionExpression
    typeExpr Expression
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
  AbstractResource struct {
    positioned
    form string
  }

  binaryExpression struct {
    positioned
    lhs Expression
    rhs Expression
  }

  booleanExpression struct {
    binaryExpression
  }

  callExpression struct {
    positioned
    rvalRequired bool
    functor Expression
    arguments []Expression
    lambda Expression
  }

  definitionExpression struct {
    positioned
  }

  literalExpression struct {
    positioned
  }

  namedDefinition struct {
    definitionExpression
    name string
    parameters []Expression
    body Expression
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
  return SearchInts(e.getLineIndex(), offset + 1)
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
  line := SearchInts(li, offset + 1)
  lineStart := li[line - 1]
  if offset == lineStart {
    return 0
  }
  return RuneCountInString(e.string[lineStart:offset])
}

func (e *positioned) AllContents(path []Expression, visitor PathVisitor) {
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

func (e *definitionExpression) ToDefinition() Definition {
  return e
}

func deepVisit(e Expression, path []Expression, visitor PathVisitor, children...interface{}) {
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

func (e *AbstractResource) Form() string {
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

func (e *AndExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *Application) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *ArithmeticExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ArithmeticExpression) Operator() string {
  return e.operator
}

func (e *AssignmentExpression) Operator() string {
  return e.operator
}

func (e *AssignmentExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

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

func (e *AttributesOperation) Expr() Expression {
  return e.expr
}

func (e *AttributesOperation) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *binaryExpression) Lhs() Expression {
  return e.lhs
}

func (e *binaryExpression) Rhs() Expression {
  return e.rhs
}

func (e *booleanExpression) ToBooleanExpression() BooleanExpression {
  return e
}

func (e *BlockExpression) Statements() []Expression {
  return e.statements
}

func (e *BlockExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.statements)
}

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

func (e *callExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.functor, e.arguments, e.lambda)
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

func (e *CaseExpression) Test() Expression {
  return e.test
}

func (e *CaseExpression) Options() []Expression {
  return e.options
}

func (e *CaseExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.test, e.options)
}

func (e *CaseOption) Values() []Expression {
  return e.values
}

func (e *CaseOption) Then() Expression {
  return e.then
}

func (e *CaseOption) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.values, e.then)
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

func (e *ComparisonExpression) Operator() string {
  return e.operator
}

func (e *ComparisonExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *ConcatenatedString) Segments() []Expression {
  return e.segments
}

func (e *ConcatenatedString) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.segments)
}

func (e *EppExpression) ParametersSpecified() bool {
  return e.parametersSpecified
}

func (e *EppExpression) Body() Expression {
  return e.body
}

func (e *EppExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.body)
}

func (e *FunctionDefinition) ReturnType() Expression {
  return e.returnType
}

func (e *FunctionDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
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

func (e *HostClassDefinition) ParentClass() string {
  return e.parentClass
}

func (e *HostClassDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
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

func (e *InExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *KeyedEntry) Key() Expression {
  return e.key
}

func (e *KeyedEntry) Value() Expression {
  return e.value
}

func (e *KeyedEntry) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.key, e.value)
}

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

func (e *LiteralBoolean) Bool() bool {
  return e.value
}

func (e *LiteralBoolean) Value() interface{} {
  return e.value
}

func (e *LiteralBoolean) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *LiteralDefault) Value() interface{} {
  return DEFAULT_INSTANCE
}

func (e *LiteralDefault) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *literalExpression) ToLiteralValue() LiteralValue {
  return e
}

func (e *literalExpression) Value() interface{} {
  return nil
}

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

func (e *LiteralList) Elements() []Expression {
  return e.elements
}

func (e *LiteralList) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.elements)
}

func (e *LiteralString) String() string {
  return e.value
}

func (e *LiteralString) Value() interface{} {
  return e.value
}

func (e *LiteralString) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *LiteralUndef) Value() interface{} {
  return nil
}

func (e *LiteralUndef) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *MatchExpression) Operator() string {
  return e.operator
}

func (e *MatchExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

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

func (e *Nop) IsNop() bool { return true }

func (e *Nop) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *NotExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *OrExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *Parameter) CapturesRest() bool {
  return e.capturesRest
}

func (e *Parameter) Name() string {
  return e.name
}

func (e *Parameter) Type() Expression {
  return e.typeExpr
}

func (e *Parameter) Value() Expression {
  return e.value
}

func (e *Parameter) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.typeExpr, e.value)
}

func (e *Program) Definitions() []Expression {
  return e.definitions
}

func (e *Program) Body() Expression {
  return e.body
}

func (e *Program) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.body)
}

func (e *QRefDefinition) Name() string {
  return e.name
}

func (e *QualifiedName) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *QualifiedName) Name() string {
  return e.name
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

func (e *QueryExpression) Expr() Expression {
  return e.expr
}

func (e *QueryExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *RegexpExpression) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *RegexpExpression) Value() interface{} {
  return e.value
}

func (e *RelationshipExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.rhs)
}

func (e *RelationshipExpression) Operator() string {
  return e.operator
}

func (e *ReservedWord) AllContents(path []Expression, visitor PathVisitor) {
  visitor(path, e)
}

func (e *ReservedWord) Name() string {
  return e.word
}

func (e *ReservedWord) Future() bool {
  return e.future
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

func (e *ResourceDefaultsExpression) TypeRef() Expression {
  return e.typeRef
}

func (e *ResourceDefaultsExpression) Operations() []Expression {
  return e.operations
}

func (e *ResourceDefaultsExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.typeRef, e.operations)
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

func (e *ResourceOverrideExpression) Resources() Expression {
  return e.resources
}

func (e *ResourceOverrideExpression) Operations() []Expression {
  return e.operations
}

func (e *ResourceOverrideExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.resources, e.operations)
}

func (e *ResourceTypeDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.parameters, e.body)
}

func (e *SelectorEntry) Matching() Expression {
  return e.matching
}

func (e *SelectorEntry) Value() Expression {
  return e.value
}

func (e *SelectorEntry) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.matching, e.value)
}

func (e *SelectorExpression) Lhs() Expression {
  return e.lhs
}

func (e *SelectorExpression) Selectors() []Expression {
  return e.selectors
}

func (e *SelectorExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.lhs, e.selectors)
}

func (e *SiteDefinition) Body() Expression {
  return e.body
}

func (e *SiteDefinition) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.body)
}

func (e *TypeAlias) Type() Expression {
  return e.typeExpr
}

func (e *TypeAlias) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.typeExpr)
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

func (e *TypeMapping) Type() Expression {
  return e.typeExpr
}

func (e *TypeMapping) Mapping() Expression {
  return e.mappingExpr
}

func (e *TypeMapping) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.typeExpr, e.mappingExpr)
}

func (e *unaryExpression) Expr() Expression {
  return e.expr
}

func (e *unaryExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *unaryExpression) ToUnaryExpression() UnaryExpression {
  return e
}

func (e *UnfoldExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}

func (e *VariableExpression) Name() string {
  return e.expr.(*QualifiedName).name
}

func (e *VariableExpression) AllContents(path []Expression, visitor PathVisitor) {
  deepVisit(e, path, visitor, e.expr)
}
