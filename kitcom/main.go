package main

import (
	"flag"
	"fmt"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
)

type ValType int

// todo check TInt size < 64
// todo check not float

const (
	TInt    ValType = 1
	TString ValType = 2
	TBool   ValType = 3
	TBlob   ValType = 4
	TArray  ValType = 5
)

type Val struct {
	Name     string
	Type     ValType
	Children []Val
}

type Method struct {
	Name   string
	Params []Val
	Ret    []Val
}

type Endpoint struct {
	Name    string
	Methods []Method
}

type Api struct {
	Endpoints []Endpoint
}

type ApiParser interface {
	Parse(sourceFile string) (*Api, error)
}

type ApiGenerator interface {
	Generate(api *Api, destFile string) error
}

func main() {
	// todo support go:generate
	//goFile := os.Getenv("GOFILE")
	//if goFile == "" {
	//	log.Panic("GOFILE must be set")
	//}

	src := flag.String("src", "", "Source file")
	dest := flag.String("dest", "", "Dest file")
	pkgName := flag.String("pkgname", "", "Package name (for go)")
	flag.Parse()

	if *src == "" || *dest == "" {
		log.Panic("source and destination must be set")
	}

	srcAbs, err := filepath.Abs(*src)
	if err != nil {
		log.Panic(err)
	}

	destAbs, err := filepath.Abs(*dest)
	if err != nil {
		log.Panic(err)
	}

	if err := checkIsFile(srcAbs); err != nil {
		log.Panic(err)
	}

	apiParser, err := apiParserByExt(srcAbs)
	if err != nil {
		log.Panic(err)
	}

	api, err := apiParser.Parse(srcAbs)
	if err != nil {
		log.Panic(err)
	}

	apiGenerator, err := apiGeneratorByExt(destAbs, *pkgName)
	if err != nil {
		log.Panic(err)
	}

	if err := apiGenerator.Generate(api, destAbs); err != nil {
		log.Panic(err)
	}
}

func checkIsFile(src string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory; directories are not supported yet", src)
	}
	return nil
}

func apiParserByExt(src string) (ApiParser, error) {
	switch path.Ext(src) {
	case ".go":
		return &GoApiParser{}, nil
	case ".ts":
		return &TypescriptApiParser{}, nil
	case ".js":
		return nil, fmt.Errorf("vanilla javascript is not supported and never will be")
	case "":
		return nil, fmt.Errorf("could not find file extension for %s", src)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", path.Ext(src))
	}
}

func apiGeneratorByExt(dest string, pkgName string) (ApiGenerator, error) {
	switch path.Ext(dest) {
	case ".go":
		if pkgName == "" {
			return nil, fmt.Errorf("package name must be set for Go generation")
		}
		if !token.IsIdentifier(pkgName) {
			return nil, fmt.Errorf("invalid package name: %s", pkgName)
		}
		return &GoApiGenerator{
			pkgName: pkgName,
		}, nil
	case ".ts":
		return &TypescriptApiGenerator{}, nil
	case ".js":
		return nil, fmt.Errorf("vanilla javascript is not supported and never will be")
	case "":
		return nil, fmt.Errorf("could not find file extension for %s", dest)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", path.Ext(dest))
	}
}
