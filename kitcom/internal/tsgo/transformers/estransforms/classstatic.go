package estransforms

import (
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ast"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/transformers"
)

type classStaticBlockTransformer struct {
	transformers.Transformer
}

func (ch *classStaticBlockTransformer) visit(node *ast.Node) *ast.Node {
	return node // !!!
}

func newClassStaticBlockTransformer(opts *transformers.TransformOptions) *transformers.Transformer {
	tx := &classStaticBlockTransformer{}
	return tx.NewTransformer(tx.visit, opts.Context)
}
