package generator

import "io/fs"

type ContentLoader interface {
	Load(p *Project, path string, entry fs.DirEntry) error
}

type ContentLoaderFunc func(p *Project, path string, entry fs.DirEntry) error

func (load ContentLoaderFunc) Load(p *Project, path string, entry fs.DirEntry) error {
	return load(p, path, entry)
}

func NewPageLoader() ContentLoaderFunc {
	return func(p *Project, path string, entry fs.DirEntry) error {
		// Skip dirs
		if entry.IsDir() {
			return nil
		}

		// skip unknown file types
		if !PageKindMarkdown.PathIs(path) {
			return nil
		}

		p.Pages = append(p.Pages, &Page{
			SourcePath: path,
			Kind:       PageKindMarkdown,
			Metadata:   make(map[string]any),
		})
		return nil
	}
}
