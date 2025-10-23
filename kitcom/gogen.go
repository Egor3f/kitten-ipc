package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

type goGenData struct {
	PkgName string
	Api     *Api
}

type GoApiGenerator struct {
	pkgName string
}

func (g *GoApiGenerator) Generate(api *Api, destFile string) error {
	tplCtx := goGenData{
		PkgName: g.pkgName,
		Api:     api,
	}

	tpl := template.New("gogen")
	tpl = tpl.Funcs(map[string]any{
		"receiver": func(name string) string {
			return strings.ToLower(name)[0:1]
		},
		"typedef": func(t ValType) (string, error) {
			td, ok := map[ValType]string{
				TInt:    "int",
				TString: "string",
				TBool:   "bool",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate type %v", t)
			}
			return td, nil
		},
		"zerovalue": func(t ValType) (string, error) {
			v, ok := map[ValType]string{
				TInt:    "0",
				TString: `""`,
				TBool:   "false",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate zero value for type %v", t)
			}
			return v, nil
		},
	})
	tpl = template.Must(tpl.ParseFiles("./go_gen.tmpl"))

	var buf bytes.Buffer

	if err := tpl.ExecuteTemplate(&buf, "go_gen.tmpl", tplCtx); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := g.writeDest(destFile, buf.Bytes()); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (g *GoApiGenerator) writeDest(destFile string, bytes []byte) error {
	f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer f.Close()

	formatted, err := format.Source(bytes)
	if err != nil {
		return fmt.Errorf("format source: %w", err)
	}

	if _, err := f.Write(formatted); err != nil {
		return fmt.Errorf("write formatted source: %w", err)
	}
	return nil
}
