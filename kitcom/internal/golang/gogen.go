package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"

	_ "embed"

	"efprojects.com/kitten-ipc/kitcom/internal/api"
	"efprojects.com/kitten-ipc/types"
)

// todo: check int overflow
// todo: check float is whole

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

	const defaultReceiver = "self"

	tpl := template.New("gogen")
	tpl = tpl.Funcs(map[string]any{
		"receiver": func(name string) string {
			return defaultReceiver
		},
		"typedef": func(t types.ValType) (string, error) {
			td, ok := map[types.ValType]string{
				types.TInt:    "int",
				types.TString: "string",
				types.TBool:   "bool",
				types.TBlob:   "[]byte",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate type %v", t)
			}
			return td, nil
		},
		"convtype": func(valDef string, t types.ValType) (string, error) {
			td, ok := map[types.ValType]string{
				types.TInt:    fmt.Sprintf("int(%s.(float64))", valDef),
				types.TString: fmt.Sprintf("%s.(string)", valDef),
				types.TBool:   fmt.Sprintf("%s.(bool)", valDef),
				types.TBlob: fmt.Sprintf(
					"%s.Ipc.ConvType(reflect.TypeOf([]byte{}), reflect.TypeOf(\"\"), %s).([]byte)",
					defaultReceiver,
					valDef,
				),
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot convert type %v for val %s", t, valDef)
			}
			return td, nil
		},
		"zerovalue": func(t types.ValType) (string, error) {
			v, ok := map[types.ValType]string{
				types.TInt:    "0",
				types.TString: `""`,
				types.TBool:   "false",
				types.TBlob:   "[]byte{}",
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
