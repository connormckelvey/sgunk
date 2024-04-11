package ssg

import (
	"testing"
	"time"

	"github.com/connormckelvey/website/parser"
	"github.com/connormckelvey/website/renderer"
	"github.com/connormckelvey/website/tree"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	t.Log(time.Now().Format("2006/01/02"))
}

func TestGenerator(t *testing.T) {
	projectFS := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1")
	config, err := loadConfigFile(projectFS)
	require.NoError(t, err)

	project := New(
		WithWorkDir("testdata/project1"),
		WithConfig(config),
	)

	err = project.Generate()
	assert.NoError(t, err)
}

func TestParser(t *testing.T) {
	site := tree.Site{
		BaseNode: tree.NewBaseNode("", true),
	}

	siteDir := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1/site")
	themeDir := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1/theme")
	buildDir := afero.NewBasePathFs(afero.NewOsFs(), "testdata/project1/_build")

	p := parser.New(
		parser.WithSiteFS(siteDir),
		parser.WithEntryParsers(
			parser.NewBlogEntryParser("blog"),
			&parser.DefaultPageParser{},
		),
	)

	err := p.Parse(".", &site)
	assert.NoError(t, err)

	r := renderer.New(
		renderer.WithEntryRenderers(
			&renderer.BlogRenderer{},
			&renderer.DefaultRenderer{},
		),
	)
	context := renderer.NewRenderContext(
		renderer.WithFS(siteDir, themeDir, buildDir),
	)
	err = r.Render(&site, context)
	assert.NoError(t, err)
}
