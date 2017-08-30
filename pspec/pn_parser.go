package pspec

import (
  . "github.com/puppetlabs/go-parser/parser"
  . "github.com/puppetlabs/go-parser/pn"
)

func ParsePN(file string, content string) PN {
  lexer := NewSimpleLexer(file, content)
  lexer.NextToken()
  return parseNext(lexer)
}

func parseNext(lexer Lexer) PN {
  switch lexer.CurrentToken() {
  case TOKEN_LB, TOKEN_LISTSTART:
    return parseArray(lexer)
  case TOKEN_LC, TOKEN_SELC:
    return parseMap(lexer)
  case TOKEN_LP, TOKEN_WSLP:
    return parseCall(lexer)
  case TOKEN_STRING, TOKEN_BOOLEAN, TOKEN_INTEGER, TOKEN_FLOAT, TOKEN_UNDEF:
    return parseLiteral(lexer)
  case TOKEN_IDENTIFIER:
    switch lexer.TokenValue().(string) {
    case `null`:
      return LiteralPN(nil)
    default:
      lexer.SyntaxError()
    }
  case TOKEN_SUBTRACT:
    switch lexer.NextToken() {
    case TOKEN_FLOAT:
      return LiteralPN(-lexer.TokenValue().(float64))
    case TOKEN_INTEGER:
      return LiteralPN(-lexer.TokenValue().(int64))
    default:
      lexer.SyntaxError()
    }
  default:
    lexer.SyntaxError()
  }
  return nil
}

func parseArray(lexer Lexer) PN {
  return ListPN(parseElements(lexer, TOKEN_RB))
}

func parseMap(lexer Lexer) PN {
  entries := make([]Entry, 0, 8)
  token := lexer.NextToken()
  for token != TOKEN_RC && token != TOKEN_END {
    lexer.AssertToken(TOKEN_COLON)
    lexer.NextToken()
    key := parseIdentifier(lexer)
    entries = append(entries, parseNext(lexer).WithName(key))
    token = lexer.CurrentToken()
  }
  lexer.AssertToken(TOKEN_RC)
  lexer.NextToken()
  return MapPN(entries)
}

func parseCall(lexer Lexer) PN {
  lexer.NextToken()
  name := parseIdentifier(lexer)
  elements := parseElements(lexer, TOKEN_RP)
  return CallPN(name, elements...)
}

func parseLiteral(lexer Lexer) PN {
  pn := LiteralPN(lexer.TokenValue())
  lexer.NextToken()
  return pn
}

func parseIdentifier(lexer Lexer) string {
  switch lexer.CurrentToken() {
  case TOKEN_END,
    TOKEN_LP, TOKEN_WSLP, TOKEN_RP,
    TOKEN_LB, TOKEN_LISTSTART, TOKEN_RB,
    TOKEN_LC, TOKEN_SELC, TOKEN_RC,
    TOKEN_EPP_END, TOKEN_EPP_END_TRIM, TOKEN_RENDER_EXPR, TOKEN_RENDER_STRING,
    TOKEN_COMMA, TOKEN_COLON, TOKEN_SEMICOLON,
    TOKEN_STRING, TOKEN_INTEGER, TOKEN_FLOAT, TOKEN_CONCATENATED_STRING, TOKEN_HEREDOC,
    TOKEN_REGEXP:
    lexer.SyntaxError()
    return ``
  case TOKEN_DEFAULT:
    lexer.NextToken()
    return `default`
  default:
    str := lexer.TokenString()
    lexer.NextToken()
    return str
  }
}

func parseElements(lexer Lexer, endToken int) []PN {
  elements := make([]PN, 0, 8)
  token := lexer.CurrentToken()
  for token != endToken && token != TOKEN_END {
    elements = append(elements, parseNext(lexer))
    token = lexer.CurrentToken()
  }
  lexer.AssertToken(endToken)
  lexer.NextToken()
  return elements
}
