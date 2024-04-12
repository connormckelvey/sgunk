package blog

import (
	"github.com/connormckelvey/sgunk"
	"github.com/connormckelvey/sgunk/parser"
	"github.com/connormckelvey/sgunk/renderer"
	"github.com/mitchellh/mapstructure"
)

const extName = "github.com/connormckelvey/sgunk/extension/blog"

type Extension struct {
}

func (be *Extension) Name() string {
	return extName
}

func (be *Extension) Register(project *sgunk.Project, c map[string]any) error {
	var config struct {
		Path string `mapstructure:"path"`
	}
	err := mapstructure.Decode(c, &config)
	if err != nil {
		return err
	}
	useEntryParsers := parser.WithEntryParsers(
		NewBlogEntryParser(config.Path),
	)
	if err := sgunk.WithParserOptions(useEntryParsers)(project); err != nil {
		return err
	}
	useEntryRenderers := renderer.WithEntryRenderers(
		&BlogRenderer{},
	)
	sgunk.WithRendererOptions(useEntryRenderers)(project)

	if err := sgunk.WithRendererOptions(useEntryRenderers)(project); err != nil {
		return err
	}
	return nil
}
