package parser

import (
	"bytes"

	"github.com/adrg/frontmatter"
	"github.com/spf13/afero"
)

type ParserContext struct {
	siteFS afero.Fs

	// sources should belong to project so it can be shared with render
	sources map[string][]byte
}

func (pc *ParserContext) Source(path string) ([]byte, error) {
	if b, ok := pc.sources[path]; ok {
		return b, nil
	}
	b, err := afero.ReadFile(pc.siteFS, path)
	if err != nil {
		return nil, err
	}
	pc.sources[path] = b
	return b, nil
}

func (pc *ParserContext) Content(path string) ([]byte, error) {
	source, err := pc.Source(path)
	if err != nil {
		return nil, err
	}
	content, err := frontmatter.Parse(bytes.NewReader(source), &(map[string]any{}))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (pc *ParserContext) FrontMatter(path string, output any) error {
	source, err := pc.Source(path)
	if err != nil {
		return err
	}
	if _, err := frontmatter.Parse(bytes.NewReader(source), output); err != nil {
		return err
	}
	return nil
}
