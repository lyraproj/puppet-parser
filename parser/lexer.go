package parser

import (
	"bytes"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/puppetlabs/go-parser/issue"
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

func (ctx *context) parseIssue(issueCode issue.Code) *issue.Reported {
	return issue.NewReported(issueCode, issue.SEVERITY_ERROR, issue.NO_ARGS, &location{ctx.locator, ctx.Pos()})
}

func (ctx *context) parseIssue2(issueCode issue.Code, args issue.H) *issue.Reported {
	return issue.NewReported(issueCode, issue.SEVERITY_ERROR, args, &location{ctx.locator, ctx.Pos()})
}

const (
	TOKEN_END = 0

	// Binary ops
	TOKEN_ASSIGN          = 1
	TOKEN_ADD_ASSIGN      = 2
	TOKEN_SUBTRACT_ASSIGN = 3

	TOKEN_MULTIPLY  = 10
	TOKEN_DIVIDE    = 11
	TOKEN_REMAINDER = 12
	TOKEN_SUBTRACT  = 13
	TOKEN_ADD       = 14

	TOKEN_LSHIFT = 20
	TOKEN_RSHIFT = 21

	TOKEN_EQUAL         = 30
	TOKEN_NOT_EQUAL     = 31
	TOKEN_LESS          = 32
	TOKEN_LESS_EQUAL    = 33
	TOKEN_GREATER       = 34
	TOKEN_GREATER_EQUAL = 35

	TOKEN_MATCH     = 40
	TOKEN_NOT_MATCH = 41

	TOKEN_LCOLLECT  = 50
	TOKEN_LLCOLLECT = 51

	TOKEN_RCOLLECT  = 60
	TOKEN_RRCOLLECT = 61

	TOKEN_FARROW = 70
	TOKEN_PARROW = 71

	TOKEN_IN_EDGE      = 72
	TOKEN_IN_EDGE_SUB  = 73
	TOKEN_OUT_EDGE     = 74
	TOKEN_OUT_EDGE_SUB = 75

	// Unary ops
	TOKEN_NOT  = 80
	TOKEN_AT   = 81
	TOKEN_ATAT = 82

	// ()
	TOKEN_LP   = 90
	TOKEN_WSLP = 91
	TOKEN_RP   = 92

	// []
	TOKEN_LB        = 100
	TOKEN_LISTSTART = 101
	TOKEN_RB        = 102

	// {}
	TOKEN_LC   = 110
	TOKEN_SELC = 111
	TOKEN_RC   = 112

	// | |
	TOKEN_PIPE = 120
	TOKEN_PIPE_END = 121

	// EPP
	TOKEN_EPP_END       = 130
	TOKEN_EPP_END_TRIM  = 131
	TOKEN_RENDER_EXPR   = 132
	TOKEN_RENDER_STRING = 133

	// Separators
	TOKEN_COMMA     = 140
	TOKEN_DOT       = 141
	TOKEN_QMARK     = 142
	TOKEN_COLON     = 143
	TOKEN_SEMICOLON = 144

	// Strings with semantics
	TOKEN_IDENTIFIER          = 150
	TOKEN_STRING              = 151
	TOKEN_INTEGER             = 152
	TOKEN_FLOAT               = 153
	TOKEN_BOOLEAN             = 154
	TOKEN_CONCATENATED_STRING = 155
	TOKEN_HEREDOC             = 156
	TOKEN_VARIABLE            = 157
	TOKEN_REGEXP              = 158
	TOKEN_TYPE_NAME           = 159

	// Keywords
	TOKEN_AND         = 200
	TOKEN_APPLICATION = 201
	TOKEN_ATTR        = 202
	TOKEN_CASE        = 203
	TOKEN_CLASS       = 204
	TOKEN_CONSUMES    = 205
	TOKEN_DEFAULT     = 206
	TOKEN_DEFINE      = 207
	TOKEN_FUNCTION    = 208
	TOKEN_IF          = 209
	TOKEN_IN          = 210
	TOKEN_INHERITS    = 211
	TOKEN_ELSE        = 212
	TOKEN_ELSIF       = 213
	TOKEN_NODE        = 214
	TOKEN_OR          = 215
	TOKEN_PLAN        = 216
	TOKEN_PRIVATE     = 217
	TOKEN_PRODUCES    = 218
	TOKEN_SITE        = 219
	TOKEN_TYPE        = 220
	TOKEN_UNDEF       = 221
	TOKEN_UNLESS      = 222
)

func IsKeywordToken(token int) bool {
	return token >= TOKEN_AND && token <= TOKEN_UNLESS
}

var tokenMap = map[int]string{
	TOKEN_END: `EOF`,

	// Binary ops
	TOKEN_ASSIGN:          `=`,
	TOKEN_ADD_ASSIGN:      `+=`,
	TOKEN_SUBTRACT_ASSIGN: `-=`,

	TOKEN_MULTIPLY:  `*`,
	TOKEN_DIVIDE:    `/`,
	TOKEN_REMAINDER: `%`,
	TOKEN_SUBTRACT:  `-`,
	TOKEN_ADD:       `+`,

	TOKEN_LSHIFT: `<<`,
	TOKEN_RSHIFT: `>>`,

	TOKEN_EQUAL:         `==`,
	TOKEN_NOT_EQUAL:     `!=`,
	TOKEN_LESS:          `<`,
	TOKEN_LESS_EQUAL:    `<=`,
	TOKEN_GREATER:       `>`,
	TOKEN_GREATER_EQUAL: `>=`,

	TOKEN_MATCH:     `=~`,
	TOKEN_NOT_MATCH: `!~`,

	TOKEN_LCOLLECT:  `<|`,
	TOKEN_LLCOLLECT: `<<|`,

	TOKEN_RCOLLECT:  `|>`,
	TOKEN_RRCOLLECT: `|>>`,

	TOKEN_FARROW: `=>`,
	TOKEN_PARROW: `+>`,

	TOKEN_IN_EDGE:      `->`,
	TOKEN_IN_EDGE_SUB:  `~>`,
	TOKEN_OUT_EDGE:     `<-`,
	TOKEN_OUT_EDGE_SUB: `<~`,

	// Unary ops
	TOKEN_NOT:  `!`,
	TOKEN_AT:   `@`,
	TOKEN_ATAT: `@@`,

	TOKEN_COMMA: `,`,

	// ()
	TOKEN_LP:   `(`,
	TOKEN_WSLP: `(`,
	TOKEN_RP:   `)`,

	// []
	TOKEN_LB:        `[`,
	TOKEN_LISTSTART: `[`,
	TOKEN_RB:        `]`,

	// {}
	TOKEN_LC:   `{`,
	TOKEN_SELC: `{`,
	TOKEN_RC:   `}`,

	// | |
	TOKEN_PIPE: `|`,
	TOKEN_PIPE_END: `|`,

	// EPP
	TOKEN_EPP_END:       `%>`,
	TOKEN_EPP_END_TRIM:  `-%>`,
	TOKEN_RENDER_EXPR:   `<%=`,
	TOKEN_RENDER_STRING: `epp text`,

	// Separators
	TOKEN_DOT:       `.`,
	TOKEN_QMARK:     `?`,
	TOKEN_COLON:     `:`,
	TOKEN_SEMICOLON: `;`,

	// Strings with semantics
	TOKEN_IDENTIFIER:          `identifier`,
	TOKEN_STRING:              `string literal`,
	TOKEN_INTEGER:             `integer literal`,
	TOKEN_FLOAT:               `float literal`,
	TOKEN_BOOLEAN:             `boolean literal`,
	TOKEN_CONCATENATED_STRING: `dq string literal`,
	TOKEN_HEREDOC:             `heredoc`,
	TOKEN_VARIABLE:            `variable`,
	TOKEN_REGEXP:              `regexp`,
	TOKEN_TYPE_NAME:           `type name`,

	// Keywords
	TOKEN_AND:         `and`,
	TOKEN_APPLICATION: `application`,
	TOKEN_ATTR:        `attr`,
	TOKEN_CASE:        `case`,
	TOKEN_CLASS:       `class`,
	TOKEN_CONSUMES:    `consumes`,
	TOKEN_DEFAULT:     `default`,
	TOKEN_DEFINE:      `define`,
	TOKEN_FUNCTION:    `function`,
	TOKEN_IF:          `if`,
	TOKEN_IN:          `in`,
	TOKEN_INHERITS:    `inherits`,
	TOKEN_ELSE:        `else`,
	TOKEN_ELSIF:       `elsif`,
	TOKEN_NODE:        `node`,
	TOKEN_OR:          `or`,
	TOKEN_PLAN:        `plan`,
	TOKEN_PRIVATE:     `private`,
	TOKEN_PRODUCES:    `produces`,
	TOKEN_SITE:        `site`,
	TOKEN_TYPE:        `type`,
	TOKEN_UNDEF:       `undef`,
	TOKEN_UNLESS:      `unless`,
}

var keywords = map[string]int{
	tokenMap[TOKEN_APPLICATION]: TOKEN_APPLICATION,
	tokenMap[TOKEN_AND]:         TOKEN_AND,
	tokenMap[TOKEN_ATTR]:        TOKEN_ATTR,
	tokenMap[TOKEN_CASE]:        TOKEN_CASE,
	tokenMap[TOKEN_CLASS]:       TOKEN_CLASS,
	tokenMap[TOKEN_CONSUMES]:    TOKEN_CONSUMES,
	tokenMap[TOKEN_DEFAULT]:     TOKEN_DEFAULT,
	tokenMap[TOKEN_DEFINE]:      TOKEN_DEFINE,
	`false`:                     TOKEN_BOOLEAN,
	tokenMap[TOKEN_FUNCTION]:    TOKEN_FUNCTION,
	tokenMap[TOKEN_ELSE]:        TOKEN_ELSE,
	tokenMap[TOKEN_ELSIF]:       TOKEN_ELSIF,
	tokenMap[TOKEN_IF]:          TOKEN_IF,
	tokenMap[TOKEN_IN]:          TOKEN_IN,
	tokenMap[TOKEN_INHERITS]:    TOKEN_INHERITS,
	tokenMap[TOKEN_NODE]:        TOKEN_NODE,
	tokenMap[TOKEN_OR]:          TOKEN_OR,
	tokenMap[TOKEN_PLAN]:        TOKEN_PLAN,
	tokenMap[TOKEN_PRIVATE]:     TOKEN_PRIVATE,
	tokenMap[TOKEN_PRODUCES]:    TOKEN_PRODUCES,
	tokenMap[TOKEN_SITE]:        TOKEN_SITE,
	`true`:                      TOKEN_BOOLEAN,
	tokenMap[TOKEN_TYPE]:        TOKEN_TYPE,
	tokenMap[TOKEN_UNDEF]:       TOKEN_UNDEF,
	tokenMap[TOKEN_UNLESS]:      TOKEN_UNLESS,
}

var DEFAULT_INSTANCE = Default{}

type Default struct{}

type context struct {
	stringReader
	locator               *Locator
	eppMode               bool
	handleBacktickStrings bool
	handleHexEscapes      bool
	tasks                 bool
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

func (ctx *context) setTokenValue(token int, value interface{}) {
	ctx.currentToken = token
	ctx.tokenValue = value
}

func (ctx *context) unterminatedQuote(start int, delimiter rune) *issue.Reported {
	ctx.SetPos(start)
	var stringType string
	if delimiter == '"' {
		stringType = `double`
	} else if delimiter == '\'' {
		stringType = `single`
	} else {
		stringType = `backtick`
	}
	return ctx.parseIssue2(LEX_UNTERMINATED_STRING, issue.H{`string_type`: stringType})
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
			panic(ctx.parseIssue(LEX_DIGIT_EXPECTED))
		}
		v, _ := strconv.ParseInt(ctx.From(start), 10, 64)
		ctx.setTokenValue(TOKEN_INTEGER, v)
		ctx.radix = 10

	case 'A' <= c && c <= 'Z':
		ctx.consumeQualifiedName(start, TOKEN_TYPE_NAME)

	case 'a' <= c && c <= 'z':
		ctx.consumeQualifiedName(start, TOKEN_IDENTIFIER)

	default:
		switch c {
		case 0:
			ctx.setToken(TOKEN_END)
		case '=':
			c, sz = ctx.Peek()
			switch c {
			case '=':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_EQUAL)
			case '~':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_MATCH)
			case '>':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_FARROW)
			default:
				ctx.setToken(TOKEN_ASSIGN)
			}
		case '{':
			if ctx.currentToken == TOKEN_QMARK {
				ctx.setToken(TOKEN_SELC)
			} else {
				ctx.setToken(TOKEN_LC)
			}

		case '}':
			ctx.setToken(TOKEN_RC)

		case '[':
			// If token is preceded by whitespace or if it's the first token to be parsed, then it's a
			// list rather than parameters to an access expression
			if scanStart < start || start == 0 {
				ctx.setToken(TOKEN_LISTSTART)
				break
			}
			ctx.setToken(TOKEN_LB)

		case ']':
			ctx.setToken(TOKEN_RB)

		case '(':
			// If token is first on line or only preceded by whitespace, then it is not start of parameters
			// in a call.
			savePos := ctx.Pos()
			ctx.SetPos(ctx.beginningOfLine)
			_, firstNonWhite := ctx.skipWhite(false)
			ctx.SetPos(savePos)
			if firstNonWhite == start {
				ctx.setToken(TOKEN_WSLP)
			} else {
				ctx.setToken(TOKEN_LP)
			}

		case ')':
			ctx.setToken(TOKEN_RP)

		case ',':
			ctx.setToken(TOKEN_COMMA)

		case ';':
			ctx.setToken(TOKEN_SEMICOLON)

		case '.':
			ctx.setToken(TOKEN_DOT)

		case '?':
			ctx.setToken(TOKEN_QMARK)

		case ':':
			ctx.setToken(TOKEN_COLON)
			c, sz = ctx.Peek()
			if c == ':' {
				ctx.Advance(sz)
				c, sz = ctx.Next()
				if isUppercaseLetter(c) {
					ctx.consumeQualifiedName(start, TOKEN_TYPE_NAME)
				} else if isLowercaseLetter(c) {
					ctx.consumeQualifiedName(start, TOKEN_IDENTIFIER)
				} else {
					ctx.SetPos(start)
					panic(ctx.parseIssue(LEX_DOUBLE_COLON_NOT_FOLLOWED_BY_NAME))
				}
			}

		case '-':
			c, sz = ctx.Peek()
			switch c {
			case '=':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_SUBTRACT_ASSIGN)
			case '>':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_IN_EDGE)
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
						panic(ctx.parseIssue2(LEX_INVALID_OPERATOR, issue.H{`op`: `-%`}))
					}
					break
				}
				fallthrough

			default:
				ctx.setToken(TOKEN_SUBTRACT)
			}

		case '+':
			c, sz = ctx.Peek()
			if c == '=' {
				ctx.Advance(sz)
				ctx.setToken(TOKEN_ADD_ASSIGN)
			} else if c == '>' {
				ctx.Advance(sz)
				ctx.setToken(TOKEN_PARROW)
			} else {
				ctx.setToken(TOKEN_ADD)
			}

		case '*':
			ctx.setToken(TOKEN_MULTIPLY)

		case '%':
			ctx.setToken(TOKEN_REMAINDER)
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
				ctx.setToken(TOKEN_NOT_EQUAL)
			} else if c == '~' {
				ctx.Advance(sz)
				ctx.setToken(TOKEN_NOT_MATCH)
			} else {
				ctx.setToken(TOKEN_NOT)
			}

		case '>':
			c, sz = ctx.Peek()
			if c == '=' {
				ctx.Advance(sz)
				ctx.setToken(TOKEN_GREATER_EQUAL)
			} else if c == '>' {
				ctx.Advance(sz)
				ctx.setToken(TOKEN_RSHIFT)
			} else {
				ctx.setToken(TOKEN_GREATER)
			}

		case '~':
			c, sz = ctx.Peek()
			if c == '>' {
				ctx.Advance(sz)
				ctx.setToken(TOKEN_IN_EDGE_SUB)
			} else {
				// Standalone tilde is not an operator in Puppet
				ctx.SetPos(start)
				panic(ctx.parseIssue2(LEX_UNEXPECTED_TOKEN, issue.H{`token`: `~`}))
			}

		case '@':
			c, sz = ctx.Peek()
			if c == '@' {
				ctx.Advance(sz)
				ctx.setToken(TOKEN_ATAT)
			} else if c == '(' {
				ctx.Advance(sz)
				ctx.consumeHeredocString()
			} else {
				ctx.setToken(TOKEN_AT)
			}

		case '<':
			c, sz = ctx.Peek()
			switch c {
			case '=':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_LESS_EQUAL)
			case '<':
				ctx.Advance(sz)
				c, sz = ctx.Peek()
				if c == '|' {
					ctx.Advance(sz)
					ctx.setToken(TOKEN_LLCOLLECT)
				} else {
					ctx.setToken(TOKEN_LSHIFT)
				}
			case '|':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_LCOLLECT)
			case '-':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_OUT_EDGE)
			case '~':
				ctx.Advance(sz)
				ctx.setToken(TOKEN_OUT_EDGE_SUB)
			case '%':
				if ctx.eppMode {
					ctx.Advance(sz)
					// <%# and <%% has been dealt with in consumeEPP so there's no need to deal with
					// that. Only <%, <%- and <%= can show up here
					c, sz = ctx.Peek()
					switch c {
					case '=':
						ctx.Advance(sz)
						ctx.setToken(TOKEN_RENDER_EXPR)
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
				ctx.setToken(TOKEN_LESS)
			}

		case '|':
			c, sz = ctx.Peek()
			switch c {
			case '>':
				ctx.Advance(sz)
				c, sz = ctx.Peek()
				if c == '>' {
					ctx.Advance(sz)
					ctx.setToken(TOKEN_RRCOLLECT)
				} else {
					ctx.setToken(TOKEN_RCOLLECT)
				}
			default:
				if ctx.currentToken == TOKEN_PIPE {
					// Empty parameter list
					ctx.setToken(TOKEN_PIPE_END)
				} else {
					pos := ctx.Pos()
					n, _ := ctx.skipWhite(false)
					ctx.SetPos(pos)
					if n == '{' || n == '>' || ctx.eppMode && (n == '%' || n == '-') {
						// A lambda parameter list cannot start with either of these tokens so
						// this must be the end (next is either block body or block return type declaration)
						ctx.setToken(TOKEN_PIPE_END)
					} else {
						ctx.setToken(TOKEN_PIPE)
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
			ctx.setToken(TOKEN_DIVIDE)

		case '$':
			c, sz = ctx.Peek()
			if c == ':' {
				ctx.Advance(sz)
				c, sz = ctx.Peek()
				if c != ':' {
					ctx.SetPos(start)
					panic(ctx.parseIssue(LEX_INVALID_VARIABLE_NAME))
				}
				ctx.Advance(sz)
				c, sz = ctx.Peek()
			}
			if isLowercaseLetter(c) {
				ctx.Advance(sz)
				ctx.consumeQualifiedName(start, TOKEN_VARIABLE)
			} else if isDecimalDigit(c) {
				ctx.Advance(sz)
				ctx.skipDecimalDigits()
				ctx.tokenValue, _ = strconv.ParseInt(ctx.From(start+1), 10, 64)
			} else if unicode.IsLetter(c) {
				panic(ctx.parseIssue(LEX_INVALID_VARIABLE_NAME))
			} else {
				ctx.tokenValue = ``
			}
			ctx.setTokenValue(TOKEN_VARIABLE, ctx.tokenValue)

		case '0':
			ctx.radix = 10
			c, sz = ctx.Peek()

			switch c {
			case 0:
				ctx.setTokenValue(TOKEN_INTEGER, int64(0))
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
					panic(ctx.parseIssue(LEX_HEXDIGIT_EXPECTED))
				}
				v, _ := strconv.ParseInt(ctx.From(hexStart), 16, 64)
				ctx.radix = 16
				ctx.setTokenValue(TOKEN_INTEGER, v)

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
					panic(ctx.parseIssue(LEX_OCTALDIGIT_EXPECTED))
				}
				if ctx.Pos() > octalStart {
					v, _ := strconv.ParseInt(ctx.From(octalStart), 8, 64)
					ctx.radix = 8
					ctx.setTokenValue(TOKEN_INTEGER, v)
				} else {
					ctx.setTokenValue(TOKEN_INTEGER, int64(0))
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
			panic(ctx.parseIssue2(LEX_UNEXPECTED_TOKEN, issue.H{`token`: string(c)}))
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
				panic(ctx.parseIssue(LEX_UNTERMINATED_COMMENT))
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
	outer: for {
		c, n := ctx.Peek()
		for isLetterOrDigit(c) {
			ctx.Advance(n)
			c, n = ctx.Peek()
		}

		if c == '-' && token == TOKEN_IDENTIFIER {
			// Valid only if a letter or digit is present before end of name
			i := ctx.Pos() + n
			for {
				c, n = ctx.PeekAt(i);
				i += n;
				if isLetterOrDigit(c) {
					hasDash = true
					ctx.SetPos(i);
					continue outer;
				}
				if c != '-' {
					break outer;
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
		if token == TOKEN_TYPE_NAME && isUppercaseLetter(c) ||
			token != TOKEN_TYPE_NAME && (isLowercaseLetter(c) ||
				token == TOKEN_VARIABLE && c == '_') {
			// Next segment starts here and only last segment is allowed to
			// start with underscore
			if !lastStartsWithUnderscore {
				ctx.Advance(n)
				lastStartsWithUnderscore = c == '_'
				continue
			}
		}

		ctx.SetPos(start)
		issueCode := issue.Code(LEX_INVALID_NAME)
		if token == TOKEN_TYPE_NAME {
			issueCode = LEX_INVALID_TYPE_NAME
		} else if token == TOKEN_VARIABLE {
			issueCode = LEX_INVALID_VARIABLE_NAME
		}
		panic(ctx.parseIssue(issueCode))
	}

	if token == TOKEN_VARIABLE {
		start++ // skip leading '$´
	}

	word := ctx.From(start)

	if token == TOKEN_IDENTIFIER {
		if hasDash {
			token = TOKEN_STRING
		} else if kwToken, ok := keywords[word]; ok {
			switch kwToken {
			case TOKEN_BOOLEAN:
				ctx.setTokenValue(kwToken, word == `true`)
				return
			case TOKEN_DEFAULT:
				ctx.setTokenValue(kwToken, DEFAULT_INSTANCE)
				return
			case TOKEN_PLAN:
				if ctx.tasks {
					token = kwToken
				}
			default:
				token = kwToken
			}
		}
	}

	ctx.setTokenValue(token, word)
}

func (ctx *context) consumeFloat(start int, d rune) {
	if ctx.skipDecimalDigits() == 0 {
		panic(ctx.parseIssue(LEX_DIGIT_EXPECTED))
	}
	c, n := ctx.Peek()
	if d == '.' {
		// Check for 'e'
		if c == 'e' || c == 'E' {
			ctx.Advance(n)
			if ctx.skipDecimalDigits() == 0 {
				panic(ctx.parseIssue(LEX_DIGIT_EXPECTED))
			}
			c, n = ctx.Peek()
		}
	}
	if unicode.IsLetter(c) {
		panic(ctx.parseIssue(LEX_DIGIT_EXPECTED))
	}
	v, _ := strconv.ParseFloat(ctx.From(start), 64)
	ctx.setTokenValue(TOKEN_FLOAT, v)
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
			ctx.setToken(TOKEN_DIVIDE)
			return

		case delimiter:
			ctx.setTokenValue(TOKEN_STRING, buf.String())
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
					panic(ctx.parseIssue(LEX_UNBALANCED_EPP_COMMENT))
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
			ctx.setTokenValue(TOKEN_RENDER_STRING, buf.String())
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
		ctx.setToken(TOKEN_END)
	} else {
		ctx.setTokenValue(TOKEN_RENDER_STRING, buf.String())
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
		expr := ctx.parse(TOKEN_RC, true)

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
			if identifier, ok := call.functor.(*QualifiedName); ok {
				expr = ctx.factory.CallMethod(
					ctx.factory.Variable(identifier, ctx.locator, start, identifier.ByteLength()+1),
					call.arguments, call.lambda, ctx.locator, start, call.ByteLength()+1)
			} else if ne, ok := call.functor.(*NamedAccessExpression); ok {
				modNe := ctx.convertNamedAccessLHS(ne, start)
				if modNe != ne {
					expr = ctx.factory.CallMethod(modNe, call.arguments, call.lambda, ctx.locator, start, call.ByteLength()+1)
				}
			}
		}
		return ctx.factory.Text(expr, ctx.locator, start, ctx.Pos()-start)
	}

	// Not delimited by curly braces. Must be a single identifier then
	ctx.setToken(TOKEN_VARIABLE)
	if c == ':' || isLowercaseLetter(c) || isDecimalDigit(c) {
		ctx.nextToken()
	}
	if ctx.currentToken != TOKEN_IDENTIFIER {
		ctx.SetPos(start)
		panic(ctx.parseIssue(LEX_MALFORMED_INTERPOLATION))
	}
	textExpr := ctx.factory.QualifiedName(ctx.tokenValue.(string), ctx.locator, start+1, ctx.Pos()-(start+1))
	return ctx.factory.Text(ctx.factory.Variable(textExpr, ctx.locator, start, ctx.Pos()-start), ctx.locator, start, ctx.Pos()-start)
}

func (ctx *context) convertNamedAccessLHS(expr *NamedAccessExpression, start int) Expression {
	lhs := expr.lhs
	switch lhs.(type) {
	case *QualifiedName:
		return ctx.factory.NamedAccess(
			ctx.factory.Variable(lhs, ctx.locator, start, lhs.ByteLength()+1),
			expr.rhs, ctx.locator, start, expr.ByteLength()+1)
	case *AccessExpression:
		access := lhs.(*AccessExpression)
		if identifier, ok := access.operand.(*QualifiedName); ok {
			lhs = ctx.factory.Access(
				ctx.factory.Variable(identifier, ctx.locator, start, identifier.ByteLength()+1),
				access.keys, ctx.locator, start, access.ByteLength()+1)
		}
		return ctx.factory.NamedAccess(lhs, expr.rhs, ctx.locator, start, expr.ByteLength()+1)
	case *NamedAccessExpression:
		return ctx.factory.NamedAccess(
			ctx.convertNamedAccessLHS(lhs.(*NamedAccessExpression), start),
			expr.rhs, ctx.locator, start, expr.ByteLength()+1)
	}
	if identifier, ok := lhs.(*QualifiedName); ok {
		return ctx.factory.NamedAccess(
			ctx.factory.Variable(identifier, ctx.locator, start, identifier.ByteLength()+1),
			expr.rhs, ctx.locator, start, expr.ByteLength()+1)
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
	ctx.setTokenValue(TOKEN_STRING, ctx.From(start))
	ctx.Advance(sz)
}

func (ctx *context) consumeDoubleQuotedString() {
	var segments []Expression
	if ctx.factory != nil {
		segments = make([]Expression, 0, 4)
	}
	segments = ctx.consumeDelimitedString('"', ctx.Pos() - 1, segments,
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
	ctx.setTokenValue(TOKEN_CONCATENATED_STRING, ctx.factory.ConcatenatedString(segments, ctx.locator, firstPos, ctx.Pos()-firstPos))
}

func (ctx *context) consumeSingleQuotedString() {
	ctx.consumeDelimitedString('\'', ctx.Pos() - 1, nil, func(buf *bytes.Buffer, ctx *context, ec rune) {
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
	ctx.consumeDelimitedString('/', start - 1, nil, func(buf *bytes.Buffer, ctx *context, ec rune) {
		buf.WriteRune('\\')
		buf.WriteRune(ec)
	})
	if ctx.currentToken == TOKEN_STRING {
		ctx.currentToken = TOKEN_REGEXP
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
			panic(ctx.parseIssue(LEX_HEREDOC_DECL_UNTERMINATED))

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
				panic(ctx.parseIssue(LEX_HEREDOC_MULTIPLE_SYNTAX))
			}
			if tag == `` {
				tag = ctx.From(start)
			}
			ctx.Advance(n)
			syntaxStart = ctx.Pos()

		case '/':
			if escapeStart > 0 {
				panic(ctx.parseIssue(LEX_HEREDOC_MULTIPLE_ESCAPE))
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
				panic(ctx.parseIssue(LEX_HEREDOC_MULTIPLE_TAG))
			}
			ctx.Advance(n)
			quoteStart = ctx.Pos()
		findEndQuote:
			for {
				c, n = ctx.Peek()
				switch c {
				case 0, '\n':
					ctx.SetPos(heredocStart)
					panic(ctx.parseIssue(LEX_HEREDOC_DECL_UNTERMINATED))
				case '"':
					break findEndQuote
				default:
					ctx.Advance(n)
				}
			}
			if quoteStart == ctx.Pos() {
				ctx.SetPos(heredocStart)
				panic(ctx.parseIssue(LEX_HEREDOC_EMPTY_TAG))
			}
			tag = ctx.From(quoteStart)
			ctx.Advance(n)
		default:
			ctx.Advance(n)
		}
	}

	if tag == `` {
		ctx.SetPos(heredocStart)
		panic(ctx.parseIssue(LEX_HEREDOC_EMPTY_TAG))
	}

	// Find where actual text starts
	heredocContentStart := -1
	c, sz := ctx.Peek()
findStartOfText:
	for {
		switch c {
		case 0:
			ctx.SetPos(heredocStart)
			panic(ctx.parseIssue(LEX_HEREDOC_UNTERMINATED))

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
			panic(ctx.parseIssue(LEX_HEREDOC_UNTERMINATED))

		case '\n':
			lineStart := ctx.Pos()
			c, n = ctx.skipWhiteInLiteral()
			switch c {
			case 0:
				ctx.SetPos(heredocStart)
				panic(ctx.parseIssue(LEX_HEREDOC_UNTERMINATED))

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
			c, n = ctx.Next()
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
		if segments != nil && len(segments) > 0 {
			if len(heredoc) > 0 {
				segments = append(segments, ctx.factory.String(heredoc, ctx.locator, ctx.tokenStartPos, ctx.Pos()-ctx.tokenStartPos))
			}
			ctx.SetPos(heredocTagEnd)          // Normal parsing continues here
			ctx.nextLineStart = heredocEnd + 1 // and next newline will jump to here
			textExpr := ctx.factory.ConcatenatedString(segments, ctx.locator, heredocContentStart, heredocContentEnd-heredocContentStart)
			ctx.setTokenValue(TOKEN_HEREDOC, ctx.factory.Heredoc(textExpr, syntax, ctx.locator, heredocStart, heredocContentEnd-heredocStart))
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
		ctx.setTokenValue(TOKEN_HEREDOC, ctx.factory.Heredoc(textExpr, syntax, ctx.locator, heredocStart, heredocContentEnd-heredocStart))
	} else {
		ctx.setTokenValue(TOKEN_STRING, heredoc)
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
			panic(ctx.parseIssue2(LEX_HEREDOC_ILLEGAL_ESCAPE, issue.H{`flag`: string(flag)}))
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
			break
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
			panic(ctx.parseIssue(LEX_MALFORMED_HEX_ESCAPE))
		}
	}
	r, _ := strconv.ParseInt(ctx.From(start), 16, 64)
	buf.WriteByte(byte(r))
	return
}

func (ctx *context) appendUnicode(buf *bytes.Buffer) {
	ec, start := ctx.Next()
	if isHexDigit(ec) {
		// Must be XXXX (a four-digit hex number)
		for i := 1; i < 4; i++ {
			digit, _ := ctx.Next()
			if !isHexDigit(digit) {
				ctx.SetPos(start - 2)
				panic(ctx.parseIssue(LEX_MALFORMED_UNICODE_ESCAPE))
			}
		}
		r, _ := strconv.ParseInt(ctx.From(start), 16, 32)
		buf.WriteRune(rune(r))
		return
	}

	if ec != '{' {
		ctx.SetPos(start - 2)
		panic(ctx.parseIssue(LEX_MALFORMED_UNICODE_ESCAPE))
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
		panic(ctx.parseIssue(LEX_MALFORMED_UNICODE_ESCAPE))
	}

	r, _ := strconv.ParseInt(ctx.From(hexStart), 16, 32)
	ctx.Advance(n) // Skip terminating '}'
	buf.WriteRune(rune(r))
}

func (ctx *context) isRegexpAcceptable() bool {
	switch ctx.currentToken {
	// Operands that can be followed by TOKEN_DIVIDE
	case TOKEN_RP, TOKEN_RB, TOKEN_TYPE_NAME, TOKEN_IDENTIFIER, TOKEN_BOOLEAN, TOKEN_INTEGER, TOKEN_FLOAT, TOKEN_STRING,
		TOKEN_HEREDOC, TOKEN_CONCATENATED_STRING, TOKEN_REGEXP, TOKEN_VARIABLE:
		return false
	default:
		return true
	}
}
