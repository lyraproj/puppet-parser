package parser_test

import (
  `testing`
  . "github.com/puppetlabs/go-parser/parser"
  . "github.com/puppetlabs/go-parser/pspec"
  "path/filepath"
  "io/ioutil"
)

func TestPrimitives(t *testing.T) {
  runPspecTests(t, `primitives.pspec`)
}

func runPspecTests(t *testing.T, name string) {
  runTests(t, NewSpecEvaluator().CreateTests(parseTestContents(t, name)))
}

func runTests(t *testing.T, tests []Test) {
  var (
    testExec *TestExecutable
    testGroup *TestGroup
    ok bool
  )
  for _, test := range tests {
    if testExec, ok = test.(*TestExecutable); ok {
      t.Run(testExec.Name(), func(t *testing.T) {
        testExec.Executable()(&assertions{t})
      })
    } else if testGroup, ok = test.(*TestGroup); ok {
      t.Run(testGroup.Name(), func(t *testing.T) {
        runTests(t, testGroup.Tests())
      })
    }
  }
}

func parseTestContents(t *testing.T, name string) Expression {
  path := filepath.Join(`testdata`, name)
  content, err := ioutil.ReadFile(path)
  if err != nil {
    t.Fatal(err)
  }

  expr, err := CreatePspecParser().Parse(path, string(content), false, false)
  if err != nil {
    t.Fatal(err.Error())
  }
  return expr
}

type assertions struct {
  t *testing.T
}

func (a *assertions)Fail(message string) {
  a.t.Error(message)
}

func (a *assertions)AssertEquals(expected interface{}, actual interface{}) {
  if expected != actual {
    a.t.Errorf("expected '%v', got '%v'\n", expected, actual)
  }
}

