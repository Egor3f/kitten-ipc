package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ast"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/core"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/parser"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/tspath"
)

const TagName = "kittenipc"
const TagComment = "api"

type TypescriptApiParser struct {
}

type apiClass struct {
}

func (t *TypescriptApiParser) Parse(sourceFilePath string) (*Api, error) {

	f, err := os.Open(sourceFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	fileContents, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName:                       sourceFilePath,
		Path:                           tspath.Path(sourceFilePath),
		CompilerOptions:                core.SourceFileAffectingCompilerOptions{},
		ExternalModuleIndicatorOptions: ast.ExternalModuleIndicatorOptions{},
		JSDocParsingMode:               ast.JSDocParsingModeParseAll,
	}, string(fileContents), core.ScriptKindTS)
	_ = sourceFile

	var apiClasses []apiClass

	sourceFile.ForEachChild(func(node *ast.Node) bool {
		if node.Kind != ast.KindClassDeclaration {
			return false
		}
		cls := node.AsClassDeclaration()

		jsDocNodes := cls.JSDoc(nil)
		if len(jsDocNodes) == 0 {
			return false
		}

		for _, jsDocNode := range jsDocNodes {
			jsDoc := jsDocNode.AsJSDoc()
			for _, tag := range jsDoc.Tags.Nodes {
				if tag.TagName().Text() == TagName {
					for _, com := range tag.Comments() {
						if strings.TrimSpace(com.Text()) == TagComment {
							apiClasses = append(apiClasses, apiClass{})
							return false
						}
					}
				}
			}
		}

		return false
	})

	if len(apiClasses) == 0 {
		return nil, fmt.Errorf("no api class found")
	}

	if len(apiClasses) > 1 {
		return nil, fmt.Errorf("multiple api classes found")
	}

	return nil, nil
}
