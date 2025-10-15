package main

import (
	"fmt"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ast"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/core"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/parser"
)

type TypescriptApiParser struct {
}

func (t *TypescriptApiParser) Parse(sourceFilePath string) (Api, error) {
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName:                       sourceFilePath,
		Path:                           "",
		CompilerOptions:                core.SourceFileAffectingCompilerOptions{},
		ExternalModuleIndicatorOptions: ast.ExternalModuleIndicatorOptions{},
		JSDocParsingMode:               ast.JSDocParsingModeParseAll,
	}, "", core.ScriptKindTS)
	_ = sourceFile

	sourceFile.ForEachChild(func(node *ast.Node) bool {
		if node.IsJSDoc() {
			jsDoc := node.AsJSDoc()
			_ = jsDoc
			fmt.Println("a")
		}
		return false
	})

	return Api{}, nil
}
