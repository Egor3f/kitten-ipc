package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	_ "embed"

	"github.com/egor3f/kitten-ipc/kitcom/internal/api"
	"github.com/egor3f/kitten-ipc/kitcom/internal/common"
)

//go:embed gogen.tmpl
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
			return strings.ToLower(name[:1])
		},
		"typedef": func(t api.ValType) (string, error) {
			td, ok := map[api.ValType]string{
				api.TInt:    "int",
				api.TString: "string",
				api.TBool:   "bool",
				api.TBlob:   "[]byte",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate type %v", t)
			}
			return td, nil
		},
		"convtype": func(valDef string, t api.ValType) (string, error) {
			td, ok := map[api.ValType]string{
				api.TInt:    fmt.Sprintf("int(%s.(float64))", valDef),
				api.TString: fmt.Sprintf("%s.(string)", valDef),
				api.TBool:   fmt.Sprintf("%s.(bool)", valDef),
				api.TBlob:   fmt.Sprintf("%s.([]byte)", valDef),
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot convert type %v for val %s", t, valDef)
			}
			return td, nil
		},
		"zerovalue": func(t api.ValType) (string, error) {
			v, ok := map[api.ValType]string{
				api.TInt:    "0",
				api.TString: `""`,
				api.TBool:   "false",
				api.TBlob:   "[]byte{}",
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

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("format source: %w", err)
	}

	if err := common.WriteFile(destFile, formatted); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
