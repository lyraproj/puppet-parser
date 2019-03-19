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
	expectIssues(t, `$1 = 'y'`, ValidateIllegalNumericAssignment)
}

func TestMultipleVariableAssign(t *testing.T) {
	expectNoIssues(t, `[$a, $b] = 'y'`)
	expectIssues(t, `[$a, $1] = 'y'`, ValidateIllegalNumericAssignment)
	expectIssues(t, `[$a, $b['h']] = 'y'`, ValidateIllegalAssignmentViaIndex)
	expectIssues(t, `[$a, $b::z] = 'y'`, ValidateCrossScopeAssignment)
}

func TestAccessAssignValidation(t *testing.T) {
	expectIssues(t, `$x['h'] = 'y'`, ValidateIllegalAssignmentViaIndex)
	expectIssues(t, `$::x = 'y'`, ValidateCrossScopeAssignment)
}

func TestAppendsDeletesValidation(t *testing.T) {
	expectIssues(t, `$x += 'y'`, ValidateAppendsDeletesNoLongerSupported)
	expectIssues(t, `$x -= 'y'`, ValidateAppendsDeletesNoLongerSupported)
}

func TestNamespaceAssignValidation(t *testing.T) {
	expectIssues(t, `$x::z = 'y'`, ValidateCrossScopeAssignment)
}

func TestAttributeAppendValidation(t *testing.T) {
	expectNoIssues(t, `Service[apache] { require +> File['apache.pem'] }`)

	expectIssues(t, `service { apache: require +> File['apache.pem'] }`, ValidateIllegalAttributeAppend)
}

func TestAttributesOpValidation(t *testing.T) {
	expectNoIssues(t, `
      file { '/tmp/foo':
        ensure => file,
        * => $file_ownership
      }`)

	expectIssues(t,
		issue.Unindent(`
      File <| mode == '0644' |> {
        * => $file_ownership
      }`),
		ValidateUnsupportedOperatorInContext)

	expectIssues(t,
		issue.Unindent(`
      File {
        ensure => file,
        * => $file_ownership
      }`),
		ValidateUnsupportedOperatorInContext)

	expectIssues(t,
		issue.Unindent(`
      File['/tmp/foo'] {
        ensure => file,
        * => $file_ownership
      }`),
		ValidateUnsupportedOperatorInContext)

	expectIssues(t,
		issue.Unindent(`
      file { '/tmp/foo':
        ensure => file,
        * => function foo() {}
      }`),
		ValidateNotTopLevel, ValidateNotRvalue)
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
		ValidateIllegalExpression)
}

func TestBinaryOpValidation(t *testing.T) {
	expectIssues(t, `notice(define foo() {} < 3)`, ValidateNotTopLevel, ValidateNotRvalue)
	expectNoIssues(t, `notice(true == !false)`)
}

func TestBlockValidation(t *testing.T) {
	expectIssues(t,
		issue.Unindent(`
      ['a', 'b']
      $x = 3
      `),
		ValidateIdemExpressionNotLast)

	expectIssues(t,
		issue.Unindent(`
      case $z {
      2: { true }
      3: { false }
      default: { false }
      }
      $x = 3
      `),
		ValidateIdemExpressionNotLast)

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
		ValidateIdemExpressionNotLast)

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
		ValidateIdemExpressionNotLast)

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
		ValidateIdemExpressionNotLast)

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
		ValidateIllegalExpression)
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
		ValidateIllegalClassref)

	expectIssues(t,
		issue.Unindent(`
      Something produces foo {}
      `),
		ValidateIllegalClassref)

	expectIssues(t,
		issue.Unindent(`
      Something['A', 'B'] produces Foo {}
      `),
		ValidateIllegalExpression)
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
		ValidateDuplicateDefault)

	expectIssues(t,
		issue.Unindent(`
      case $x {
        function foo() {}: { true }
        default: { false }
      }`),
		ValidateNotTopLevel, ValidateNotRvalue)
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
		ValidateIllegalQueryExpression)

	expectIssues(t,
		issue.Unindent(`
      User <| groups >= 'admin' |>`),
		ValidateIllegalQueryExpression)

	expectIssues(t,
		issue.Unindent(`
      user <| groups == 'admin' |>`),
		ValidateIllegalExpression)

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
		ValidateIllegalQueryExpression)

	expectIssues(t,
		issue.Unindent(`
      User <<| groups >= 'admin' |>>`),
		ValidateIllegalQueryExpression)

	expectIssues(t,
		issue.Unindent(`
      user <<| groups == 'admin' |>>`),
		ValidateIllegalExpression)
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
		ValidateDuplicateParameter)

	expectIssuesEPP(t,
		issue.Unindent(`
      <%-| $a, *$b |-%>
      This is $a <%= $a %>`),
		ValidateCapturesRestNotSupported)
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
		ValidateCapturesRestNotLast)

	expectIssues(t,
		issue.Unindent(`
      function foo($a::b) {}`),
		ValidateIllegalParameterName)

	expectIssues(t,
		issue.Unindent(`
      function foo($a = ($x = 3)) {}`),
		ValidateIllegalAssignmentContext)

	expectNoIssues(t,
		issue.Unindent(`
      function foo($a = bar() |$p| { $p = 1 }) {}`))

	expectIssues(t,
		issue.Unindent(`
      function foo() {
        function bar() {}
      }`),
		ValidateNotAbsoluteTopLevel)

	expectIssues(t,
		issue.Unindent(`
      function foo() >> Application {
      }`),
		ValidateFutureReservedWord)

	expectIssues(t,
		issue.Unindent(`
      function foo() >> Application[X] {
      }`),
		ValidateFutureReservedWord)

	expectIssues(t,
		issue.Unindent(`
      function variant() {
      }`),
		ValidateReservedTypeName)

	expectIssues(t,
		issue.Unindent(`
      function Foo() {
      }`),
		ValidateIllegalDefinitionName)
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
		ValidateCapturesRestNotSupported)

	expectIssues(t,
		issue.Unindent(`
      class foo($title, $b) {
      }`),
		ValidateReservedParameter)

	expectIssues(t,
		issue.Unindent(`
      class foo($name, $b) {
      }`),
		ValidateReservedParameter)

	expectIssues(t,
		issue.Unindent(`
      class foo($a) {
        [$a]
      }`),
		ValidateIdemNotAllowedLast)

	expectIssues(t,
		issue.Unindent(`
      class variant {
      }`),
		ValidateReservedTypeName)

	expectIssues(t,
		issue.Unindent(`
      class Foo {
      }`),
		ValidateIllegalDefinitionName)
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
		ValidateDuplicateKey)

	expectIssues(t,
		issue.Unindent(`
      $x = {
        'func' => define foo() {}
      }`),
		ValidateNotTopLevel, ValidateNotRvalue)
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
		ValidateNotTopLevel, ValidateNotRvalue)
}

func TestResourceDefinitionValidation(t *testing.T) {
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
		ValidateNotTopLevel)

	expectIssues(t,
		issue.Unindent(`
      define foo($a, *$b) {
      }`),
		ValidateCapturesRestNotSupported)

	expectIssues(t,
		issue.Unindent(`
      define foo($title, $b) {
      }`),
		ValidateReservedParameter)

	expectIssues(t,
		issue.Unindent(`
      define foo($name, $b) {
      }`),
		ValidateReservedParameter)

	expectIssues(t,
		issue.Unindent(`
      define foo($a) {
        [$a]
      }`),
		ValidateIdemNotAllowedLast)

	expectIssues(t,
		issue.Unindent(`
      define variant() {
      }`),
		ValidateReservedTypeName)

	expectIssues(t,
		issue.Unindent(`
      define Foo() {
      }`),
		ValidateIllegalDefinitionName)
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
		ValidateNotRvalue, ValidateNotTopLevel)

	expectIssues(t,
		issue.Unindent(`
      class my_class {} ->
      package { 'openssh-server':
        ensure => present,
      }`),
		ValidateNotRvalue, ValidateNotTopLevel)
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
      }`), ValidateMultipleAttributesUnfold)
}

func TestResourceValidation(t *testing.T) {
	expectNoIssues(t, `class { my: message => 'syntax ok' }`)

	expectNoIssues(t, `@foo { my: message => 'syntax ok' }`)

	expectNoIssues(t, `@@foo { my: message => 'syntax ok' }`)

	expectIssues(t, `@class { my: message => 'syntax ok' }`, ValidateNotVirtualizable)

	expectIssues(t, `@@class { my: message => 'syntax ok' }`, ValidateNotVirtualizable)
}

func TestResourceDefaultValidation(t *testing.T) {
	expectNoIssues(t, `Something { message => 'syntax ok' }`)

	expectIssues(t, `@Something { message => 'syntax ok' }`, ValidateNotVirtualizable)

	expectIssues(t, `@@Something { message => 'syntax ok' }`, ValidateNotVirtualizable)
}

func TestResourceOverrideValidation(t *testing.T) {
	expectNoIssues(t, `Something['here'] { message => 'syntax ok' }`)

	expectIssues(t, `@Something['here'] { message => 'syntax ok' }`, ValidateNotVirtualizable)

	expectIssues(t, `@@Something['here'] { message => 'syntax ok' }`, ValidateNotVirtualizable)
}

func TestNodeDefinitionValidation(t *testing.T) {
	expectNoIssues(t, `node foo {}`)

	expectNoIssues(t, `node 'foo' {}`)

	expectNoIssues(t, `node "foo" {}`)

	expectNoIssues(t, `node MyNode {}`)

	expectNoIssues(t, `node 192.168.0.10 {}`)

	expectNoIssues(t, `node /.*\.example\.com/ {}`)

	expectNoIssues(t, `node example.com {}`)

	expectIssues(t, `node 'bad char' {}`, ValidateIllegalHostnameChars)

	expectIssues(t, `node "bad char" {}`, ValidateIllegalHostnameChars)

	expectIssues(t, `node "not${here}" {}`, ValidateIllegalHostnameInterpolation)
}

func TestReservedWordValidation(t *testing.T) {
	expectIssues(t, `$x = private`, ValidateReservedWord)
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
      }`), ValidateDuplicateDefault)
}

func TestTypeAliasValidation(t *testing.T) {
	expectNoIssues(t, `type MyType = Integer`)

	expectNoIssues(t, `type MyType = Variant[Integer, String]`)

	expectIssues(t, `type Variant = MyType`, ValidateReservedTypeName)

	expectIssues(t, `type ::MyType = Integer`, ValidateIllegalDefinitionName)
}

func TestTypeMappingValidation(t *testing.T) {
	expectNoIssues(t, `type Runtime[ruby, 'MyModule::MyObject'] = MyPackage::MyObject`)

	expectNoIssues(t, `type Runtime[ruby, [/^MyPackage::(\w+)$/, 'MyModule::\1']] = [/^MyModule::(\w+)$/, 'MyPackage::\1']`)

	expectIssues(t,
		`type Runtime[ruby, [/^MyPackage::(\w+)$/, 'MyModule::\1']] = MyPackage::MyObject`,
		ValidateIllegalRegexpTypeMapping)

	expectIssues(t,
		`type Runtime[ruby, 'MyModule::MyObject'] = [/^MyModule::(\w+)$/, 'MyPackage::\1']`,
		ValidateIllegalSingleTypeMapping)

	expectIssues(t,
		`type Runtime[ruby, [/^MyPackage::(\w+)$/, 'MyModule::\1']] = $x`,
		ValidateIllegalRegexpTypeMapping)

	expectIssues(t,
		`type Runtime[ruby, 'MyModule::MyObject'] = $x`,
		ValidateIllegalSingleTypeMapping)

	expectIssues(t,
		`type Pattern[/^MyPackage::(\w+)$/, 'MyModule::\1'] = [/^MyModule::(\w+)$/, 'MyPackage::\1']`,
		ValidateUnsupportedExpression)
}

func expectNoIssues(t *testing.T, str string) {
	expectIssuesX(t, str, []parser.Option{})
}

func expectNoIssuesEPP(t *testing.T, str string) {
	expectIssuesX(t, str, []parser.Option{parser.EppMode})
}

func expectIssues(t *testing.T, str string, expectedIssueCodes ...issue.Code) {
	expectIssuesX(t, str, []parser.Option{}, expectedIssueCodes...)
}

func expectIssuesEPP(t *testing.T, str string, expectedIssueCodes ...issue.Code) {
	expectIssuesX(t, str, []parser.Option{parser.EppMode}, expectedIssueCodes...)
}

func expectIssuesX(t *testing.T, str string, parserOptions []parser.Option, expectedIssueCodes ...issue.Code) {
	issues := parseAndValidate(t, str, parserOptions...)
	if issues == nil {
		return
	}
	fail := false
nextCode:
	for _, expectedIssueCode := range expectedIssueCodes {
		for _, i := range issues {
			if expectedIssueCode == i.Code() {
				continue nextCode
			}
		}
		fail = true
		t.Logf(`Expected issue '%s' but it was not produced`, expectedIssueCode)
	}

nextIssue:
	for _, i := range issues {
		for _, expectedIssueCode := range expectedIssueCodes {
			if expectedIssueCode == i.Code() {
				continue nextIssue
			}
		}
		fail = true
		t.Logf(`Unexpected issue %s: '%s'`, i.Code(), i.String())
	}
	if fail {
		t.Fail()
	}
}

func parseAndValidate(t *testing.T, str string, parserOptions ...parser.Option) []issue.Reported {
	if PuppetTasks {
		parserOptions = append([]parser.Option{parser.TasksEnabled}, parserOptions...)
	}
	if PuppetWorkflow {
		parserOptions = append([]parser.Option{parser.WorkflowEnabled}, parserOptions...)
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
		v = ValidatePuppet(expr, StrictError)
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
