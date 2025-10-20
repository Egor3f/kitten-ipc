package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
)

var (
	Src  string
	Dest string
)

func parseFlags() {
	flag.StringVar(&Src, "src", "", "Source file")
	flag.StringVar(&Dest, "dest", "", "Dest file")
	flag.Parse()
}

type Method struct {
	Name string
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

	parseFlags()
	if Src == "" || Dest == "" {
		log.Panic("source and destination must be set")
	}

	srcAbs, err := filepath.Abs(Src)
	if err != nil {
		log.Panic(err)
	}

	destAbs, err := filepath.Abs(Dest)
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

	apiGenerator, err := apiGeneratorByExt(destAbs)
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

func apiGeneratorByExt(dest string) (ApiGenerator, error) {
	switch path.Ext(dest) {
	case ".go":
		return &GoApiGenerator{}, nil
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
