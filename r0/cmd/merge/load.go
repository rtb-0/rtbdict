package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errf("go.mod not found")
		}
		dir = parent
	}
}

func readTax(root, rel string) (taxDoc, error) {
	body, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return taxDoc{}, err
	}
	var doc taxDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return taxDoc{}, errf("%s: %w", rel, err)
	}
	return doc, nil
}

func readMap(root, rel string) (mapDoc, error) {
	body, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return mapDoc{}, err
	}
	var doc mapDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return mapDoc{}, errf("%s: %w", rel, err)
	}
	return doc, nil
}
