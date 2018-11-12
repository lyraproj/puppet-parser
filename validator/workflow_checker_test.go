package validator

import "testing"

func TestWorkflowResourceValidation(t *testing.T) {
	PuppetWorkflow = true
	defer func() { PuppetWorkflow = false }()

	expectIssues(t, `class { my: message => 'syntax ok' }`, VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED)

	expectIssues(t, `foo { my: message => 'syntax ok' }`, VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED)

	expectIssues(t, `@foo { my: message => 'syntax ok' }`, VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED)

	expectIssues(t, `@@foo { my: message => 'syntax ok' }`, VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED)

	expectIssues(t, `@class { my: message => 'syntax ok' }`, VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED)

	expectIssues(t, `@@class { my: message => 'syntax ok' }`, VALIDATE_CATALOG_OPERATION_NOT_SUPPORTED)

	expectNoIssues(t, `workflow foo {}`)

}
