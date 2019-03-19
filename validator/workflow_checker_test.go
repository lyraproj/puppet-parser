package validator

import "testing"

func TestWorkflowResourceValidation(t *testing.T) {
	PuppetWorkflow = true
	defer func() { PuppetWorkflow = false }()

	expectIssues(t, `class { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `foo { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@foo { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@@foo { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@class { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@@class { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectNoIssues(t, `workflow foo {}`)

}
