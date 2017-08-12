// +build !go1.7

package parser

import (
  "encoding/json"
  "io"
)

// Special version for Go < 1.7 where the encoder lacks function SetEscapeHTML
func ToJson(expr Expression, result io.Writer) {
  enc := json.NewEncoder(result)
  enc.Encode(expr.ToPN().ToData())
}
