package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
)

func jsonMarshalString(s string) ([]byte, error) {
	return json.Marshal(s)
}

func writeDoc(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
