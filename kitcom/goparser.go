package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
)

var decorComment = regexp.MustCompile(`^//\s?kittenipc:api$`)

type apiStruct struct {
	pkgName string
	name    string
	methods []*ast.FuncDecl
}

type GoApiParser struct {
	apiStructs []*apiStruct
}

func (g *GoApiParser) Parse(sourceFile string) (Api, error) {

	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, sourceFile, nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return Api{}, fmt.Errorf("parse file: %w", err)
	}

	pkgName := astFile.Name.Name

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

		g.apiStructs = append(g.apiStructs, &apiStruct{
			name:    typeSpec.Name.Name,
			pkgName: pkgName,
		})
	}

	if len(g.apiStructs) == 0 {
		// todo support arbitrary order of input files
		return Api{}, fmt.Errorf("no api struct found")
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

		for _, apiStrct := range g.apiStructs {
			if recvIdent.Name == apiStrct.name && pkgName == apiStrct.pkgName {
				apiStrct.methods = append(apiStrct.methods, funcDecl)
			}
		}
	}

	return Api{}, nil
}
