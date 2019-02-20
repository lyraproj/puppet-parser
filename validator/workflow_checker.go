package validator

import (
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/puppet-parser/parser"
)

type workflowChecker struct {
	tasksChecker
}

func NewWorkflowChecker() Checker {
	wfChecker := &workflowChecker{}
	wfChecker.initialize(STRICT_ERROR)
	return wfChecker
}

func (v *workflowChecker) Validate(e parser.Expression) {
	Check(v, e)
}

func (v *workflowChecker) check_ActivityExpression(e *parser.ActivityExpression) {
	switch e.Style() {
	case parser.ActivityStyleAction:
		v.checkAction(e)
	case parser.ActivityStyleResource:
		v.checkResource(e)
	case parser.ActivityStyleStateHandler:
		v.checkStateHandler(e)
	case parser.ActivityStyleWorkflow:
		v.checkWorkflow(e)
	default:
		v.Accept(VALIDATE_INVALID_ACTIVITY_STYLE, e, issue.H{`style`: e.Style()})
	}
}

func (v *workflowChecker) checkAction(e *parser.ActivityExpression) {
}

func (v *workflowChecker) checkStateHandler(e *parser.ActivityExpression) {
}

func (v *workflowChecker) checkResource(e *parser.ActivityExpression) {
}

func (v *workflowChecker) checkWorkflow(e *parser.ActivityExpression) {
}

func (v *workflowChecker) assertValidEntries(e *parser.ActivityExpression, entryNames ...string) {
}
