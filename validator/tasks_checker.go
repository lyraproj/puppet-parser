package validator

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-parser/parser"
)

type tasksChecker struct {
	basicChecker
}

func NewTasksChecker() Checker {
	tasksChecker := &tasksChecker{}
	tasksChecker.initialize(StrictError)
	return tasksChecker
}

func (v *tasksChecker) Validate(e parser.Expression) {
	Check(v, e)
}

func (v *tasksChecker) illegalTasksExpression(e parser.Expression) {
	v.Accept(ValidateCatalogOperationNotSupported, e, issue.H{`operation`: e})
}

func (v *tasksChecker) checkApplication(e *parser.Application) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkCapabilityMapping(e *parser.CapabilityMapping) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkCollectExpression(e *parser.CollectExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkHostClassDefinition(e *parser.HostClassDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkNodeDefinition(e *parser.NodeDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkRelationshipExpression(e *parser.RelationshipExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkResourceDefaultsExpression(e *parser.ResourceDefaultsExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkResourceExpression(e *parser.ResourceExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkResourceOverrideExpression(e *parser.ResourceOverrideExpression) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkResourceTypeDefinition(e *parser.ResourceTypeDefinition) {
	v.illegalTasksExpression(e)
}

func (v *tasksChecker) checkSiteDefinition(e *parser.SiteDefinition) {
	v.illegalTasksExpression(e)
}
