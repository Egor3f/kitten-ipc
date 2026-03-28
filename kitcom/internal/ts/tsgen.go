package ts

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"text/template"

	_ "embed"

	"github.com/egor3f/kitten-ipc/kitcom/internal/api"
	"github.com/egor3f/kitten-ipc/kitcom/internal/common"
)

//go:embed tsgen.tmpl
var templateString string

type tsGenData struct {
	Api *api.Api
}

type TypescriptApiGenerator struct {
}

func (g *TypescriptApiGenerator) Generate(apis *api.Api, destFile string) error {
	tplCtx := tsGenData{
		Api: apis,
	}

	tpl := template.New("tsgen")
	tpl = tpl.Funcs(map[string]any{
		"typedef": func(t api.ValType) (string, error) {
			td, ok := map[api.ValType]string{
				api.TInt:    "number",
				api.TString: "string",
				api.TBool:   "boolean",
				api.TBlob:   "Buffer",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate type %v", t)
			}
			return td, nil
		},
		"convtype": func(valDef string, t api.ValType) (string, error) {
			td, ok := map[api.ValType]string{
				api.TInt:    fmt.Sprintf("%s as number", valDef),
				api.TString: fmt.Sprintf("%s as string", valDef),
				api.TBool:   fmt.Sprintf("%s as boolean", valDef),
				api.TBlob:   fmt.Sprintf("Buffer.from(%s, 'base64')", valDef),
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot convert type %v for val %s", t, valDef)
			}
			return td, nil
		},
	})
	tpl = template.Must(tpl.Parse(templateString))

	var buf bytes.Buffer

	if err := tpl.ExecuteTemplate(&buf, "tsgen", tplCtx); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := common.WriteFile(destFile, buf.Bytes()); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	prettierCmd := exec.Command("npx", "prettier", destFile, "--write")
	if out, err := prettierCmd.CombinedOutput(); err != nil {
		log.Printf("Prettier returned error: %v", err)
		log.Printf("Output: \n%s", string(out))
	}

	return nil
}
