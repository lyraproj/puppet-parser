package parser

import "github.com/lyraproj/puppet-parser/pn"

type StepStyle string

const StepStyleAction = StepStyle(`action`)
const StepStyleResource = StepStyle(`resource`)
const StepStyleStateHandler = StepStyle(`stateHandler`)
const StepStyleWorkflow = StepStyle(`workflow`)

type StepExpression struct {
	Positioned
	name       string
	style      StepStyle
	properties Expression
	definition Expression
}

func (e *StepExpression) AllContents(path []Expression, visitor PathVisitor) {
	DeepVisit(e, path, visitor, e.properties, e.definition)
}

func (e *StepExpression) Contents(path []Expression, visitor PathVisitor) {
	ShallowVisit(e, path, visitor, e.properties, e.definition)
}

func (e *StepExpression) Name() string {
	return e.name
}

func (e *StepExpression) Style() StepStyle {
	return e.style
}

func (e *StepExpression) Definition() Expression {
	return e.definition
}

func (e *StepExpression) Properties() Expression {
	return e.properties
}

func (e *StepExpression) ToDefinition() Definition {
	return e
}

func (e *StepExpression) ToPN() pn.PN {
	entries := []pn.Entry{
		pn.Literal(e.name).WithName(`name`),
		pn.Literal(string(e.style)).WithName(`style`)}

	if e.properties != nil {
		entries = append(entries, e.properties.ToPN().WithName(`properties`))
	}
	if e.definition != nil {
		entries = append(entries, e.definition.ToPN().WithName(`definition`))
	}
	return pn.Map(entries).AsCall(`step`)
}
