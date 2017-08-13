// +build go1.7
package main

import (
  . "os"
  . "fmt"
  . "github.com/puppetlabs/go-parser/parser"
  . "strings"
  "io/ioutil"
  "flag"
  "bytes"
  "github.com/puppetlabs/go-parser/validator"
)

// Program to parse and validate a .pp or .epp file
var validateOnly = flag.Bool("v", false, "validator only")

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

  v := validator.ValidatePuppet(expr)
  if len(v.Issues()) > 0 {
    severity := Severity(SEVERITY_IGNORE)
    for _, issue := range v.Issues() {
      Fprintln(Stderr, issue.String())
      if issue.Severity() > severity {
        severity = issue.Severity()
      }
    }
    if severity == SEVERITY_ERROR {
      Exit(1)
    }
  }

  if !*validateOnly {
    result := bytes.NewBufferString(``)
    ToJson(expr, result)
    Print(result.String())
  }
}
