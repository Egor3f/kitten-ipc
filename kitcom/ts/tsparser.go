package ts

import (
	"fmt"
	"io"
	"os"
	"strings"

	"efprojects.com/kitten-ipc/kitcom/api"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/ast"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/core"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/parser"
	"efprojects.com/kitten-ipc/kitcom/internal/tsgo/tspath"
)

const TagName = "kittenipc"
const TagComment = "api"

type TypescriptApiParser struct {
	files []string
}

func (t *TypescriptApiParser) AddFile(path string) {
	t.files = append(t.files, path)
}

func (t *TypescriptApiParser) Parse() (*api.Api, error) {

	var apis api.Api

	for _, f := range t.files {
		endpoints, err := t.parseFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse file: %w", err)
		}
		apis.Endpoints = append(apis.Endpoints, endpoints...)
	}

	if len(apis.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints found")
	}

	return &apis, nil
}

func (t *TypescriptApiParser) parseFile(sourceFilePath string) ([]api.Endpoint, error) {
	var endpoints []api.Endpoint

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

	sourceFile.ForEachChild(func(node *ast.Node) bool {
		if node.Kind != ast.KindClassDeclaration {
			return false
		}
		cls := node.AsClassDeclaration()

		jsDocNodes := cls.JSDoc(nil)
		if len(jsDocNodes) == 0 {
			return false
		}

		var isApi bool

	outer:
		for _, jsDocNode := range jsDocNodes {
			jsDoc := jsDocNode.AsJSDoc()
			if jsDoc.Tags == nil {
				continue
			}
			for _, tag := range jsDoc.Tags.Nodes {
				if tag.TagName().Text() == TagName {
					for _, com := range tag.Comments() {
						if strings.TrimSpace(com.Text()) == TagComment {
							isApi = true
							break outer
						}
					}
				}
			}
		}

		if !isApi {
			return false
		}

		var endpoint api.Endpoint

		endpoint.Name = cls.Name().Text()

		for _, member := range cls.MemberList().Nodes {
			if member.Kind != ast.KindMethodDeclaration {
				continue
			}

			method := member.AsMethodDeclaration()
			if method.ModifierFlags()&ast.ModifierFlagsPrivate > 0 || method.ModifierFlags()&ast.ModifierFlagsProtected > 0 {
				continue
			}

			var apiMethod api.Method
			apiMethod.Name = method.Name().Text()
			for _, parNode := range method.ParameterList().Nodes {
				par := parNode.AsParameterDeclaration()
				var apiPar api.Val
				apiPar.Name = par.Name().Text()
				switch par.Type.Kind {
				case ast.KindNumberKeyword:
					apiPar.Type = api.TInt
				case ast.KindStringKeyword:
					apiPar.Type = api.TString
				case ast.KindBooleanKeyword:
					apiPar.Type = api.TBool
				default:
					err = fmt.Errorf("parameter type %s is not supported yet", par.Type.Kind)
					return false
				}
				apiMethod.Params = append(apiMethod.Params, apiPar)
			}
			if method.Type != nil {
				var apiRet api.Val
				switch method.Type.Kind {
				case ast.KindNumberKeyword:
					apiRet.Type = api.TInt
				case ast.KindStringKeyword:
					apiRet.Type = api.TString
				case ast.KindBooleanKeyword:
					apiRet.Type = api.TBool
				default:
					err = fmt.Errorf("return type %s is not supported yet", method.Type.Kind)
					return false
				}
				apiMethod.Ret = []api.Val{apiRet}
			}
			endpoint.Methods = append(endpoint.Methods, apiMethod)
		}

		endpoints = append(endpoints, endpoint)

		return false
	})

	if err != nil {
		return nil, err
	}

	return endpoints, nil
}
