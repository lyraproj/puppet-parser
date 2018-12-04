package validator

import (
	"testing"

	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-parser/parser"
)

var PuppetTasks = false
var PuppetWorkflow = false

func TestVariableAssignValidation(t *testing.T) {
	expectNoIssues(t, `$x = 'y'`)
}

func TestNumericVariableAssignValidation(t *testing.T) {
	expectIssues(t, `$1 = 'y'`, VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT)
}

func TestMultipleVariableAssign(t *testing.T) {
	expectNoIssues(t, `[$a, $b] = 'y'`)
	expectIssues(t, `[$a, $1] = 'y'`, VALIDATE_ILLEGAL_NUMERIC_ASSIGNMENT)
	expectIssues(t, `[$a, $b['h']] = 'y'`, VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX)
	expectIssues(t, `[$a, $b::z] = 'y'`, VALIDATE_CROSS_SCOPE_ASSIGNMENT)
}

func TestAccessAssignValidation(t *testing.T) {
	expectIssues(t, `$x['h'] = 'y'`, VALIDATE_ILLEGAL_ASSIGNMENT_VIA_INDEX)
	expectIssues(t, `$::x = 'y'`, VALIDATE_CROSS_SCOPE_ASSIGNMENT)
}

func TestAppendsDeletesValidation(t *testing.T) {
	expectIssues(t, `$x += 'y'`, VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED)
	expectIssues(t, `$x -= 'y'`, VALIDATE_APPENDS_DELETES_NO_LONGER_SUPPORTED)
}

func TestNamespaceAssignValidation(t *testing.T) {
	expectIssues(t, `$x::z = 'y'`, VALIDATE_CROSS_SCOPE_ASSIGNMENT)
}

func TestAttributeAppendValidation(t *testing.T) {
	expectNoIssues(t, `Service[apache] { require +> File['apache.pem'] }`)

	expectIssues(t, `service { apache: require +> File['apache.pem'] }`, VALIDATE_ILLEGAL_ATTRIBUTE_APPEND)
}

func TestAttributesOpValidation(t *testing.T) {
	expectNoIssues(t,
		(`
      file { '/tmp/foo':
        ensure => file,
        * => $file_ownership
      }`))

	expectIssues(t,
		issue.Unindent(`
      File <| mode == '0644' |> {
        * => $file_ownership
      }`),
		VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT)

	expectIssues(t,
		issue.Unindent(`
      File {
        ensure => file,
        * => $file_ownership
      }`),
		VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT)

	expectIssues(t,
		issue.Unindent(`
      File['/tmp/foo'] {
        ensure => file,
        * => $file_ownership
      }`),
		VALIDATE_UNSUPPORTED_OPERATOR_IN_CONTEXT)

	expectIssues(t,
		issue.Unindent(`
      file { '/tmp/foo':
        ensure => file,
        * => function foo() {}
      }`),
		VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
}

func TestCallNamedFunctionValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      include apache
      `))

	expectNoIssues(t,
		issue.Unindent(`
      $x = String(123, 16)
      `))

	expectNoIssues(t,
		issue.Unindent(`
      $x = Enum['a', 'b']('a')
      `))

	expectIssues(t,
		issue.Unindent(`
      $x = enum['a', 'b']('a')
      `),
		VALIDATE_ILLEGAL_EXPRESSION)
}

func TestBinaryOpValidation(t *testing.T) {
	expectIssues(t, `notice(define foo() {} < 3)`, VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
	expectNoIssues(t, `notice(true == !false)`)
}

func TestBlockValidation(t *testing.T) {
	expectIssues(t,
		issue.Unindent(`
      ['a', 'b']
      $x = 3
      `),
		VALIDATE_IDEM_EXPRESSION_NOT_LAST)

	expectIssues(t,
		issue.Unindent(`
      case $z {
      2: { true }
      3: { false }
      default: { false }
      }
      $x = 3
      `),
		VALIDATE_IDEM_EXPRESSION_NOT_LAST)

	expectNoIssues(t,
		issue.Unindent(`
      case $z {
      2: { true }
      3: { false }
      default: { $v = 1 }
      }
      $x = 3
      `))

	expectNoIssues(t,
		issue.Unindent(`
      case ($z = 2) {
      2: { true }
      3: { false }
      default: { false }
      }
      $x = 3
      `))

	expectNoIssues(t,
		issue.Unindent(`
      case $z {
      ($y = 2): { true }
      3: { false }
      default: { false }
      }
      $x = 3
      `))

	expectIssues(t,
		issue.Unindent(`
      if $z { 3 } else { 4 }
      $x = 3
      `),
		VALIDATE_IDEM_EXPRESSION_NOT_LAST)

	expectNoIssues(t,
		issue.Unindent(`
      if $z { $v = 3 } else { $v = 4 }
      $x = 3
      `))

	expectIssues(t,
		issue.Unindent(`
      unless $z { 3 }
      $x = 3
      `),
		VALIDATE_IDEM_EXPRESSION_NOT_LAST)

	expectNoIssues(t,
		issue.Unindent(`
      unless $z { $v = 3 }
      $x = 3
      `))

	expectIssues(t,
		issue.Unindent(`
      (3)
      $x = 3
      `),
		VALIDATE_IDEM_EXPRESSION_NOT_LAST)

	expectNoIssues(t,
		issue.Unindent(`
      ($v = 3)
      $x = 3
      `))
}

func TestCallMethodValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      $x = $y.size()
    `))

	expectIssues(t,
		issue.Unindent(`
      $x = $y.Size()`),
		VALIDATE_ILLEGAL_EXPRESSION)
}

func TestCapabilityMappingValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      Something produces Foo {}
      `))

	expectNoIssues(t,
		issue.Unindent(`
      Something[A] produces Foo {}
      `))

	expectIssues(t,
		issue.Unindent(`
      something produces Foo {}
      `),
		VALIDATE_ILLEGAL_CLASSREF)

	expectIssues(t,
		issue.Unindent(`
      Something produces foo {}
      `),
		VALIDATE_ILLEGAL_CLASSREF)

	expectIssues(t,
		issue.Unindent(`
      Something['A', 'B'] produces Foo {}
      `),
		VALIDATE_ILLEGAL_EXPRESSION)
}

func TestCaseExpressionValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      case $x {
        'a': { true }
        default: { 'false' }
      }`))

	expectIssues(t,
		issue.Unindent(`
      case $x {
        'a': { true }
        default: { 'false' }
        default: { 'true' }
      }`),
		VALIDATE_DUPLICATE_DEFAULT)

	expectIssues(t,
		issue.Unindent(`
      case $x {
        function foo() {}: { true }
        default: { false }
      }`),
		VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
}

func TestCollectValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      User <| groups == 'admin' |>`))

	expectNoIssues(t,
		issue.Unindent(`
      User <| (groups == 'admin') |>`))

	expectNoIssues(t,
		issue.Unindent(`
      User <| present |>`))

	expectNoIssues(t,
		issue.Unindent(`
      User <| present and groups == 'admin' |>`))

	expectIssues(t,
		issue.Unindent(`
      User <| $x + 1 |>`),
		VALIDATE_ILLEGAL_QUERY_EXPRESSION)

	expectIssues(t,
		issue.Unindent(`
      User <| groups >= 'admin' |>`),
		VALIDATE_ILLEGAL_QUERY_EXPRESSION)

	expectIssues(t,
		issue.Unindent(`
      user <| groups == 'admin' |>`),
		VALIDATE_ILLEGAL_EXPRESSION)

	expectNoIssues(t,
		issue.Unindent(`
      User <<| groups == 'admin' |>>`))

	expectNoIssues(t,
		issue.Unindent(`
      User <<| (groups == 'admin') |>>`))

	expectNoIssues(t,
		issue.Unindent(`
      User <<| present |>>`))

	expectNoIssues(t,
		issue.Unindent(`
      User <<| present and groups == 'admin' |>>`))

	expectIssues(t,
		issue.Unindent(`
      User <<| $x + 1 |>>`),
		VALIDATE_ILLEGAL_QUERY_EXPRESSION)

	expectIssues(t,
		issue.Unindent(`
      User <<| groups >= 'admin' |>>`),
		VALIDATE_ILLEGAL_QUERY_EXPRESSION)

	expectIssues(t,
		issue.Unindent(`
      user <<| groups == 'admin' |>>`),
		VALIDATE_ILLEGAL_EXPRESSION)
}

func TestEppValidation(t *testing.T) {
	expectNoIssuesEPP(t,
		issue.Unindent(`
      <%-| $a, $b |-%>
      This is $a <%= $a %>`))

	expectIssuesEPP(t,
		issue.Unindent(`
      <%-| $a, $b, $a |-%>
      This is $a <%= $a %>`),
		VALIDATE_DUPLICATE_PARAMETER)

	expectIssuesEPP(t,
		issue.Unindent(`
      <%-| $a, *$b |-%>
      This is $a <%= $a %>`),
		VALIDATE_CAPTURES_REST_NOT_SUPPORTED)
}

func TestFunctionDefinitionValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      function foo() {}`))

	expectNoIssues(t,
		issue.Unindent(`
      function foo($a, *$b) {}`))

	expectIssues(t,
		issue.Unindent(`
      function foo($a, *$b, $c) {}`),
		VALIDATE_CAPTURES_REST_NOT_LAST)

	expectIssues(t,
		issue.Unindent(`
      function foo($a::b) {}`),
		VALIDATE_ILLEGAL_PARAMETER_NAME)

	expectIssues(t,
		issue.Unindent(`
      function foo($a = ($x = 3)) {}`),
		VALIDATE_ILLEGAL_ASSIGNMENT_CONTEXT)

	expectNoIssues(t,
		issue.Unindent(`
      function foo($a = bar() |$p| { $p = 1 }) {}`))

	expectIssues(t,
		issue.Unindent(`
      function foo() {
        function bar() {}
      }`),
		VALIDATE_NOT_ABSOLUTE_TOP_LEVEL)

	expectIssues(t,
		issue.Unindent(`
      function foo() >> Application {
      }`),
		VALIDATE_FUTURE_RESERVED_WORD)

	expectIssues(t,
		issue.Unindent(`
      function foo() >> Application[X] {
      }`),
		VALIDATE_FUTURE_RESERVED_WORD)

	expectIssues(t,
		issue.Unindent(`
      function variant() {
      }`),
		VALIDATE_RESERVED_TYPE_NAME)

	expectIssues(t,
		issue.Unindent(`
      function Foo() {
      }`),
		VALIDATE_ILLEGAL_DEFINITION_NAME)
}

func TestHostClassDefinitionValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      class foo {}`))

	expectNoIssues(t,
		issue.Unindent(`
      class foo {
        class bar {}
      }`))

	expectIssues(t,
		issue.Unindent(`
      class foo($a, *$b) {
      }`),
		VALIDATE_CAPTURES_REST_NOT_SUPPORTED)

	expectIssues(t,
		issue.Unindent(`
      class foo($title, $b) {
      }`),
		VALIDATE_RESERVED_PARAMETER)

	expectIssues(t,
		issue.Unindent(`
      class foo($name, $b) {
      }`),
		VALIDATE_RESERVED_PARAMETER)

	expectIssues(t,
		issue.Unindent(`
      class foo($a) {
        [$a]
      }`),
		VALIDATE_IDEM_NOT_ALLOWED_LAST)

	expectIssues(t,
		issue.Unindent(`
      class variant {
      }`),
		VALIDATE_RESERVED_TYPE_NAME)

	expectIssues(t,
		issue.Unindent(`
      class Foo {
      }`),
		VALIDATE_ILLEGAL_DEFINITION_NAME)
}

func TestLiteralHashValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      $x = {
        1 => 'one',
        '2' => 'two',
        3.0 => 'three',
        true => 'true',
        undef => 'undef',
        default => 'default'
      }`))

	expectIssues(t,
		issue.Unindent(`
      $x = {
        1 => 'one',
        '2' => 'two',
        1 => 'one again'
      }`),
		VALIDATE_DUPLICATE_KEY)

	expectIssues(t,
		issue.Unindent(`
      $x = {
        'func' => define foo() {}
      }`),
		VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
}

func TestLiteralListValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      $x = [
        1, '2', 3.0, four, 'five', true, undef, default, [1, 2], {'one' => 1}
      ]`))

	expectIssues(t,
		issue.Unindent(`
      $x = [define foo() {}]`),
		VALIDATE_NOT_TOP_LEVEL, VALIDATE_NOT_RVALUE)
}

func TestReourceDefinitionValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      define foo() {}`))

	expectNoIssues(t,
		issue.Unindent(`
      class foo() {
        define bar() {}
      }`))

	expectIssues(t,
		issue.Unindent(`
      define foo() {
        define bar() {}
      }`),
		VALIDATE_NOT_TOP_LEVEL)

	expectIssues(t,
		issue.Unindent(`
      define foo($a, *$b) {
      }`),
		VALIDATE_CAPTURES_REST_NOT_SUPPORTED)

	expectIssues(t,
		issue.Unindent(`
      define foo($title, $b) {
      }`),
		VALIDATE_RESERVED_PARAMETER)

	expectIssues(t,
		issue.Unindent(`
      define foo($name, $b) {
      }`),
		VALIDATE_RESERVED_PARAMETER)

	expectIssues(t,
		issue.Unindent(`
      define foo($a) {
        [$a]
      }`),
		VALIDATE_IDEM_NOT_ALLOWED_LAST)

	expectIssues(t,
		issue.Unindent(`
      define variant() {
      }`),
		VALIDATE_RESERVED_TYPE_NAME)

	expectIssues(t,
		issue.Unindent(`
      define Foo() {
      }`),
		VALIDATE_ILLEGAL_DEFINITION_NAME)
}

func TestRelationshipValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      package { 'openssh-server':
        ensure => present,
      } ->
      file { '/etc/ssh/sshd_config':
        ensure => file,
        mode   => '0600',
        source => 'puppet:///modules/sshd/sshd_config',
      }`))

	expectIssues(t,
		issue.Unindent(`
      package { 'openssh-server':
        ensure => present,
      } ->
      node example.com {}`),
		VALIDATE_NOT_RVALUE, VALIDATE_NOT_TOP_LEVEL)

	expectIssues(t,
		issue.Unindent(`
      class my_class {} ->
      package { 'openssh-server':
        ensure => present,
      }`),
		VALIDATE_NOT_RVALUE, VALIDATE_NOT_TOP_LEVEL)
}

func TestResourceBodyValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      file { '/tmp/foo':
        ensure => file,
        * => $file_ownership
      }`))

	expectIssues(t,
		issue.Unindent(`
      file { '/tmp/foo':
        ensure => file,
        * => $file_ownership,
        * => $file_mode_content
      }`), VALIDATE_MULTIPLE_ATTRIBUTES_UNFOLD)
}

func TestResourceValidation(t *testing.T) {
	expectNoIssues(t, `class { my: message => 'syntax ok' }`)

	expectNoIssues(t, `@foo { my: message => 'syntax ok' }`)

	expectNoIssues(t, `@@foo { my: message => 'syntax ok' }`)

	expectIssues(t, `@class { my: message => 'syntax ok' }`, VALIDATE_NOT_VIRTUALIZABLE)

	expectIssues(t, `@@class { my: message => 'syntax ok' }`, VALIDATE_NOT_VIRTUALIZABLE)
}

func TestResourceDefaultValidation(t *testing.T) {
	expectNoIssues(t, `Something { message => 'syntax ok' }`)

	expectIssues(t, `@Something { message => 'syntax ok' }`, VALIDATE_NOT_VIRTUALIZABLE)

	expectIssues(t, `@@Something { message => 'syntax ok' }`, VALIDATE_NOT_VIRTUALIZABLE)
}

func TestResourceOverrideValidation(t *testing.T) {
	expectNoIssues(t, `Something['here'] { message => 'syntax ok' }`)

	expectIssues(t, `@Something['here'] { message => 'syntax ok' }`, VALIDATE_NOT_VIRTUALIZABLE)

	expectIssues(t, `@@Something['here'] { message => 'syntax ok' }`, VALIDATE_NOT_VIRTUALIZABLE)
}

func TestNodeDefinitionValidation(t *testing.T) {
	expectNoIssues(t, `node foo {}`)

	expectNoIssues(t, `node 'foo' {}`)

	expectNoIssues(t, `node "foo" {}`)

	expectNoIssues(t, `node MyNode {}`)

	expectNoIssues(t, `node 192.168.0.10 {}`)

	expectNoIssues(t, `node /.*\.example\.com/ {}`)

	expectNoIssues(t, `node example.com {}`)

	expectIssues(t, `node 'bad char' {}`, VALIDATE_ILLEGAL_HOSTNAME_CHARS)

	expectIssues(t, `node "bad char" {}`, VALIDATE_ILLEGAL_HOSTNAME_CHARS)

	expectIssues(t, `node "not${here}" {}`, VALIDATE_ILLEGAL_HOSTNAME_INTERPOLATION)
}

func TestReservedWordValidation(t *testing.T) {
	expectIssues(t, `$x = private`, VALIDATE_RESERVED_WORD)
}

func TestSelectorExpressionValidation(t *testing.T) {
	expectNoIssues(t,
		issue.Unindent(`
      $role = $facts['os']['name'] ? {
        'Solaris'           => role::solaris,
        'RedHat'            => role::redhat,
        /^(Debian|Ubuntu)$/ => role::debian,
        default             => role::generic,
      }`))

	expectIssues(t,
		issue.Unindent(`
      $role = $facts['os']['name'] ? {
        'Solaris'           => role::solaris,
        default             => role::generic,
        'RedHat'            => role::redhat,
        default             => role::generic,
      }`), VALIDATE_DUPLICATE_DEFAULT)
}

func TestTypeAliasValidation(t *testing.T) {
	expectNoIssues(t, `type MyType = Integer`)

	expectNoIssues(t, `type MyType = Variant[Integer, String]`)

	expectIssues(t, `type Variant = MyType`, VALIDATE_RESERVED_TYPE_NAME)

	expectIssues(t, `type ::MyType = Integer`, VALIDATE_ILLEGAL_DEFINITION_NAME)
}

func TestTypeMappingValidation(t *testing.T) {
	expectNoIssues(t, `type Runtime[ruby, 'MyModule::MyObject'] = MyPackage::MyObject`)

	expectNoIssues(t, `type Runtime[ruby, [/^MyPackage::(\w+)$/, 'MyModule::\1']] = [/^MyModule::(\w+)$/, 'MyPackage::\1']`)

	expectIssues(t,
		`type Runtime[ruby, [/^MyPackage::(\w+)$/, 'MyModule::\1']] = MyPackage::MyObject`,
		VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING)

	expectIssues(t,
		`type Runtime[ruby, 'MyModule::MyObject'] = [/^MyModule::(\w+)$/, 'MyPackage::\1']`,
		VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING)

	expectIssues(t,
		`type Runtime[ruby, [/^MyPackage::(\w+)$/, 'MyModule::\1']] = $x`,
		VALIDATE_ILLEGAL_REGEXP_TYPE_MAPPING)

	expectIssues(t,
		`type Runtime[ruby, 'MyModule::MyObject'] = $x`,
		VALIDATE_ILLEGAL_SINGLE_TYPE_MAPPING)

	expectIssues(t,
		`type Pattern[/^MyPackage::(\w+)$/, 'MyModule::\1'] = [/^MyModule::(\w+)$/, 'MyPackage::\1']`,
		VALIDATE_UNSUPPORTED_EXPRESSION)
}

func expectNoIssues(t *testing.T, str string) {
	expectIssuesX(t, str, []parser.Option{})
}

func expectNoIssuesEPP(t *testing.T, str string) {
	expectIssuesX(t, str, []parser.Option{parser.PARSER_EPP_MODE})
}

func expectIssues(t *testing.T, str string, expectedIssueCodes ...issue.Code) {
	expectIssuesX(t, str, []parser.Option{}, expectedIssueCodes...)
}

func expectIssuesEPP(t *testing.T, str string, expectedIssueCodes ...issue.Code) {
	expectIssuesX(t, str, []parser.Option{parser.PARSER_EPP_MODE}, expectedIssueCodes...)
}

func expectIssuesX(t *testing.T, str string, parserOptions []parser.Option, expectedIssueCodes ...issue.Code) {
	issues := parseAndValidate(t, str, parserOptions...)
	if issues == nil {
		return
	}
	fail := false
nextCode:
	for _, expectedIssueCode := range expectedIssueCodes {
		for _, issue := range issues {
			if expectedIssueCode == issue.Code() {
				continue nextCode
			}
		}
		fail = true
		t.Logf(`Expected issue '%s' but it was not produced`, expectedIssueCode)
	}

nextIssue:
	for _, issue := range issues {
		for _, expectedIssueCode := range expectedIssueCodes {
			if expectedIssueCode == issue.Code() {
				continue nextIssue
			}
		}
		fail = true
		t.Logf(`Unexpected issue %s: '%s'`, issue.Code(), issue.String())
	}
	if fail {
		t.Fail()
	}
}

func parseAndValidate(t *testing.T, str string, parserOptions ...parser.Option) []issue.Reported {
	if PuppetTasks {
		parserOptions = append([]parser.Option{parser.PARSER_TASKS_ENABLED}, parserOptions...)
	}
	if PuppetWorkflow {
		parserOptions = append([]parser.Option{parser.PARSER_WORKFLOW_ENABLED}, parserOptions...)
	}

	var v Validator
	if PuppetTasks || PuppetWorkflow {
		if expr := parse(t, str, parserOptions...); expr != nil {
			if PuppetWorkflow {
				v = ValidateWorkflow(expr)
			} else {
				v = ValidateTasks(expr)
			}
			return v.Issues()
		}
	} else if expr := parse(t, str, parserOptions...); expr != nil {
		v = ValidatePuppet(expr, STRICT_ERROR)
		return v.Issues()
	}
	return nil
}

func parse(t *testing.T, str string, parserOptions ...parser.Option) *parser.Program {
	expr, err := parser.CreateParser(parserOptions...).Parse(``, str, false)
	if err != nil {
		t.Errorf(err.Error())
		return nil
	}
	block, ok := expr.(*parser.Program)
	if !ok {
		t.Errorf("'%s' did not parse to a program", str)
		return nil
	}
	return block
}
