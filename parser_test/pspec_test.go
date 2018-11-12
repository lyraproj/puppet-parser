package parser_test

import (
	"testing"
	"github.com/puppetlabs/go-pspec/pspec"
)

func TestAll(t *testing.T) {
	pspec.RunPspecTests(t, `testdata`, nil)
}
