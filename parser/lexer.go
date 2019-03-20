package parser

import (
	"bytes"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/lyraproj/issue/issue"
)

// Recursive descent lexer for the Puppet language.

type location struct {
	locator    *Locator
	byteOffset int
}

func (l *location) File() string {
	return l.locator.file
}

func (l *location) Line() int {
	return l.locator.LineForOffset(l.byteOffset)
}

func (l *location) Pos() int {
	return l.locator.PosOnLine(l.byteOffset)
}

func (ctx *context) parseIssue(issueCode issue.Code) issue.Reported {
	return issue.NewReported(issueCode, issue.SeverityError, issue.NoArgs, &location{ctx.locator, ctx.Pos()})
}

func (ctx *context) parseIssue2(issueCode issue.Code, args issue.H) issue.Reported {
	return issue.NewReported(issueCode, issue.SeverityError, args, &location{ctx.locator, ctx.Pos()})
}

const (
	tokenEnd = 0

	// Binary ops
	tokenAssign         = 1
	tokenAddAssign      = 2
	tokenSubtractAssign = 3

	tokenMultiply  = 10
	tokenDivide    = 11
	tokenRemainder = 12
	tokenSubtract  = 13
	tokenAdd       = 14

	tokenLshift = 20
	tokenRshift = 21

	tokenEqual        = 30
	tokenNotEqual     = 31
	tokenLess         = 32
	tokenLessEqual    = 33
	tokenGreater      = 34
	tokenGreaterEqual = 35

	tokenMatch    = 40
	tokenNotMatch = 41

	tokenLcollect  = 50
	tokenLlcollect = 51

	tokenRcollect  = 60
	tokenRrcollect = 61

	tokenFarrow = 70
	tokenParrow = 71

	tokenInEdge     = 72
	tokenInEdgeSub  = 73
	tokenOutEdge    = 74
	tokenOutEdgeSub = 75

	// Unary ops
	tokenNot  = 80
	tokenAt   = 81
	tokenAtat = 82

	// ()
	tokenLp   = 90
	tokenWslp = 91
	tokenRp   = 92

	// []
	tokenLb        = 100
	tokenListstart = 101
	tokenRb        = 102

	// {}
	tokenLc   = 110
	tokenSelc = 111
	tokenRc   = 112

	// | |
	tokenPipe    = 120
	tokenPipeEnd = 121

	// EPP
	tokenEppEnd       = 130
	tokenEppEndTrim   = 131
	tokenRenderExpr   = 132
	tokenRenderString = 133

	// Separators
	tokenComma     = 140
	tokenDot       = 141
	tokenQmark     = 142
	tokenColon     = 143
	tokenSemicolon = 144

	// Strings with semantics
	tokenIdentifier         = 150
	tokenString             = 151
	tokenInteger            = 152
	tokenFloat              = 153
	tokenBoolean            = 154
	tokenConcatenatedString = 155
	tokenHeredoc            = 156
	tokenVariable           = 157
	tokenRegexp             = 158
	tokenTypeName           = 159

	// Keywords
	tokenAnd         = 200
	tokenApplication = 201
	tokenAttr        = 202
	tokenCase        = 203
	tokenClass       = 204
	tokenConsumes    = 205
	tokenDefault     = 206
	tokenDefine      = 207
	tokenFunction    = 208
	tokenIf          = 209
	tokenIn          = 210
	tokenInherits    = 211
	tokenElse        = 212
	tokenElsif       = 213
	tokenNode        = 214
	tokenOr          = 215
	tokenPlan        = 216
	tokenPrivate     = 217
	tokenProduces    = 218
	tokenSite        = 219
	tokenType        = 220
	tokenUndef       = 221
	tokenUnless      = 222
)

func IsKeywordToken(token int) bool {
	return token >= tokenAnd && token <= tokenUnless
}

var tokenMap = map[int]string{
	tokenEnd: `EOF`,

	// Binary ops
	tokenAssign:         `=`,
	tokenAddAssign:      `+=`,
	tokenSubtractAssign: `-=`,

	tokenMultiply:  `*`,
	tokenDivide:    `/`,
	tokenRemainder: `%`,
	tokenSubtract:  `-`,
	tokenAdd:       `+`,

	tokenLshift: `<<`,
	tokenRshift: `>>`,

	tokenEqual:        `==`,
	tokenNotEqual:     `!=`,
	tokenLess:         `<`,
	tokenLessEqual:    `<=`,
	tokenGreater:      `>`,
	tokenGreaterEqual: `>=`,

	tokenMatch:    `=~`,
	tokenNotMatch: `!~`,

	tokenLcollect:  `<|`,
	tokenLlcollect: `<<|`,

	tokenRcollect:  `|>`,
	tokenRrcollect: `|>>`,

	tokenFarrow: `=>`,
	tokenParrow: `+>`,

	tokenInEdge:     `->`,
	tokenInEdgeSub:  `~>`,
	tokenOutEdge:    `<-`,
	tokenOutEdgeSub: `<~`,

	// Unary ops
	tokenNot:  `!`,
	tokenAt:   `@`,
	tokenAtat: `@@`,

	tokenComma: `,`,

	// ()
	tokenLp:   `(`,
	tokenWslp: `(`,
	tokenRp:   `)`,

	// []
	tokenLb:        `[`,
	tokenListstart: `[`,
	tokenRb:        `]`,

	// {}
	tokenLc:   `{`,
	tokenSelc: `{`,
	tokenRc:   `}`,

	// | |
	tokenPipe:    `|`,
	tokenPipeEnd: `|`,

	// EPP
	tokenEppEnd:       `%>`,
	tokenEppEndTrim:   `-%>`,
	tokenRenderExpr:   `<%=`,
	tokenRenderString: `epp text`,

	// Separators
	tokenDot:       `.`,
	tokenQmark:     `?`,
	tokenColon:     `:`,
	tokenSemicolon: `;`,

	// Strings with semantics
	tokenIdentifier:         `identifier`,
	tokenString:             `string literal`,
	tokenInteger:            `integer literal`,
	tokenFloat:              `float literal`,
	tokenBoolean:            `boolean literal`,
	tokenConcatenatedString: `dq string literal`,
	tokenHeredoc:            `heredoc`,
	tokenVariable:           `variable`,
	tokenRegexp:             `regexp`,
	tokenTypeName:           `type name`,

	// Keywords
	tokenAnd:         `and`,
	tokenApplication: `application`,
	tokenAttr:        `attr`,
	tokenCase:        `case`,
	tokenClass:       `class`,
	tokenConsumes:    `consumes`,
	tokenDefault:     `default`,
	tokenDefine:      `define`,
	tokenFunction:    `function`,
	tokenIf:          `if`,
	tokenIn:          `in`,
	tokenInherits:    `inherits`,
	tokenElse:        `else`,
	tokenElsif:       `elsif`,
	tokenNode:        `node`,
	tokenOr:          `or`,
	tokenPlan:        `plan`,
	tokenPrivate:     `private`,
	tokenProduces:    `produces`,
	tokenSite:        `site`,
	tokenType:        `type`,
	tokenUndef:       `undef`,
	tokenUnless:      `unless`,
}

var keywords = map[string]int{
	tokenMap[tokenApplication]: tokenApplication,
	tokenMap[tokenAnd]:         tokenAnd,
	tokenMap[tokenAttr]:        tokenAttr,
	tokenMap[tokenCase]:        tokenCase,
	tokenMap[tokenClass]:       tokenClass,
	tokenMap[tokenConsumes]:    tokenConsumes,
	tokenMap[tokenDefault]:     tokenDefault,
	tokenMap[tokenDefine]:      tokenDefine,
	`false`:                    tokenBoolean,
	tokenMap[tokenFunction]:    tokenFunction,
	tokenMap[tokenElse]:        tokenElse,
	tokenMap[tokenElsif]:       tokenElsif,
	tokenMap[tokenIf]:          tokenIf,
	tokenMap[tokenIn]:          tokenIn,
	tokenMap[tokenInherits]:    tokenInherits,
	tokenMap[tokenNode]:        tokenNode,
	tokenMap[tokenOr]:          tokenOr,
	tokenMap[tokenPlan]:        tokenPlan,
	tokenMap[tokenPrivate]:     tokenPrivate,
	tokenMap[tokenProduces]:    tokenProduces,
	tokenMap[tokenSite]:        tokenSite,
	`true`:                     tokenBoolean,
	tokenMap[tokenType]:        tokenType,
	tokenMap[tokenUndef]:       tokenUndef,
	tokenMap[tokenUnless]:      tokenUnless,
}

var DefaultInstance = Default{}

type Default struct{}

type context struct {
	stringReader
	locator               *Locator
	eppMode               bool
	handleBacktickStrings bool
	handleHexEscapes      bool
	tasks                 bool
	workflow              bool
	nextLineStart         int
	currentToken          int
	beginningOfLine       int
	tokenStartPos         int
	tokenValue            interface{}
	radix                 int
	factory               ExpressionFactory
	nameStack             []string
	definitions           []Definition
}

func (ctx *context) setToken(token int) {
	ctx.currentToken = token
	ctx.tokenValue = nil
}

func (ctx *context) settokenValue(token int, value interface{}) {
	ctx.currentToken = token
	ctx.tokenValue = value
}

func (ctx *context) unterminatedQuote(start int, delimiter rune) issue.Reported {
	ctx.SetPos(start)
	var stringType string
	if delimiter == '"' {
		stringType = `double`
	} else if delimiter == '\'' {
		stringType = `single`
	} else {
		stringType = `backtick`
	}
	return ctx.parseIssue2(lexUnterminatedString, issue.H{`string_type`: stringType})
}

func (ctx *context) nextToken() {
	sz := 0
	scanStart := ctx.Pos()

	c, start := ctx.skipWhite(false)
	ctx.tokenStartPos = start

	switch {
	case '1' <= c && c <= '9':
		ctx.skipDecimalDigits()
		c, sz = ctx.Peek()
		if c == '.' || c == 'e' || c == 'E' {
			ctx.Advance(sz)
			ctx.consumeFloat(start, c)
			break
		}
		if unicode.IsLetter(c) {
			panic(ctx.parseIssue(lexDigitExpected))
		}
		v, _ := strconv.ParseInt(ctx.From(start), 10, 64)
		ctx.settokenValue(tokenInteger, v)
		ctx.radix = 10

	case 'A' <= c && c <= 'Z':
		ctx.consumeQualifiedName(start, tokenTypeName)

	case 'a' <= c && c <= 'z':
		ctx.consumeQualifiedName(start, tokenIdentifier)

	default:
		switch c {
		case 0:
			ctx.setToken(tokenEnd)
		case '=':
			c, sz = ctx.Peek()
			switch c {
			case '=':
				ctx.Advance(sz)
				ctx.setToken(tokenEqual)
			case '~':
				ctx.Advance(sz)
				ctx.setToken(tokenMatch)
			case '>':
				ctx.Advance(sz)
				ctx.setToken(tokenFarrow)
			default:
				ctx.setToken(tokenAssign)
			}
		case '{':
			if ctx.currentToken == tokenQmark {
				ctx.setToken(tokenSelc)
			} else {
				ctx.setToken(tokenLc)
			}

		case '}':
			ctx.setToken(tokenRc)

		case '[':
			// If token is preceded by whitespace or if it's the first token to be parsed, then it's a
			// list rather than parameters to an access expression
			if scanStart < start || start == 0 {
				ctx.setToken(tokenListstart)
				break
			}
			ctx.setToken(tokenLb)

		case ']':
			ctx.setToken(tokenRb)

		case '(':
			// If token is first on line or only preceded by whitespace, then it is not start of parameters
			// in a call.
			savePos := ctx.Pos()
			ctx.SetPos(ctx.beginningOfLine)
			_, firstNonWhite := ctx.skipWhite(false)
			ctx.SetPos(savePos)
			if firstNonWhite == start {
				ctx.setToken(tokenWslp)
			} else {
				ctx.setToken(tokenLp)
			}

		case ')':
			ctx.setToken(tokenRp)

		case ',':
			ctx.setToken(tokenComma)

		case ';':
			ctx.setToken(tokenSemicolon)

		case '.':
			ctx.setToken(tokenDot)

		case '?':
			ctx.setToken(tokenQmark)

		case ':':
			ctx.setToken(tokenColon)
			c, sz = ctx.Peek()
			if c == ':' {
				ctx.Advance(sz)
				c, _ = ctx.Next()
				if isUppercaseLetter(c) {
					ctx.consumeQualifiedName(start, tokenTypeName)
				} else if isLowercaseLetter(c) {
					ctx.consumeQualifiedName(start, tokenIdentifier)
				} else {
					ctx.SetPos(start)
					panic(ctx.parseIssue(lexDoubleColonNotFollowedByName))
				}
			}

		case '-':
			c, sz = ctx.Peek()
			switch c {
			case '=':
				ctx.Advance(sz)
				ctx.setToken(tokenSubtractAssign)
			case '>':
				ctx.Advance(sz)
				ctx.setToken(tokenInEdge)
			case '%':
				if ctx.eppMode {
					ctx.Advance(sz)
					c, sz = ctx.Peek()
					if c == '>' {
						ctx.Advance(sz)
						for c, sz = ctx.Peek(); c == ' ' || c == '\t'; c, sz = ctx.Peek() {
							ctx.Advance(sz)
						}
						if c == '\n' {
							ctx.Advance(sz)
						}
						ctx.consumeEPP()
					} else {
						panic(ctx.parseIssue2(lexInvalidOperator, issue.H{`op`: `-%`}))
					}
					break
				}
				fallthrough

			default:
				ctx.setToken(tokenSubtract)
			}

		case '+':
			c, sz = ctx.Peek()
			if c == '=' {
				ctx.Advance(sz)
				ctx.setToken(tokenAddAssign)
			} else if c == '>' {
				ctx.Advance(sz)
				ctx.setToken(tokenParrow)
			} else {
				ctx.setToken(tokenAdd)
			}

		case '*':
			ctx.setToken(tokenMultiply)

		case '%':
			ctx.setToken(tokenRemainder)
			if ctx.eppMode {
				c, sz = ctx.Peek()
				if c == '>' {
					ctx.Advance(sz)
					ctx.consumeEPP()
				}
			}

		case '!':
			c, sz = ctx.Peek()
			if c == '=' {
				ctx.Advance(sz)
				ctx.setToken(tokenNotEqual)
			} else if c == '~' {
				ctx.Advance(sz)
				ctx.setToken(tokenNotMatch)
			} else {
				ctx.setToken(tokenNot)
			}

		case '>':
			c, sz = ctx.Peek()
			if c == '=' {
				ctx.Advance(sz)
				ctx.setToken(tokenGreaterEqual)
			} else if c == '>' {
				ctx.Advance(sz)
				ctx.setToken(tokenRshift)
			} else {
				ctx.setToken(tokenGreater)
			}

		case '~':
			c, sz = ctx.Peek()
			if c == '>' {
				ctx.Advance(sz)
				ctx.setToken(tokenInEdgeSub)
			} else {
				// Standalone tilde is not an operator in Puppet
				ctx.SetPos(start)
				panic(ctx.parseIssue2(lexUnexpectedToken, issue.H{`token`: `~`}))
			}

		case '@':
			c, sz = ctx.Peek()
			if c == '@' {
				ctx.Advance(sz)
				ctx.setToken(tokenAtat)
			} else if c == '(' {
				ctx.Advance(sz)
				ctx.consumeHeredocString()
			} else {
				ctx.setToken(tokenAt)
			}

		case '<':
			c, sz = ctx.Peek()
			switch c {
			case '=':
				ctx.Advance(sz)
				ctx.setToken(tokenLessEqual)
			case '<':
				ctx.Advance(sz)
				c, sz = ctx.Peek()
				if c == '|' {
					ctx.Advance(sz)
					ctx.setToken(tokenLlcollect)
				} else {
					ctx.setToken(tokenLshift)
				}
			case '|':
				ctx.Advance(sz)
				ctx.setToken(tokenLcollect)
			case '-':
				ctx.Advance(sz)
				ctx.setToken(tokenOutEdge)
			case '~':
				ctx.Advance(sz)
				ctx.setToken(tokenOutEdgeSub)
			case '%':
				if ctx.eppMode {
					ctx.Advance(sz)
					// <%# and <%% has been dealt with in consumeEPP so there's no need to deal with
					// that. Only <%, <%- and <%= can show up here
					c, sz = ctx.Peek()
					switch c {
					case '=':
						ctx.Advance(sz)
						ctx.setToken(tokenRenderExpr)
					case '-':
						ctx.Advance(sz)
						ctx.nextToken()
					default:
						ctx.nextToken()
					}
					break
				}
				fallthrough
			default:
				ctx.setToken(tokenLess)
			}

		case '|':
			c, sz = ctx.Peek()
			switch c {
			case '>':
				ctx.Advance(sz)
				c, sz = ctx.Peek()
				if c == '>' {
					ctx.Advance(sz)
					ctx.setToken(tokenRrcollect)
				} else {
					ctx.setToken(tokenRcollect)
				}
			default:
				if ctx.currentToken == tokenPipe {
					// Empty parameter list
					ctx.setToken(tokenPipeEnd)
				} else {
					pos := ctx.Pos()
					n, _ := ctx.skipWhite(false)
					ctx.SetPos(pos)
					if n == '{' || n == '>' || ctx.eppMode && (n == '%' || n == '-') {
						// A lambda parameter list cannot start with either of these tokens so
						// this must be the end (next is either block body or block return type declaration)
						ctx.setToken(tokenPipeEnd)
					} else {
						ctx.setToken(tokenPipe)
					}
				}
			}

		case '"':
			ctx.consumeDoubleQuotedString()

		case '\'':
			ctx.consumeSingleQuotedString()

		case '/':
			if ctx.isRegexpAcceptable() && ctx.consumeRegexp() {
				return
			}
			ctx.setToken(tokenDivide)

		case '$':
			c, sz = ctx.Peek()
			if c == ':' {
				ctx.Advance(sz)
				c, sz = ctx.Peek()
				if c != ':' {
					ctx.SetPos(start)
					panic(ctx.parseIssue(lexInvalidVariableName))
				}
				ctx.Advance(sz)
				c, sz = ctx.Peek()
			}
			if isLowercaseLetter(c) {
				ctx.Advance(sz)
				ctx.consumeQualifiedName(start, tokenVariable)
			} else if isDecimalDigit(c) {
				ctx.Advance(sz)
				ctx.skipDecimalDigits()
				ctx.tokenValue, _ = strconv.ParseInt(ctx.From(start+1), 10, 64)
			} else if unicode.IsLetter(c) {
				panic(ctx.parseIssue(lexInvalidVariableName))
			} else {
				ctx.tokenValue = ``
			}
			ctx.settokenValue(tokenVariable, ctx.tokenValue)

		case '0':
			ctx.radix = 10
			c, sz = ctx.Peek()

			switch c {
			case 0:
				ctx.settokenValue(tokenInteger, int64(0))
				return

			case 'x', 'X':
				ctx.Advance(sz) // consume 'x'
				hexStart := ctx.Pos()
				c, sz = ctx.Peek()
				for isHexDigit(c) {
					ctx.Advance(sz)
					c, sz = ctx.Peek()
				}
				if ctx.Pos() == hexStart || isLetter(c) {
					panic(ctx.parseIssue(lexHexdigitExpected))
				}
				v, _ := strconv.ParseInt(ctx.From(hexStart), 16, 64)
				ctx.radix = 16
				ctx.settokenValue(tokenInteger, v)

			case '.', 'e', 'E':
				// 0[.eE]<something>
				ctx.Advance(sz)
				ctx.consumeFloat(start, c)

			default:
				octalStart := ctx.Pos()
				for isOctalDigit(c) {
					ctx.Advance(sz)
					c, sz = ctx.Peek()
				}
				if isDecimalDigit(c) || unicode.IsLetter(c) {
					panic(ctx.parseIssue(lexOctaldigitExpected))
				}
				if ctx.Pos() > octalStart {
					v, _ := strconv.ParseInt(ctx.From(octalStart), 8, 64)
					ctx.radix = 8
					ctx.settokenValue(tokenInteger, v)
				} else {
					ctx.settokenValue(tokenInteger, int64(0))
				}
			}

		case '`':
			if ctx.handleBacktickStrings {
				ctx.consumeBacktickedString()
				break
			}
			fallthrough

		default:
			ctx.SetPos(start)
			panic(ctx.parseIssue2(lexUnexpectedToken, issue.H{`token`: string(c)}))
		}
	}
}

// Skips to next non-whitespace character and returns that character and its start position. Comments are treated
// as whitespaces and will be skipped over
func (ctx *context) skipWhite(breakOnNewLine bool) (c rune, start int) {
	commentStart := 0
	commentStartPos := 0
	for {
		c, start = ctx.Next()
		switch c {
		case 0:
			if commentStart == '*' {
				ctx.SetPos(commentStartPos)
				panic(ctx.parseIssue(lexUnterminatedComment))
			}
			return
		case '\n':
			if commentStart == '*' {
				continue
			}
			if breakOnNewLine {
				ctx.SetPos(start)
				return
			}
			if ctx.nextLineStart >= 0 {
				ctx.SetPos(ctx.nextLineStart)
				ctx.nextLineStart = -1
			}
			if commentStart == '#' {
				commentStart = 0
			}
			ctx.beginningOfLine = ctx.Pos()

		case '#':
			if commentStart == 0 {
				commentStart = '#'
				commentStartPos = start
			}

		case '/':
			if commentStart == 0 {
				tc, sz := ctx.Peek()
				if tc == '*' {
					ctx.Advance(sz)
					commentStart = '*'
					commentStartPos = start
					continue
				}
				return
			}

		case '*':
			if commentStart == '#' {
				continue
			}
			if commentStart == '*' {
				tc, sz := ctx.Peek()
				if tc == '/' {
					ctx.Advance(sz)
					commentStart = 0
				}
				continue
			}
			return

		case ' ', '\r', '\t':
			continue

		default:
			if commentStart == 0 {
				return
			}
		}
	}
}

// Skips to next non-whitespace or newline character and returns that character and its start position without
// comment recognition
func (ctx *context) skipWhiteInLiteral() (c rune, start int) {
	for {
		c, start = ctx.Next()
		switch c {
		case 0:
			return
		case ' ', '\r', '\t':
			continue
		default:
			return
		}
	}
}

func isDecimalDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isOctalDigit(c rune) bool {
	return c >= '0' && c <= '7'
}

func isHexDigit(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f'
}

func isLetter(c rune) bool {
	return isUppercaseLetter(c) || isLowercaseLetter(c) || c == '_'
}

func isLetterOrDigit(c rune) bool {
	return isDecimalDigit(c) || isLetter(c)
}

func isLowercaseLetter(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isUppercaseLetter(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func (ctx *context) consumeQualifiedName(start int, token int) {
	lastStartsWithUnderscore := false
	hasDash := false
outer:
	for {
		c, n := ctx.Peek()
		for isLetterOrDigit(c) {
			ctx.Advance(n)
			c, n = ctx.Peek()
		}

		if c == '-' && token == tokenIdentifier {
			// Valid only if a letter or digit is present before end of name
			i := ctx.Pos() + n
			for {
				c, n = ctx.PeekAt(i)
				i += n
				if isLetterOrDigit(c) {
					hasDash = true
					ctx.SetPos(i)
					continue outer
				}
				if c != '-' {
					break outer
				}
			}
		}

		if c != ':' {
			break
		}

		nameEnd := ctx.Pos()
		ctx.Advance(n)
		c, n = ctx.Peek()

		if c != ':' {
			// Single ':' after a name is ok. Should not be consumed
			ctx.SetPos(nameEnd)
			break
		}

		ctx.Advance(n)
		c, n = ctx.Peek()
		if token == tokenTypeName && isUppercaseLetter(c) ||
			token != tokenTypeName && (isLowercaseLetter(c) ||
				token == tokenVariable && c == '_') {
			// Next segment starts here and only last segment is allowed to
			// start with underscore
			if !lastStartsWithUnderscore {
				ctx.Advance(n)
				lastStartsWithUnderscore = c == '_'
				continue
			}
		}

		ctx.SetPos(start)
		issueCode := issue.Code(lexInvalidName)
		if token == tokenTypeName {
			issueCode = lexInvalidTypeName
		} else if token == tokenVariable {
			issueCode = lexInvalidVariableName
		}
		panic(ctx.parseIssue(issueCode))
	}

	if token == tokenVariable {
		start++ // skip leading '$´
	}

	word := ctx.From(start)

	if token == tokenIdentifier {
		if hasDash {
			token = tokenString
		} else if kwToken, ok := keywords[word]; ok {
			switch kwToken {
			case tokenBoolean:
				ctx.settokenValue(kwToken, word == `true`)
				return
			case tokenDefault:
				ctx.settokenValue(kwToken, DefaultInstance)
				return
			case tokenPlan:
				if ctx.tasks {
					token = kwToken
				}
			default:
				token = kwToken
			}
		}
	}

	ctx.settokenValue(token, word)
}

func (ctx *context) consumeFloat(start int, d rune) {
	if ctx.skipDecimalDigits() == 0 {
		panic(ctx.parseIssue(lexDigitExpected))
	}
	c, n := ctx.Peek()
	if d == '.' {
		// Check for 'e'
		if c == 'e' || c == 'E' {
			ctx.Advance(n)
			if ctx.skipDecimalDigits() == 0 {
				panic(ctx.parseIssue(lexDigitExpected))
			}
			c, _ = ctx.Peek()
		}
	}
	if unicode.IsLetter(c) {
		panic(ctx.parseIssue(lexDigitExpected))
	}
	v, _ := strconv.ParseFloat(ctx.From(start), 64)
	ctx.settokenValue(tokenFloat, v)
}

func (ctx *context) skipDecimalDigits() (digitCount int) {
	digitCount = 0
	c, n := ctx.Peek()
	if c == '-' || c == '+' {
		ctx.Advance(n)
		c, n = ctx.Peek()
	}
	for isDecimalDigit(c) {
		ctx.Advance(n)
		c, n = ctx.Peek()
		digitCount++
	}
	return
}

type escapeHandler func(buffer *bytes.Buffer, ctx *context, c rune)

func (ctx *context) consumeDelimitedString(delimiter rune, delimiterStart int, interpolateSegments []Expression, handler escapeHandler) (segments []Expression) {
	buf := bytes.NewBufferString(``)
	ec, start := ctx.Next()
	segments = interpolateSegments
	for {
		switch ec {
		case 0:
			if delimiter != '/' {
				panic(ctx.unterminatedQuote(delimiterStart, delimiter))
			}
			ctx.setToken(tokenDivide)
			return

		case delimiter:
			ctx.settokenValue(tokenString, buf.String())
			return

		case '\\':
			ec, _ = ctx.Next()
			switch ec {
			case 0:
				panic(ctx.unterminatedQuote(delimiterStart, delimiter))

			case delimiter:
				buf.WriteRune(delimiter)
				ec, _ = ctx.Next()
				continue

			default:
				handler(buf, ctx, ec)
				ec, _ = ctx.Next()
				continue
			}

		case '$':
			if segments != nil {
				segments = ctx.handleInterpolation(start, segments, buf)
				ec, start = ctx.Next()
				continue
			}

			// treat '$' just like any other character when segments is nil
			fallthrough
		default:
			buf.WriteRune(ec)
			ec, _ = ctx.Next()
		}
	}
}

func (ctx *context) consumeEPP() {
	buf := bytes.NewBufferString(``)
	lastNonWS := 0
	var sz int
	for ec, start := ctx.Next(); ec != 0; ec, start = ctx.Next() {
		switch ec {
		case '<':
			ec, sz = ctx.Peek()
			if ec != '%' {
				buf.WriteByte('<')
				lastNonWS = buf.Len()
				continue
			}
			ctx.Advance(sz)

			ec, sz = ctx.Peek()
			switch ec {
			case '%':
				// <%% is verbatim <%
				ctx.Advance(sz)
				buf.WriteString(`<%`)
				lastNonWS = buf.Len()
				continue

			case '#':
				ctx.Advance(sz)
				prev := ec
				foundEnd := false
				for ec, _ = ctx.Next(); ec != 0; ec, _ = ctx.Next() {
					if ec == '%' {
						ec, sz = ctx.Peek()
						if ec == '>' && prev != '%' {
							ctx.Advance(sz)
							foundEnd = true
							break
						}
					}
					prev = ec
				}
				if !foundEnd {
					ctx.SetPos(start)
					panic(ctx.parseIssue(lexUnbalancedEppComment))
				}
				continue

			case '-':
				// trim whitespaces leading up to <%-
				ctx.Advance(sz)
				buf.Truncate(lastNonWS)

			case '=':
				ctx.Advance(sz)
			}
			ctx.SetPos(start) // Next token will be TOKEN_RENDER_EXPR
			ctx.settokenValue(tokenRenderString, buf.String())
			if buf.Len() == 0 {
				ctx.nextToken()
			}
			return

		case ' ', '\t':
			buf.WriteRune(ec)

		case '%':
			// %%> is verbatim %>
			buf.WriteByte('%')
			ec, sz = ctx.Peek()
			if ec == '%' {
				ctx.Advance(sz)
				ec, sz = ctx.Peek()
				if ec == '>' {
					ctx.Advance(sz)
					buf.WriteByte('>')
				} else {
					buf.WriteByte('%')
				}
			}
			lastNonWS = buf.Len()

		default:
			buf.WriteRune(ec)
			lastNonWS = buf.Len()
		}
	}
	if buf.Len() == 0 {
		ctx.setToken(tokenEnd)
	} else {
		ctx.settokenValue(tokenRenderString, buf.String())
	}
}

// Called after a '$' has been encountered on input.
//   - Extracts the preceding string from the buf and resets buf.
//   - Unless the string is empty, adds a StringExpression that represents the string to the segments slice
//   - Asks the context to perform interpolation and adds the resulting expression to the segments slice
//   - Sets the tokenStartPos to the position just after the end of the interpolation expression
//
func (ctx *context) handleInterpolation(start int, segments []Expression, buf *bytes.Buffer) []Expression {
	precedingString := buf.String()
	buf.Reset()

	if precedingString != `` {
		segments = append(segments, ctx.factory.String(precedingString, ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos))
	}
	segments = append(segments, ctx.interpolate(start))
	ctx.tokenStartPos = ctx.Pos()
	return segments
}

// Performs interpolation starting at the current position (which must point at the starting '$' character)
// and returns the resulting expression
func (ctx *context) interpolate(start int) Expression {
	c, sz := ctx.Peek()
	if c == '{' {
		ctx.Advance(sz)

		// Call context recursively and expect the ending token to be the ending curly brace
		ctx.nextToken()
		expr := ctx.parse(tokenRc, true)

		// If the result is a single QualifiedName or an AccessExpression or CallMemberExpression with a QualifiedName
		// as the LHS, then it's actually a variable since the `${var}` is the same as `$var`
		switch expr.(type) {
		case *QualifiedName:
			expr = ctx.factory.Variable(expr, ctx.locator, start, ctx.Pos()-start)
		case *AccessExpression:
			access := expr.(*AccessExpression)
			if identifier, ok := access.operand.(*QualifiedName); ok {
				expr = ctx.factory.Access(
					ctx.factory.Variable(identifier, ctx.locator, start, identifier.ByteLength()+1),
					access.keys, ctx.locator, start, access.ByteLength()+1)
			}
		case *CallMethodExpression:
			call := expr.(*CallMethodExpression)
			if ne, ok := call.functor.(*NamedAccessExpression); ok {
				modNe := ctx.convertNamedAccessLHS(ne, start)
				if modNe != ne {
					expr = ctx.factory.CallMethod(modNe, call.arguments, call.lambda, ctx.locator, start, call.ByteLength()+1)
				}
			}
		}
		return ctx.factory.Text(expr, ctx.locator, start, ctx.Pos()-start)
	}

	// Not delimited by curly braces. Must be a single identifier then
	ctx.setToken(tokenVariable)
	if c == ':' || isLowercaseLetter(c) || isDecimalDigit(c) {
		ctx.nextToken()
	}
	if ctx.currentToken != tokenIdentifier {
		ctx.SetPos(start)
		panic(ctx.parseIssue(lexMalformedInterpolation))
	}
	textExpr := ctx.factory.QualifiedName(ctx.tokenValue.(string), ctx.locator, start+1, ctx.Pos()-(start+1))
	return ctx.factory.Text(ctx.factory.Variable(textExpr, ctx.locator, start, ctx.Pos()-start), ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) convertNamedAccessLHS(expr *NamedAccessExpression, start int) Expression {
	switch lhs := expr.lhs.(type) {
	case *QualifiedName:
		return ctx.factory.NamedAccess(
			ctx.factory.Variable(lhs, ctx.locator, start, lhs.ByteLength()+1),
			expr.rhs, ctx.locator, start, expr.ByteLength()+1)
	case *AccessExpression:
		var lhe Expression = lhs
		if identifier, ok := lhs.operand.(*QualifiedName); ok {
			lhe = ctx.factory.Access(
				ctx.factory.Variable(identifier, ctx.locator, start, identifier.ByteLength()+1),
				lhs.keys, ctx.locator, start, lhs.ByteLength()+1)
		}
		return ctx.factory.NamedAccess(lhe, expr.rhs, ctx.locator, start, expr.ByteLength()+1)
	case *NamedAccessExpression:
		return ctx.factory.NamedAccess(
			ctx.convertNamedAccessLHS(lhs, start),
			expr.rhs, ctx.locator, start, expr.ByteLength()+1)
	case *CallMethodExpression:
		lhe := ctx.factory.CallMethod(
			ctx.convertNamedAccessLHS(lhs.functor.(*NamedAccessExpression), start),
			lhs.arguments, lhs.lambda, ctx.locator, start, lhs.ByteLength()+1).(*CallMethodExpression)
		return ctx.factory.NamedAccess(lhe, expr.rhs, ctx.locator, start, expr.ByteLength()+1)
	}
	return expr
}

func (ctx *context) consumeBacktickedString() {
	start := ctx.Pos()
	c, sz := ctx.Peek()
	for c != 0 && c != '`' {
		ctx.Advance(sz)
		c, sz = ctx.Peek()
	}
	if c == 0 {
		panic(ctx.unterminatedQuote(start-1, '`'))
	}
	ctx.settokenValue(tokenString, ctx.From(start))
	ctx.Advance(sz)
}

func (ctx *context) consumeDoubleQuotedString() {
	var segments []Expression
	if ctx.factory != nil {
		segments = make([]Expression, 0, 4)
	}
	segments = ctx.consumeDelimitedString('"', ctx.Pos()-1, segments,
		func(buf *bytes.Buffer, ctx *context, ec rune) {
			switch ec {
			case '\\', '\'':
				buf.WriteRune(ec)
			case '$':
				if ctx.factory == nil {
					buf.WriteRune('\\')
				}
				buf.WriteRune(ec)
			case 'n':
				buf.WriteRune('\n')
			case 'r':
				buf.WriteRune('\r')
			case 't':
				buf.WriteRune('\t')
			case 's':
				buf.WriteRune(' ')
			case 'u':
				ctx.appendUnicode(buf)
			case 'x':
				if ctx.handleHexEscapes {
					ctx.appendHexadec(buf)
					break
				}
				fallthrough
			default:
				// Unrecognized escape sequence. Treat as literal backslash
				buf.WriteRune('\\')
				buf.WriteRune(ec)
			}
		})
	if ctx.factory == nil {
		// currentToken will be TOKEN_STRING
		return
	}

	if len(segments) > 0 {
		// Result of the consumeDelimitedString is just the tail
		tail := ctx.tokenValue.(string)
		if tail != `` {
			segments = append(segments, ctx.factory.String(tail, ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos))
		}
	} else {
		segments = append(segments, ctx.factory.String(ctx.tokenValue.(string), ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos))
	}
	firstPos := segments[0].ByteOffset()
	if len(segments) == 1 {
		if _, ok := segments[0].(*LiteralString); ok {
			// Avoid turning a single string literal into a concatenated string
			return
		}
	}
	ctx.settokenValue(tokenConcatenatedString, ctx.factory.ConcatenatedString(segments, ctx.locator, firstPos, ctx.Pos()-firstPos))
}

func (ctx *context) consumeSingleQuotedString() {
	ctx.consumeDelimitedString('\'', ctx.Pos()-1, nil, func(buf *bytes.Buffer, ctx *context, ec rune) {
		buf.WriteRune('\\')
		if ec != '\\' {
			buf.WriteRune(ec)
		}
	})
}

// Finds end of regexp. If found, sets the tokenValue to the string starting at the position that was current
// when this function was called (should be pointing at the character just after leading '/') and the position
// of the ending '/' (not included in string).
// The method returns true if a regexp was found, false otherwise
func (ctx *context) consumeRegexp() bool {
	start := ctx.Pos()
	ctx.consumeDelimitedString('/', start-1, nil, func(buf *bytes.Buffer, ctx *context, ec rune) {
		buf.WriteRune('\\')
		buf.WriteRune(ec)
	})
	if ctx.currentToken == tokenString {
		ctx.currentToken = tokenRegexp
		return true
	}
	ctx.SetPos(start)
	return false
}

func (ctx *context) consumeHeredocString() {
	var (
		c     rune
		n     int
		tag   string
		flags []byte
	)
	escapeStart := -1
	quoteStart := -1
	syntaxStart := -1
	heredocTagEnd := -1
	syntax := ``
	start := ctx.Pos()
	heredocStart := ctx.Pos() - 2 // Backtrack '@' and '('

findTagEnd:
	for {
		c, n = ctx.Peek()
		switch c {
		case 0, '\n':
			ctx.SetPos(heredocStart)
			panic(ctx.parseIssue(lexHeredocDeclUnterminated))

		case ')':
			if syntaxStart > 0 {
				syntax = ctx.From(syntaxStart)
			}
			if escapeStart > 0 {
				flags = ctx.extractFlags(escapeStart)
			}
			if tag == `` {
				tag = ctx.From(start)
			}
			ctx.Advance(n)
			heredocTagEnd = ctx.Pos()
			break findTagEnd

		case ':':
			if syntaxStart > 0 {
				panic(ctx.parseIssue(lexHeredocMultipleSyntax))
			}
			if tag == `` {
				tag = ctx.From(start)
			}
			ctx.Advance(n)
			syntaxStart = ctx.Pos()

		case '/':
			if escapeStart > 0 {
				panic(ctx.parseIssue(lexHeredocMultipleEscape))
			}
			if tag == `` {
				tag = ctx.From(start)
			} else if syntaxStart > 0 {
				syntax = ctx.From(syntaxStart)
				syntaxStart = -1
			}
			ctx.Advance(n)
			escapeStart = ctx.Pos()

		case '"':
			if tag != `` {
				panic(ctx.parseIssue(lexHeredocMultipleTag))
			}
			ctx.Advance(n)
			quoteStart = ctx.Pos()
		findEndQuote:
			for {
				c, n = ctx.Peek()
				switch c {
				case 0, '\n':
					ctx.SetPos(heredocStart)
					panic(ctx.parseIssue(lexHeredocDeclUnterminated))
				case '"':
					break findEndQuote
				default:
					ctx.Advance(n)
				}
			}
			if quoteStart == ctx.Pos() {
				ctx.SetPos(heredocStart)
				panic(ctx.parseIssue(lexHeredocEmptyTag))
			}
			tag = ctx.From(quoteStart)
			ctx.Advance(n)
		default:
			ctx.Advance(n)
		}
	}

	if tag == `` {
		ctx.SetPos(heredocStart)
		panic(ctx.parseIssue(lexHeredocEmptyTag))
	}

	// Find where actual text starts
	heredocContentStart := -1
	c, sz := ctx.Peek()
findStartOfText:
	for {
		switch c {
		case 0:
			ctx.SetPos(heredocStart)
			panic(ctx.parseIssue(lexHeredocUnterminated))

		case '#':
			c, _ = ctx.skipWhite(true)

		case '/':
			n = ctx.Pos()
			ctx.Advance(sz)
			c, _ = ctx.Next()
			if c == '*' {
				ctx.SetPos(n) // rewind to comment start
				c, _ = ctx.skipWhite(true)
			}

		case '\n':
			if ctx.nextLineStart >= 0 {
				ctx.SetPos(ctx.nextLineStart)
				ctx.nextLineStart = -1
			} else {
				ctx.Advance(sz)
			}
			heredocContentStart = ctx.Pos()
			break findStartOfText

		default:
			ctx.Advance(sz)
			c, sz = ctx.Peek()
		}
	}

	suppressLastNL := false
	heredocContentEnd := -1
	heredocEnd := -1
	indentStrip := 0
	tagLen := len(tag)

	// Find end of heredoc and heredoc content
	tagStart, _ := utf8.DecodeRuneInString(tag)
findEndOfText:
	for {
		switch c {
		case 0:
			ctx.SetPos(heredocStart)
			panic(ctx.parseIssue(lexHeredocUnterminated))

		case '\n':
			lineStart := ctx.Pos()
			c, n = ctx.skipWhiteInLiteral()
			switch c {
			case 0:
				ctx.SetPos(heredocStart)
				panic(ctx.parseIssue(lexHeredocUnterminated))

			case '|':
				indentStrip = n - lineStart
				c, n = ctx.skipWhiteInLiteral()
				if c != '-' {
					break
				}
				fallthrough

			case '-':
				suppressLastNL = true
				c, n = ctx.skipWhiteInLiteral()
			}

			if c != tagStart {
				continue
			}

			expr := ctx.Text()
			tagEnd := n + tagLen
			if tagEnd <= len(expr) && tag == expr[n:tagEnd] {
				// tag found if rest of line is whitespace
				ctx.SetPos(tagEnd)
				c, n = ctx.skipWhiteInLiteral()
				heredocEnd = n
				if c == '\n' || c == 0 {
					heredocContentEnd = lineStart
					if suppressLastNL {
						heredocContentEnd--
						if expr[heredocContentEnd-1] == '\r' {
							heredocContentEnd--
						}
					}
					break findEndOfText
				}
			}
		default:
			c, _ = ctx.Next()
		}
	}

	var heredoc string
	if flags != nil || quoteStart >= 0 || indentStrip > 0 {
		ctx.SetPos(heredocContentStart)
		var segments []Expression
		if quoteStart >= 0 && ctx.factory != nil {
			segments = make([]Expression, 0, 4)
		}
		heredoc, segments = ctx.applyEscapes(heredocContentEnd, indentStrip, flags, segments)
		if len(segments) > 0 {
			if len(heredoc) > 0 {
				segments = append(segments, ctx.factory.String(heredoc, ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos))
			}
			ctx.SetPos(heredocTagEnd)          // Normal parsing continues here
			ctx.nextLineStart = heredocEnd + 1 // and next newline will jump to here
			textExpr := ctx.factory.ConcatenatedString(segments, ctx.locator, heredocContentStart, heredocContentEnd-heredocContentStart)
			ctx.settokenValue(tokenHeredoc, ctx.factory.Heredoc(textExpr, syntax, ctx.locator, heredocStart, heredocContentEnd-heredocStart))
			return
		}
	} else {
		ctx.SetPos(heredocContentEnd)
		heredoc = ctx.From(heredocContentStart)
	}

	ctx.SetPos(heredocTagEnd)          // Normal parsing continues here
	ctx.nextLineStart = heredocEnd + 1 // and next newline will jump to here
	if ctx.factory != nil {
		textExpr := ctx.factory.String(heredoc, ctx.locator, heredocContentStart, heredocContentEnd-heredocContentStart)
		ctx.settokenValue(tokenHeredoc, ctx.factory.Heredoc(textExpr, syntax, ctx.locator, heredocStart, heredocContentEnd-heredocStart))
	} else {
		ctx.settokenValue(tokenString, heredoc)
	}
}

func (ctx *context) extractFlags(start int) []byte {
	s := ctx.From(start)
	top := len(s)
	flags := make([]byte, top)
	for idx := 0; idx < top; idx++ {
		flag := s[idx]
		switch flag {
		case 't', 'r', 'n', 's', 'u', '$':
			flags[idx] = flag
		case 'L':
			flags[idx] = '\n'
		default:
			ctx.SetPos(start)
			panic(ctx.parseIssue2(lexHeredocIllegalEscape, issue.H{`flag`: string(flag)}))
		}
	}
	return flags
}

func (ctx *context) applyEscapes(end int, indentStrip int, flags []byte, interpolateSegments []Expression) (heredoc string, segments []Expression) {
	bld := bytes.NewBufferString(``)
	segments = interpolateSegments
	ctx.stripIndent(indentStrip)
	for c, start := ctx.Next(); c != 0 && start < end; c, start = ctx.Next() {
		if c != '\\' {
			if c == '$' && segments != nil {
				segments = ctx.handleInterpolation(start, segments, bld)
			} else {
				bld.WriteRune(c)
				if c == '\n' {
					ctx.stripIndent(indentStrip)
				}
			}
			continue
		}

		c, start = ctx.Next()
		if start >= end {
			bld.WriteByte('\\')
			break
		}

		escaped := false
		if c < utf8.RuneSelf {
			bc := byte(c)
			fi := len(flags) - 1
			for fi >= 0 {
				if flags[fi] == bc {
					escaped = true
					break
				}
				fi--
			}
		}
		if !escaped {
			bld.WriteRune('\\')
			if c == '$' && segments != nil {
				segments = ctx.handleInterpolation(start, segments, bld)
			} else {
				bld.WriteRune(c)
				if c == '\n' {
					ctx.stripIndent(indentStrip)
				}
			}
			continue
		}

		switch c {
		case 'r':
			bld.WriteRune('\r')
		case 'n':
			bld.WriteRune('\n')
		case 't':
			bld.WriteRune('\t')
		case 's':
			bld.WriteRune(' ')
		case 'u':
			ctx.appendUnicode(bld)
		case '\n':
			ctx.stripIndent(indentStrip)
		default:
			bld.WriteRune(c)
		}
	}
	heredoc = bld.String()
	return
}

func (ctx *context) stripIndent(indentStrip int) {
	start := ctx.Pos()
	for indentStrip > 0 {
		if c, s := ctx.Peek(); c == '\t' || c == ' ' {
			ctx.Advance(s)
			indentStrip--
			continue
		}
		// Lines that cannot have their indent stripped i full, does not
		// get it stripped at all
		ctx.SetPos(start)
		break
	}
}

func (ctx *context) appendHexadec(buf *bytes.Buffer) {
	// Must be XX (a two-digit hex number)
	d, start := ctx.Next()
	if isHexDigit(d) {
		d, _ := ctx.Next()
		if !isHexDigit(d) {
			ctx.SetPos(start - 2)
			panic(ctx.parseIssue(lexMalformedHexEscape))
		}
	}
	r, _ := strconv.ParseInt(ctx.From(start), 16, 64)
	buf.WriteByte(byte(r))
}

func (ctx *context) appendUnicode(buf *bytes.Buffer) {
	ec, start := ctx.Next()
	if isHexDigit(ec) {
		// Must be XXXX (a four-digit hex number)
		for i := 1; i < 4; i++ {
			digit, _ := ctx.Next()
			if !isHexDigit(digit) {
				ctx.SetPos(start - 2)
				panic(ctx.parseIssue(lexMalformedUnicodeEscape))
			}
		}
		r, _ := strconv.ParseInt(ctx.From(start), 16, 32)
		buf.WriteRune(rune(r))
		return
	}

	if ec != '{' {
		ctx.SetPos(start - 2)
		panic(ctx.parseIssue(lexMalformedUnicodeEscape))
	}

	// Must be {XXxxxx} (a hex number between two and six digits
	hexStart := ctx.Pos()
	ec, n := ctx.Peek()
	for isHexDigit(ec) {
		ctx.Advance(n)
		ec, n = ctx.Peek()
	}
	uLen := ctx.Pos() - hexStart
	if !(uLen >= 2 && uLen <= 6 && ec == '}') {
		ctx.SetPos(start - 2)
		panic(ctx.parseIssue(lexMalformedUnicodeEscape))
	}

	r, _ := strconv.ParseInt(ctx.From(hexStart), 16, 32)
	ctx.Advance(n) // Skip terminating '}'
	buf.WriteRune(rune(r))
}

func (ctx *context) isRegexpAcceptable() bool {
	switch ctx.currentToken {
	// Operands that can be followed by TOKEN_DIVIDE
	case tokenRp, tokenRb, tokenTypeName, tokenIdentifier, tokenBoolean, tokenInteger, tokenFloat, tokenString,
		tokenHeredoc, tokenConcatenatedString, tokenRegexp, tokenVariable:
		return false
	default:
		return true
	}
}
