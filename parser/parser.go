package parser

import (
	"fmt"
	. "strconv"
	. "strings"

	. "github.com/puppetlabs/go-parser/issue"
)

// Recursive descent context for the Puppet language.
//
// This is actually the lexer with added functionality. Having the lexer and context being the
// same instance is very beneficial when the lexer must parse expressions (as is the case when
// it encounters double quoted strings or heredoc with interpolation).

type (
	ExpressionParser interface {
		Parse(filename string, source string, eppMode bool, singleExpression bool) (expr Expression, err error)
	}

	// For argument lists that are not within parameters
	commaSeparatedList struct {
		LiteralList
	}
)

// Set of names that will be treated as top level function calls rather than just identifiers
// when followed by a single expression that is not within parenthesis.
var statementCalls = map[string]bool{
	`require`: true,
	`realize`: true,
	`include`: true,
	`contain`: true,
	`tag`:     true,

	`debug`:   true,
	`info`:    true,
	`notice`:  true,
	`warning`: true,
	`err`:     true,

	`fail`:   true,
	`import`: true,
	`break`:  true,
	`next`:   true,
	`return`: true,
}

type Lexer interface {
	CurrentToken() int

	NextToken() int

	SyntaxError()

	TokenValue() interface{}

	TokenString() string

	AssertToken(token int)
}

type lexer struct {
	context
}

func NewSimpleLexer(filename string, source string) Lexer {
	// Essentially a lexer that has no knowledge of interpolations
	return &lexer{context{
		stringReader:          stringReader{text: source},
		factory:               nil,
		locator:               &Locator{string: source, file: filename},
		handleBacktickStrings: false,
		handleHexEscapes:      false}}
}

func (l *lexer) CurrentToken() int {
	return l.context.currentToken
}

func (l *lexer) NextToken() int {
	l.context.nextToken()
	return l.context.currentToken
}

func (l *lexer) SyntaxError() {
	panic(l.context.parseIssue2(LEX_UNEXPECTED_TOKEN, H{`token`: tokenMap[l.context.currentToken]}))
}

func (l *lexer) TokenString() string {
	return l.context.tokenString()
}

func (l *lexer) TokenValue() interface{} {
	return l.context.tokenValue
}

func (l *lexer) AssertToken(token int) {
	l.context.assertToken(token)
}

// CreatePspecParser returns a parser that is capable of lexing backticked strings and that
// will recognize \xNN escapes in double qouted strings
func CreatePspecParser() ExpressionParser {
	return &context{factory: DefaultFactory(), handleBacktickStrings: true, handleHexEscapes: true}
}

func CreateParser() ExpressionParser {
	return &context{factory: DefaultFactory(), handleBacktickStrings: false, handleHexEscapes: false}
}

// Parse the contents of the given source. The filename is optional and will be used
// in warnings and errors issued by the context.
//
// If eppMode is true, the context will treat the given source as text with embedded puppet
// expressions.
func (ctx *context) Parse(filename string, source string, eppMode bool, singleExpression bool) (expr Expression, err error) {
	ctx.stringReader = stringReader{text: source}
	ctx.locator = &Locator{string: source, file: filename}
	ctx.definitions = make([]Definition, 0, 8)
	ctx.eppMode = eppMode
	ctx.nextLineStart = -1

	expr, err = ctx.parseTopExpression(filename, source, eppMode, singleExpression)
	if err == nil && !singleExpression {
		expr = ctx.factory.Program(expr, ctx.definitions, ctx.locator, 0, ctx.Pos())
	}
	return
}

func (ctx *context) parseTopExpression(filename string, source string, eppMode bool, singleExpression bool) (expr Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(*ReportedIssue); !ok {
				if err, ok = r.(*ParseError); !ok {
					panic(r)
				}
			}
		}
	}()

	if eppMode {
		ctx.consumeEPP()

		var text string
		if ctx.currentToken == TOKEN_RENDER_STRING {
			text = ctx.tokenString()
			ctx.nextToken()
		}

		if ctx.currentToken == TOKEN_END {
			// No EPP in the source.
			te := ctx.factory.RenderString(text, ctx.locator, 0, ctx.Pos())
			expr = ctx.factory.Block([]Expression{te}, ctx.locator, 0, ctx.Pos())
			return
		}

		if ctx.currentToken == TOKEN_PIPE {
			if text != `` {
				panic(ctx.parseIssue(PARSE_ILLEGAL_EPP_PARAMETERS))
			}
			eppParams := ctx.lambdaParameterList()
			expr = ctx.factory.Block([]Expression{
				ctx.factory.EppExpression(eppParams, ctx.parse(TOKEN_END, false), ctx.locator, 0, ctx.Pos())}, ctx.locator, 0, ctx.Pos())
			return
		}

		expressions := make([]Expression, 0, 10)
		if text != `` {
			expressions = append(expressions, ctx.factory.RenderString(text, ctx.locator, 0, ctx.tokenStartPos))
		}

		for {
			if ctx.currentToken == TOKEN_END {
				expr = ctx.factory.Block(ctx.transformCalls(expressions, 0), ctx.locator, 0, ctx.Pos())
				return
			}
			expressions = append(expressions, ctx.expression())
		}
	}

	ctx.nextToken()
	expr = ctx.parse(TOKEN_END, singleExpression)
	return
}

func (ctx *context) parse(expectedEnd int, singleExpression bool) (expr Expression) {
	_, start := ctx.skipWhite(false)
	ctx.SetPos(start)
	if singleExpression {
		if ctx.currentToken == expectedEnd {
			expr = ctx.factory.Undef(ctx.locator, start, 0)
		} else {
			expr = ctx.assignment()
			ctx.assertToken(expectedEnd)
		}
		return
	}

	expressions := make([]Expression, 0, 10)
	for ctx.currentToken != expectedEnd {
		expressions = append(expressions, ctx.syntacticStatement())
		if ctx.currentToken == TOKEN_SEMICOLON {
			ctx.nextToken()
		}
	}
	expr = ctx.factory.Block(ctx.transformCalls(expressions, start), ctx.locator, start, ctx.Pos()-start)
	return
}

func (ctx *context) assertToken(token int) {
	if ctx.currentToken != token {
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue2(PARSE_EXPECTED_TOKEN, H{`expected`: tokenMap[token], `actual`: tokenMap[ctx.currentToken]}))
	}
}

func (ctx *context) tokenString() string {
	if ctx.tokenValue == nil {
		return tokenMap[ctx.currentToken]
	}
	if str, ok := ctx.tokenValue.(string); ok {
		return str
	}
	panic(fmt.Sprintf("Token '%s' has no string representation", tokenMap[ctx.currentToken]))
}

// Iterates all statements in a block and transforms qualified names that names a "statement call" and are followed
// by an argument, into a calls. I.e. `warning "some message"` is transformed into `warning("some message")`
func (ctx *context) transformCalls(exprs []Expression, start int) (result []Expression) {
	top := len(exprs)
	if top == 0 {
		return exprs
	}

	memo := exprs[0]
	result = make([]Expression, 0, top)
	idx := 1
	for ; idx < top; idx++ {
		expr := exprs[idx]
		if qname, ok := memo.(*QualifiedName); ok && statementCalls[qname.name] {
			var args []Expression
			if csList, ok := expr.(*commaSeparatedList); ok {
				args = csList.elements
			} else {
				args = []Expression{expr}
			}
			cn := ctx.factory.CallNamed(memo, false, args, nil, ctx.locator, memo.byteOffset(), (expr.byteOffset()+expr.byteLength())-memo.byteOffset())
			if cnFunc, ok := expr.(*CallNamedFunctionExpression); ok {
				cnFunc.rvalRequired = true
			}
			result = append(result, cn)
			idx++
			if idx == top {
				return
			}
			memo = exprs[idx]
		} else {
			if cnFunc, ok := memo.(*CallNamedFunctionExpression); ok {
				cnFunc.rvalRequired = false
			}
			result = append(result, memo)
			memo = expr
		}
	}
	if cnFunc, ok := memo.(*CallNamedFunctionExpression); ok {
		cnFunc.rvalRequired = false
	}
	result = append(result, memo)
	return
}

type producerFunc func(ctx *context) Expression

// producerFunc to call ctx.expression
func expression(ctx *context) Expression {
	return ctx.expression()
}

func (ctx *context) expressions(endToken int, producer producerFunc) (exprs []Expression) {
	exprs = make([]Expression, 0, 4)
	for {
		if ctx.currentToken == endToken {
			ctx.nextToken()
			return
		}
		exprs = append(exprs, producer(ctx))
		if ctx.currentToken != TOKEN_COMMA {
			if ctx.currentToken != endToken {
				ctx.SetPos(ctx.tokenStartPos)
				panic(ctx.parseIssue2(PARSE_EXPECTED_ONE_OF_TOKENS, H{
					`expected`: fmt.Sprintf(`'%s' or '%s'`, tokenMap[TOKEN_COMMA], tokenMap[endToken]),
					`actual`: tokenMap[ctx.currentToken]}))
			}
			ctx.nextToken()
			return
		}
		ctx.nextToken()
	}
}

func (ctx *context) syntacticStatement() (expr Expression) {
	var args []Expression
	expr = ctx.assignment()
	for ctx.currentToken == TOKEN_COMMA {
		ctx.nextToken()
		if args == nil {
			args = make([]Expression, 0, 2)
			args = append(args, expr)
		}
		args = append(args, ctx.assignment())
	}
	if args != nil {
		expr = &commaSeparatedList{LiteralList{positioned{ctx.locator, expr.byteOffset(), ctx.Pos() - expr.byteOffset()}, args}}
	}
	return
}

func (ctx *context) collectionEntry() (expr Expression) {
	return ctx.handleKeyword(ctx.argument)
}

func (ctx *context) argument() (expr Expression) {
	expr = ctx.assignment()
	if ctx.currentToken == TOKEN_FARROW {
		ctx.nextToken()
		value := ctx.assignment()
		expr = ctx.factory.KeyedEntry(expr, value, ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
	}
	return
}

func (ctx *context) hashEntry() (expr Expression) {
	return ctx.handleKeyword(ctx.assignment)
}

func (ctx *context) handleKeyword(next func() Expression) (expr Expression) {
	switch ctx.currentToken {
	case TOKEN_TYPE, TOKEN_FUNCTION, TOKEN_APPLICATION, TOKEN_CONSUMES, TOKEN_PRODUCES, TOKEN_SITE:
		expr = ctx.factory.QualifiedName(ctx.tokenString(), ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos)
		ctx.nextToken()
		if ctx.currentToken == TOKEN_LP {
			expr = ctx.callFunctionExpression(expr)
		}
	default:
		expr = next()
	}
	return
}

func (ctx *context) assignment() (expr Expression) {
	expr = ctx.relationship()
	for {
		switch ctx.currentToken {
		case TOKEN_ASSIGN, TOKEN_ADD_ASSIGN, TOKEN_SUBTRACT_ASSIGN:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Assignment(op, expr, ctx.assignment(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
		case TOKEN_PIPE:
			switch expr.(type) {
			case *QualifiedName:
				expr = ctx.callFunctionExpression(expr)
			case *CallMethodExpression:
				cm := expr.(*CallMethodExpression)
				if cm.lambda == nil {
					cm.lambda = ctx.lambda()
				}
			default:
				return expr
			}
		default:
			return expr
		}
	}
}

func (ctx *context) relationship() (expr Expression) {
	expr = ctx.resource()
	for {
		switch ctx.currentToken {
		case TOKEN_IN_EDGE, TOKEN_IN_EDGE_SUB, TOKEN_OUT_EDGE, TOKEN_OUT_EDGE_SUB:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.RelOp(op, expr, ctx.resource(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
		default:
			return expr
		}
	}
}

func (ctx *context) resource() (expr Expression) {
	expr = ctx.expression()
	if ctx.currentToken == TOKEN_LC {
		expr = ctx.resourceExpression(expr.byteOffset(), expr, `regular`)
	}
	return
}

func (ctx *context) expression() (expr Expression) {
	expr = ctx.selectExpression()
	for {
		switch ctx.currentToken {
		case TOKEN_PRODUCES, TOKEN_CONSUMES:
			// Must be preceded by name of class
			capToken := ctx.tokenString()
			switch expr.(type) {
			case *QualifiedName, *QualifiedReference, *ReservedWord, *AccessExpression:
				expr = ctx.capabilityMapping(expr, capToken)
			}
		default:
			if namedAccess, ok := expr.(*NamedAccessExpression); ok {
				// Transform into method call
				expr = ctx.factory.CallMethod(namedAccess, make([]Expression, 0), nil, ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
			}
		}
		break
	}
	return
}

func (ctx *context) selectExpression() (expr Expression) {
	expr = ctx.orExpression()
	for {
		switch ctx.currentToken {
		case TOKEN_QMARK:
			expr = ctx.selectorsExpression(expr)
		default:
			return
		}
	}
}

func (ctx *context) orExpression() (expr Expression) {
	expr = ctx.andExpression()
	for {
		switch ctx.currentToken {
		case TOKEN_OR:
			ctx.nextToken()
			expr = ctx.factory.Or(expr, ctx.orExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
		default:
			return
		}
	}
}

func (ctx *context) andExpression() (expr Expression) {
	expr = ctx.compareExpression()
	for {
		switch ctx.currentToken {
		case TOKEN_AND:
			ctx.nextToken()
			expr = ctx.factory.And(expr, ctx.andExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
		default:
			return
		}
	}
}

func (ctx *context) compareExpression() (expr Expression) {
	expr = ctx.equalExpression()
	for {
		switch ctx.currentToken {
		case TOKEN_LESS, TOKEN_LESS_EQUAL, TOKEN_GREATER, TOKEN_GREATER_EQUAL:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Comparison(op, expr, ctx.compareExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())

		default:
			return
		}
	}
}

func (ctx *context) equalExpression() (expr Expression) {
	expr = ctx.shiftExpression()
	for {
		t := ctx.currentToken
		switch t {
		case TOKEN_EQUAL, TOKEN_NOT_EQUAL:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Comparison(op, expr, ctx.equalExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())

		default:
			return
		}
	}
}

func (ctx *context) shiftExpression() (expr Expression) {
	expr = ctx.additiveExpression()
	for {
		t := ctx.currentToken
		switch t {
		case TOKEN_LSHIFT, TOKEN_RSHIFT:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Arithmetic(op, expr, ctx.shiftExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())

		default:
			return
		}
	}
}

func (ctx *context) additiveExpression() (expr Expression) {
	expr = ctx.multiplicativeExpression()
	for {
		t := ctx.currentToken
		switch t {
		case TOKEN_ADD, TOKEN_SUBTRACT:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Arithmetic(op, expr, ctx.additiveExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())

		default:
			return
		}
	}
}

func (ctx *context) multiplicativeExpression() (expr Expression) {
	expr = ctx.matchExpression()
	for {
		t := ctx.currentToken
		switch t {
		case TOKEN_MULTIPLY, TOKEN_DIVIDE, TOKEN_REMAINDER:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Arithmetic(op, expr, ctx.multiplicativeExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())

		default:
			return
		}
	}
}

func (ctx *context) matchExpression() (expr Expression) {
	expr = ctx.inExpression()
	for {
		t := ctx.currentToken
		switch t {
		case TOKEN_MATCH, TOKEN_NOT_MATCH:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Match(op, expr, ctx.matchExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())

		default:
			return
		}
	}
}

func (ctx *context) inExpression() (expr Expression) {
	expr = ctx.unaryExpression()
	for {
		switch ctx.currentToken {
		case TOKEN_IN:
			ctx.nextToken()
			expr = ctx.factory.In(expr, ctx.inExpression(), ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())

		default:
			return expr
		}
	}
}

func (ctx *context) arrayExpression() (elements []Expression) {
	return ctx.joinHashEntries(ctx.expressions(TOKEN_RB, collectionEntry))
}

func keyedEntry(ctx *context) Expression {
	key := ctx.hashEntry()
	if ctx.currentToken != TOKEN_FARROW {
		panic(ctx.parseIssue(PARSE_EXPECTED_FARROW_AFTER_KEY))
	}
	ctx.nextToken()
	value := ctx.hashEntry()
	return ctx.factory.KeyedEntry(key, value, ctx.locator, key.byteOffset(), ctx.Pos()-key.byteOffset())
}

func (ctx *context) hashExpression() (entries []Expression) {
	return ctx.expressions(TOKEN_RC, keyedEntry)
}

func (ctx *context) unaryExpression() Expression {
	unaryStart := ctx.tokenStartPos
	switch ctx.currentToken {
	case TOKEN_SUBTRACT:
		if c, _ := ctx.Peek(); isDecimalDigit(c) {
			ctx.nextToken()
			if ctx.currentToken == TOKEN_INTEGER {
				ctx.setTokenValue(ctx.currentToken, -ctx.tokenValue.(int64))
			} else {
				ctx.setTokenValue(ctx.currentToken, -ctx.tokenValue.(float64))
			}
			expr := ctx.primaryExpression()
			expr.updateOffsetAndLength(unaryStart, ctx.Pos()-unaryStart)
			return expr
		}
		ctx.nextToken()
		expr := ctx.primaryExpression()
		return ctx.factory.Negate(expr, ctx.locator, unaryStart, ctx.Pos()-unaryStart)

	case TOKEN_ADD:
		// Allow '+' prefix for constant numbers
		if c, _ := ctx.Peek(); isDecimalDigit(c) {
			ctx.nextToken()
			expr := ctx.primaryExpression()
			expr.updateOffsetAndLength(unaryStart, ctx.Pos()-unaryStart)
			return expr
		}
		panic(ctx.parseIssue2(LEX_UNEXPECTED_TOKEN, H{`token`: `+`}))

	case TOKEN_NOT:
		ctx.nextToken()
		expr := ctx.unaryExpression()
		return ctx.factory.Not(expr, ctx.locator, unaryStart, ctx.Pos()-unaryStart)

	case TOKEN_MULTIPLY:
		ctx.nextToken()
		expr := ctx.unaryExpression()
		return ctx.factory.Unfold(expr, ctx.locator, unaryStart, ctx.Pos()-unaryStart)

	case TOKEN_AT, TOKEN_ATAT:
		kind := `virtual`
		if ctx.currentToken == TOKEN_ATAT {
			kind = `exported`
		}
		ctx.nextToken()
		expr := ctx.primaryExpression()
		ctx.assertToken(TOKEN_LC)
		return ctx.resourceExpression(unaryStart, expr, kind)

	default:
		return ctx.primaryExpression()
	}
}

func (ctx *context) primaryExpression() (expr Expression) {
	expr = ctx.atomExpression()
	for {
		switch ctx.currentToken {
		case TOKEN_LP:
			expr = ctx.callFunctionExpression(expr)
		case TOKEN_LCOLLECT, TOKEN_LLCOLLECT:
			expr = ctx.collectExpression(expr)
		case TOKEN_LB:
			ctx.nextToken()
			params := ctx.arrayExpression()
			expr = ctx.factory.Access(expr, params, ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
		case TOKEN_DOT:
			ctx.nextToken()
			var rhs Expression
			if ctx.currentToken == TOKEN_TYPE {
				rhs = ctx.factory.QualifiedName(ctx.tokenString(), ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos)
				ctx.nextToken()
			} else {
				rhs = ctx.atomExpression()
			}
			expr = ctx.factory.NamedAccess(expr, rhs, ctx.locator, expr.byteOffset(), ctx.Pos()-expr.byteOffset())
		default:
			return
		}
	}
}

func (ctx *context) atomExpression() (expr Expression) {
	atomStart := ctx.tokenStartPos
	switch ctx.currentToken {
	case TOKEN_LP, TOKEN_WSLP:
		ctx.nextToken()
		expr = ctx.factory.Parenthesized(ctx.assignment(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.assertToken(TOKEN_RP)
		ctx.nextToken()

	case TOKEN_LB, TOKEN_LISTSTART:
		ctx.nextToken()
		expr = ctx.factory.Array(ctx.arrayExpression(), ctx.locator, atomStart, ctx.Pos()-atomStart)

	case TOKEN_LC:
		ctx.nextToken()
		expr = ctx.factory.Hash(ctx.hashExpression(), ctx.locator, atomStart, ctx.Pos()-atomStart)

	case TOKEN_BOOLEAN:
		expr = ctx.factory.Boolean(ctx.tokenValue.(bool), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_INTEGER:
		expr = ctx.factory.Integer(ctx.tokenValue.(int64), ctx.radix, ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_FLOAT:
		expr = ctx.factory.Float(ctx.tokenValue.(float64), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_STRING:
		expr = ctx.factory.String(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_ATTR, TOKEN_PRIVATE:
		expr = ctx.factory.ReservedWord(ctx.tokenString(), false, ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_DEFAULT:
		expr = ctx.factory.Default(ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_HEREDOC, TOKEN_CONCATENATED_STRING:
		expr = ctx.tokenValue.(Expression)
		ctx.nextToken()

	case TOKEN_REGEXP:
		expr = ctx.factory.Regexp(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_UNDEF:
		expr = ctx.factory.Undef(ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_TYPE_NAME:
		expr = ctx.factory.QualifiedReference(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_IDENTIFIER:
		expr = ctx.factory.QualifiedName(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_VARIABLE:
		vni := ctx.tokenValue
		ctx.nextToken()
		var name Expression
		if s, ok := vni.(string); ok {
			name = ctx.factory.QualifiedName(s, ctx.locator, atomStart+1, len(s))
		} else {
			name = ctx.factory.Integer(vni.(int64), 10, ctx.locator, atomStart+1, ctx.Pos()-(atomStart+1))
		}
		expr = ctx.factory.Variable(name, ctx.locator, atomStart, ctx.Pos()-atomStart)

	case TOKEN_CASE:
		expr = ctx.caseExpression()

	case TOKEN_IF:
		expr = ctx.ifExpression(false)

	case TOKEN_UNLESS:
		expr = ctx.ifExpression(true)

	case TOKEN_CLASS:
		name := ctx.tokenString()
		ctx.nextToken()
		if ctx.currentToken == TOKEN_LC {
			// Class resource
			expr = ctx.factory.QualifiedName(name, ctx.locator, atomStart, ctx.Pos()-atomStart)
		} else {
			expr = ctx.classExpression(atomStart)
		}

	case TOKEN_TYPE:
		// look ahead for '(' in which case this is a named function call
		name := ctx.tokenString()
		ctx.nextToken()
		if ctx.currentToken == TOKEN_TYPE_NAME {
			expr = ctx.typeAliasOrDefinition()
		} else {
			// Not a type definition. Just treat the 'type' keyword as a qualfied name
			expr = ctx.factory.QualifiedName(name, ctx.locator, atomStart, ctx.Pos()-atomStart)
		}

	case TOKEN_FUNCTION:
		expr = ctx.functionDefinition()

	case TOKEN_NODE:
		expr = ctx.nodeDefinition()

	case TOKEN_DEFINE, TOKEN_APPLICATION:
		expr = ctx.resourceDefinition(ctx.currentToken)

	case TOKEN_SITE:
		expr = ctx.siteDefinition()

	case TOKEN_RENDER_STRING:
		expr = ctx.factory.RenderString(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case TOKEN_RENDER_EXPR:
		ctx.nextToken()
		expr = ctx.factory.RenderExpression(ctx.expression(), ctx.locator, atomStart, ctx.Pos()-atomStart)

	default:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue2(LEX_UNEXPECTED_TOKEN, H{`token`: tokenMap[ctx.currentToken]}))
	}
	return
}

func (ctx *context) ifExpression(unless bool) (expr Expression) {
	start := ctx.tokenStartPos // start of if, elsif, or unless keyword
	ctx.nextToken()
	condition := ctx.orExpression()
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	thenPart := ctx.parse(TOKEN_RC, false)
	ctx.nextToken()

	var elsePart Expression
	switch ctx.currentToken {
	case TOKEN_ELSE:
		ctx.nextToken()
		ctx.assertToken(TOKEN_LC)
		ctx.nextToken()
		elsePart = ctx.parse(TOKEN_RC, false)
		ctx.nextToken()
	case TOKEN_ELSIF:
		if unless {
			panic(ctx.parseIssue(PARSE_ELSIF_IN_UNLESS))
		}
		elsePart = ctx.ifExpression(false)
	default:
		elsePart = ctx.factory.Nop(ctx.locator, ctx.tokenStartPos, 0)
	}

	if unless {
		expr = ctx.factory.Unless(condition, thenPart, elsePart, ctx.locator, start, ctx.Pos()-start)
	} else {
		expr = ctx.factory.If(condition, thenPart, elsePart, ctx.locator, start, ctx.Pos()-start)
	}
	return
}

func (ctx *context) selectorsExpression(test Expression) (expr Expression) {
	var selectors []Expression
	ctx.nextToken()
	if ctx.currentToken == TOKEN_SELC {
		ctx.nextToken()
		selectors = ctx.expressions(TOKEN_RC, selectorEntry)
	} else {
		selectors = []Expression{selectorEntry(ctx)}
	}
	return ctx.factory.Select(test, selectors, ctx.locator, test.byteOffset(), ctx.Pos()-test.byteOffset())
}

func selectorEntry(ctx *context) (expr Expression) {
	start := ctx.tokenStartPos
	lhs := ctx.expression()
	ctx.assertToken(TOKEN_FARROW)
	ctx.nextToken()
	return ctx.factory.Selector(lhs, ctx.expression(), ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) caseExpression() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	test := ctx.expression()
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	caseOptions := ctx.caseOptions()
	return ctx.factory.Case(test, caseOptions, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) caseOptions() (exprs []Expression) {
	exprs = make([]Expression, 0, 4)
	for {
		exprs = append(exprs, ctx.caseOption())
		if ctx.currentToken == TOKEN_RC {
			ctx.nextToken()
			return
		}
	}
}

func (ctx *context) caseOption() Expression {
	start := ctx.tokenStartPos
	expressions := ctx.expressions(TOKEN_COLON, expression)
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	block := ctx.parse(TOKEN_RC, false)
	ctx.nextToken()
	return ctx.factory.When(expressions, block, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) resourceExpression(start int, first Expression, form string) (expr Expression) {
	bodiesStart := ctx.Pos()
	ctx.nextToken()
	titleStart := ctx.Pos()
	firstTitle := ctx.expression()
	if ctx.currentToken != TOKEN_COLON {
		// Resource body without title
		ctx.SetPos(titleStart)
		switch ctx.resourceShape(first) {
		case `resource`:
			// This is just LHS followed by a hash. It only makes sense when LHS is an identifier equal
			// to one of the known "statement calls". For all other cases, this is an error
			fqn, ok := first.(*QualifiedName)
			name := ``
			if ok {
				name = fqn.name
				if _, ok := statementCalls[name]; ok {
					args := make([]Expression, 1)
					ctx.SetPos(bodiesStart)
					ctx.nextToken()
					args[0] = ctx.factory.Hash(ctx.hashExpression(), ctx.locator, bodiesStart, ctx.Pos()-bodiesStart)
					expr = ctx.factory.CallNamed(first, true, args, nil, ctx.locator, start, ctx.Pos()-start)
					return
				}
			}
			ctx.SetPos(start)
			panic(ctx.parseIssue2(PARSE_RESOURCE_WITHOUT_TITLE, H{`name`: name}))
		case `defaults`:
			ctx.SetPos(bodiesStart)
			ctx.nextToken()
			ops := ctx.attributeOperations()
			expr = ctx.factory.ResourceDefaults(form, first, ops, ctx.locator, start, ctx.Pos()-start)
		case `override`:
			ctx.SetPos(bodiesStart)
			ctx.nextToken()
			ops := ctx.attributeOperations()
			expr = ctx.factory.ResourceOverride(form, first, ops, ctx.locator, start, ctx.Pos()-start)
		default:
			ctx.SetPos(first.byteOffset())
			panic(ctx.parseIssue(PARSE_INVALID_RESOURCE))
		}
	} else {
		bodies := ctx.resourceBodies(firstTitle)
		expr = ctx.factory.Resource(form, first, bodies, ctx.locator, start, ctx.Pos()-start)
	}

	ctx.assertToken(TOKEN_RC)
	ctx.nextToken()
	return
}

func (ctx *context) resourceShape(expr Expression) string {
	if _, ok := expr.(*QualifiedName); ok {
		return "resource"
	}
	if _, ok := expr.(*QualifiedReference); ok {
		return "defaults"
	}
	if accessExpr, ok := expr.(*AccessExpression); ok {
		if qn, ok := accessExpr.operand.(*QualifiedReference); ok && qn.String() == `Resource` && len(accessExpr.keys) == 1 {
			return "defaults"
		}
		return "override"
	}
	return "error"
}

func (ctx *context) resourceBodies(title Expression) (result []Expression) {
	result = make([]Expression, 0, 1)
	for ctx.currentToken != TOKEN_RC {
		result = append(result, ctx.resourceBody(title))
		if ctx.currentToken != TOKEN_SEMICOLON {
			break
		}
		ctx.nextToken()
		if ctx.currentToken != TOKEN_RC {
			title = ctx.expression()
		}
	}
	return
}

func (ctx *context) resourceBody(title Expression) Expression {
	if ctx.currentToken != TOKEN_COLON {
		ctx.SetPos(title.byteOffset())
		panic(ctx.parseIssue(PARSE_EXPECTED_TITLE))
	}
	ctx.nextToken()
	ops := ctx.attributeOperations()
	return ctx.factory.ResourceBody(title, ops, ctx.locator, title.byteOffset(), ctx.Pos()-title.byteOffset())
}

func (ctx *context) attributeOperations() (result []Expression) {
	result = make([]Expression, 0, 5)
	for {
		switch ctx.currentToken {
		case TOKEN_SEMICOLON, TOKEN_RC:
			return
		default:
			result = append(result, ctx.attributeOperation())
			if ctx.currentToken != TOKEN_COMMA {
				return
			}
			ctx.nextToken()
		}
	}
}

func (ctx *context) attributeOperation() (op Expression) {
	start := ctx.tokenStartPos
	splat := ctx.currentToken == TOKEN_MULTIPLY
	if splat {
		ctx.nextToken()
		ctx.assertToken(TOKEN_FARROW)
		ctx.nextToken()
		return ctx.factory.AttributesOp(ctx.expression(), ctx.locator, start, ctx.Pos()-start)
	}

	name := ctx.attributeName()

	switch ctx.currentToken {
	case TOKEN_FARROW, TOKEN_PARROW:
		op := ctx.tokenString()
		ctx.nextToken()
		return ctx.factory.AttributeOp(op, name, ctx.expression(), ctx.locator, start, ctx.Pos()-start)
	default:
		panic(ctx.parseIssue(PARSE_INVALID_ATTRIBUTE))
	}
}

func (ctx *context) attributeName() (name string) {
	switch ctx.currentToken {
	case TOKEN_IDENTIFIER:
		name = ctx.tokenString()
		ctx.nextToken()
		return
	default:
		if word, ok := ctx.keyword(); ok {
			name = word
			ctx.nextToken()
			return
		}
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(PARSE_EXPECTED_ATTRIBUTE_NAME))
	}
}

func (ctx *context) collectExpression(lhs Expression) Expression {
	var collectQuery Expression
	queryStart := ctx.tokenStartPos
	if ctx.currentToken == TOKEN_LCOLLECT {
		ctx.nextToken()
		var queryExpr Expression
		if ctx.currentToken == TOKEN_RCOLLECT {
			queryExpr = ctx.factory.Nop(ctx.locator, ctx.tokenStartPos, 0)
		} else {
			queryExpr = ctx.expression()
			ctx.assertToken(TOKEN_RCOLLECT)
		}
		ctx.nextToken()
		collectQuery = ctx.factory.VirtualQuery(queryExpr, ctx.locator, queryStart, ctx.Pos()-queryStart)
	} else {
		ctx.nextToken()
		var queryExpr Expression
		if ctx.currentToken == TOKEN_RRCOLLECT {
			queryExpr = ctx.factory.Nop(ctx.locator, queryStart, ctx.tokenStartPos-queryStart)
		} else {
			queryExpr = ctx.expression()
			ctx.assertToken(TOKEN_RRCOLLECT)
		}
		ctx.nextToken()
		collectQuery = ctx.factory.ExportedQuery(queryExpr, ctx.locator, queryStart, ctx.Pos()-queryStart)
	}

	var attributeOps []Expression
	if ctx.currentToken != TOKEN_LC {
		attributeOps = make([]Expression, 0, 0)
	} else {
		ctx.nextToken()
		attributeOps = ctx.attributeOperations()
		ctx.assertToken(TOKEN_RC)
		ctx.nextToken()
	}
	return ctx.factory.Collect(lhs, collectQuery, attributeOps, ctx.locator, lhs.byteOffset(), ctx.Pos()-lhs.byteOffset())
}

func (ctx *context) typeAliasOrDefinition() Expression {
	start := ctx.tokenStartPos
	typeExpr := ctx.parameterType()
	fqr, ok := typeExpr.(*QualifiedReference)
	if !ok {
		if _, ok = typeExpr.(*AccessExpression); ok {
			if ctx.currentToken == TOKEN_ASSIGN {
				ctx.nextToken()
				return ctx.addDefinition(ctx.factory.TypeMapping(typeExpr, ctx.expression(), ctx.locator, start, ctx.Pos()-start))
			}
		}
		panic(ctx.parseIssue(PARSE_EXPECTED_TYPE_NAME_AFTER_TYPE))
	}

	parent := ``
	switch ctx.currentToken {
	case TOKEN_ASSIGN:
		ctx.nextToken()
		bodyStart := ctx.tokenStartPos
		body := ctx.expression()
		switch body.(type) {
		case *QualifiedReference:
			if ctx.currentToken == TOKEN_LC {
				body = ctx.factory.KeyedEntry(body, ctx.expression(),  ctx.locator, bodyStart, ctx.Pos()-bodyStart)
			}
			return ctx.addDefinition(ctx.factory.TypeAlias(fqr.name, body, ctx.locator, start, ctx.Pos()-start))
		default:
			return ctx.addDefinition(ctx.factory.TypeAlias(fqr.name, body, ctx.locator, start, ctx.Pos()-start))
		}
	case TOKEN_INHERITS:
		ctx.nextToken()
		nameExpr := ctx.typeName()
		if nameExpr == nil {
			panic(ctx.parseIssue(PARSE_INHERITS_MUST_BE_TYPE_NAME))
		}
		parent = nameExpr.(*QualifiedReference).name
		ctx.assertToken(TOKEN_LC)
		fallthrough

	case TOKEN_LC:
		ctx.nextToken()
		body := ctx.parse(TOKEN_RC, false)
		ctx.nextToken() // consume TOKEN_RC
		return ctx.addDefinition(ctx.factory.TypeDefinition(fqr.name, parent, body, ctx.locator, start, ctx.Pos()-start))

	default:
		panic(ctx.parseIssue2(LEX_UNEXPECTED_TOKEN, H{`token`: tokenMap[ctx.currentToken]}))
	}
}

func (ctx *context) callFunctionExpression(functorExpr Expression) Expression {
	var args []Expression
	if ctx.currentToken != TOKEN_PIPE {
		ctx.nextToken()
		args = ctx.arguments()
	}
	var block Expression
	if ctx.currentToken == TOKEN_PIPE {
		block = ctx.lambda()
	}
	start := functorExpr.byteOffset()
	if namedAccess, ok := functorExpr.(*NamedAccessExpression); ok {
		return ctx.factory.CallMethod(namedAccess, args, block, ctx.locator, start, ctx.Pos()-start)
	}
	return ctx.factory.CallNamed(functorExpr, true, args, block, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) lambda() (result Expression) {
	start := ctx.tokenStartPos
	parameterList := ctx.lambdaParameterList()
	var returnType Expression
	if ctx.currentToken == TOKEN_RSHIFT {
		ctx.nextToken()
		returnType = ctx.parameterType()
	}

	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	block := ctx.parse(TOKEN_RC, false)
	ctx.nextToken() // consume TOKEN_RC
	return ctx.factory.Lambda(parameterList, block, returnType, ctx.locator, start, ctx.Pos()-start)
}

func collectionEntry(ctx *context) Expression {
	return ctx.collectionEntry()
}

func argument(ctx *context) Expression {
	return ctx.argument()
}

func (ctx *context) joinHashEntries(exprs []Expression) (result []Expression) {
	// Assume that this is a no-op
	result = exprs
	for _, expr := range exprs {
		if _, ok := expr.(*KeyedEntry); ok {
			result = ctx.processHashEntries(exprs)
			break
		}
	}
	return
}

// Convert keyed entry occurrences into hashes. Adjacent entries are merged into
// one hash.
func (ctx *context) processHashEntries(exprs []Expression) (result []Expression) {
	result = make([]Expression, 0, len(exprs))
	var collector []Expression
	for _, expr := range exprs {
		if ke, ok := expr.(*KeyedEntry); ok {
			if collector == nil {
				collector = make([]Expression, 0, 8)
			}
			collector = append(collector, ke)
		} else {
			if collector != nil {
				result = append(result, ctx.newHashWithoutBraces(collector))
				collector = nil
			}
			result = append(result, expr)
		}
	}
	if collector != nil {
		result = append(result, ctx.newHashWithoutBraces(collector))
	}
	return
}

func (ctx *context) newHashWithoutBraces(entries []Expression) Expression {
	start := entries[0].byteOffset()
	last := entries[len(entries)-1]
	end := last.byteOffset() + last.byteLength()
	return ctx.factory.Hash(entries, ctx.locator, start, end-start)
}

func (ctx *context) arguments() (result []Expression) {
	return ctx.joinHashEntries(ctx.expressions(TOKEN_RP, argument))
}

func (ctx *context) functionDefinition() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	var name string
	switch ctx.currentToken {
	case TOKEN_IDENTIFIER, TOKEN_TYPE_NAME:
		name = ctx.tokenString()
	default:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(PARSE_EXPECTED_NAME_AFTER_FUNCTION))
	}
	ctx.nextToken()
	parameterList := ctx.parameterList()

	var returnType Expression
	if ctx.currentToken == TOKEN_RSHIFT {
		ctx.nextToken()
		returnType = ctx.parameterType()
	}

	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	block := ctx.parse(TOKEN_RC, false)
	ctx.nextToken() // consume TOKEN_RC
	return ctx.addDefinition(ctx.factory.Function(name, parameterList, block, returnType, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) nodeDefinition() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	hostnames := ctx.hostnames()
	var nodeParent Expression
	if ctx.currentToken == TOKEN_INHERITS {
		ctx.nextToken()
		nodeParent = ctx.hostname()
	}
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	block := ctx.parse(TOKEN_RC, false)
	ctx.nextToken()
	return ctx.addDefinition(ctx.factory.Node(hostnames, nodeParent, block, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) hostnames() (hostnames []Expression) {
	hostnames = make([]Expression, 0, 4)
	for {
		hostnames = append(hostnames, ctx.hostname())
		if ctx.currentToken != TOKEN_COMMA {
			return
		}
		ctx.nextToken()
		switch ctx.currentToken {
		case TOKEN_INHERITS, TOKEN_LC:
			return
		}
	}
}

func (ctx *context) hostname() (hostname Expression) {
	start := ctx.tokenStartPos
	switch ctx.currentToken {
	case TOKEN_IDENTIFIER, TOKEN_TYPE_NAME, TOKEN_INTEGER, TOKEN_FLOAT:
		hostname = ctx.dottedName()
	case TOKEN_REGEXP:
		hostname = ctx.factory.Regexp(ctx.tokenString(), ctx.locator, start, ctx.Pos()-start)
		ctx.nextToken()
	case TOKEN_STRING:
		hostname = ctx.factory.String(ctx.tokenString(), ctx.locator, start, ctx.Pos()-start)
		ctx.nextToken()
	case TOKEN_DEFAULT:
		hostname = ctx.factory.Default(ctx.locator, start, ctx.Pos()-start)
		ctx.nextToken()
	case TOKEN_CONCATENATED_STRING, TOKEN_HEREDOC:
		hostname = ctx.tokenValue.(Expression)
		ctx.nextToken()
	default:
		panic(ctx.parseIssue(PARSE_EXPECTED_HOSTNAME))
	}
	return
}

func (ctx *context) dottedName() Expression {
	start := ctx.tokenStartPos
	names := make([]string, 0, 8)
	for {
		switch ctx.currentToken {
		case TOKEN_IDENTIFIER, TOKEN_TYPE_NAME:
			names = append(names, ctx.tokenString())
		case TOKEN_INTEGER:
			names = append(names, FormatInt(ctx.tokenValue.(int64), 10))
		case TOKEN_FLOAT:
			names = append(names, FormatFloat(ctx.tokenValue.(float64), 'g', -1, 64))
		default:
			panic(ctx.parseIssue(PARSE_EXPECTED_NAME_OR_NUMBER_AFTER_DOT))
		}

		ctx.nextToken()
		if ctx.currentToken != TOKEN_DOT {
			return ctx.factory.String(Join(names, `.`), ctx.locator, start, ctx.Pos()-start)
		}
		ctx.nextToken()
	}
}

func (ctx *context) parameterList() (result []Expression) {
	switch ctx.currentToken {
	case TOKEN_LP, TOKEN_WSLP:
		ctx.nextToken()
		return ctx.expressions(TOKEN_RP, parameter)
	default:
		return []Expression{}
	}
}

func (ctx *context) lambdaParameterList() (result []Expression) {
	ctx.nextToken()
	return ctx.expressions(TOKEN_PIPE, parameter)
}

func parameter(ctx *context) Expression {
	var typeExpr, defaultExpression Expression

	start := ctx.tokenStartPos
	if ctx.currentToken == TOKEN_TYPE_NAME {
		typeExpr = ctx.parameterType()
	}

	capturesRest := ctx.currentToken == TOKEN_MULTIPLY
	if capturesRest {
		ctx.nextToken()
	}

	if ctx.currentToken != TOKEN_VARIABLE {
		panic(ctx.parseIssue(PARSE_EXPECTED_VARIABLE))
	}
	variable, ok := ctx.tokenValue.(string)
	if !ok {
		panic(ctx.parseIssue(PARSE_EXPECTED_VARIABLE))
	}
	ctx.nextToken()

	if ctx.currentToken == TOKEN_ASSIGN {
		ctx.nextToken()
		defaultExpression = ctx.expression()
	}
	return ctx.factory.Parameter(
		variable,
		defaultExpression, typeExpr, capturesRest, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) parameterType() Expression {
	start := ctx.tokenStartPos
	typeName := ctx.typeName()
	if typeName == nil {
		panic(ctx.parseIssue(PARSE_EXPECTED_TYPE_NAME))
	}

	if ctx.currentToken == TOKEN_LB {
		ctx.nextToken()
		typeArgs := ctx.arrayExpression()
		return ctx.factory.Access(typeName, typeArgs, ctx.locator, start, ctx.Pos()-start)
	}
	return typeName
}

func (ctx *context) typeName() Expression {
	if ctx.currentToken == TOKEN_TYPE_NAME {
		name := ctx.factory.QualifiedReference(ctx.tokenString(), ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos)
		ctx.nextToken()
		return name
	}
	return nil
}

func (ctx *context) classExpression(start int) Expression {
	name := ctx.className()
	if HasPrefix(name, `::`) {
		name = name[2:]
	}

	// Push to namestack
	ctx.nameStack = append(ctx.nameStack, name)

	params := ctx.parameterList()
	var parent string
	if ctx.currentToken == TOKEN_INHERITS {
		ctx.nextToken()
		if ctx.currentToken == TOKEN_DEFAULT {
			parent = tokenMap[TOKEN_DEFAULT]
			ctx.nextToken()
		} else {
			parent = ctx.className()
		}
	}
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	body := ctx.parse(TOKEN_RC, false)
	ctx.nextToken()

	// Pop namestack
	ctx.nameStack = ctx.nameStack[:len(ctx.nameStack)-1]
	return ctx.addDefinition(ctx.factory.Class(ctx.qualifiedName(name), params, parent, body, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) className() (name string) {
	switch ctx.currentToken {
	case TOKEN_TYPE_NAME, TOKEN_IDENTIFIER:
		name = ctx.tokenString()
		ctx.nextToken()
		return
	case TOKEN_STRING, TOKEN_CONCATENATED_STRING:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(PARSE_QUOTED_NOT_VALID_NAME))
	case TOKEN_CLASS:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(PARSE_CLASS_NOT_VALID_HERE))
	default:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(PARSE_EXPECTED_CLASS_NAME))
	}
}

func (ctx *context) keyword() (word string, ok bool) {
	if ctx.currentToken != TOKEN_BOOLEAN {
		str := tokenMap[ctx.currentToken]
		if _, ok = keywords[str]; ok {
			word = str
		}
	}
	return
}

func (ctx *context) qualifiedName(name string) string {
	return Join(append(ctx.nameStack, name), `::`)
}

func (ctx *context) capabilityMapping(component Expression, kind string) Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	capName := ctx.className()
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	mappings := ctx.attributeOperations()
	ctx.assertToken(TOKEN_RC)
	ctx.nextToken()

	switch component.(type) {
	case *QualifiedReference, *QualifiedName:
		// No action
	case *ReservedWord:
		// All reserved words are lowercase only
		component = ctx.factory.QualifiedName(ctx.qualifiedName(component.(*ReservedWord).Name()), ctx.locator, component.byteOffset(), component.byteLength())
	}
	return ctx.addDefinition(ctx.factory.CapabilityMapping(kind, component, ctx.qualifiedName(capName), mappings, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) siteDefinition() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	block := ctx.parse(TOKEN_RC, false)
	ctx.nextToken()
	return ctx.addDefinition(ctx.factory.Site(block, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) resourceDefinition(resourceToken int) Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	name := ctx.className()
	params := ctx.parameterList()
	ctx.assertToken(TOKEN_LC)
	ctx.nextToken()
	body := ctx.parse(TOKEN_RC, false)
	ctx.nextToken()
	var def Expression
	if resourceToken == TOKEN_APPLICATION {
		def = ctx.factory.Application(name, params, body, ctx.locator, start, ctx.Pos()-start)
	} else {
		def = ctx.factory.Definition(name, params, body, ctx.locator, start, ctx.Pos()-start)
	}
	return ctx.addDefinition(def)
}

func (ctx *context) addDefinition(expr Expression) Expression {
	ctx.definitions = append(ctx.definitions, expr.(Definition))
	return expr
}
