package renderer

import "github.com/connormckelvey/ssg/tree"

type EntryRenderer interface {
	Test(tree.Node) (bool, error)
	Open(node tree.Node, context *RenderContext) error
	Close(node tree.Node, context *RenderContext) error
}
