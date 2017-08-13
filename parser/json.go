// +build go1.7

package parser

import (
  "encoding/json"
  "io"
)

func ToJson(value interface{}, result io.Writer) {
  enc := json.NewEncoder(result)
  enc.SetEscapeHTML(false)
  enc.Encode(value)
}
