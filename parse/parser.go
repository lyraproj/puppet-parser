// +build go1.7
package main

import (
  . "os"
  . "fmt"
  . "github.com/puppetlabs/go-parser/issue"
  . "github.com/puppetlabs/go-parser/json"
  . "github.com/puppetlabs/go-parser/parser"
  . "strings"
  . "github.com/puppetlabs/go-parser/validator"
  "io/ioutil"
  "flag"
  "bytes"
)

// Program to parse and validate a .pp or .epp file
var validateOnly = flag.Bool("v", false, "validate only")
var json = flag.Bool("j", false, "json output")
var strict = flag.String("s", `off`, "strict (off, warning, or error)")

func main() {
  flag.Parse()

  args := flag.Args()
  if len(args) != 1 {
    Fprintf(Stderr, "Usage: parse [options] <pp or epp file to parse>\nValid options are:\n")
    flag.PrintDefaults()
    Exit(1)
  }

  fileName := args[0]
  content, err := ioutil.ReadFile(fileName)
  if err != nil {
    panic(err)
  }

  var result map[string]interface{}
  if *json {
    result = make(map[string]interface{}, 2)
  }

  strictness := Strict(*strict)

  expr, err := CreateParser().Parse(args[0], string(content), HasSuffix(fileName, `.epp`), false)
  if *json {
    if err != nil {
      if issue, ok := err.(*ReportedIssue); ok {
        result[`issues`] = []interface{}{issue.ToPN().ToData()}
      } else {
        result[`error`] = err.Error()
      }
      emitJson(result)
      // Parse error is always SEVERITY_ERROR
      Exit(1)
    }

    v := ValidatePuppet(expr, strictness)
    if len(v.Issues()) > 0 {
      severity := Severity(SEVERITY_IGNORE)
      issues := make([]interface{}, len(v.Issues()))
      for idx, issue := range v.Issues() {
        if issue.Severity() > severity {
          severity = issue.Severity()
        }
        issues[idx] = issue.ToPN().ToData()
      }
      result[`issues`] = issues
      if severity == SEVERITY_ERROR {
        emitJson(result)
        Exit(1)
      }
    }

    if !*validateOnly {
      result[`ast`] = expr.ToPN().ToData()
    }
    emitJson(result)
    return
  }

  if err != nil {
    Fprintln(Stderr, err.Error())
    // Parse error is always SEVERITY_ERROR
    Exit(1)
  }

  v := ValidatePuppet(expr, strictness)
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
    b := bytes.NewBufferString(``)
    expr.ToPN().Format(b)
    Println(b)
  }
}

func emitJson(value interface{}) {
  b := bytes.NewBufferString(``)
  ToJson(value, b)
  Println(b.String())
}
