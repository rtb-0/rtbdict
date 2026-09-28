// Command gendict writes Go dictionaries from iab/datasets into matching iab/dict directories.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	datasets := filepath.Join(root, "iab", "datasets")
	outRoot := filepath.Join(root, "iab", "dict")
	families := map[string]string{}
	err = filepath.WalkDir(datasets, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		rel, err := filepath.Rel(datasets, path)
		if err != nil {
			return err
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) != 2 {
			return fmt.Errorf("unexpected dataset path %s", rel)
		}
		family, file := parts[0], parts[1]
		pkg, ok := packageName[family]
		if !ok {
			return fmt.Errorf("unknown dataset directory %s", family)
		}
		families[family] = pkg
		kind, src, err := generateFile(path, pkg)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if fileKind[family] == "" {
			fileKind[family] = kind
		} else if fileKind[family] != kind {
			return fmt.Errorf("%s mixes %s and %s", family, fileKind[family], kind)
		}
		out := filepath.Join(outRoot, family, strings.TrimSuffix(file, ".json")+".go")
		return writeGo(out, src)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for family, pkg := range families {
		src := lookupSource(pkg, fileKind[family])
		if err := writeGo(filepath.Join(outRoot, family, "lookup.go"), src); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

var packageName = map[string]string{
	"ad-product": "adproduct",
	"content":    "content",
	"audience":   "audience",
	"mappings":   "mappings",
}

var fileKind = map[string]string{}

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
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

type meta struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Cattax     int    `json:"cattax"`
	Deprecated bool   `json:"deprecated"`
	Publisher  string `json:"publisher"`
	Page       string `json:"page"`
	Source     string `json:"source"`
	Released   string `json:"released"`
	From       string `json:"from"`
	FromCattax int    `json:"from_cattax"`
	To         string `json:"to"`
	ToCattax   int    `json:"to_cattax"`
	Notes      string `json:"notes"`
}

type node struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Children  []node `json:"children"`
}

type ref struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Path []string `json:"path"`
}

type entry struct {
	From ref    `json:"from"`
	To   *ref   `json:"to"`
	Note string `json:"note"`
}

type taxonomyDoc struct {
	Meta  meta   `json:"meta"`
	Nodes []node `json:"nodes"`
}

type mappingDoc struct {
	Meta    meta    `json:"meta"`
	Entries []entry `json:"entries"`
}

type rawNode struct {
	ID        string
	Name      string
	Extension string
	Parent    int
}

func generateFile(path, pkg string) (kind string, src []byte, err error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	var probe struct {
		Meta meta `json:"meta"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return "", nil, err
	}
	switch probe.Meta.Kind {
	case "taxonomy":
		var doc taxonomyDoc
		if err := json.Unmarshal(body, &doc); err != nil {
			return "", nil, err
		}
		return "taxonomy", renderTaxonomy(pkg, doc), nil
	case "mapping":
		var doc mappingDoc
		if err := json.Unmarshal(body, &doc); err != nil {
			return "", nil, err
		}
		return "mapping", renderMapping(pkg, doc), nil
	default:
		return "", nil, fmt.Errorf("unknown kind %q", probe.Meta.Kind)
	}
}

func renderTaxonomy(pkg string, doc taxonomyDoc) []byte {
	var raw []rawNode
	flatten(doc.Nodes, -1, &raw)
	var b bytes.Buffer
	writeHeader(&b, pkg)
	b.WriteString("func init() {\n\tregister(dict.MustTaxonomy(")
	writeMeta(&b, doc.Meta)
	b.WriteString(", []dict.RawNode{\n")
	for _, n := range raw {
		fmt.Fprintf(&b, "\t\t{ID: %s, Name: %s", strconv.Quote(n.ID), strconv.Quote(n.Name))
		if n.Extension != "" {
			fmt.Fprintf(&b, ", Extension: %s", strconv.Quote(n.Extension))
		}
		fmt.Fprintf(&b, ", Parent: %d},\n", n.Parent)
	}
	b.WriteString("\t}))\n}\n")
	return b.Bytes()
}

func flatten(nodes []node, parent int, out *[]rawNode) {
	for _, n := range nodes {
		idx := len(*out)
		*out = append(*out, rawNode{ID: n.ID, Name: n.Name, Extension: n.Extension, Parent: parent})
		flatten(n.Children, idx, out)
	}
}

func renderMapping(pkg string, doc mappingDoc) []byte {
	var b bytes.Buffer
	writeHeader(&b, pkg)
	b.WriteString("func init() {\n\tregister(dict.NewMapping(")
	writeMeta(&b, doc.Meta)
	b.WriteString(", []dict.Entry{\n")
	for _, e := range doc.Entries {
		b.WriteString("\t\t{From: ")
		writeRef(&b, e.From, false)
		if e.To != nil {
			b.WriteString(", To: ")
			writeRef(&b, *e.To, true)
		}
		if e.Note != "" {
			fmt.Fprintf(&b, ", Note: %s", strconv.Quote(e.Note))
		}
		b.WriteString("},\n")
	}
	b.WriteString("\t}))\n}\n")
	return b.Bytes()
}

func writeHeader(b *bytes.Buffer, pkg string) {
	b.WriteString("// Code generated by gendict. DO NOT EDIT.\n\n")
	fmt.Fprintf(b, "package %s\n\n", pkg)
	b.WriteString("import \"github.com/rtb-0/rtbdict/iab/dict\"\n\n")
}

func writeMeta(b *bytes.Buffer, m meta) {
	b.WriteString("dict.Meta{")
	fields := []struct{ name, value string }{
		{"ID", m.ID},
		{"Kind", m.Kind},
		{"Name", m.Name},
		{"Version", m.Version},
		{"Publisher", m.Publisher},
		{"Page", m.Page},
		{"Source", m.Source},
		{"Released", m.Released},
		{"From", m.From},
		{"To", m.To},
		{"Notes", m.Notes},
	}
	for _, f := range fields {
		if f.value == "" {
			continue
		}
		fmt.Fprintf(b, "%s: %s, ", f.name, strconv.Quote(f.value))
	}
	if m.Cattax != 0 {
		fmt.Fprintf(b, "Cattax: %d, ", m.Cattax)
	}
	if m.Deprecated {
		b.WriteString("Deprecated: true, ")
	}
	if m.FromCattax != 0 {
		fmt.Fprintf(b, "FromCattax: %d, ", m.FromCattax)
	}
	if m.ToCattax != 0 {
		fmt.Fprintf(b, "ToCattax: %d, ", m.ToCattax)
	}
	b.WriteString("}")
}

func writeRef(b *bytes.Buffer, r ref, pointer bool) {
	if pointer {
		b.WriteByte('&')
	}
	b.WriteString("dict.Ref{")
	if r.ID != "" {
		fmt.Fprintf(b, "ID: %s, ", strconv.Quote(r.ID))
	}
	if r.Name != "" {
		fmt.Fprintf(b, "Name: %s, ", strconv.Quote(r.Name))
	}
	if len(r.Path) > 0 {
		b.WriteString("Path: []string{")
		for i, p := range r.Path {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(strconv.Quote(p))
		}
		b.WriteString("}, ")
	}
	b.WriteString("}")
}

func lookupSource(pkg, kind string) []byte {
	var b bytes.Buffer
	b.WriteString("// Code generated by gendict. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	b.WriteString("import \"github.com/rtb-0/rtbdict/iab/dict\"\n\n")
	if kind == "mapping" {
		b.WriteString(`var mappings []*dict.Mapping

func register(m *dict.Mapping) {
	mappings = append(mappings, m)
}

// ByID returns the mapping with meta.id.
func ByID(id string) (*dict.Mapping, bool) {
	for _, m := range mappings {
		if m.Meta.ID == id {
			return m, true
		}
	}
	return nil, false
}

// ByTaxonomy returns the mapping from one taxonomy id to another.
func ByTaxonomy(fromID, toID string) (*dict.Mapping, bool) {
	var found *dict.Mapping
	for _, m := range mappings {
		if m.Meta.From != fromID || m.Meta.To != toID {
			continue
		}
		if found != nil {
			return nil, false
		}
		found = m
	}
	if found == nil {
		return nil, false
	}
	return found, true
}
`)
		return b.Bytes()
	}
	b.WriteString(`var taxonomies []*dict.Taxonomy

func register(t *dict.Taxonomy) {
	taxonomies = append(taxonomies, t)
}

// ByID returns the taxonomy with meta.id.
func ByID(id string) (*dict.Taxonomy, bool) {
	for _, t := range taxonomies {
		if t.Meta.ID == id {
			return t, true
		}
	}
	return nil, false
}

// ByCattax returns the taxonomy whose cattax is set to cattax.
func ByCattax(cattax int) (*dict.Taxonomy, bool) {
	if cattax == 0 {
		return nil, false
	}
	var found *dict.Taxonomy
	for _, t := range taxonomies {
		if t.Meta.Cattax != cattax {
			continue
		}
		if found != nil {
			return nil, false
		}
		found = t
	}
	if found == nil {
		return nil, false
	}
	return found, true
}
`)
	return b.Bytes()
}

func writeGo(path string, src []byte) error {
	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, formatted, 0o644)
}
