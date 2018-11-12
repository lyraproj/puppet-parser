// +build go1.7

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-parser/json"
	"github.com/puppetlabs/go-parser/parser"
	"github.com/puppetlabs/go-parser/validator"
	"github.com/puppetlabs/go-parser/pn"
)

// Program to parse and validate a .pp or .epp file
var validateOnly = flag.Bool("v", false, "validate only")
var jsonOuput = flag.Bool("j", false, "json output")
var strict = flag.String("s", `off`, "strict (off, warning, or error)")
var tasks = flag.Bool("t", false, "tasks")
var workflow = flag.Bool("w", false, "workflow")

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: parse [options] <pp or epp file to parse>\nValid options are:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fileName := args[0]
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	if *jsonOuput {
		result = make(map[string]interface{}, 2)
	}

	strictness := validator.Strict(*strict)

	parseOpts := []parser.Option{}
	if strings.HasSuffix(fileName, `.epp`) {
		parseOpts = append(parseOpts, parser.PARSER_EPP_MODE)
	}
	if *tasks {
		parseOpts = append(parseOpts, parser.PARSER_TASKS_ENABLED)
	}
	if *workflow {
		parseOpts = append(parseOpts, parser.PARSER_WORKFLOW_ENABLED)
	}

	expr, err := parser.CreateParser(parseOpts...).Parse(args[0], string(content), false)
	if *jsonOuput {
		if err != nil {
			if issue, ok := err.(issue.Reported); ok {
				result[`issues`] = []interface{}{pn.ReportedToPN(issue).ToData()}
			} else {
				result[`error`] = err.Error()
			}
			emitJson(result)
			// Parse error is always SEVERITY_ERROR
			os.Exit(1)
		}

		v := validator.ValidatePuppet(expr, strictness)
		if len(v.Issues()) > 0 {
			severity := issue.Severity(issue.SEVERITY_IGNORE)
			issues := make([]interface{}, len(v.Issues()))
			for idx, issue := range v.Issues() {
				if issue.Severity() > severity {
					severity = issue.Severity()
				}
				issues[idx] = pn.ReportedToPN(issue).ToData()
			}
			result[`issues`] = issues
			if severity == issue.SEVERITY_ERROR {
				emitJson(result)
				os.Exit(1)
			}
		}

		if !*validateOnly {
			result[`ast`] = expr.ToPN().ToData()
		}
		emitJson(result)
		return
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		// Parse error is always SEVERITY_ERROR
		os.Exit(1)
	}

	v := validator.ValidatePuppet(expr, strictness)
	if len(v.Issues()) > 0 {
		severity := issue.Severity(issue.SEVERITY_IGNORE)
		for _, issue := range v.Issues() {
			fmt.Fprintln(os.Stderr, issue.String())
			if issue.Severity() > severity {
				severity = issue.Severity()
			}
		}
		if severity == issue.SEVERITY_ERROR {
			os.Exit(1)
		}
	}

	if !*validateOnly {
		b := bytes.NewBufferString(``)
		expr.ToPN().Format(b)
		fmt.Println(b)
	}
}

func emitJson(value interface{}) {
	b := bytes.NewBufferString(``)
	json.ToJson(value, b)
	fmt.Println(b.String())
}
