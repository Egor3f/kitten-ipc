package golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"

	"efprojects.com/kitten-ipc/kitcom/internal/api"
	"efprojects.com/kitten-ipc/kitcom/internal/common"
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
					apiPar, err := fieldToVal(param)
					if err != nil {
						return nil, fmt.Errorf("parse parameter %s: %w", param.Names[0].Name, err)
					}

					if len(param.Names) != 1 {
						return nil, fmt.Errorf("all parameters in method %s should be named", apiMethod.Name)
					}
					apiPar.Name = param.Names[0].Name

					apiMethod.Params = append(apiMethod.Params, apiPar)
				}
				for _, ret := range funcDecl.Type.Results.List {
					apiRet, err := fieldToVal(ret)
					if err != nil {
						return nil, fmt.Errorf("parse return value %s: %w", ret.Names[0].Name, err)
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

func fieldToVal(param *ast.Field) (api.Val, error) {
	var val api.Val
	switch paramType := param.Type.(type) {
	case *ast.Ident:
		switch paramType.Name {
		case "int":
			val.Type = api.TInt
		case "string":
			val.Type = api.TString
		case "bool":
			val.Type = api.TBool
		default:
			return val, fmt.Errorf("parameter type %s is not supported yet", paramType.Name)
		}
	case *ast.ArrayType:
		switch elementType := paramType.Elt.(type) {
		case *ast.Ident:
			switch elementType.Name {
			case "byte":
				val.Type = api.TBlob
			default:
				return val, fmt.Errorf("parameter type %s is not supported yet", elementType.Name)
			}
		}
	default:
		return val, fmt.Errorf("parameter type %T is not supported yet", paramType)
	}
	return val, nil
}
