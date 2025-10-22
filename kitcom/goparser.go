package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
)

var decorComment = regexp.MustCompile(`^//\s?kittenipc:api$`)

type GoApiParser struct {
}

func (g *GoApiParser) Parse(sourceFile string) (*Api, error) {

	var api Api

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

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		_ = structType

		api.Endpoints = append(api.Endpoints, Endpoint{
			Name: typeSpec.Name.Name,
		})
	}

	if len(api.Endpoints) == 0 {
		return nil, fmt.Errorf("no api struct found")
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

		for i, endpoint := range api.Endpoints {
			if recvIdent.Name == endpoint.Name {
				var apiMethod Method
				apiMethod.Name = funcDecl.Name.Name
				for _, param := range funcDecl.Type.Params.List {
					var apiPar Val
					ident := param.Type.(*ast.Ident)
					switch ident.Name {
					case "int":
						apiPar.Type = TInt
					case "string":
						apiPar.Type = TString
					case "bool":
						apiPar.Type = TBool
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
					var apiRet Val
					ident := ret.Type.(*ast.Ident)
					switch ident.Name {
					case "int":
						apiRet.Type = TInt
					case "string":
						apiRet.Type = TString
					case "bool":
						apiRet.Type = TBool
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
				api.Endpoints[i].Methods = append(api.Endpoints[i].Methods, apiMethod)
			}
		}
	}

	return &api, nil
}
