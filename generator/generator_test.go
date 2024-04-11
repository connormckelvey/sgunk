package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	p := NewProject(
		WithProjectDir("testdata/project1"),
		WithContentLoader(NewPageLoader()),
		WithPageTransformers(
			NewFrontMatterTransformer(),
			NewBlogPostTransformer(),
			NewTMPLRunTransformer(),
			NewMarkdownTransformer(),
		),
	)
	err := p.Load()
	assert.NoError(t, err)

	assert.NoError(t, p.Build())
	// spew.Dump(p)

	for _, page := range p.Pages {
		t.Log(page.Content.String())
	}
}
