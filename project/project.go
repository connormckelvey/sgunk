package project

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type EntryKind string

const (
	EntryKindFile EntryKind = "file"
	EntryKindDir  EntryKind = "dir"
)

// implement Entry
type FileEntry struct {
	fsys fs.FS
	path string
}

func (fe *FileEntry) Kind() EntryKind {
	return EntryKindFile
}

func (fe *FileEntry) Path() string {
	return fe.path
}

func (fe *FileEntry) Content() ([]byte, error) {
	return fs.ReadFile(fe.fsys, fe.path)
}

type Entry interface {
	Kind() EntryKind
	Path() string
}

type PageMetadata interface {
	Kind() PageKind
	Metadata() map[string]any
}
type PageParser interface {
	Parse(base *BasePage) (PageMetadata, error)
}

type Project struct {
	options          []ProjectOption
	config           ProjectConfig
	dir              string
	configFile       string
	extensionLoaders map[string]ExtensionLoader
	pageParsers      map[PageKind]PageParser
	pages            map[string]*BasePage
}

func New(opts ...ProjectOption) *Project {
	return &Project{
		options:          opts,
		extensionLoaders: make(map[string]ExtensionLoader),
		pageParsers:      make(map[PageKind]PageParser),
		pages:            make(map[string]*BasePage),
	}
}

func (p *Project) applyOptions() error {
	// TODO apply defaults

	for _, opt := range p.options {
		if err := opt.Apply(p); err != nil {
			return nil
		}
	}
	return nil
}

var configFiles = map[string]func([]byte, any) error{
	"project.json": json.Unmarshal,
	"project.yml":  yaml.Unmarshal,
	"project.yaml": yaml.Unmarshal,
}

func (p *Project) loadConfigFile() error {
	for name, unmarshal := range configFiles {
		b, err := os.ReadFile(filepath.Join(p.dir, name))
		if err != nil && os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		p.configFile = name
		var c ProjectConfig
		if err := unmarshal(b, &c); err != nil {
			return err
		}
		p.config = c
		return nil
	}

	return nil
}

const defaultSiteDir = "site"

func (p *Project) siteDir() string {
	siteDir := p.config.Site.Dir
	if siteDir == "" {
		siteDir = defaultSiteDir
	}
	return filepath.Join(p.dir, siteDir)
}

const defaultBuildDir = "_build"

func (p *Project) buildDir() string {
	buildDir := p.config.Build.Dir
	if buildDir == "" {
		buildDir = defaultBuildDir
	}
	return filepath.Join(p.dir, buildDir)
}

func (p *Project) RegisterPageParser(kind PageKind, parser PageParser) error {
	p.pageParsers[kind] = parser
	return nil
}

func (p *Project) Load() error {
	if err := p.applyOptions(); err != nil {
		return err
	}

	if err := p.loadConfigFile(); err != nil {
		return err
	}

	for _, extConfig := range p.config.Uses {
		loader, ok := p.extensionLoaders[extConfig.Name]
		if !ok {
			continue
		}
		ext, err := loader.LoadExtension(extConfig.Config)
		if err != nil {
			return err
		}
		if err := ext.Register(p); err != nil {
			return err
		}
	}

	siteDir := os.DirFS(p.siteDir())
	return fs.WalkDir(siteDir, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			log.Println("err walking:", err)
			return nil
		}

		// skip site root dir entry
		rel := filepath.Join(p.siteDir(), path)
		if filepath.Clean(rel) == filepath.Clean(p.siteDir()) {
			log.Println("skipping", path)
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		fileEntry := &FileEntry{
			fsys: siteDir,
			path: path,
		}
		page, err := parsePage(path, fileEntry)
		if err != nil {
			return err
		}
		if page == nil {
			return errors.New("nil page")
		}

		p.pages[path] = page
		for _, parser := range p.pageParsers {
			// try parse
			metadata, err := parser.Parse(page)
			if err != nil {
				return err
			}
			if metadata != nil {
				page.setKind(metadata.Kind())
				page.metadata[metadata.Kind()] = metadata
				break
			}
		}
		return nil
	})
}

func (p *Project) Build() error {
	// load extensions
	buildFS := afero.NewBasePathFs(afero.NewOsFs(), p.buildDir())

	for _, page := range p.pages {
		err := buildFS.MkdirAll(page.Dir(), 0755)
		if err != nil {
			return err
		}

		buildFile, err := buildFS.Create(page.Path())
		if err != nil {
			return err
		}
		defer buildFile.Close()

	}
	return nil
}
