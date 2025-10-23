package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
)

type tsGenData struct {
	Api *Api
}

type TypescriptApiGenerator struct {
}

func (g *TypescriptApiGenerator) Generate(api *Api, destFile string) error {
	tplCtx := tsGenData{
		Api: api,
	}

	tpl := template.New("gogen")
	tpl = tpl.Funcs(map[string]any{
		"typedef": func(t ValType) (string, error) {
			td, ok := map[ValType]string{
				TInt:    "number",
				TString: "string",
				TBool:   "bool",
			}[t]
			if !ok {
				return "", fmt.Errorf("cannot generate type %v", t)
			}
			return td, nil
		},
	})
	tpl = template.Must(tpl.ParseFiles("./ts_gen.tmpl"))

	var buf bytes.Buffer

	if err := tpl.ExecuteTemplate(&buf, "ts_gen.tmpl", tplCtx); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := g.writeDest(destFile, buf.Bytes()); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (g *TypescriptApiGenerator) writeDest(destFile string, bytes []byte) error {
	f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(bytes); err != nil {
		return fmt.Errorf("write formatted source: %w", err)
	}

	prettierCmd := exec.Command("npx", "prettier", destFile, "--write")
	if err := prettierCmd.Run(); err != nil {
		log.Printf("Prettier returned error: %v", err)
	}

	return nil
}
