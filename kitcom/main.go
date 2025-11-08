package main

import (
	"flag"
	"fmt"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"efprojects.com/kitten-ipc/kitcom/api"
	"efprojects.com/kitten-ipc/kitcom/golang"
	"efprojects.com/kitten-ipc/kitcom/ts"
)

type ApiParser interface {
	AddFile(path string)
	Parse() (*api.Api, error)
}

type ApiGenerator interface {
	Generate(api *api.Api, destFile string) error
}

func main() {
	src := flag.String("src", "", "Source file/dir")
	dest := flag.String("dest", "", "Dest file")
	pkgName := flag.String("pkg", "", "Package name (for go)")
	flag.Parse()

	if *src == "" || *dest == "" {
		log.Fatalln("source and destination must be set")
	}

	srcAbs, err := filepath.Abs(*src)
	if err != nil {
		log.Fatalln(err)
	}

	destAbs, err := filepath.Abs(*dest)
	if err != nil {
		log.Fatalln(err)
	}

	apiParser, err := apiParserByPath(srcAbs)
	if err != nil {
		log.Fatalln(err)
	}

	apiGenerator, err := apiGeneratorByPath(destAbs, *pkgName)
	if err != nil {
		log.Fatalln(err)
	}

	apis, err := apiParser.Parse()
	if err != nil {
		log.Fatalln(err)
	}

	if err := apiGenerator.Generate(apis, destAbs); err != nil {
		log.Fatalln(err)
	}
}

func apiParserByPath(src string) (ApiParser, error) {
	pathFI, err := os.Stat(src)
	if err != nil {
		return nil, fmt.Errorf("stat src: %w", err)
	}

	var parser ApiParser
	var ext string

	if pathFI.IsDir() {
		if err := filepath.Walk(src, func(curPath string, fileinfo fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fileinfo.IsDir() {
				return nil
			}

			p, err := apiParserByFilePath(fileinfo.Name())
			if err == nil {
				if parser == nil {
					parser = p
					ext = path.Ext(fileinfo.Name())
				} else if path.Ext(fileinfo.Name()) != ext {
					return fmt.Errorf("path contain multiple supported filetypes")
				}
				parser.AddFile(curPath)
			}

			return nil
		}); err != nil {
			return nil, fmt.Errorf("walk dir: %w", err)
		}
	} else {
		parser, err = apiParserByFilePath(src)
		if err != nil {
			return nil, err
		}
		parser.AddFile(src)
	}

	if parser == nil {
		return nil, fmt.Errorf("could not find supported files in %s", src)
	}
	return parser, nil
}

func apiParserByFilePath(src string) (ApiParser, error) {
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

func apiGeneratorByPath(dest string, pkgName string) (ApiGenerator, error) {
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
