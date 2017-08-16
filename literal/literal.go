package literal

import (
  . "github.com/puppetlabs/go-parser/parser"
)

const notLiteral = `not literal`

func ToLiteral(e Expression)  (value interface{}, ok bool) {
  defer func() {
    if err := recover(); err != nil {
      if err == notLiteral {
        ok = false
      } else {
        panic(err)
      }
    }
  }()

  value = toLiteral(e)
  ok = true
  return
}

func toLiteral(e Expression)  interface{} {
  switch e.(type) {
  case *Program:
    return toLiteral(e.(*Program).Body())
  case *LiteralList:
    elements := e.(*LiteralList).Elements()
    result := make([]interface{}, len(elements))
    for idx, elem := range elements {
      result[idx] = toLiteral(elem)
    }
    return result
  case *LiteralHash:
    entries := e.(*LiteralHash).Entries()
    result := make(map[interface{}]interface{}, len(entries))
    for _, entry := range entries {
      kh := entry.(*KeyedEntry)
      result[toLiteral(kh.Key())] = toLiteral(kh.Value())
    }
    return result
  case *ConcatenatedString:
    segments := e.(*ConcatenatedString).Segments()
    if len(segments) == 1 {
      if ls, ok := segments[0].(*LiteralString); ok {
        return ls.Value()
      }
    }
    panic(notLiteral)
  case *HeredocExpression:
    return toLiteral(e.(*HeredocExpression).Text())
  case LiteralValue:
    return e.(LiteralValue).Value()
  default:
    panic(notLiteral)
  }
}
