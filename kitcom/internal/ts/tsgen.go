package ts

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"

	_ "embed"

	"efprojects.com/kitten-ipc/kitcom/internal/api"
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
	})
	tpl = template.Must(tpl.Parse(templateString))

	var buf bytes.Buffer

	if err := tpl.ExecuteTemplate(&buf, "tsgen", tplCtx); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := g.writeDest(destFile, buf.Bytes()); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (g *TypescriptApiGenerator) writeDest(destFile string, bytes []byte) error {
	f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(bytes); err != nil {
		return fmt.Errorf("write formatted source: %w", err)
	}

	prettierCmd := exec.Command("npx", "prettier", destFile, "--write")
	if out, err := prettierCmd.CombinedOutput(); err != nil {
		log.Printf("Prettier returned error: %v", err)
		log.Printf("Output: \n%s", string(out))
	}

	return nil
}
