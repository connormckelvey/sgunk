package renderer

import "github.com/connormckelvey/website/tree"

type EntryRenderer interface {
	Test(tree.Node) (bool, error)
	Props(node tree.Node, context *RenderContext) (map[string]any, error)
	Open(node tree.Node, context *RenderContext) error
	Close(node tree.Node, context *RenderContext) error
}
