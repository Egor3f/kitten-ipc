package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"efprojects.com/kitten-ipc/kitcom/api"
	_ "embed"
)

//go:embed go_gen.tmpl
var templateString string

type goGenData struct {
	PkgName string
	Api     *api.Api
}

type GoApiGenerator struct {
	PkgName string
}

func (g *GoApiGenerator) Generate(apis *api.Api, destFile string) error {
	tplCtx := goGenData{
		PkgName: g.PkgName,
		Api:     apis,
	}

	tpl := template.New("gogen")
	tpl = tpl.Funcs(map[string]any{
		"receiver": func(name string) string {
			return strings.ToLower(name)[0:1]
		},
		"typedef": func(t api.ValType) (string, error) {
			td, ok := map[api.ValType]string{
				api.TInt:    "int",
				api.TString: "string",
				api.TBool:   "bool",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate type %v", t)
			}
			return td, nil
		},
		"zerovalue": func(t api.ValType) (string, error) {
			v, ok := map[api.ValType]string{
				api.TInt:    "0",
				api.TString: `""`,
				api.TBool:   "false",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate zero value for type %v", t)
			}
			return v, nil
		},
	})
	tpl = template.Must(tpl.Parse(templateString))

	var buf bytes.Buffer

	if err := tpl.ExecuteTemplate(&buf, "gogen", tplCtx); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := g.writeDest(destFile, buf.Bytes()); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (g *GoApiGenerator) writeDest(destFile string, bytes []byte) error {
	f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
