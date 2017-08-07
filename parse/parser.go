package main

import (
  . "os"
  . "fmt"
  . "github.com/puppetlabs/go-parser/parser"
  . "strings"
  "io/ioutil"
  "flag"
  "encoding/json"
  "bytes"
)

var validateOnly = flag.Bool("v", false, "validate only")

func main() {
  flag.Parse()

  args := flag.Args()
  if len(args) != 1 {
    Fprintln(Stderr, "usage: parser [-v] <pp or epp file to parse>")
    Exit(1)
  }

  fileName := args[0]
  content, err := ioutil.ReadFile(fileName)
  if err != nil {
    panic(err)
  }

  expr, err := Parse(args[0], string(content), HasSuffix(fileName, `.epp`))
  if err != nil {
    Fprintln(Stderr, err.Error())
    Exit(1)
  }

  v := NewValidator()
  v.Validate(expr)
  if len(v.Issues()) > 0 {
    for _, issue := range v.Issues() {
      Fprintln(Stderr, issue.String())
    }
    Exit(1)
  }

  if !*validateOnly {
    result := bytes.NewBufferString(``)
    enc := json.NewEncoder(result)
    enc.SetEscapeHTML(false)
    enc.Encode(expr.ToPN().ToData())
    result.Truncate(result.Len() - 1)
    Println(result.String())
  }
}
