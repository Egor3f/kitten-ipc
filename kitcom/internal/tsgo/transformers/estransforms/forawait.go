package estransforms

import (
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ast"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/transformers"
)

type forawaitTransformer struct {
	transformers.Transformer
}

func (ch *forawaitTransformer) visit(node *ast.Node) *ast.Node {
	return node // !!!
}

func newforawaitTransformer(opts *transformers.TransformOptions) *transformers.Transformer {
	tx := &forawaitTransformer{}
	return tx.NewTransformer(tx.visit, opts.Context)
}
