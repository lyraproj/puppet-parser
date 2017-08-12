package parser

import (
  "testing"
  "bytes"
  "encoding/json"
  . "github.com/puppetlabs/go-parser/testutils"
)

func TestManifest(t *testing.T) {
  expectJSON(t,
    Unindent(`
      file { '/tmp/foo':
        mode => '0640',
        ensure => present;
      '/tmp/bar':
        mode => '0640',
        ensure => present;
      }

      $rootgroup = $facts['os']['family'] ? 'Solaris' => 'wheel'

      function foo(Integer[2,3] $in, String $n = 'vi') >> Float[0.0] {
          notice("show the ${n}")
        $in * 3.14
      }`),
    `["block",[`+
      `["resource",{`+
        `"bodies":[`+
          `{"ops":[["=>",[["mode"],"0640"]],["=>",[["ensure"],["qn","present"]]]],"title":"/tmp/foo"},`+
          `{"ops":[["=>",[["mode"],"0640"]],["=>",[["ensure"],["qn","present"]]]],"title":"/tmp/bar"}],`+
        `"type":["qn","file"]}],`+
      `["=",[["$","rootgroup"],["?",[["[]",[["[]",[["$","facts"],"os"]],"family"]],[["=>",["Solaris","wheel"]]]]]]],`+
      `["function",{`+
        `"body":[`+
          `["invoke",{`+
            `"args":[["concat",["show the ",["str",["$","n"]]]]],`+
            `"functor":["qn","notice"]}],`+
          `["*",[["$","in"],3.14]]],`+
        `"name":"foo",`+
        `"params":[`+
          `{"name":"in","type":["[]",[["qr","Integer"],2,3]]},`+
          `{"name":"n","type":["qr","String"],"value":"vi"}],`+
        `"returns":["[]",[["qr","Float"],0]]}]]]`)
}

func toJSON(e Expression) string {
  result := bytes.NewBufferString(``)
  enc := json.NewEncoder(result)
  enc.SetEscapeHTML(false)
  enc.Encode(e.ToPN().ToData())
  result.Truncate(result.Len() - 1)
  return result.String()
}

func expectJSON(t *testing.T, source string, expected string) {
  expr, err := Parse(``, source, false)
  if err != nil {
    t.Errorf(err.Error())
  } else {
    actual := toJSON(expr)
    if expected != actual {
      t.Errorf("expected '%s', got '%s'", expected, actual)
    }
  }
}
