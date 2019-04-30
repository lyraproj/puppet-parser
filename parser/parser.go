package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lyraproj/issue/issue"
)

// Recursive descent context for the Puppet language.
//
// This is actually the lexer with added functionality. Having the lexer and context being the
// same instance is very beneficial when the lexer must parse expressions (as is the case when
// it encounters double quoted strings or heredoc with interpolation).

type (
	ExpressionParser interface {
		Parse(filename string, source string, singleExpression bool) (expr Expression, err error)
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

var workflowStyles = map[string]StepStyle{
	`action`:       StepStyleAction,
	`resource`:     StepStyleResource,
	`stateHandler`: StepStyleStateHandler,
	`workflow`:     StepStyleWorkflow,
}

type Lexer interface {
	CurrentToken() int

	NextToken() int

	SetPos(pos int)

	SyntaxError()

	tokenStartPos() int

	tokenValue() interface{}

	tokenString() string

	AssertToken(token int)
}

type lexer struct {
	context
}

type Option int

const HandleBacktickStrings = Option(1)
const HandleHexEscapes = Option(2)
const TasksEnabled = Option(3)
const WorkflowEnabled = Option(4)
const EppMode = Option(5)

func NewSimpleLexer(filename string, source string) Lexer {
	// Essentially a lexer that has no knowledge of interpolations
	return &lexer{context{
		stringReader:          stringReader{text: source},
		factory:               nil,
		locator:               &Locator{string: source, file: filename},
		handleBacktickStrings: false,
		handleHexEscapes:      false,
		tasks:                 false,
		workflow:              false}}
}

func (l *lexer) CurrentToken() int {
	return l.context.currentToken
}

func (l *lexer) NextToken() int {
	l.context.nextToken()
	return l.context.currentToken
}

func (l *lexer) SetPos(pos int) {
	l.context.SetPos(pos)
}

func (l *lexer) SyntaxError() {
	panic(l.context.parseIssue2(lexUnexpectedToken, issue.H{`token`: tokenMap[l.context.currentToken]}))
}

func (l *lexer) tokenString() string {
	return l.context.tokenString()
}

func (l *lexer) tokenValue() interface{} {
	return l.context.tokenValue
}

func (l *lexer) tokenStartPos() int {
	return l.context.tokenStartPos
}

func (l *lexer) AssertToken(token int) {
	l.context.assertToken(token)
}

// CreatePspecParser returns a parser that is capable of lexing backticked strings and that
// will recognize \xNN escapes in double quoted strings
func CreatePspecParser() ExpressionParser {
	return CreateParser(HandleBacktickStrings, HandleHexEscapes)
}

func CreateParser(parserOptions ...Option) ExpressionParser {
	ctx := &context{factory: DefaultFactory(), handleBacktickStrings: false, handleHexEscapes: false, tasks: false, workflow: false}
	for _, option := range parserOptions {
		switch option {
		case EppMode:
			ctx.eppMode = true
		case HandleBacktickStrings:
			ctx.handleBacktickStrings = true
		case HandleHexEscapes:
			ctx.handleHexEscapes = true
		case TasksEnabled:
			ctx.tasks = true
		case WorkflowEnabled:
			ctx.workflow = true
		}
	}
	return ctx
}

// Parse the contents of the given source. The filename is optional and will be used
// in warnings and errors issued by the context.
//
// If eppMode is true, the context will treat the given source as text with embedded puppet
// expressions.
func (ctx *context) Parse(filename string, source string, singleExpression bool) (expr Expression, err error) {
	ctx.stringReader = stringReader{text: source}
	ctx.locator = &Locator{string: source, file: filename}
	ctx.definitions = make([]Definition, 0, 8)
	ctx.nextLineStart = -1

	expr, err = ctx.parseTopExpression(filename, source, singleExpression)
	if err == nil && !singleExpression {
		expr = ctx.factory.Program(expr, ctx.definitions, ctx.locator, 0, ctx.Pos())
	}
	return
}

func (ctx *context) parseTopExpression(filename string, source string, singleExpression bool) (expr Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(issue.Reported); !ok {
				if err, ok = r.(*parseError); !ok {
					panic(r)
				}
			}
		}
	}()

	if ctx.eppMode {
		ctx.consumeEPP()

		var text string
		if ctx.currentToken == tokenRenderString {
			text = ctx.tokenString()
			ctx.nextToken()
		}

		asEppLambda := func(e Expression) Expression {
			if l, ok := e.(*LambdaExpression); ok {
				if _, ok = l.body.(*EppExpression); ok {
					return e
				}
			}
			if _, ok := e.(*BlockExpression); !ok {
				e = ctx.factory.Block([]Expression{e}, ctx.locator, 0, ctx.Pos())
			}
			return ctx.factory.EppExpression([]Expression{}, e, ctx.locator, 0, ctx.Pos())
		}

		if ctx.currentToken == tokenEnd {
			// No EPP in the source.
			expr = asEppLambda(ctx.factory.RenderString(text, ctx.locator, 0, ctx.Pos()))
			return
		}

		if ctx.currentToken == tokenPipe {
			if text != `` {
				panic(ctx.parseIssue(parseIllegalEppParameters))
			}
			params := ctx.lambdaParameterList()
			ctx.nextToken()
			expr = asEppLambda(
				ctx.factory.EppExpression(
					params, ctx.parse(tokenEnd, false), ctx.locator, 0, ctx.Pos()))
			return
		}

		expressions := make([]Expression, 0, 10)
		if text != `` {
			expressions = append(expressions, ctx.factory.RenderString(text, ctx.locator, 0, ctx.tokenStartPos))
		}

		for {
			if ctx.currentToken == tokenEnd {
				expr = asEppLambda(ctx.factory.Block(ctx.transformCalls(expressions, 0), ctx.locator, 0, ctx.Pos()))
				return
			}
			expressions = append(expressions, ctx.expression())
		}
	}

	ctx.nextToken()
	expr = ctx.parse(tokenEnd, singleExpression)
	return
}

func (ctx *context) parse(expectedEnd int, singleExpression bool) (expr Expression) {
	_, start := ctx.skipWhite(false)
	ctx.SetPos(start)
	if singleExpression {
		if ctx.currentToken == expectedEnd {
			expr = ctx.factory.Undef(ctx.locator, start, 0)
		} else {
			expr = ctx.relationship()
			ctx.assertToken(expectedEnd)
		}
		return
	}

	expressions := make([]Expression, 0, 10)
	for ctx.currentToken != expectedEnd {
		expressions = append(expressions, ctx.syntacticStatement())
		if ctx.currentToken == tokenSemicolon {
			ctx.nextToken()
		}
	}
	expr = ctx.factory.Block(ctx.transformCalls(expressions, start), ctx.locator, start, ctx.Pos()-start)
	return
}

func (ctx *context) assertToken(token int) {
	if ctx.currentToken != token {
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue2(parseExpectedToken, issue.H{`expected`: tokenMap[token], `actual`: tokenMap[ctx.currentToken]}))
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
			cn := ctx.factory.CallNamed(memo, false, args, nil, ctx.locator, memo.ByteOffset(), (expr.ByteOffset()+expr.ByteLength())-memo.ByteOffset())
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
	for _, ex := range result {
		if csl, ok := ex.(*commaSeparatedList); ok {
			// This happens when a block contains extraneous commas between statements. The
			// location of the comma is estimated to be right after the first statement in
			// the list
			f := csl.elements[0]
			p := f.ByteOffset() + f.ByteLength()
			l := ctx.locator
			loc := issue.NewLocation(f.File(), l.LineForOffset(p), l.PosOnLine(p))
			panic(issue.NewReported(parseExtraneousComma, issue.SeverityError, issue.NoArgs, loc))
		}
	}
	return
}

func (ctx *context) expressions(endToken int, producerFunc func() Expression) (exprs []Expression) {
	exprs = make([]Expression, 0, 4)
	for {
		if ctx.currentToken == endToken {
			return
		}
		exprs = append(exprs, producerFunc())
		if ctx.currentToken != tokenComma {
			if ctx.currentToken != endToken {
				ctx.SetPos(ctx.tokenStartPos)
				panic(ctx.parseIssue2(parseExpectedOneOfTokens, issue.H{
					`expected`: fmt.Sprintf(`'%s' or '%s'`, tokenMap[tokenComma], tokenMap[endToken]),
					`actual`:   tokenMap[ctx.currentToken]}))
			}
			return
		}
		ctx.nextToken()
	}
}

func (ctx *context) syntacticStatement() (expr Expression) {
	var args []Expression
	expr = ctx.relationship()
	for ctx.currentToken == tokenComma {
		ctx.nextToken()
		if args == nil {
			args = make([]Expression, 0, 2)
			args = append(args, expr)
		}
		args = append(args, ctx.relationship())
	}
	if args != nil {
		expr = &commaSeparatedList{LiteralList{Positioned{ctx.locator, expr.ByteOffset(), ctx.Pos() - expr.ByteOffset()}, args}}
	}
	return
}

func (ctx *context) collectionEntry() (expr Expression) {
	return ctx.argument()
}

func (ctx *context) argument() (expr Expression) {
	expr = ctx.handleKeyword(ctx.relationship)
	if ctx.currentToken == tokenFarrow {
		ctx.nextToken()
		value := ctx.handleKeyword(ctx.relationship)
		expr = ctx.factory.KeyedEntry(expr, value, ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())
	}
	return
}

func (ctx *context) hashEntry() (expr Expression) {
	return ctx.handleKeyword(ctx.relationship)
}

func (ctx *context) handleKeyword(next func() Expression) (expr Expression) {
	switch ctx.currentToken {
	case tokenType, tokenFunction, tokenPlan, tokenApplication, tokenConsumes, tokenProduces, tokenSite:
		expr = ctx.factory.QualifiedName(ctx.tokenString(), ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos)
		ctx.nextToken()
		if ctx.currentToken == tokenLp {
			expr = ctx.callFunctionExpression(expr)
		}
	default:
		expr = next()
	}
	return
}

func (ctx *context) relationship() (expr Expression) {
	expr = ctx.assignment()
	for {
		switch ctx.currentToken {
		case tokenInEdge, tokenInEdgeSub, tokenOutEdge, tokenOutEdgeSub:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.RelOp(op, expr, ctx.assignment(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())
		default:
			return expr
		}
	}
}

func (ctx *context) assignment() (expr Expression) {
	expr = ctx.step()
	for {
		switch ctx.currentToken {
		case tokenAssign, tokenAddAssign, tokenSubtractAssign:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Assignment(op, expr, ctx.assignment(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())
		default:
			return expr
		}
	}
}

func (ctx *context) step() (expr Expression) {
	start := ctx.Pos()
	expr = ctx.resource()
	if ctx.workflow {
		if qn, ok := expr.(*QualifiedName); ok {
			s := qn.Name()
			if style, ok := workflowStyles[s]; ok {
				if name, ok := ctx.identifier(); ok {
					expr = ctx.stepDeclaration(start, style, name, true)
				}
			}
		}
	}
	return
}

func (ctx *context) resource() (expr Expression) {
	expr = ctx.expression()
	if ctx.currentToken == tokenLc {
		expr = ctx.resourceExpression(expr.ByteOffset(), expr, REGULAR)
	}
	return
}

func (ctx *context) expression() (expr Expression) {
	expr = ctx.selectExpression()
	switch ctx.currentToken {
	case tokenProduces, tokenConsumes:
		// Must be preceded by name of class
		capToken := ctx.tokenString()
		switch expr.(type) {
		case *QualifiedName, *QualifiedReference, *ReservedWord, *AccessExpression:
			expr = ctx.capabilityMapping(expr, capToken)
		}
	}
	return
}

func (ctx *context) convertLhsToCall(ne *NamedAccessExpression, args []Expression, lambda Expression, start, len int) Expression {
	f := ctx.factory
	if nal, ok := ne.lhs.(*NamedAccessExpression); ok {
		ne = f.NamedAccess(ctx.convertLhsToCall(nal, []Expression{}, nil, nal.ByteOffset(), nal.ByteLength()),
			ne.rhs, ctx.locator, ne.ByteOffset(), ne.ByteLength()).(*NamedAccessExpression)
	}
	return f.CallMethod(ne, args, lambda, ctx.locator, start, len)
}

func (ctx *context) selectExpression() (expr Expression) {
	expr = ctx.orExpression()
	for {
		switch ctx.currentToken {
		case tokenQmark:
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
		case tokenOr:
			ctx.nextToken()
			expr = ctx.factory.Or(expr, ctx.andExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())
		default:
			return
		}
	}
}

func (ctx *context) andExpression() (expr Expression) {
	expr = ctx.compareExpression()
	for {
		switch ctx.currentToken {
		case tokenAnd:
			ctx.nextToken()
			expr = ctx.factory.And(expr, ctx.compareExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())
		default:
			return
		}
	}
}

func (ctx *context) compareExpression() (expr Expression) {
	expr = ctx.equalExpression()
	for {
		switch ctx.currentToken {
		case tokenLess, tokenLessEqual, tokenGreater, tokenGreaterEqual:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Comparison(op, expr, ctx.equalExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())

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
		case tokenEqual, tokenNotEqual:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Comparison(op, expr, ctx.shiftExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())

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
		case tokenLshift, tokenRshift:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Arithmetic(op, expr, ctx.additiveExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())

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
		case tokenAdd, tokenSubtract:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Arithmetic(op, expr, ctx.multiplicativeExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())

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
		case tokenMultiply, tokenDivide, tokenRemainder:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Arithmetic(op, expr, ctx.matchExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())

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
		case tokenMatch, tokenNotMatch:
			op := ctx.tokenString()
			ctx.nextToken()
			expr = ctx.factory.Match(op, expr, ctx.inExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())

		default:
			return
		}
	}
}

func (ctx *context) inExpression() (expr Expression) {
	expr = ctx.unaryExpression()
	for {
		switch ctx.currentToken {
		case tokenIn:
			ctx.nextToken()
			expr = ctx.factory.In(expr, ctx.unaryExpression(), ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())

		default:
			return expr
		}
	}
}

func (ctx *context) arrayExpression() (elements []Expression) {
	return ctx.joinHashEntries(ctx.expressions(tokenRb, ctx.collectionEntry))
}

func (ctx *context) keyedEntry() Expression {
	key := ctx.hashEntry()
	if ctx.currentToken != tokenFarrow {
		panic(ctx.parseIssue(parseExpectedFarrowAfterKey))
	}
	ctx.nextToken()
	value := ctx.hashEntry()
	return ctx.factory.KeyedEntry(key, value, ctx.locator, key.ByteOffset(), ctx.Pos()-key.ByteOffset())
}

func (ctx *context) hashExpression() (entries []Expression) {
	return ctx.expressions(tokenRc, ctx.keyedEntry)
}

func (ctx *context) unaryExpression() Expression {
	unaryStart := ctx.tokenStartPos
	switch ctx.currentToken {
	case tokenSubtract:
		if c, _ := ctx.Peek(); isDecimalDigit(c) {
			ctx.nextToken()
			if ctx.currentToken == tokenInteger {
				ctx.settokenValue(ctx.currentToken, -ctx.tokenValue.(int64))
			} else {
				ctx.settokenValue(ctx.currentToken, -ctx.tokenValue.(float64))
			}
			expr := ctx.primaryExpression()
			expr.updateOffsetAndLength(unaryStart, ctx.Pos()-unaryStart)
			return expr
		}
		ctx.nextToken()
		expr := ctx.primaryExpression()
		return ctx.factory.Negate(expr, ctx.locator, unaryStart, ctx.Pos()-unaryStart)

	case tokenAdd:
		// Allow '+' prefix for constant numbers
		if c, _ := ctx.Peek(); isDecimalDigit(c) {
			ctx.nextToken()
			expr := ctx.primaryExpression()
			expr.updateOffsetAndLength(unaryStart, ctx.Pos()-unaryStart)
			return expr
		}
		panic(ctx.parseIssue2(lexUnexpectedToken, issue.H{`token`: `+`}))

	case tokenNot:
		ctx.nextToken()
		expr := ctx.unaryExpression()
		return ctx.factory.Not(expr, ctx.locator, unaryStart, ctx.Pos()-unaryStart)

	case tokenMultiply:
		ctx.nextToken()
		expr := ctx.unaryExpression()
		return ctx.factory.Unfold(expr, ctx.locator, unaryStart, ctx.Pos()-unaryStart)

	case tokenAt, tokenAtat:
		kind := VIRTUAL
		if ctx.currentToken == tokenAtat {
			kind = EXPORTED
		}
		ctx.nextToken()
		expr := ctx.primaryExpression()
		ctx.assertToken(tokenLc)
		return ctx.resourceExpression(unaryStart, expr, kind)

	default:
		return ctx.primaryExpression()
	}
}

func (ctx *context) primaryExpression() (expr Expression) {
	expr = ctx.atomExpression()
	for {
		switch ctx.currentToken {
		case tokenLp, tokenPipe:
			expr = ctx.callFunctionExpression(expr)
		case tokenLcollect, tokenLlcollect:
			expr = ctx.collectExpression(expr)
		case tokenLb:
			ctx.nextToken()
			params := ctx.arrayExpression()
			isCall := false
			if qn, ok := expr.(*QualifiedName); ok {
				_, isCall = statementCalls[qn.name]
			}
			l := ctx.Pos() - expr.ByteOffset()
			if isCall {
				expr = ctx.factory.CallNamed(expr, false, []Expression{ctx.factory.Array(params, ctx.locator, expr.ByteOffset(), l)}, nil, ctx.locator, expr.ByteOffset(), l)
			} else {
				expr = ctx.factory.Access(expr, params, ctx.locator, expr.ByteOffset(), l)
			}
			ctx.nextToken()
		case tokenDot:
			ctx.nextToken()
			var rhs Expression
			if ctx.currentToken == tokenType {
				rhs = ctx.factory.QualifiedName(ctx.tokenString(), ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos)
				ctx.nextToken()
			} else {
				rhs = ctx.atomExpression()
			}
			expr = ctx.factory.NamedAccess(expr, rhs, ctx.locator, expr.ByteOffset(), ctx.Pos()-expr.ByteOffset())
		default:
			if namedAccess, ok := expr.(*NamedAccessExpression); ok {
				// Transform into method calls
				expr = ctx.convertLhsToCall(namedAccess, []Expression{}, nil, expr.ByteOffset(), expr.ByteLength())
			}
			return
		}
	}
}

func (ctx *context) atomExpression() (expr Expression) {
	atomStart := ctx.tokenStartPos
	switch ctx.currentToken {
	case tokenLp, tokenWslp:
		ctx.nextToken()
		expr = ctx.factory.Parenthesized(ctx.relationship(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.assertToken(tokenRp)
		ctx.nextToken()

	case tokenLb, tokenListstart:
		ctx.nextToken()
		expr = ctx.factory.Array(ctx.arrayExpression(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenLc:
		ctx.nextToken()
		expr = ctx.factory.Hash(ctx.hashExpression(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenBoolean:
		expr = ctx.factory.Boolean(ctx.tokenValue.(bool), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenInteger:
		expr = ctx.factory.Integer(ctx.tokenValue.(int64), ctx.radix, ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenFloat:
		expr = ctx.factory.Float(ctx.tokenValue.(float64), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenString:
		expr = ctx.factory.String(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenAttr, tokenPrivate:
		expr = ctx.factory.ReservedWord(ctx.tokenString(), false, ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenDefault:
		expr = ctx.factory.Default(ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenHeredoc, tokenConcatenatedString:
		expr = ctx.tokenValue.(Expression)
		ctx.nextToken()

	case tokenRegexp:
		expr = ctx.factory.Regexp(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenUndef:
		expr = ctx.factory.Undef(ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenTypeName:
		expr = ctx.factory.QualifiedReference(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenIdentifier:
		expr = ctx.factory.QualifiedName(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenVariable:
		vni := ctx.tokenValue
		ctx.nextToken()
		var name Expression
		if s, ok := vni.(string); ok {
			name = ctx.factory.QualifiedName(s, ctx.locator, atomStart+1, len(s))
		} else {
			name = ctx.factory.Integer(vni.(int64), 10, ctx.locator, atomStart+1, ctx.Pos()-(atomStart+1))
		}
		expr = ctx.factory.Variable(name, ctx.locator, atomStart, ctx.Pos()-atomStart)

	case tokenCase:
		expr = ctx.caseExpression()

	case tokenIf:
		expr = ctx.ifExpression(false)

	case tokenUnless:
		expr = ctx.ifExpression(true)

	case tokenClass:
		name := ctx.tokenString()
		ctx.nextToken()
		if ctx.currentToken == tokenLc {
			// Class resource
			expr = ctx.factory.QualifiedName(name, ctx.locator, atomStart, ctx.Pos()-atomStart)
		} else {
			expr = ctx.classExpression(atomStart)
		}

	case tokenType:
		// look ahead for '(' in which case this is a named function call
		name := ctx.tokenString()
		ctx.nextToken()
		if ctx.currentToken == tokenTypeName {
			expr = ctx.typeAliasOrDefinition()
		} else {
			// Not a type definition. Just treat the 'type' keyword as a qualified name
			expr = ctx.factory.QualifiedName(name, ctx.locator, atomStart, ctx.Pos()-atomStart)
		}

	case tokenPlan:
		expr = ctx.planDefinition()

	case tokenFunction:
		expr = ctx.functionDefinition()

	case tokenNode:
		expr = ctx.nodeDefinition()

	case tokenDefine, tokenApplication:
		expr = ctx.resourceDefinition(ctx.currentToken)

	case tokenSite:
		expr = ctx.siteDefinition()

	case tokenRenderString:
		expr = ctx.factory.RenderString(ctx.tokenString(), ctx.locator, atomStart, ctx.Pos()-atomStart)
		ctx.nextToken()

	case tokenRenderExpr:
		ctx.nextToken()
		expr = ctx.factory.RenderExpression(ctx.expression(), ctx.locator, atomStart, ctx.Pos()-atomStart)

	default:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue2(lexUnexpectedToken, issue.H{`token`: tokenMap[ctx.currentToken]}))
	}
	return
}

func (ctx *context) ifExpression(unless bool) (expr Expression) {
	start := ctx.tokenStartPos // start of if, elsif, or unless keyword
	ctx.nextToken()
	condition := ctx.orExpression()
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	thenPart := ctx.parse(tokenRc, false)
	ctx.nextToken()

	var elsePart Expression
	switch ctx.currentToken {
	case tokenElse:
		ctx.nextToken()
		ctx.assertToken(tokenLc)
		ctx.nextToken()
		elsePart = ctx.parse(tokenRc, false)
		ctx.nextToken()
	case tokenElsif:
		if unless {
			panic(ctx.parseIssue(parseElsifInUnless))
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
	needNext := false
	if ctx.currentToken == tokenSelc {
		ctx.nextToken()
		selectors = ctx.expressions(tokenRc, ctx.selectorEntry)
		needNext = true
	} else {
		selectors = []Expression{ctx.selectorEntry()}
	}
	expr = ctx.factory.Select(test, selectors, ctx.locator, test.ByteOffset(), ctx.Pos()-test.ByteOffset())
	if needNext {
		ctx.nextToken()
	}
	return
}

func (ctx *context) selectorEntry() (expr Expression) {
	start := ctx.tokenStartPos
	lhs := ctx.expression()
	ctx.assertToken(tokenFarrow)
	ctx.nextToken()
	return ctx.factory.Selector(lhs, ctx.expression(), ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) caseExpression() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	test := ctx.expression()
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	caseOptions := ctx.caseOptions()
	expr := ctx.factory.Case(test, caseOptions, ctx.locator, start, ctx.Pos()-start)
	ctx.nextToken()
	return expr
}

func (ctx *context) caseOptions() (exprs []Expression) {
	exprs = make([]Expression, 0, 4)
	for {
		exprs = append(exprs, ctx.caseOption())
		if ctx.currentToken == tokenRc {
			return
		}
	}
}

func (ctx *context) caseOption() Expression {
	start := ctx.tokenStartPos
	expressions := ctx.expressions(tokenColon, ctx.expression)
	ctx.nextToken()
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	block := ctx.parse(tokenRc, false)
	ctx.nextToken()
	return ctx.factory.When(expressions, block, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) resourceExpression(start int, first Expression, form ResourceForm) (expr Expression) {
	bodiesStart := ctx.Pos()
	ctx.nextToken()
	titleStart := ctx.Pos()
	var firstTitle Expression

	// First attribute might be a * => operator. No attempt should be made
	// to read it as an expression.
	if ctx.currentToken != tokenMultiply {
		firstTitle = ctx.expression()
	}

	if ctx.currentToken != tokenColon {
		// Resource body without title
		ctx.SetPos(titleStart)
		switch ctx.resourceShape(first) {
		case `resource`:
			// This is just LHS followed by a hash. It only makes sense when LHS is an identifier equal
			// to one of the known "statement calls" or, if workflow is enabled, to one of the keywords
			// "workflow", "action", or "resource". For all other cases, this is an error
			fqn, ok := first.(*QualifiedName)
			name := ``
			if ok {
				name = fqn.name
				if _, ok := statementCalls[name]; ok {
					// Handle the call here and set lexer position to where the next expression (the one starting
					// with a curly brace) starts.
					args := make([]Expression, 1)
					ctx.SetPos(bodiesStart)
					ctx.nextToken()
					args[0] = ctx.factory.Hash(ctx.hashExpression(), ctx.locator, bodiesStart, ctx.Pos()-bodiesStart)
					expr = ctx.factory.CallNamed(first, true, args, nil, ctx.locator, start, ctx.Pos()-start)
					ctx.nextToken()
					return
				}
			}
			ctx.SetPos(start)
			panic(ctx.parseIssue2(parseResourceWithoutTitle, issue.H{`name`: name}))
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
			ctx.SetPos(first.ByteOffset())
			panic(ctx.parseIssue(parseInvalidResource))
		}
	} else {
		bodies := ctx.resourceBodies(firstTitle)
		expr = ctx.factory.Resource(form, first, bodies, ctx.locator, start, ctx.Pos()-start)
	}

	ctx.assertToken(tokenRc)
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
	for ctx.currentToken != tokenRc {
		result = append(result, ctx.resourceBody(title))
		if ctx.currentToken != tokenSemicolon {
			break
		}
		ctx.nextToken()
		if ctx.currentToken != tokenRc {
			title = ctx.expression()
		}
	}
	return
}

func (ctx *context) resourceBody(title Expression) Expression {
	if ctx.currentToken != tokenColon {
		ctx.SetPos(title.ByteOffset())
		panic(ctx.parseIssue(parseExpectedTitle))
	}
	ctx.nextToken()
	ops := ctx.attributeOperations()
	return ctx.factory.ResourceBody(title, ops, ctx.locator, title.ByteOffset(), ctx.Pos()-title.ByteOffset())
}

func (ctx *context) attributeOperations() (result []Expression) {
	result = make([]Expression, 0, 5)
	for {
		switch ctx.currentToken {
		case tokenSemicolon, tokenRc:
			return
		default:
			result = append(result, ctx.attributeOperation())
			if ctx.currentToken != tokenComma {
				return
			}
			ctx.nextToken()
		}
	}
}

func (ctx *context) attributeOperation() (op Expression) {
	start := ctx.tokenStartPos
	splat := ctx.currentToken == tokenMultiply
	if splat {
		ctx.nextToken()
		ctx.assertToken(tokenFarrow)
		ctx.nextToken()
		return ctx.factory.AttributesOp(ctx.expression(), ctx.locator, start, ctx.Pos()-start)
	}

	name := ctx.attributeName()

	switch ctx.currentToken {
	case tokenFarrow, tokenParrow:
		op := ctx.tokenString()
		ctx.nextToken()
		return ctx.factory.AttributeOp(op, name, ctx.expression(), ctx.locator, start, ctx.Pos()-start)
	default:
		panic(ctx.parseIssue(parseInvalidAttribute))
	}
}

func (ctx *context) attributeName() string {
	if name, ok := ctx.identifier(); ok {
		return name
	}
	panic(ctx.parseIssue(parseExpectedAttributeName))
}

func (ctx *context) identifier() (string, bool) {
	start := ctx.tokenStartPos
	switch ctx.currentToken {
	case tokenIdentifier:
		name := ctx.tokenString()
		ctx.nextToken()
		return name, true
	default:
		if word, ok := ctx.keyword(); ok {
			ctx.nextToken()
			return word, ok
		}
		ctx.SetPos(start)
		return ``, false
	}
}

func (ctx *context) identifierExpr() (Expression, bool) {
	start := ctx.tokenStartPos
	switch ctx.currentToken {
	case tokenIdentifier:
		name := ctx.factory.QualifiedName(ctx.tokenString(), ctx.locator, start, start-ctx.Pos())
		ctx.nextToken()
		return name, true
	default:
		if word, ok := ctx.keyword(); ok {
			name := ctx.factory.QualifiedName(word, ctx.locator, start, start-ctx.Pos())
			ctx.nextToken()
			return name, ok
		}
		ctx.SetPos(start)
		return nil, false
	}
}

func (ctx *context) collectExpression(lhs Expression) Expression {
	var collectQuery Expression
	queryStart := ctx.tokenStartPos
	if ctx.currentToken == tokenLcollect {
		ctx.nextToken()
		var queryExpr Expression
		if ctx.currentToken == tokenRcollect {
			queryExpr = ctx.factory.Nop(ctx.locator, ctx.tokenStartPos, 0)
		} else {
			queryExpr = ctx.expression()
			ctx.assertToken(tokenRcollect)
		}
		ctx.nextToken()
		collectQuery = ctx.factory.VirtualQuery(queryExpr, ctx.locator, queryStart, ctx.Pos()-queryStart)
	} else {
		ctx.nextToken()
		var queryExpr Expression
		if ctx.currentToken == tokenRrcollect {
			queryExpr = ctx.factory.Nop(ctx.locator, queryStart, ctx.tokenStartPos-queryStart)
		} else {
			queryExpr = ctx.expression()
			ctx.assertToken(tokenRrcollect)
		}
		ctx.nextToken()
		collectQuery = ctx.factory.ExportedQuery(queryExpr, ctx.locator, queryStart, ctx.Pos()-queryStart)
	}

	var attributeOps []Expression
	if ctx.currentToken != tokenLc {
		attributeOps = make([]Expression, 0)
	} else {
		ctx.nextToken()
		attributeOps = ctx.attributeOperations()
		ctx.assertToken(tokenRc)
		ctx.nextToken()
	}
	return ctx.factory.Collect(lhs, collectQuery, attributeOps, ctx.locator, lhs.ByteOffset(), ctx.Pos()-lhs.ByteOffset())
}

func (ctx *context) typeAliasOrDefinition() Expression {
	start := ctx.tokenStartPos
	typeExpr := ctx.parameterType()
	fqr, ok := typeExpr.(*QualifiedReference)
	if !ok {
		if _, ok = typeExpr.(*AccessExpression); ok {
			if ctx.currentToken == tokenAssign {
				ctx.nextToken()
				return ctx.addDefinition(ctx.factory.TypeMapping(typeExpr, ctx.expression(), ctx.locator, start, ctx.Pos()-start))
			}
		}
		panic(ctx.parseIssue(parseExpectedTypeNameAfterType))
	}

	parent := ``
	switch ctx.currentToken {
	case tokenAssign:
		ctx.nextToken()
		bodyStart := ctx.tokenStartPos
		body := ctx.expression()
		switch bt := body.(type) {
		case *QualifiedReference:
			if ctx.currentToken == tokenLc {
				hash := ctx.expression().(*LiteralHash)
				if bt.name == `Object` || bt.name == `TypeSet` {
					body = ctx.factory.Access(bt, []Expression{hash}, ctx.locator, bodyStart, ctx.Pos()-bodyStart)
				} else {
					pref := ctx.factory.String(`parent`, ctx.locator, bt.ByteOffset(), bt.ByteLength())
					hash := ctx.factory.Hash(
						append([]Expression{ctx.factory.KeyedEntry(pref, bt, ctx.locator, bt.ByteOffset(), bt.ByteLength())}, hash.entries...),
						ctx.locator, bodyStart, ctx.Pos()-bodyStart)
					body = ctx.factory.Access(ctx.factory.QualifiedReference(`Object`, ctx.locator, bodyStart, 0), []Expression{hash}, ctx.locator, bodyStart, ctx.Pos()-bodyStart)
				}
			}
		case *LiteralList:
			if len(bt.elements) == 1 {
				body = ctx.factory.Access(ctx.factory.QualifiedReference(`Object`, ctx.locator, bodyStart, 0), bt.elements, ctx.locator, bodyStart, ctx.Pos()-bodyStart)
			}
		case *LiteralHash:
			body = ctx.factory.Access(ctx.factory.QualifiedReference(`Object`, ctx.locator, bodyStart, 0), []Expression{body}, ctx.locator, bodyStart, ctx.Pos()-bodyStart)
		}
		return ctx.addDefinition(ctx.factory.TypeAlias(fqr.name, body, ctx.locator, start, ctx.Pos()-start))
	case tokenInherits:
		ctx.nextToken()
		nameExpr := ctx.typeName()
		if nameExpr == nil {
			panic(ctx.parseIssue(parseInheritsMustBeTypeName))
		}
		parent = nameExpr.(*QualifiedReference).name
		ctx.assertToken(tokenLc)
		fallthrough

	case tokenLc:
		ctx.nextToken()
		body := ctx.parse(tokenRc, false)
		ctx.nextToken() // consume TOKEN_RC
		return ctx.addDefinition(ctx.factory.TypeDefinition(fqr.name, parent, body, ctx.locator, start, ctx.Pos()-start))

	default:
		panic(ctx.parseIssue2(lexUnexpectedToken, issue.H{`token`: tokenMap[ctx.currentToken]}))
	}
}

func (ctx *context) callFunctionExpression(functorExpr Expression) Expression {
	var args []Expression
	start := functorExpr.ByteOffset()
	end := start + functorExpr.ByteLength()
	if ctx.currentToken != tokenPipe {
		ctx.nextToken()
		args = ctx.arguments()
		end = ctx.Pos()
		ctx.nextToken()
	}
	var block Expression
	if ctx.currentToken == tokenPipe {
		block = ctx.lambda()
		end = block.ByteOffset() + block.ByteLength()
	}
	if namedAccess, ok := functorExpr.(*NamedAccessExpression); ok {
		return ctx.convertLhsToCall(namedAccess, args, block, start, end-start)
	}
	return ctx.factory.CallNamed(functorExpr, true, args, block, ctx.locator, start, end-start)
}

func (ctx *context) stepStyle() StepStyle {
	switch ctx.currentToken {
	case tokenIdentifier:
		if style, ok := workflowStyles[ctx.tokenString()]; ok {
			ctx.nextToken()
			return style
		}
	}
	panic(ctx.parseIssue(parseExpectedStepStyle))
}

func (ctx *context) stepName(step StepStyle) string {
	if tn, ok := ctx.identifier(); ok {
		return tn
	}
	panic(ctx.parseIssue2(parseExpectedStepName, issue.H{`step`: step}))
}

// stepEntry is a hash entry with some specific constraints

func (ctx *context) stepProperty() Expression {
	start := ctx.Pos()
	key, ok := ctx.identifierExpr()
	if !ok {
		panic(ctx.parseIssue(parseExpectedAttributeName))
	}
	if ctx.currentToken != tokenFarrow {
		panic(ctx.parseIssue(parseExpectedFarrowAfterKey))
	}
	ctx.nextToken()

	vstart := ctx.Pos()
	name := key.(*QualifiedName).name
	var value Expression
	switch name {
	case `parameters`:
		// TODO: Allow non condensed declaration using array of hashes where everything is
		// spelled out (type and value expressed in hash)
		value = ctx.factory.Array(ctx.parameterList(), ctx.locator, vstart, ctx.Pos()-vstart)
		ctx.nextToken()
	case `returns`:
		// TODO: Allow non condensed declaration using array of hashes where everything is
		// spelled out (type and value expressed in hash)
		params := ctx.returnParameters()
		value = ctx.factory.Array(params, ctx.locator, vstart, ctx.Pos()-vstart)
		ctx.nextToken()

	default:
		value = ctx.hashEntry()
	}
	return ctx.factory.KeyedEntry(key, value, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) stateHash(start int) []Expression {
	entries := ctx.expressions(tokenRc, ctx.stateAttribute)

	// All variable references in this hash must be converted to Deferred references to prevent that evaluation happens
	// prematurely.
	for i, e := range entries {
		entries[i] = convertToDeferred(ctx.factory, e)
	}
	return entries
}

func convertSliceToDeferred(f ExpressionFactory, be []Expression) []Expression {
	ae := make([]Expression, len(be))
	for i, e := range be {
		ae[i] = convertToDeferred(f, e)
	}
	return ae
}

func convertToDeferred(f ExpressionFactory, e Expression) Expression {
	if e == nil {
		return nil
	}
	l := e.Locator()
	bo := e.ByteOffset()
	bl := e.ByteLength()
	switch et := e.(type) {
	case *LiteralList:
		e = f.Array(convertSliceToDeferred(f, et.elements), l, bo, bl)
	case *LiteralHash:
		e = f.Hash(convertSliceToDeferred(f, et.entries), l, bo, bl)
	case *KeyedEntry:
		e = f.KeyedEntry(convertToDeferred(f, et.key), convertToDeferred(f, et.value), l, bo, bl)
	case *CallNamedFunctionExpression:
		switch et.functor.(type) {
		case *QualifiedName:
			n := et.functor.(*QualifiedName).Name()
			nw := f.QualifiedName(`new`, l, bo, 0)
			e = f.CallMethod(f.NamedAccess(f.QualifiedReference(`Deferred`, l, bo, 0), nw, l, bo, 0),
				[]Expression{f.String(n, l, e.ByteOffset(), e.ByteLength()), f.Array(convertSliceToDeferred(f, et.arguments), l, bo, 0)}, nil, l, bo, bl)
		case *QualifiedReference:
			nw := f.QualifiedName(`new`, l, bo, 0)
			args := append([]Expression{et.functor}, et.arguments...)
			e = f.CallMethod(f.NamedAccess(f.QualifiedReference(`Deferred`, l, bo, 0), nw, l, bo, 0),
				[]Expression{nw, f.Array(convertSliceToDeferred(f, args), l, bo, 0)}, nil, l, bo, bl)
		}
	case *VariableExpression:
		ve := e.(*VariableExpression)
		n, _ := ve.Name()
		e = f.CallMethod(f.NamedAccess(f.QualifiedReference(`Deferred`, l, bo, 0), f.QualifiedName(`new`, l, bo, 0), l, bo, 0),
			[]Expression{f.String(`$`+n, l, bo, bl)}, nil, l, bo, bl)
	}
	return e
}

func (ctx *context) stateAttribute() (op Expression) {
	start := ctx.tokenStartPos
	name, ok := ctx.identifierExpr()
	if !ok {
		panic(ctx.parseIssue(parseExpectedAttributeName))
	}

	switch ctx.currentToken {
	case tokenFarrow:
		ctx.nextToken()
		return ctx.factory.KeyedEntry(name, ctx.expression(), ctx.locator, start, ctx.Pos()-start)
	default:
		panic(ctx.parseIssue(parseInvalidAttribute))
	}
}

func (ctx *context) stepExpression() Expression {
	start := ctx.Pos()
	if ctx.currentToken == tokenFunction {
		return ctx.functionDefinition()
	}
	style := ctx.stepStyle()
	name := ctx.stepName(style)
	return ctx.stepDeclaration(start, style, name, false)
}

func (ctx *context) activities() []Expression {
	activities := make([]Expression, 0)
	for ctx.currentToken != tokenRc {
		activities = append(activities, ctx.stepExpression())
	}
	ctx.nextToken()
	return activities
}

func (ctx *context) stepDeclaration(start int, style StepStyle, name string, atTop bool) Expression {
	if style == StepStyleWorkflow {
		// Push to name stack
		ctx.nameStack = append(ctx.nameStack, name)
	}
	hstart := ctx.tokenStartPos
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	propEntries := ctx.expressions(tokenRc, ctx.stepProperty)
	hEnd := ctx.Pos()
	ctx.nextToken()

	f := ctx.factory
	l := ctx.locator

	iterName := name
	if ctx.currentToken == tokenVariable {
		iterName = ctx.tokenString()
		ctx.nextToken()
		ctx.assertToken(tokenAssign)
		ctx.nextToken()
		ctx.assertToken(tokenIdentifier)
	}

	if ctx.currentToken == tokenIdentifier {
		switch ctx.tokenString() {
		case `times`, `range`, `each`:
			iterFunc := ctx.tokenString()
			nl := len(iterFunc)
			fs := ctx.tokenStartPos
			ps := ctx.Pos()
			ctx.nextToken()
			iterParams := ctx.parameterList()
			ctx.nextToken()
			vs := ctx.Pos()
			pl := ps - vs
			iterVars := ctx.lambdaParameterList()
			ctx.nextToken()
			vl := ctx.Pos() - vs
			fn := ctx.Pos() - fs
			iter := f.Hash(
				[]Expression{
					f.KeyedEntry(
						f.QualifiedName(`name`, l, fs, 0),
						f.QualifiedName(iterName, l, fs, nl), l, fs, nl),
					f.KeyedEntry(
						f.QualifiedName(`function`, l, fs, 0),
						f.QualifiedName(iterFunc, l, fs, nl), l, fs, nl),
					f.KeyedEntry(
						f.QualifiedName(`params`, l, ps, 0),
						f.Array(iterParams, l, ps, pl), l, ps, pl),
					f.KeyedEntry(
						f.QualifiedName(`vars`, l, vs, 0),
						f.Array(iterVars, l, vs, vl), l, vs, vl),
				}, l, fs, fn)
			propEntries = append(propEntries, f.KeyedEntry(f.QualifiedName(`iteration`, l, fs, 0), iter, l, fs, fn))
		default:
			panic(ctx.parseIssue2(parseExpectedIteratorStyle, issue.H{`style`: ctx.tokenString()}))
		}
	}
	var properties Expression
	if len(propEntries) > 0 {
		properties = f.Hash(propEntries, l, hstart, hEnd-hstart)
	}

	var block Expression

	switch style {
	case StepStyleWorkflow:
		if ctx.currentToken == tokenLc {
			hstart := ctx.tokenStartPos
			ctx.nextToken()
			activities := ctx.activities()
			if len(activities) > 0 {
				block = ctx.factory.Block(activities, ctx.locator, hstart, ctx.Pos()-hstart)
			}
		}

		// Pop name stack
		ctx.nameStack = ctx.nameStack[:len(ctx.nameStack)-1]
	case StepStyleResource:
		if ctx.currentToken == tokenLc {
			hstart := ctx.tokenStartPos
			ctx.nextToken()
			entries := ctx.stateHash(hstart)
			if len(entries) > 0 {
				block = ctx.factory.Hash(entries, ctx.locator, start, ctx.Pos()-start).(*LiteralHash)
			}
			ctx.nextToken()
		}
	default: // StepStyleAction or StepStyleStateHandler
		ctx.assertToken(tokenLc)
		ctx.nextToken()
		block = ctx.parse(tokenRc, false)
		ctx.nextToken()
	}
	step := f.Step(ctx.qualifiedName(name), style, properties, block, l, start, ctx.Pos()-start)
	if atTop {
		ctx.addDefinition(step)
	}
	return step
}

func (ctx *context) lambda() (result Expression) {
	start := ctx.tokenStartPos
	parameterList := ctx.lambdaParameterList()
	ctx.nextToken()
	var returnType Expression
	if ctx.currentToken == tokenRshift {
		ctx.nextToken()
		returnType = ctx.parameterType()
	}

	ctx.assertToken(tokenLc)
	ctx.nextToken()
	block := ctx.parse(tokenRc, false)
	result = ctx.factory.Lambda(parameterList, block, returnType, ctx.locator, start, ctx.Pos()-start)
	ctx.nextToken() // consume TOKEN_RC
	return
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
	start := entries[0].ByteOffset()
	last := entries[len(entries)-1]
	end := last.ByteOffset() + last.ByteLength()
	return ctx.factory.Hash(entries, ctx.locator, start, end-start)
}

func (ctx *context) arguments() (result []Expression) {
	return ctx.joinHashEntries(ctx.expressions(tokenRp, ctx.argument))
}

func (ctx *context) functionDefinition() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	var name string
	switch ctx.currentToken {
	case tokenIdentifier, tokenTypeName:
		name = ctx.tokenString()
	default:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(parseExpectedNameAfterFunction))
	}

	ctx.nextToken()

	var parameterList []Expression
	switch ctx.currentToken {
	case tokenLp, tokenWslp:
		parameterList = ctx.parameterList()
		ctx.nextToken()
	default:
		parameterList = []Expression{}
	}

	var returnType Expression
	if ctx.currentToken == tokenRshift {
		ctx.nextToken()
		returnType = ctx.parameterType()
	}

	ctx.assertToken(tokenLc)
	ctx.nextToken()
	block := ctx.parse(tokenRc, false)
	ctx.nextToken() // consume TOKEN_RC
	return ctx.addDefinition(ctx.factory.Function(name, parameterList, block, returnType, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) planDefinition() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	var name string
	switch ctx.currentToken {
	case tokenIdentifier, tokenTypeName:
		name = ctx.tokenString()
	default:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(parseExpectedNameAfterPlan))
	}
	ctx.nextToken()

	// Push to namestack
	ctx.nameStack = append(ctx.nameStack, name)

	var parameterList []Expression
	switch ctx.currentToken {
	case tokenLp, tokenWslp:
		parameterList = ctx.parameterList()
		ctx.nextToken()
	default:
		parameterList = []Expression{}
	}

	var returnType Expression
	if ctx.currentToken == tokenRshift {
		ctx.nextToken()
		returnType = ctx.parameterType()
	}

	ctx.assertToken(tokenLc)
	ctx.nextToken()
	block := ctx.parse(tokenRc, false)
	ctx.nextToken() // consume TOKEN_RC

	// Pop namestack
	ctx.nameStack = ctx.nameStack[:len(ctx.nameStack)-1]
	return ctx.addDefinition(ctx.factory.Plan(name, parameterList, block, returnType, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) nodeDefinition() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	hostnames := ctx.hostnames()
	var nodeParent Expression
	if ctx.currentToken == tokenInherits {
		ctx.nextToken()
		nodeParent = ctx.hostname()
	}
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	block := ctx.parse(tokenRc, false)
	ctx.nextToken()
	return ctx.addDefinition(ctx.factory.Node(hostnames, nodeParent, block, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) hostnames() (hostnames []Expression) {
	hostnames = make([]Expression, 0, 4)
	for {
		hostnames = append(hostnames, ctx.hostname())
		if ctx.currentToken != tokenComma {
			return
		}
		ctx.nextToken()
		switch ctx.currentToken {
		case tokenInherits, tokenLc:
			return
		}
	}
}

func (ctx *context) hostname() (hostname Expression) {
	start := ctx.tokenStartPos
	switch ctx.currentToken {
	case tokenIdentifier, tokenTypeName, tokenInteger, tokenFloat:
		hostname = ctx.dottedName()
	case tokenRegexp:
		hostname = ctx.factory.Regexp(ctx.tokenString(), ctx.locator, start, ctx.Pos()-start)
		ctx.nextToken()
	case tokenString:
		hostname = ctx.factory.String(ctx.tokenString(), ctx.locator, start, ctx.Pos()-start)
		ctx.nextToken()
	case tokenDefault:
		hostname = ctx.factory.Default(ctx.locator, start, ctx.Pos()-start)
		ctx.nextToken()
	case tokenConcatenatedString, tokenHeredoc:
		hostname = ctx.tokenValue.(Expression)
		ctx.nextToken()
	default:
		panic(ctx.parseIssue(parseExpectedHostname))
	}
	return
}

func (ctx *context) dottedName() Expression {
	start := ctx.tokenStartPos
	names := make([]string, 0, 8)
	for {
		switch ctx.currentToken {
		case tokenIdentifier, tokenTypeName:
			names = append(names, ctx.tokenString())
		case tokenInteger:
			names = append(names, strconv.FormatInt(ctx.tokenValue.(int64), 10))
		case tokenFloat:
			names = append(names, strconv.FormatFloat(ctx.tokenValue.(float64), 'g', -1, 64))
		default:
			panic(ctx.parseIssue(parseExpectedNameOrNumberAfterDot))
		}

		ctx.nextToken()
		if ctx.currentToken != tokenDot {
			return ctx.factory.String(strings.Join(names, `.`), ctx.locator, start, ctx.Pos()-start)
		}
		ctx.nextToken()
	}
}

func (ctx *context) parameterList() (result []Expression) {
	switch ctx.currentToken {
	case tokenLp, tokenWslp:
		ctx.nextToken()
		return ctx.expressions(tokenRp, ctx.parameter)
	default:
		return []Expression{}
	}
}

func (ctx *context) lambdaParameterList() (result []Expression) {
	ctx.nextToken()
	return ctx.expressions(tokenPipeEnd, ctx.parameter)
}

func (ctx *context) parameter() Expression {
	var typeExpr, defaultExpression Expression

	start := ctx.tokenStartPos
	if ctx.currentToken == tokenTypeName {
		typeExpr = ctx.parameterType()
	}

	capturesRest := ctx.currentToken == tokenMultiply
	if capturesRest {
		ctx.nextToken()
	}

	if ctx.currentToken != tokenVariable {
		panic(ctx.parseIssue(parseExpectedVariable))
	}
	variable, ok := ctx.tokenValue.(string)
	if !ok {
		panic(ctx.parseIssue(parseExpectedVariable))
	}
	ctx.nextToken()

	if ctx.currentToken == tokenAssign {
		ctx.nextToken()
		defaultExpression = ctx.expression()
	}
	return ctx.factory.Parameter(
		variable,
		defaultExpression, typeExpr, capturesRest, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) returnParameters() (result []Expression) {
	switch ctx.currentToken {
	case tokenLp, tokenWslp:
		ctx.nextToken()
		return ctx.expressions(tokenRp, ctx.returnParameter)
	default:
		return []Expression{}
	}
}

func (ctx *context) attributeAlias() Expression {
	s := ctx.tokenStartPos
	if i, ok := ctx.identifier(); ok {
		return ctx.factory.String(i, ctx.locator, s, len(i))
	}
	panic(ctx.parseIssue(parseExpectedAttributeName))
}

func (ctx *context) returnParameter() Expression {
	var typeExpr, defaultExpression Expression

	start := ctx.tokenStartPos

	if ctx.currentToken == tokenTypeName {
		typeExpr = ctx.parameterType()
	}
	if ctx.currentToken != tokenVariable {
		panic(ctx.parseIssue(parseExpectedVariable))
	}
	variable, ok := ctx.tokenValue.(string)
	if !ok {
		panic(ctx.parseIssue(parseExpectedVariable))
	}
	ctx.nextToken()

	if ctx.currentToken == tokenAssign {
		ctx.nextToken()
		switch ctx.currentToken {
		case tokenLp, tokenWslp:
			ps := ctx.tokenStartPos
			ctx.nextToken()
			defaultExpression = ctx.factory.Array(ctx.expressions(tokenRp, ctx.attributeAlias), ctx.locator, ps, ps-ctx.Pos())
			ctx.nextToken()
		default:
			defaultExpression = ctx.attributeAlias()
		}
	}
	return ctx.factory.Parameter(
		variable,
		defaultExpression, typeExpr, false, ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) parameterType() Expression {
	start := ctx.tokenStartPos
	typeName := ctx.typeName()
	if typeName == nil {
		panic(ctx.parseIssue(parseExpectedTypeName))
	}

	if ctx.currentToken == tokenLb {
		ctx.nextToken()
		typeArgs := ctx.arrayExpression()
		expr := ctx.factory.Access(typeName, typeArgs, ctx.locator, start, ctx.Pos()-start)
		ctx.nextToken()
		return expr
	}
	return typeName
}

func (ctx *context) typeName() Expression {
	if ctx.currentToken == tokenTypeName {
		name := ctx.factory.QualifiedReference(ctx.tokenString(), ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos)
		ctx.nextToken()
		return name
	}
	return nil
}

func (ctx *context) classExpression(start int) Expression {
	name := strings.TrimPrefix(ctx.className(), `::`)

	// Push to namestack
	ctx.nameStack = append(ctx.nameStack, name)

	var parameterList []Expression
	switch ctx.currentToken {
	case tokenLp, tokenWslp:
		parameterList = ctx.parameterList()
		ctx.nextToken()
	default:
		parameterList = []Expression{}
	}

	var parent string
	if ctx.currentToken == tokenInherits {
		ctx.nextToken()
		if ctx.currentToken == tokenDefault {
			parent = tokenMap[tokenDefault]
			ctx.nextToken()
		} else {
			parent = ctx.className()
		}
	}
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	body := ctx.parse(tokenRc, false)
	ctx.nextToken()

	// Pop namestack
	ctx.nameStack = ctx.nameStack[:len(ctx.nameStack)-1]
	return ctx.addDefinition(ctx.factory.Class(ctx.qualifiedName(name), parameterList, parent, body, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) className() (name string) {
	switch ctx.currentToken {
	case tokenTypeName, tokenIdentifier:
		name = ctx.tokenString()
		ctx.nextToken()
		return
	case tokenString, tokenConcatenatedString:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(parseQuotedNotValidName))
	case tokenClass:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(parseClassNotValidHere))
	default:
		ctx.SetPos(ctx.tokenStartPos)
		panic(ctx.parseIssue(parseExpectedClassName))
	}
}

func (ctx *context) keyword() (word string, ok bool) {
	if ctx.currentToken != tokenBoolean {
		str := tokenMap[ctx.currentToken]
		if _, ok = keywords[str]; ok {
			word = str
		}
	}
	return
}

func (ctx *context) qualifiedName(name string) string {
	return strings.Join(append(ctx.nameStack, name), `::`)
}

func (ctx *context) capabilityMapping(component Expression, kind string) Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	capName := ctx.className()
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	mappings := ctx.attributeOperations()
	ctx.assertToken(tokenRc)
	ctx.nextToken()

	switch ct := component.(type) {
	case *QualifiedReference, *QualifiedName:
		// No action
	case *ReservedWord:
		// All reserved words are lowercase only
		component = ctx.factory.QualifiedName(ctx.qualifiedName(ct.Name()), ctx.locator, component.ByteOffset(), component.ByteLength())
	}
	return ctx.addDefinition(ctx.factory.CapabilityMapping(kind, component, ctx.qualifiedName(capName), mappings, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) siteDefinition() Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	ctx.assertToken(tokenLc)
	ctx.nextToken()
	block := ctx.parse(tokenRc, false)
	ctx.nextToken()
	return ctx.addDefinition(ctx.factory.Site(block, ctx.locator, start, ctx.Pos()-start))
}

func (ctx *context) resourceDefinition(resourceToken int) Expression {
	start := ctx.tokenStartPos
	ctx.nextToken()
	name := ctx.className()

	var parameterList []Expression
	switch ctx.currentToken {
	case tokenLp, tokenWslp:
		parameterList = ctx.parameterList()
		ctx.nextToken()
	default:
		parameterList = []Expression{}
	}

	ctx.assertToken(tokenLc)
	ctx.nextToken()
	body := ctx.parse(tokenRc, false)
	ctx.nextToken()
	var def Expression
	if resourceToken == tokenApplication {
		def = ctx.factory.Application(name, parameterList, body, ctx.locator, start, ctx.Pos()-start)
	} else {
		def = ctx.factory.Definition(name, parameterList, body, ctx.locator, start, ctx.Pos()-start)
	}
	return ctx.addDefinition(def)
}

func (ctx *context) addDefinition(expr Expression) Expression {
	ctx.definitions = append(ctx.definitions, expr.(Definition))
	return expr
}
