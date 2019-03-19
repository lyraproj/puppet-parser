package parser

import "github.com/lyraproj/puppet-parser/pn"

type ActivityStyle string

const ActivityStyleAction = ActivityStyle(`action`)
const ActivityStyleResource = ActivityStyle(`resource`)
const ActivityStyleStateHandler = ActivityStyle(`stateHandler`)
const ActivityStyleWorkflow = ActivityStyle(`workflow`)

type ActivityExpression struct {
	Positioned
	name       string
	style      ActivityStyle
	properties Expression
	definition Expression
}

func (e *ActivityExpression) AllContents(path []Expression, visitor PathVisitor) {
	DeepVisit(e, path, visitor, e.properties, e.definition)
}

func (e *ActivityExpression) Contents(path []Expression, visitor PathVisitor) {
	ShallowVisit(e, path, visitor, e.properties, e.definition)
}

func (e *ActivityExpression) Name() string {
	return e.name
}

func (e *ActivityExpression) Style() ActivityStyle {
	return e.style
}

func (e *ActivityExpression) Definition() Expression {
	return e.definition
}

func (e *ActivityExpression) Properties() Expression {
	return e.properties
}

func (e *ActivityExpression) ToDefinition() Definition {
	return e
}

func (e *ActivityExpression) ToPN() pn.PN {
	entries := []pn.Entry{
		pn.Literal(e.name).WithName(`name`),
		pn.Literal(string(e.style)).WithName(`style`)}

	if e.properties != nil {
		entries = append(entries, e.properties.ToPN().WithName(`properties`))
	}
	if e.definition != nil {
		entries = append(entries, e.definition.ToPN().WithName(`definition`))
	}
	return pn.Map(entries).AsCall(`activity`)
}
