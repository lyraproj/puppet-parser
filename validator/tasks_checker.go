package validator

import (
	"github.com/puppetlabs/go-issues/issue"
	"github.com/puppetlabs/go-parser/parser"
)

type tasksChecker struct {
	basicChecker
}

func NewTasksChecker() Checker {
	tasksChecker := &tasksChecker{}
	tasksChecker.initialize(STRICT_ERROR)
	return tasksChecker
}

func (v *tasksChecker) Validate(e parser.Expression) {
	Check(v, e)
}

func (v *tasksChecker) illegalTasksExpression(e parser.Expression) {
	v.Accept(VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED, e, issue.H{`operation`: e})
}

func (v *tasksChecker) check_Application(e *parser.Application) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_CapabilityMapping(e *parser.CapabilityMapping) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_CollectExpression(e *parser.CollectExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_HostClassDefinition(e *parser.HostClassDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_NodeDefinition(e *parser.NodeDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_RelationshipExpression(e *parser.RelationshipExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceDefaultsExpression(e *parser.ResourceDefaultsExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceExpression(e *parser.ResourceExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceOverrideExpression(e *parser.ResourceOverrideExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceTypeDefinition(e *parser.ResourceTypeDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_SiteDefinition(e *parser.SiteDefinition) {
	v.illegalTasksExpression(e)
}
