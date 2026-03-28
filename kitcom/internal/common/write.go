package common

import (
	"fmt"
	"os"
)

func WriteFile(destFile string, content []byte) error {
	f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
