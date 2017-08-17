package parser

type ExpressionFactory interface {
  Access(operand Expression, keys []Expression, locator *Locator, offset int, length int) Expression
  And(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  Application(name string, params []Expression, body Expression, locator *Locator, offset int, length int) Expression
  Array(expressions []Expression, locator *Locator, offset int, length int) Expression
  Arithmetic(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  Assignment(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  AttributeOp(op string, name string, value Expression, locator *Locator, offset int, length int) Expression
  AttributesOp(valueExpr Expression, locator *Locator, offset int, length int) Expression
  Block(expressions []Expression, locator *Locator, offset int, length int) Expression
  Boolean(value bool, locator *Locator, offset int, length int) Expression
  CallMethod(functorExpr Expression, args []Expression, lambda Expression, locator *Locator, offset int, length int) Expression
  CallNamed(functorExpr Expression, rvalRequired bool, args []Expression, lambda Expression, locator *Locator, offset int, length int) Expression
  CapabilityMapping(kind string, component Expression, capability string, mappings []Expression, locator *Locator, offset int, length int) Expression
  Case(test Expression, options []Expression, locator *Locator, offset int, length int) Expression
  Class(name string, parameters []Expression, parent string, body Expression, locator *Locator, offset int, length int) Expression
  Collect(resourceType Expression, query Expression, operations []Expression, locator *Locator, offset int, length int) Expression
  Comparison(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  ConcatenatedString(segments []Expression, locator *Locator, offset int, length int) Expression
  Default(locator *Locator, offset int, length int) Expression
  Definition(name string, params []Expression, body Expression, locator *Locator, offset int, length int) Expression
  EppExpression(params []Expression, body Expression, locator *Locator, offset int, length int) Expression
  ExportedQuery(queryExpr Expression, locator *Locator, offset int, length int) Expression
  Float(value float64, locator *Locator, offset int, length int) Expression
  Function(name string, parameters []Expression, body Expression, returnType Expression, locator *Locator, offset int, length int) Expression
  Hash(entries []Expression, locator *Locator, offset int, length int) Expression
  Heredoc(text Expression, syntax string, locator *Locator, offset int, length int) Expression
  If(condition Expression, thenPart Expression, elsePart Expression, locator *Locator, offset int, length int) Expression
  In(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  Integer(value int64, radix int, locator *Locator, offset int, length int) Expression
  KeyedEntry(key Expression, value Expression, locator *Locator, offset int, length int) Expression
  Lambda(parameters []Expression, body Expression, returnType Expression, locator *Locator, offset int, length int) Expression
  Match(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  NamedAccess(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  Negate(expr Expression, locator *Locator, offset int, length int) Expression
  Node(hostnames []Expression, parent Expression, statements Expression, locator *Locator, offset int, length int) Expression
  Nop(locator *Locator, offset int, length int) Expression
  Not(expr Expression, locator *Locator, offset int, length int) Expression
  Or(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  Parameter(name string, expr Expression, typeExpr Expression, capturesRest bool, locator *Locator, offset int, length int) Expression
  Parenthesized(expr Expression, locator *Locator, offset int, length int) Expression
  Program(body Expression, definitions []Definition, locator *Locator, offset int, length int) Expression
  QualifiedName(name string, locator *Locator, offset int, length int) Expression
  QualifiedReference(name string, locator *Locator, offset int, length int) Expression
  Regexp(value string, locator *Locator, offset int, length int) Expression
  RelOp(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression
  RenderExpression(expr Expression, locator *Locator, offset int, length int) Expression
  RenderString(text string, locator *Locator, offset int, length int) Expression
  ReservedWord(value string, future bool, locator *Locator, offset int, length int) Expression
  Resource(form string, typeName Expression, bodies []Expression, locator *Locator, offset int, length int) Expression
  ResourceBody(title Expression, operations []Expression, locator *Locator, offset int, length int) Expression
  ResourceDefaults(form string, typeRef Expression, operations []Expression, locator *Locator, offset int, length int) Expression
  ResourceOverride(form string, resources Expression, operations []Expression, locator *Locator, offset int, length int) Expression
  Select(rval Expression, entries []Expression, locator *Locator, offset int, length int) Expression
  Selector(key Expression, value Expression, locator *Locator, offset int, length int) Expression
  Site(statements Expression, locator *Locator, offset int, length int) Expression
  String(value string, locator *Locator, offset int, length int) Expression
  Text(expr Expression, locator *Locator, offset int, length int) Expression
  TypeAlias(name string, typeExpr Expression, locator *Locator, offset int, length int) Expression
  TypeDefinition(name string, parent string, body Expression, locator *Locator, offset int, length int) Expression
  TypeMapping(typeExpr Expression, mapping Expression, locator *Locator, offset int, length int) Expression
  Undef(locator *Locator, offset int, length int) Expression
  Unfold(expr Expression, locator *Locator, offset int, length int) Expression
  Unless(condition Expression, thenPart Expression, elsePart Expression, locator *Locator, offset int, length int) Expression
  Variable(expr Expression, locator *Locator, offset int, length int) Expression
  VirtualQuery(queryExpr Expression, locator *Locator, offset int, length int) Expression
  When(values []Expression, thenExpr Expression, locator *Locator, offset int, length int) Expression
}

type defaultExpressionFactory struct {
}

func DefaultFactory() ExpressionFactory {
  return &defaultExpressionFactory{}
}

func (f *defaultExpressionFactory) And(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &AndExpression{ binaryExpression{positioned{locator, offset, length}, lhs, rhs }}
}

func (f *defaultExpressionFactory) Access(operand Expression, keys []Expression, locator *Locator, offset int, length int) Expression {
  return &AccessExpression{positioned{locator, offset, length}, operand, keys }
}

func (f *defaultExpressionFactory) Application(name string, params []Expression, body Expression, locator *Locator, offset int, length int) Expression {
  return &Application{namedDefinition{positioned{locator, offset, length}, name, params, body}}
}

func (f *defaultExpressionFactory) Arithmetic(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &ArithmeticExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }, op }
}

func (f *defaultExpressionFactory) Array(expressions []Expression, locator *Locator, offset int, length int) Expression {
  return &LiteralList{positioned{locator, offset, length}, expressions }
}

func (f *defaultExpressionFactory) Assignment(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &AssignmentExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }, op }
}

func (f *defaultExpressionFactory) AttributeOp(op string, name string, value Expression, locator *Locator, offset int, length int) Expression {
  return &AttributeOperation{positioned{locator, offset, length}, op, name, value}
}

func (f *defaultExpressionFactory) AttributesOp(valueExpr Expression, locator *Locator, offset int, length int) Expression {
  return &AttributesOperation{positioned{locator, offset, length}, valueExpr}
}

func (f *defaultExpressionFactory) Block(expressions []Expression, locator *Locator, offset int, length int) Expression {
  return &BlockExpression{positioned{locator, offset, length}, expressions }
}

func (f *defaultExpressionFactory) Boolean(value bool, locator *Locator, offset int, length int) Expression {
  return &LiteralBoolean{positioned{locator, offset, length}, value }
}

func (f *defaultExpressionFactory) CallMethod(functorExpr Expression, args []Expression, lambda Expression, locator *Locator, offset int, length int) Expression {
  return &CallMethodExpression{callExpression{positioned{locator, offset, length}, true, functorExpr, args, lambda}}
}

func (f *defaultExpressionFactory) CallNamed(functorExpr Expression, rvalRequired bool, args []Expression, lambda Expression, locator *Locator, offset int, length int) Expression {
  return &CallNamedFunctionExpression{callExpression{positioned{locator, offset, length}, rvalRequired, functorExpr, args, lambda}}
}

func (f *defaultExpressionFactory) CapabilityMapping(kind string, component Expression, capability string, mappings []Expression, locator *Locator, offset int, length int) Expression {
  return &CapabilityMapping{positioned{locator, offset, length}, kind, capability, component, mappings}
}

func (f *defaultExpressionFactory) Case(test Expression, options []Expression, locator *Locator, offset int, length int) Expression {
  return &CaseExpression{positioned{locator, offset, length}, test, options}
}

func (f *defaultExpressionFactory) Class(name string, parameters []Expression, parent string, body Expression, locator *Locator, offset int, length int) Expression {
  return &HostClassDefinition{namedDefinition{positioned{locator, offset, length}, name, parameters, body}, parent}
}

func (f *defaultExpressionFactory) Collect(resourceType Expression, query Expression, operations []Expression, locator *Locator, offset int, length int) Expression {
  return &CollectExpression{positioned{locator, offset, length}, resourceType, query, operations}
}

func (f *defaultExpressionFactory) Comparison(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &ComparisonExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }, op }
}

func (f *defaultExpressionFactory) ConcatenatedString(segments []Expression, locator *Locator, offset int, length int) Expression {
  return &ConcatenatedString{positioned{locator, offset, length}, segments }
}

func (f *defaultExpressionFactory) Default(locator *Locator, offset int, length int) Expression {
  return &LiteralDefault{positioned{locator, offset, length} }
}

func (f *defaultExpressionFactory) Definition(name string, params []Expression, body Expression, locator *Locator, offset int, length int) Expression {
  return &ResourceTypeDefinition{namedDefinition{positioned{locator, offset, length}, name, params, body}}
}

func (f *defaultExpressionFactory) EppExpression(params []Expression, body Expression, locator *Locator, offset int, length int) Expression {
  return f.Lambda(params, &EppExpression{positioned{locator, offset, length}, params != nil, body}, nil, locator, offset, length)
}

func (f *defaultExpressionFactory) ExportedQuery(queryExpr Expression, locator *Locator, offset int, length int) Expression {
  return &ExportedQuery{queryExpression{positioned{locator, offset, length}, queryExpr}}
}

func (f *defaultExpressionFactory) Float(value float64, locator *Locator, offset int, length int) Expression {
  return &LiteralFloat{positioned{locator, offset, length}, value }
}

func (f *defaultExpressionFactory) Function(name string, parameters []Expression, body Expression, returnType Expression, locator *Locator, offset int, length int) Expression {
  return &FunctionDefinition{namedDefinition{positioned{locator, offset, length}, name, parameters, body }, returnType}
}

func (f *defaultExpressionFactory) Heredoc(text Expression, syntax string, locator *Locator, offset int, length int) Expression {
  return &HeredocExpression{positioned{locator, offset, length}, syntax, text }
}

func (f *defaultExpressionFactory) Hash(entries []Expression, locator *Locator, offset int, length int) Expression {
  return &LiteralHash{positioned{locator, offset, length}, entries }
}

func (f *defaultExpressionFactory) If(test Expression, thenExpr Expression, elseExpr Expression, locator *Locator, offset int, length int) Expression {
  return &IfExpression{positioned{locator, offset, length}, test, thenExpr, elseExpr}
}

func (f *defaultExpressionFactory) In(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &InExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }}
}

func (f *defaultExpressionFactory) Integer(value int64, radix int, locator *Locator, offset int, length int) Expression {
  return &LiteralInteger{positioned{locator, offset, length}, radix, value }
}

func (f *defaultExpressionFactory) KeyedEntry(key Expression, value Expression, locator *Locator, offset int, length int) Expression {
  return &KeyedEntry{positioned{locator, offset, length}, key, value }
}

func (f *defaultExpressionFactory) Lambda(parameters []Expression, body Expression, returnType Expression, locator *Locator, offset int, length int) Expression {
  return &LambdaExpression{positioned{locator, offset, length}, parameters, body, returnType}
}

func (f *defaultExpressionFactory) Selector(key Expression, value Expression, locator *Locator, offset int, length int) Expression {
  return &SelectorEntry{positioned{locator, offset, length}, key, value }
}

func (f *defaultExpressionFactory) Match(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &MatchExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }, op }
}

func (f *defaultExpressionFactory) NamedAccess(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &NamedAccessExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }}
}

func (f *defaultExpressionFactory) Negate(expr Expression, locator *Locator, offset int, length int) Expression {
  return &UnaryMinusExpression{unaryExpression{positioned{locator, offset, length}, expr }}
}

func (f *defaultExpressionFactory) Node(hostMatches []Expression, parent Expression, statements Expression, locator *Locator, offset int, length int) Expression {
  return &NodeDefinition{positioned{locator, offset, length}, parent, hostMatches, statements}
}

func (f *defaultExpressionFactory) Nop(locator *Locator, offset int, length int) Expression {
  return &Nop{positioned{locator, offset, length}}
}

func (f *defaultExpressionFactory) Not(expr Expression, locator *Locator, offset int, length int) Expression {
  return &NotExpression{unaryExpression{positioned{locator, offset, length}, expr }}
}

func (f *defaultExpressionFactory) Or(lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &OrExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }}
}

func (f *defaultExpressionFactory) Parameter(name string, expr Expression, typeExpr Expression, capturesRest bool, locator *Locator, offset int, length int)  Expression {
  return &Parameter{positioned{locator, offset, length}, name, expr, typeExpr, capturesRest }
}

func (f *defaultExpressionFactory) Parenthesized(expr Expression, locator *Locator, offset int, length int) Expression {
  return &ParenthesizedExpression{unaryExpression{positioned{locator, offset, length}, expr }}
}

func (f *defaultExpressionFactory) Program(body Expression, definitions []Definition, locator *Locator, offset int, length int) Expression {
  return &Program{positioned{locator, offset, length}, body, definitions}
}

func (f *defaultExpressionFactory) QualifiedName(name string, locator *Locator, offset int, length int) Expression {
  return &QualifiedName{positioned{locator, offset, length}, name }
}

func (f *defaultExpressionFactory) QualifiedReference(name string, locator *Locator, offset int, length int) Expression {
  return &QualifiedReference{QualifiedName{positioned{locator, offset, length}, name }, ""}
}

func (f *defaultExpressionFactory) Regexp(value string, locator *Locator, offset int, length int) Expression {
  return &RegexpExpression{positioned{locator, offset, length}, value}
}

func (f *defaultExpressionFactory) RelOp(op string, lhs Expression, rhs Expression, locator *Locator, offset int, length int) Expression {
  return &RelationshipExpression{binaryExpression{positioned{locator, offset, length}, lhs, rhs }, op }
}

func (f *defaultExpressionFactory) RenderExpression(expr Expression, locator *Locator, offset int, length int) Expression {
  return &RenderExpression{unaryExpression{positioned{locator, offset, length}, expr}}
}

func (f *defaultExpressionFactory) RenderString(text string, locator *Locator, offset int, length int) Expression {
  return &RenderStringExpression{LiteralString{positioned{locator, offset, length}, text}}
}

func (f *defaultExpressionFactory) ReservedWord(value string, future bool, locator *Locator, offset int, length int) Expression {
  return &ReservedWord{positioned{locator, offset, length}, value, future}
}

func (f *defaultExpressionFactory) Resource(form string, typeName Expression, bodies []Expression, locator *Locator, offset int, length int) Expression {
  return &ResourceExpression{abstractResource{positioned{locator, offset, length}, form}, typeName, bodies}
}

func (f *defaultExpressionFactory) ResourceBody(title Expression, operations []Expression, locator *Locator, offset int, length int) Expression {
  return &ResourceBody{positioned{locator, offset, length}, title, operations}
}

func (f *defaultExpressionFactory) ResourceDefaults(form string, typeRef Expression, operations []Expression, locator *Locator, offset int, length int) Expression {
  return &ResourceDefaultsExpression{abstractResource{positioned{locator, offset, length}, form}, typeRef, operations}
}

func (f *defaultExpressionFactory) ResourceOverride(form string, resources Expression, operations []Expression, locator *Locator, offset int, length int) Expression {
  return &ResourceOverrideExpression{abstractResource{positioned{locator, offset, length},form}, resources, operations}
}

func (f *defaultExpressionFactory) Select(lhs Expression, entries []Expression, locator *Locator, offset int, length int) Expression {
  return &SelectorExpression{positioned{locator, offset, length}, lhs, entries}
}

func (f *defaultExpressionFactory) Site(statements Expression, locator *Locator, offset int, length int) Expression {
  return &SiteDefinition{positioned{locator, offset, length}, statements}
}

func (f *defaultExpressionFactory) String(value string, locator *Locator, offset int, length int) Expression {
  return &LiteralString{positioned{locator, offset, length}, value }
}

func (f *defaultExpressionFactory) Text(expr Expression, locator *Locator, offset int, length int) Expression {
  return &TextExpression{unaryExpression{positioned{locator, offset, length}, expr}}
}

func (f *defaultExpressionFactory) TypeAlias(name string, typeExpr Expression, locator *Locator, offset int, length int) Expression {
  return &TypeAlias{qRefDefinition{positioned{locator, offset, length}, name}, typeExpr}
}

func (f *defaultExpressionFactory) TypeDefinition(name string, parent string, body Expression, locator *Locator, offset int, length int) Expression {
  return &TypeDefinition{qRefDefinition{positioned{locator, offset, length}, name}, parent, body}
}

func (f *defaultExpressionFactory) TypeMapping(typeExpr Expression, mapping Expression, locator *Locator, offset int, length int) Expression {
  return &TypeMapping{positioned{locator, offset, length}, typeExpr, mapping}
}

func (f *defaultExpressionFactory) Undef(locator *Locator, offset int, length int) Expression {
  return &LiteralUndef{positioned{locator, offset, length} }
}

func (f *defaultExpressionFactory) Unfold(expr Expression, locator *Locator, offset int, length int) Expression {
  return &UnfoldExpression{unaryExpression{positioned{locator, offset, length}, expr}}
}

func (f *defaultExpressionFactory) Unless(test Expression, thenExpr Expression, elseExpr Expression, locator *Locator, offset int, length int) Expression {
  return &UnlessExpression{IfExpression{positioned{locator, offset, length}, test, thenExpr, elseExpr}}
}

func (f *defaultExpressionFactory) Variable(expr Expression, locator *Locator, offset int, length int) Expression {
  return &VariableExpression{unaryExpression{positioned{locator, offset, length}, expr }}
}

func (f *defaultExpressionFactory) VirtualQuery(queryExpr Expression, locator *Locator, offset int, length int) Expression {
  return &VirtualQuery{queryExpression{positioned{locator, offset, length}, queryExpr}}
}

func (f *defaultExpressionFactory) When(values []Expression, thenExpr Expression, locator *Locator, offset int, length int) Expression {
  return &CaseOption{positioned{locator, offset, length}, values, thenExpr}
}
