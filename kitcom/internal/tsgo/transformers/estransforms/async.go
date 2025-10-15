package estransforms

import (
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ast"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/transformers"
)

type asyncTransformer struct {
	transformers.Transformer
}

func (ch *asyncTransformer) visit(node *ast.Node) *ast.Node {
	return node // !!!
}

func newAsyncTransformer(opts *transformers.TransformOptions) *transformers.Transformer {
	tx := &asyncTransformer{}
	return tx.NewTransformer(tx.visit, opts.Context)
}
