// +build go1.7

package json

import (
	"encoding/json"
	"io"
)

func ToJson(value interface{}, result io.Writer) {
	enc := json.NewEncoder(result)
	enc.SetEscapeHTML(false)
	enc.Encode(value)
}
