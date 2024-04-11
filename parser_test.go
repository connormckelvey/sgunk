package ssg

import (
	"testing"
	"time"

	"github.com/connormckelvey/website/parser"
	"github.com/connormckelvey/website/renderer"
	"github.com/connormckelvey/website/tree"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	t.Log(time.Now().Format("2006/01/02"))
}

func TestParser(t *testing.T) {
	site := tree.Site{
		BaseNode: tree.NewBaseNode("", true),
	}

	siteDir := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1/site")
	themeDir := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1/theme")
	buildDir := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1/_build")
	err := parser.Parse(siteDir, ".", &site, []parser.EntryParser{
		parser.NewBlogEntryParser("blog"),
		&parser.DefaultPageParser{},
	})
	assert.NoError(t, err)
	// spew.Dump(site)

	err = renderer.Render(siteDir, buildDir, &site,
		[]renderer.Renderer{
			&renderer.BlogRenderer{},
			&renderer.DefaultRenderer{},
		},
		renderer.NewRenderContext(siteDir, themeDir, buildDir),
	)
	assert.NoError(t, err)
}
