package validator

import (
	. "github.com/puppetlabs/go-parser/parser"
	. "github.com/puppetlabs/go-parser/issue"
)

type tasksChecker struct {
	basicChecker
}

func NewTasksChecker() Checker {
	tasksChecker := &tasksChecker{}
	tasksChecker.initialize(STRICT_ERROR)
	return tasksChecker
}

func (v *tasksChecker) Validate(e Expression) {
	Check(v, e)
}

func (v *tasksChecker) illegalTasksExpression(e Expression) {
	v.Accept(VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED_WHEN_SCRIPTING, e, H{`operation`: e})
}

func (v *tasksChecker) check_Application(e *Application) {
  v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_CapabilityMapping(e *CapabilityMapping) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_CollectExpression(e *CollectExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_HostClassDefinition(e *HostClassDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_NodeDefinition(e *NodeDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_RelationshipExpression(e *RelationshipExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceDefaultsExpression(e *ResourceDefaultsExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceExpression(e *ResourceExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceOverrideExpression(e *ResourceOverrideExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_ResourceTypeDefinition(e *ResourceTypeDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) check_SiteDefinition(e *SiteDefinition) {
	v.illegalTasksExpression(e)
}
