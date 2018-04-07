package utils

import (
	"os"
	"path/filepath"

	"gopkg.in/securityfirst/tent.v3/component"
)

func WriteCmp(base string, c component.Component) error {
	path := filepath.Join(base, c.Path())
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(c.Contents()); err != nil {
		return err
	}
	return nil
}
