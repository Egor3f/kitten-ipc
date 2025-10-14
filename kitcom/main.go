package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

type ProgLang string

const (
	Golang     ProgLang = "Golang"
	TypeScript          = "Typescript"
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

type Api struct {
}

type ApiParser interface {
	Parse(sourceFile string) (Api, error)
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

	if err := checkIsFile(Src); err != nil {
		log.Panic(err)
	}

	apiParser, err := apiParserByExt(Src)
	if err != nil {
		log.Panic(err)
	}

	_, err = apiParser.Parse(Src)
	if err != nil {
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
