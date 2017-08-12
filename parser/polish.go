package parser

import (
  . "fmt"
  . "bytes"
)

type (
  // Roughly Polish notation of an expression, i.e. an operand in string form
  // and Operator()s.
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
  PN interface {
    // Produces a compact, LISP like syntax. Suitable for tests.
    Format(b *Buffer)

    // Produces an object that where all values are of primitive type or
    // slices or maps. This format is suitable for output as JSON or YAML
    //
    ToData() interface{}

    // Give this PN a name, or change the name if it already has one
    withName(name string) entry
  }

  // Entry in hash
  entry interface {
    PN
    formatEntry(b *Buffer)
    key() string
    value() interface{}
  }

  hash struct {
    entries []entry
  }

  array struct {
    elements []PN
  }

  stringValue struct {
    val string
  }

  literal struct {
    val interface{}
  }

  namedArray struct {
    name string
    elements []PN
  }

  namedValue struct {
    name string
    val PN
  }

  namedStringValue struct {
    name string
    val string
  }
)

func formatElements(elements []PN, b *Buffer) {
  top := len(elements)
  if top > 0 {
    elements[0].Format(b)
    for idx := 1; idx < top; idx++ {
      b.WriteByte(' ')
      elements[idx].Format(b)
    }
  }
}

func formatEntries(elements []entry, b *Buffer) {
  top := len(elements)
  if top > 0 {
    elements[0].formatEntry(b)
    for idx := 1; idx < top; idx++ {
      b.WriteByte(' ')
      elements[idx].formatEntry(b)
    }
  }
}

func (pn *array) Format(b *Buffer) {
  b.WriteByte('[')
  formatElements(pn.elements, b)
  b.WriteByte(']')
}

func (pn *array) ToData() interface{} {
  me := make([]interface{}, 0, 1 + len(pn.elements))
  for _, op := range pn.elements {
    me = append(me, op.ToData())
  }
  return me
}

func (pn *array) withName(name string) entry {
  return &namedArray{name, pn.elements}
}

func (pn *namedArray) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  if len(pn.elements) > 0 {
    b.WriteByte(' ')
    formatElements(pn.elements, b)
  }
  b.WriteByte(')')
}

func (pn *namedArray) formatEntry(b *Buffer) {
  b.WriteByte(':')
  b.WriteString(pn.name)
  b.WriteString(` [`)
  formatElements(pn.elements, b)
  b.WriteByte(']')
}

func (pn *namedArray) ToData() interface{} {
  me := make([]interface{}, 0, len(pn.elements))
  for _, op := range pn.elements {
    me = append(me, op.ToData())
  }
  return []interface{} { pn.name, me }
}

func (pn *namedArray) withName(name string) entry {
  return &namedArray{name, pn.elements}
}

func (pn *namedArray) key() string {
  return pn.name
}

func (pn *namedArray) value() interface{} {
  de := make([]interface{}, len(pn.elements))
  for idx := range pn.elements {
    de[idx] = pn.elements[idx].ToData()
  }
  return de
}

func (pn *namedValue) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  pn.val.Format(b)
  b.WriteByte(')')
}

func (pn *namedValue) formatEntry(b *Buffer) {
  b.WriteByte(':')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  pn.val.Format(b)
}

func (pn *namedValue) ToData() interface{} {
  return []interface{} { pn.name, pn.val.ToData() }
}

func (pn *namedValue) withName(name string) entry {
  return &namedValue{name, pn.val}
}

func (pn *namedValue) key() string {
  return pn.name
}

func (pn *namedValue) value() interface{} {
  return pn.val.ToData()
}

func (pn *hash) Format(b *Buffer) {
  b.WriteByte('{')
  formatEntries(pn.entries, b)
  b.WriteByte('}')
}

func (pn *hash) ToData() interface{} {
  me := make(map[string]interface{}, len(pn.entries))
  for _, op := range pn.entries {
    entry, _ := op.(entry)
    me[entry.key()] = entry.value()
  }
  return me
}

func (pn *hash) withName(name string) entry {
  return &namedValue{name, pn}
}

func (pn *stringValue) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.val)
  b.WriteByte(')')
}

func (pn *stringValue) ToData() interface{} {
  return []interface{} { pn.val }
}

func (pn *stringValue) withName(name string) entry {
  return &namedStringValue{name, pn.val}
}

func (pn *namedStringValue) Format(b *Buffer) {
  b.WriteByte('(')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  b.WriteString(pn.val)
  b.WriteByte(')')
}

func (pn *namedStringValue) ToData() interface{} {
  return []interface{} { pn.name, pn.val }
}

func (pn *namedStringValue) formatEntry(b *Buffer) {
  b.WriteByte(':')
  b.WriteString(pn.name)
  b.WriteByte(' ')
  b.WriteString(pn.val)
}

func (pn *namedStringValue) withName(name string) entry {
  return &namedStringValue{name, pn.val}
}

func (pn *namedStringValue) key() string {
  return pn.name
}

func (pn *namedStringValue) value() interface{} {
  return pn.val
}

func (pn *literal) Format(b *Buffer) {
  switch pn.val.(type) {
  case string:
    doubleQuote(pn.val.(string), b)
  case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
    Fprintf(b, `%d`, pn.val)
  case float32, float64:
    Fprintf(b, `%g`, pn.val)
  case bool:
    Fprintf(b, `%t`, pn.val)
  default:
    Fprintf(b, `%v`, pn.val)
  }
}

func (pn *literal) ToData() interface{} {
  return pn.val
}

func (pn *literal) withName(name string) entry {
  return &namedValue{name, pn}
}

// Abstract, should not get called but needed to cast abstract struct to Expression
func (e *positioned) ToPN() PN { return &stringValue{`positioned`} }

// Concrete

func (e *AccessExpression) ToPN() PN     { return &namedArray{`[]`, append(pnArgs(e.Operand()), pnElems(e.Keys())...)}}
func (e *AndExpression) ToPN() PN        { return e.binaryOp(`and`) }
func (e *ArithmeticExpression) ToPN() PN { return e.binaryOp(e.Operator()) }
func (e *Application) ToPN() PN          { return e.pnNamedDefinition(`application`) }
func (e *AssignmentExpression) ToPN() PN { return e.binaryOp(e.Operator()) }
func (e *AttributeOperation) ToPN() PN   { return &namedArray{e.Operator(), []PN{ &stringValue{e.Name()}, e.Value().ToPN()}}}
func (e *AttributesOperation) ToPN() PN  { return &namedArray{`=>`, []PN { &stringValue{`*`}, e.Expr().ToPN()}}}
func (e *BlockExpression) ToPN() PN      { return &namedArray{`block`, pnElems(e.Statements())}}

func (e *CallMethodExpression) ToPN() PN {
  s := `invoke_method`
  if e.RvalRequired() {
    s = `call_method`
  }
  entries := []entry{ &namedValue{ `functor`, e.Functor().ToPN() }, &namedArray{`args`, pnElems(e.Arguments())}}
  if e.Lambda() != nil {
    entries = append(entries, e.Lambda().ToPN().withName(`block`))
  }
  return &namedValue{s, &hash{entries}}
}

func (e *CallNamedFunctionExpression) ToPN() PN {
  s := `invoke`
  if e.RvalRequired() {
    s = `call`
  }
  entries := []entry{ &namedValue{ `functor`, e.Functor().ToPN() }, &namedArray{`args`, pnElems(e.Arguments())}}
  if e.Lambda() != nil {
    entries = append(entries, e.Lambda().ToPN().withName(`block`))
  }
  return &namedValue{s, &hash{entries}}
}

func (e *CapabilityMapping) ToPN() PN    { return &namedArray{e.Kind(), []PN{ e.Component().ToPN(), &namedArray{ e.Capability(), pnElems(e.Mappings())}}}}
func (e *CaseExpression) ToPN() PN       { return &namedArray{`case`, pnElems(e.Options())}}

func (e *CaseOption) ToPN() PN {
  return &hash{[]entry{
    &namedArray{`when`, pnElems(e.Values())},
    &namedValue{`then`, e.Then().ToPN()}}}}

func (e *CollectExpression) ToPN() PN {
  entries := make([]entry, 0, 3)
  entries = append(entries, &namedValue{`type`, e.ResourceType().ToPN()}, &namedValue{`query`, e.Query().ToPN()})
  if len(e.Operations()) > 0 {
    entries = append(entries, &namedArray{`ops`, pnElems(e.Operations())})
  }
  return &namedValue{`collect`, &hash{entries}}
}

func (e *ComparisonExpression) ToPN() PN { return e.binaryOp(e.Operator()) }
func (e *ConcatenatedString) ToPN() PN   { return &namedArray{`concat`, pnElems(e.Segments())}}

func (e *EppExpression) ToPN() PN {
  return e.Body().ToPN().withName(`epp`)
}

func (e *ExportedQuery) ToPN() PN {
  if e.Expr().IsNop() {
    return &stringValue{`<<| |>>`}
  }
  return &namedValue{`<<| |>>`, e.Expr().ToPN()}
}

func (e *FunctionDefinition) ToPN() PN {
  entries := make([]entry, 0, 4)
  entries = append(entries, &namedStringValue{`name`, e.Name()})
  if len(e.Parameters()) > 0 {
    entries = append(entries, &namedArray{`params`, pnElems(e.Parameters())})
  }
  if e.ReturnType() != nil {
    entries = append(entries, &namedValue{`returns`, e.ReturnType().ToPN()})
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().withName(`body`))
  }
  return &namedValue{`function`, &hash{entries}}
}

func (e *HeredocExpression) ToPN() PN {
  entries := make([]entry, 0, 2)
  if e.Syntax() != `` {
    entries = append(entries, &namedStringValue{`syntax`, e.Syntax()})
  }
  entries = append(entries, &namedValue{ `text`, e.Text().ToPN()} )
  return &namedValue{`heredoc`, &hash{entries}}
}

func (e *HostClassDefinition) ToPN() PN {
  entries := make([]entry, 0, 3)
  entries = append(entries, &namedStringValue{`name`, e.Name()})
  if e.ParentClass() != `` {
    entries = append(entries, &namedStringValue{`parent`, e.ParentClass()})
  }
  if len(e.Parameters()) > 0 {
    entries = append(entries, &namedArray{`params`, pnElems(e.Parameters())})
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().withName(`body`))
  }
  return &namedValue{`class`, &hash{entries}}
}

func (e *IfExpression) ToPN() PN { return e.pnIf(`if`) }
func (e *InExpression) ToPN() PN { return e.binaryOp(`in`) }
func (e *KeyedEntry) ToPN() PN   { return &namedArray{`=>`, pnArgs(e.Key(), e.Value())}}

func (e *LambdaExpression) ToPN() PN {
  entries := make([]entry, 0, 4)
  if len(e.Parameters()) > 0 {
    entries = append(entries, &namedArray{`params`, pnElems(e.Parameters()) })
  }
  if e.ReturnType() != nil {
    entries = append(entries, &namedValue{`returns`, e.ReturnType().ToPN()})
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().withName(`body`))
  }
  return &namedValue{`lambda`, &hash{entries}}
}

func (e *LiteralBoolean) ToPN() PN        { return &literal{e.Value()}}
func (e *LiteralDefault) ToPN() PN        { return &stringValue{`default`} }
func (e *LiteralFloat) ToPN() PN          { return &literal{e.Value()}}
func (e *LiteralHash) ToPN() PN           { return &namedArray{`hash`, pnElems(e.Entries())}}
func (e *LiteralInteger) ToPN() PN        { return &literal{e.Value()}}
func (e *LiteralList) ToPN() PN           { return &namedArray{`array`, pnElems(e.Elements())}}
func (e *LiteralString) ToPN() PN         { return &literal{e.Value()}}
func (e *LiteralUndef) ToPN() PN          { return &stringValue{`undef`}}
func (e *MatchExpression) ToPN() PN       { return e.binaryOp(e.Operator()) }
func (e *NamedAccessExpression) ToPN() PN { return e.binaryOp(`.`) }

func (e *NodeDefinition) ToPN() PN {
  entries := make([]entry, 0, 4)
  entries = append(entries, &namedArray{`matches`, pnElems(e.HostMatches())})
  if e.Parent() != nil {
    entries = append(entries, &namedValue{`parent`, e.Parent().ToPN() })
  }
  if e.Body() != nil {
    entries = append(entries, e.Body().ToPN().withName(`body`))
  }
  return &namedValue{`node`, &hash{entries}}
}

func (e *Nop) ToPN() PN           { return &stringValue{`nop`}}
func (e *NotExpression) ToPN() PN { return &namedValue{`!`, e.Expr().ToPN() }}
func (e *OrExpression) ToPN() PN  { return e.binaryOp(`or`) }

func (e *Parameter) ToPN() PN {
  entries := make([]entry, 0, 3)
  entries = append(entries, &namedStringValue{`name`, e.Name()})
  if e.Type() != nil {
    entries = append(entries, &namedValue{`type`, e.Type().ToPN()})
  }
  if e.CapturesRest() {
    entries = append(entries, &namedValue{`splat`, &literal{true}})
  }
  if e.Value() != nil {
    entries = append(entries, &namedValue{`value`, e.Value().ToPN() })
  }
  return &hash{entries}
}

func (e *Program) ToPN() PN                { return e.Body().ToPN() }
func (e *QualifiedName) ToPN() PN          { return &namedStringValue{`qn`, e.Name()}}
func (e *QualifiedReference) ToPN() PN     { return &namedStringValue{`qr`, e.Name()}}
func (e *QueryExpression) ToPN() PN        { return &namedValue{`query`, e.Expr().ToPN()}}
func (e *RelationshipExpression) ToPN() PN { return e.binaryOp(e.Operator()) }
func (e *RenderExpression) ToPN() PN       { return &namedValue{`render`, e.Expr().ToPN()}}
func (e *RenderStringExpression) ToPN() PN { return &namedValue{`render-s`, &literal{e.Value()}}}
func (e *RegexpExpression) ToPN() PN       { return &namedValue{`regexp`, &literal{e.Value()} }}
func (e *ReservedWord) ToPN() PN           { return &namedStringValue{`reserved`, e.Name() }}

func (e *ResourceBody) ToPN() PN {
  return &hash{[]entry{
    &namedValue{`title`, e.Title().ToPN()},
    &namedArray{`ops`, pnElems(e.Operations())}}}
}

func (e *ResourceDefaultsExpression) ToPN() PN {
  entries := make([]entry, 0, 3)
  if e.Form() != `regular` {
    entries = append(entries, &namedStringValue{`form`, e.Form()})
  }
  entries = append(entries, &namedValue{`type`, e.TypeRef().ToPN()} )
  entries = append(entries, &namedArray{`ops`, pnElems(e.Operations())})
  return &namedValue{`resource-defaults`, &hash{entries}}
}

func (e *ResourceExpression) ToPN() PN {
  entries := make([]entry, 0, 3)
  if e.Form() != `regular` {
    entries = append(entries, &namedStringValue{`form`, e.Form()})
  }
  entries = append(entries, &namedValue{`type`, e.TypeName().ToPN()})
  entries = append(entries, &namedArray{`bodies`, pnElems(e.Bodies())})
  return &namedValue{`resource`, &hash{entries}}
}

func (e *ResourceOverrideExpression) ToPN() PN {
  entries := make([]entry, 0, 3)
  if e.Form() != `regular` {
    entries = append(entries, &namedStringValue{`form`, e.Form()})
  }
  entries = append(entries, &namedValue{`resources`, e.Resources().ToPN()})
  entries = append(entries, &namedArray{`ops`, pnElems(e.Operations())})
  return &namedValue{`resource-override`, &hash{entries}}
}

func (e *ResourceTypeDefinition) ToPN() PN          { return e.pnNamedDefinition(`define`) }

func (e *SelectorEntry) ToPN() PN { return &namedArray{`=>`, pnArgs(e.Matching(), e.Value())}}

func (e *SelectorExpression) ToPN() PN { return &namedArray{`?`, []PN {e.Lhs().ToPN(), &array{pnElems(e.Selectors())}}}}
func (e *SiteDefinition) ToPN() PN {
  return &namedValue{`site`, e.Body().ToPN()}
}

func (e *TextExpression) ToPN() PN       { return &namedValue{`str`, e.Expr().ToPN()}}
func (e *TypeAlias) ToPN() PN            { return &namedArray{`type-alias`, []PN{ &stringValue{e.Name()}, e.Type().ToPN()}}}
func (e *TypeDefinition) ToPN() PN       { return &namedArray{`type-definition`, []PN{ &stringValue{e.Name()}, &stringValue{e.Parent()}, e.Body().ToPN() }}}
func (e *TypeMapping) ToPN() PN          { return &namedArray{`type-mapping`, pnArgs(e.Type(), e.Mapping())}}
func (e *UnaryMinusExpression) ToPN() PN { return &namedValue{`-`, e.Expr().ToPN()}}
func (e *UnfoldExpression) ToPN() PN     { return &namedValue{`unfold`, e.Expr().ToPN()}}
func (e *UnlessExpression) ToPN() PN     { return e.pnIf(`unless`) }
func (e *VariableExpression) ToPN() PN   { return &namedStringValue{`$`, e.Expr().(*QualifiedName).Name()}}

func (e *VirtualQuery) ToPN() PN {
  if e.Expr().IsNop() {
    return &stringValue{`<| |>`}
  }
  return &namedValue{`<| |>`, e.Expr().ToPN()}
}

func (e* IfExpression) pnIf(name string) PN {
  entries := make([]entry, 0, 3)
  entries = append(entries, &namedValue{`test`, e.Test().ToPN()})
  if !e.Then().IsNop() {
    entries = append(entries, e.Then().ToPN().withName(`then`))
  }
  if !e.Else().IsNop() {
    entries = append(entries, e.Else().ToPN().withName(`else`))
  }
  return &namedValue{name, &hash{entries}}
}

func (e*namedDefinition) pnNamedDefinition(name string) PN {
  return &namedValue{name, &hash{[]entry{
    &namedStringValue{`name`, e.Name() },
    &namedArray{`params`, pnElems(e.Parameters())},
    e.Body().ToPN().withName(`body`)}}}
}

func (e *binaryExpression) binaryOp(op string) PN { return &namedArray{op, pnArgs(e.Lhs(), e.Rhs())}}

func pnElems(elements []Expression) []PN {
  result := make([]PN, len(elements))
  for idx, element := range elements {
    result[idx] = element.ToPN()
  }
  return result
}

func pnArgs(elements...Expression) []PN {
  result := make([]PN, len(elements))
  for idx, element := range elements {
    result[idx] = element.ToPN()
  }
  return result
}

func doubleQuote(str string, b *Buffer) {
  b.WriteByte('"')
  for _, c := range str {
    switch c {
    case '\t':
      b.WriteString(`\t`)
    case '\n':
      b.WriteString(`\n`)
    case '\r':
      b.WriteString(`\r`)
    case '"':
      b.WriteString(`\"`)
    case '\\':
      b.WriteString(`\\`)
    default:
      if c < 0x20 {
        Fprintf(b, `\u{%X}`, c)
      } else {
        b.WriteRune(c)
      }
    }
  }
  b.WriteByte('"')
}

