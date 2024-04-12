package renderer

import "github.com/connormckelvey/sgunk/tree"

type EntryRenderer interface {
	Kind() tree.NodeKind
	Open(node tree.Node, context *RenderContext) error
	Close(node tree.Node, context *RenderContext) error
}
