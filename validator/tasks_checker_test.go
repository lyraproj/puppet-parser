package validator

import "testing"

func TestTasksResourceValidation(t *testing.T) {
	PuppetTasks = true
	defer func() { PuppetTasks = false }()

	expectIssues(t, `class { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `foo { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@foo { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@@foo { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@class { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)

	expectIssues(t, `@@class { my: message => 'syntax ok' }`, ValidateCatalogOperationNotSupported)
}
