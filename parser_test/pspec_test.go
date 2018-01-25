package parser

import (
	"testing"
	"github.com/puppetlabs/go-pspec/pspec"
)

func TestPrimitives(t *testing.T) {
	pspec.RunPspecTests(t, `testdata`)
}
