package main

import (
  "os"
  "fmt"
  "io/ioutil"
  "encoding/json"
  "bytes"
  . "github.com/puppetlabs/parser"
  "strings"
  "flag"
)

var validateOnly = flag.Bool("v", false, "validate only")

func main() {
  flag.Parse()

  args := flag.Args()
  if len(args) != 1 {
    fmt.Fprintln(os.Stderr, "usage: parser [-v] <pp or epp file to parse>")
    os.Exit(1)
  }

  fileName := args[0]
  content, err := ioutil.ReadFile(fileName)
  if err != nil {
    panic(err)
  }

  expr, err := Parse(args[0], string(content), strings.HasSuffix(fileName, `.epp`))
  if err != nil {
    fmt.Fprintln(os.Stderr, err.Error())
    os.Exit(1)
  }

  v := NewValidator()
  v.Validate(expr)
  if len(v.Issues()) > 0 {
    for _, issue := range v.Issues() {
      fmt.Fprintln(os.Stderr, issue.String())
    }
    os.Exit(1)
  }

  if !*validateOnly {
    result := bytes.NewBufferString(``)
    enc := json.NewEncoder(result)
    enc.SetEscapeHTML(false)
    enc.Encode(expr.ToPN().ToData())
    result.Truncate(result.Len() - 1)
    fmt.Println(result.String())
  }
}
