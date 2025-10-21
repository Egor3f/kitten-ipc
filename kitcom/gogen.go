package main

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type GoApiGenerator struct {
}

var tpl = template.Must(template.New("gotpl").Parse(strings.TrimSpace(`



`)))

func (g *GoApiGenerator) Generate(api *Api, destFile string) error {
	destFileBak := filepath.Join(destFile, ".bak")
	_, err := os.Stat(destFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat destination file: %w", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		if err := os.Rename(destFile, destFileBak); err != nil {
			return fmt.Errorf("backup destination file: %w", err)
		}
	}

	f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer f.Close()

	if err := tpl.Execute(f, api); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if _, err := os.Stat(destFileBak); err == nil {
		if err := os.Remove(destFileBak); err != nil {
			return fmt.Errorf("remove backup file: %w", err)
		}
	}
	return nil
}
