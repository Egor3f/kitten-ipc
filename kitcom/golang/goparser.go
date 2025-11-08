package golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"

	"efprojects.com/kitten-ipc/kitcom/api"
	"efprojects.com/kitten-ipc/kitcom/common"
)

var decorComment = regexp.MustCompile(`^//\s?kittenipc:api$`)

type GoApiParser struct {
	*common.Parser
}

func (p *GoApiParser) Parse() (*api.Api, error) {
	return p.MapFiles(p.parseFile)
}

func (p *GoApiParser) parseFile(sourceFile string) ([]api.Endpoint, error) {
	var endpoints []api.Endpoint

	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, sourceFile, nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if genDecl.Doc == nil {
			continue
		}

		// use only last comment. https://tip.golang.org/doc/comment#syntax
		lastComment := genDecl.Doc.List[len(genDecl.Doc.List)-1]
		if !decorComment.MatchString(lastComment.Text) {
			continue
		}

		typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		_, isStruct := typeSpec.Type.(*ast.StructType)
		_, isIface := typeSpec.Type.(*ast.InterfaceType)
		if !isStruct && !isIface {
			continue
		}

		endpoints = append(endpoints, api.Endpoint{
			Name: typeSpec.Name.Name,
		})
	}

	if len(endpoints) == 0 {
		return nil, nil
	}

	for _, decl := range astFile.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if !funcDecl.Name.IsExported() {
			continue
		}

		if funcDecl.Recv == nil {
			continue
		}

		reciever := funcDecl.Recv.List[0]
		recvType := reciever.Type

		star, isPointer := recvType.(*ast.StarExpr)
		if isPointer {
			recvType = star.X
		}

		recvIdent, ok := recvType.(*ast.Ident)
		if !ok {
			continue
		}

		for i, endpoint := range endpoints {
			if recvIdent.Name == endpoint.Name {
				var apiMethod api.Method
				apiMethod.Name = funcDecl.Name.Name
				for _, param := range funcDecl.Type.Params.List {
					var apiPar api.Val
					ident := param.Type.(*ast.Ident)
					switch ident.Name {
					case "int":
						apiPar.Type = api.TInt
					case "string":
						apiPar.Type = api.TString
					case "bool":
						apiPar.Type = api.TBool
					default:
						return nil, fmt.Errorf("parameter type %s is not supported yet", ident.Name)
					}
					if len(param.Names) != 1 {
						return nil, fmt.Errorf("all parameters in method %s should be named", apiMethod.Name)
					}
					apiPar.Name = param.Names[0].Name
					apiMethod.Params = append(apiMethod.Params, apiPar)
				}
				for _, ret := range funcDecl.Type.Results.List {
					var apiRet api.Val
					ident := ret.Type.(*ast.Ident)
					switch ident.Name {
					case "int":
						apiRet.Type = api.TInt
					case "string":
						apiRet.Type = api.TString
					case "bool":
						apiRet.Type = api.TBool
					case "error":
						// errors are processed other way
						continue
					default:
						return nil, fmt.Errorf("return type %s is not supported yet", ident.Name)
					}
					if len(ret.Names) > 0 {
						apiRet.Name = ret.Names[0].Name
					}
					apiMethod.Ret = append(apiMethod.Ret, apiRet)
				}
				endpoints[i].Methods = append(endpoints[i].Methods, apiMethod)
			}
		}
	}
	return endpoints, nil
}
