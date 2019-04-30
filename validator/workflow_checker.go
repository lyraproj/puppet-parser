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
	wfChecker.initialize(StrictError)
	return wfChecker
}

func (v *workflowChecker) Validate(e parser.Expression) {
	Check(v, e)
}

func (v *workflowChecker) checkStepExpression(e *parser.StepExpression) {
	switch e.Style() {
	case parser.StepStyleAction:
		v.checkAction(e)
	case parser.StepStyleResource:
		v.checkResource(e)
	case parser.StepStyleStateHandler:
		v.checkStateHandler(e)
	case parser.StepStyleWorkflow:
		v.checkWorkflow(e)
	default:
		v.Accept(ValidateInvalidStepStyle, e, issue.H{`style`: e.Style()})
	}
}

func (v *workflowChecker) checkAction(e *parser.StepExpression) {
}

func (v *workflowChecker) checkStateHandler(e *parser.StepExpression) {
}

func (v *workflowChecker) checkResource(e *parser.StepExpression) {
}

func (v *workflowChecker) checkWorkflow(e *parser.StepExpression) {
}
