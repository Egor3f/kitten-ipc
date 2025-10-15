package estransforms

import (
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ast"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/transformers"
)

type esDecoratorTransformer struct {
	transformers.Transformer
}

func (ch *esDecoratorTransformer) visit(node *ast.Node) *ast.Node {
	return node // !!!
}

func newESDecoratorTransformer(opts *transformers.TransformOptions) *transformers.Transformer {
	tx := &esDecoratorTransformer{}
	return tx.NewTransformer(tx.visit, opts.Context)
}
