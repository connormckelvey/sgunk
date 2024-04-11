package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	p := New(
		WithDir("../testdata/project1"),
		WithExtensions(&BlogExtensionLoader{}),
	)
	err := p.Load()
	assert.NoError(t, err)

	assert.NoError(t, p.Load())
	assert.NoError(t, p.Build())
	// spew.Dump(p.pages)
}
