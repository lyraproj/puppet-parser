// +build go1.7

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-parser/json"
	"github.com/lyraproj/puppet-parser/parser"
	"github.com/lyraproj/puppet-parser/pn"
	"github.com/lyraproj/puppet-parser/validator"
)

// Program to parse and validate a .pp or .epp file
var validateOnly = flag.Bool("v", false, "validate only")
var jsonOutput = flag.Bool("j", false, "json output")
var strict = flag.String("s", `off`, "strict (off, warning, or error)")
var tasks = flag.Bool("t", false, "tasks")
var workflow = flag.Bool("w", false, "workflow")

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		pn.Fprintln(os.Stderr, "Usage: parse [options] <pp or epp file to parse>\nValid options are:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fileName := args[0]
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var result map[string]interface{}
	if *jsonOutput {
		result = make(map[string]interface{}, 2)
	}

	strictness := validator.Strict(*strict)

	parseOpts := make([]parser.Option, 0)
	if strings.HasSuffix(fileName, `.epp`) {
		parseOpts = append(parseOpts, parser.EppMode)
	}
	if *tasks {
		parseOpts = append(parseOpts, parser.TasksEnabled)
	}
	if *workflow {
		parseOpts = append(parseOpts, parser.WorkflowEnabled)
	}

	expr, err := parser.CreateParser(parseOpts...).Parse(args[0], string(content), false)
	if *jsonOutput {
		if err != nil {
			if i, ok := err.(issue.Reported); ok {
				result[`issues`] = []interface{}{pn.ReportedToPN(i).ToData()}
			} else {
				result[`error`] = err.Error()
			}
			emitJson(result)
			// Parse error is always SeverityError
			os.Exit(1)
		}

		v := validator.ValidatePuppet(expr, strictness)
		if len(v.Issues()) > 0 {
			severity := issue.Severity(issue.SeverityIgnore)
			issues := make([]interface{}, len(v.Issues()))
			for idx, i := range v.Issues() {
				if i.Severity() > severity {
					severity = i.Severity()
				}
				issues[idx] = pn.ReportedToPN(i).ToData()
			}
			result[`issues`] = issues
			if severity == issue.SeverityError {
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
		pn.Fprintln(os.Stderr, err.Error())
		// Parse error is always SeverityError
		os.Exit(1)
	}

	v := validator.ValidatePuppet(expr, strictness)
	if len(v.Issues()) > 0 {
		severity := issue.Severity(issue.SeverityIgnore)
		for _, i := range v.Issues() {
			pn.Fprintln(os.Stderr, i.String())
			if i.Severity() > severity {
				severity = i.Severity()
			}
		}
		if severity == issue.SeverityError {
			os.Exit(1)
		}
	}

	if !*validateOnly {
		b := bytes.NewBufferString(``)
		expr.ToPN().Format(b)
		pn.Println(b)
	}
}

func emitJson(value interface{}) {
	b := bytes.NewBufferString(``)
	json.ToJson(value, b)
	pn.Println(b.String())
}
