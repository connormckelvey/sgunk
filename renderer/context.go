package renderer

import (
	"io"
	"io/fs"
	"path/filepath"

	"github.com/connormckelvey/website/tree"
	"github.com/spf13/afero"
)

type RenderContext struct {
	siteFS    afero.Fs
	themeFS   afero.Fs
	buildFS   afero.Fs
	dirstack  []string
	openFiles []afero.File
	templater *Evaluator
}

type RenderContextOption interface {
	Apply(*RenderContext)
}

type RenderContextOptionFunc func(*RenderContext)

func (apply RenderContextOptionFunc) Apply(r *RenderContext) {
	apply(r)
}

func WithSiteFS(siteFS afero.Fs) RenderContextOptionFunc {
	return func(r *RenderContext) {
		r.siteFS = siteFS
		r.templater = NewEvaluator(afero.NewIOFS(siteFS))
	}
}
func WithFS(siteFS afero.Fs, themeFS afero.Fs, buildFS afero.Fs) RenderContextOptionFunc {
	return func(r *RenderContext) {
		WithSiteFS(siteFS)(r)
		r.themeFS = themeFS
		r.buildFS = buildFS
	}
}

func NewRenderContext(opts ...RenderContextOption) *RenderContext {
	rc := &RenderContext{
		dirstack: []string{},
	}
	for _, opt := range opts {
		opt.Apply(rc)
	}
	return rc
}

func (rc *RenderContext) Source(node tree.Node) ([]byte, error) {
	return afero.ReadFile(rc.siteFS, node.Path())
}

func (rc *RenderContext) WorkDir() string {
	return filepath.Join(rc.dirstack...)
}

func (rc *RenderContext) MkdirAll(path string, perm fs.FileMode) error {
	return rc.buildFS.MkdirAll(
		filepath.Join(rc.WorkDir(), path),
		perm,
	)
}

func (rc *RenderContext) CurrentFile() afero.File {
	if len(rc.openFiles) == 0 {
		return nil
	}
	return rc.openFiles[len(rc.openFiles)-1]
}

func (rc *RenderContext) CreateFile(path string) (io.Writer, error) {
	file, err := rc.buildFS.Create(
		filepath.Join(rc.WorkDir(), path),
	)
	if err != nil {
		return nil, err
	}
	rc.openFiles = append(rc.openFiles, file)
	return file, nil
}

func (rc *RenderContext) PopFile() afero.File {
	popped := rc.openFiles[len(rc.openFiles)-1]
	rc.openFiles = rc.openFiles[0 : len(rc.openFiles)-1]
	return popped
}

func (rc *RenderContext) PushDir(dir string) {
	rc.dirstack = append(rc.dirstack, dir)
}

func (rc *RenderContext) PopDir() string {
	popped := rc.dirstack[len(rc.dirstack)-1]
	rc.dirstack = rc.dirstack[0 : len(rc.dirstack)-1]
	return popped
}
