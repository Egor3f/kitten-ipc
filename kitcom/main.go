package main

import (
	"flag"
	"fmt"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"

	"efprojects.com/kitten-ipc/kitcom/api"
	"efprojects.com/kitten-ipc/kitcom/golang"
	"efprojects.com/kitten-ipc/kitcom/ts"
)

type ApiParser interface {
	Parse(sourceFile string) (*api.Api, error)
}

type ApiGenerator interface {
	Generate(api *api.Api, destFile string) error
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
		return &golang.GoApiParser{}, nil
	case ".ts":
		return &ts.TypescriptApiParser{}, nil
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
		return &golang.GoApiGenerator{
			PkgName: pkgName,
		}, nil
	case ".ts":
		return &ts.TypescriptApiGenerator{}, nil
	case ".js":
		return nil, fmt.Errorf("vanilla javascript is not supported and never will be")
	case "":
		return nil, fmt.Errorf("could not find file extension for %s", dest)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", path.Ext(dest))
	}
}
