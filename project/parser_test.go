package project

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	t.Log(time.Now().Format("2006/01/02"))
}

func TestParser(t *testing.T) {
	site := Site{
		BaseNode: NewBaseNode("", true),
	}
	siteDir := afero.NewBasePathFs(afero.NewOsFs(), "../testdata/project1/site")
	themeDir := afero.NewBasePathFs(afero.NewOsFs(), "../testdata/project1/theme")
	buildDir := afero.NewBasePathFs(afero.NewOsFs(), "../testdata/project1/_build")
	err := Parse(siteDir, ".", &site, []EntryParser{&BlogEntryParser{root: "blog"}, &DefaultPageParser{}})
	assert.NoError(t, err)
	// spew.Dump(site)

	assert.NoError(t, Render(siteDir, buildDir, &site, []Renderer{&BlogRenderer{}, &DefaultRenderer{}}, &RenderContext{
		themeFS:  themeDir,
		buildFS:  buildDir,
		siteFS:   siteDir,
		dirstack: []string{},
	}))
}
