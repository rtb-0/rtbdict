// Command genlight writes compact Go dictionaries from iab/datasets into matching iab/lightdict directories.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	datasets := filepath.Join(root, "iab", "datasets")
	outRoot := filepath.Join(root, "iab", "lightdict")
	var specs []spec
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
		s, err := buildSpec(path, family, pkg, strings.TrimSuffix(file, ".json"))
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		specs = append(specs, s)
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sort.Slice(specs, func(i, j int) bool {
		if specs[i].family != specs[j].family {
			return specs[i].family < specs[j].family
		}
		return specs[i].file < specs[j].file
	})
	byFamily := map[string][]spec{}
	seenVar := map[string]string{}
	seenKind := map[string]string{}
	for _, s := range specs {
		key := s.family + "/" + s.varName
		if prev, ok := seenVar[key]; ok {
			fmt.Fprintf(os.Stderr, "var %s used by %s and %s\n", s.varName, prev, s.file)
			os.Exit(1)
		}
		seenVar[key] = s.file
		if prev, ok := seenKind[s.family]; ok && prev != s.kind {
			fmt.Fprintf(os.Stderr, "%s mixes %s and %s\n", s.family, prev, s.kind)
			os.Exit(1)
		}
		seenKind[s.family] = s.kind
		out := filepath.Join(outRoot, s.family, s.file+".go")
		if err := writeGo(out, s.src); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		byFamily[s.family] = append(byFamily[s.family], s)
	}
	for family, group := range byFamily {
		src := lookupSource(group)
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

type spec struct {
	family  string
	pkg     string
	file    string
	varName string
	kind    string
	id      string
	cattax  int
	from    string
	to      string
	src     []byte
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
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

type meta struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Version string `json:"version"`
	Cattax  int    `json:"cattax"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type node struct {
	ID       string `json:"id"`
	Children []node `json:"children"`
}

type ref struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type entry struct {
	From ref  `json:"from"`
	To   *ref `json:"to"`
}

type taxonomyDoc struct {
	Meta  meta   `json:"meta"`
	Nodes []node `json:"nodes"`
}

type mappingDoc struct {
	Meta    meta    `json:"meta"`
	Entries []entry `json:"entries"`
}

func buildSpec(path, family, pkg, file string) (spec, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return spec{}, err
	}
	var probe struct {
		Meta meta `json:"meta"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return spec{}, err
	}
	s := spec{
		family: family, pkg: pkg, file: file,
		varName: goName(file), kind: probe.Meta.Kind, id: probe.Meta.ID,
		cattax: probe.Meta.Cattax, from: probe.Meta.From, to: probe.Meta.To,
	}
	switch probe.Meta.Kind {
	case "taxonomy":
		var doc taxonomyDoc
		if err := json.Unmarshal(body, &doc); err != nil {
			return spec{}, err
		}
		s.src = renderTaxonomy(pkg, s.varName, doc)
	case "mapping":
		var doc mappingDoc
		if err := json.Unmarshal(body, &doc); err != nil {
			return spec{}, err
		}
		s.src = renderMapping(pkg, s.varName, doc)
	default:
		return spec{}, fmt.Errorf("unknown kind %q", probe.Meta.Kind)
	}
	return s, nil
}

func goName(file string) string {
	var b strings.Builder
	b.WriteByte('v')
	upper := true
	for _, r := range file {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if upper {
				r = unicode.ToUpper(r)
			}
			upper = false
			b.WriteRune(r)
			continue
		}
		upper = true
	}
	return b.String()
}

type flatNode struct {
	id     string
	parent int32
}

func renderTaxonomy(pkg, varName string, doc taxonomyDoc) []byte {
	var flat []flatNode
	flatten(doc.Nodes, -1, &flat)
	n := len(flat)
	ids := make([]string, n)
	parent := make([]int32, n)
	kids := make([][]int32, n)
	childCount := 0
	for i, node := range flat {
		ids[i] = node.id
		parent[i] = node.parent
		if node.parent >= 0 {
			kids[node.parent] = append(kids[node.parent], int32(i))
			childCount++
		}
	}
	childAt := make([]int32, 0, childCount)
	childOf := make([]int32, n+1)
	for i, ch := range kids {
		childOf[i] = int32(len(childAt))
		childAt = append(childAt, ch...)
	}
	childOf[n] = int32(len(childAt))

	var order []string
	groups := map[string][]int32{}
	for i, id := range ids {
		if _, ok := groups[id]; !ok {
			order = append(order, id)
			groups[id] = nil
		}
		groups[id] = append(groups[id], int32(i))
	}
	slots := make([]int32, 0, n)
	at := make(map[string]uint32, len(order))
	for _, id := range order {
		g := groups[id]
		start := len(slots)
		if start > 0xffff || len(g) > 0xffff {
			panic("lightdict: id span does not fit uint16")
		}
		slots = append(slots, g...)
		at[id] = uint32(start)<<16 | uint32(len(g))
	}

	var b bytes.Buffer
	writeHeader(&b, pkg)
	fmt.Fprintf(&b, "var %s = lightdict.NewTaxonomy(\n", varName)
	fmt.Fprintf(&b, "\t%s,\n", strconv.Quote(doc.Meta.ID))
	fmt.Fprintf(&b, "\t%s,\n", strconv.Quote(doc.Meta.Version))
	fmt.Fprintf(&b, "\t%d,\n", doc.Meta.Cattax)
	writeStrings(&b, ids)
	b.WriteString(",\n")
	writeInts(&b, parent)
	b.WriteString(",\n")
	writeInts(&b, childAt)
	b.WriteString(",\n")
	writeInts(&b, childOf)
	b.WriteString(",\n")
	writePacked(&b, at)
	b.WriteString(",\n")
	writeInts(&b, slots)
	b.WriteString(",\n)\n")
	return b.Bytes()
}

func flatten(nodes []node, parent int32, out *[]flatNode) {
	for _, n := range nodes {
		idx := int32(len(*out))
		*out = append(*out, flatNode{id: n.ID, parent: parent})
		flatten(n.Children, idx, out)
	}
}

func renderMapping(pkg, varName string, doc mappingDoc) []byte {
	var order []string
	groups := map[string][]string{}
	for _, e := range doc.Entries {
		if e.To == nil {
			continue
		}
		key := e.From.ID
		if key == "" {
			key = e.From.Name
		}
		val := e.To.ID
		if val == "" {
			val = e.To.Name
		}
		if key == "" || val == "" {
			continue
		}
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], val)
	}
	to := make([]string, 0)
	at := make(map[string]uint32, len(order))
	for _, key := range order {
		g := groups[key]
		start := len(to)
		if start > 0xffff || len(g) > 0xffff {
			panic("lightdict: mapping span does not fit uint16")
		}
		to = append(to, g...)
		at[key] = uint32(start)<<16 | uint32(len(g))
	}

	var b bytes.Buffer
	writeHeader(&b, pkg)
	fmt.Fprintf(&b, "var %s = lightdict.NewMapping(\n", varName)
	fmt.Fprintf(&b, "\t%s,\n", strconv.Quote(doc.Meta.ID))
	writeStrings(&b, to)
	b.WriteString(",\n")
	writePacked(&b, at)
	b.WriteString(",\n)\n")
	return b.Bytes()
}

func writeHeader(b *bytes.Buffer, pkg string) {
	b.WriteString("// Code generated by genlight. DO NOT EDIT.\n\n")
	fmt.Fprintf(b, "package %s\n\n", pkg)
	b.WriteString("import \"github.com/rtb-0/rtbdict/iab/lightdict\"\n\n")
}

func writeStrings(b *bytes.Buffer, vals []string) {
	b.WriteString("\t[]string{")
	if len(vals) == 0 {
		b.WriteString("}")
		return
	}
	b.WriteByte('\n')
	for _, v := range vals {
		fmt.Fprintf(b, "\t\t%s,\n", strconv.Quote(v))
	}
	b.WriteString("\t}")
}

func writeInts(b *bytes.Buffer, vals []int32) {
	b.WriteString("\t[]int32{")
	if len(vals) == 0 {
		b.WriteString("}")
		return
	}
	b.WriteByte('\n')
	for i, v := range vals {
		if i%16 == 0 {
			b.WriteString("\t\t")
		}
		b.WriteString(strconv.FormatInt(int64(v), 10))
		b.WriteByte(',')
		if i%16 == 15 || i == len(vals)-1 {
			b.WriteByte('\n')
		} else {
			b.WriteByte(' ')
		}
	}
	b.WriteString("\t}")
}

func writePacked(b *bytes.Buffer, at map[string]uint32) {
	keys := make([]string, 0, len(at))
	for k := range at {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.WriteString("\tmap[string]uint32{")
	if len(keys) == 0 {
		b.WriteString("}")
		return
	}
	b.WriteByte('\n')
	for _, k := range keys {
		packed := at[k]
		start := packed >> 16
		count := packed & 0xffff
		fmt.Fprintf(b, "\t\t%s: %d<<16 | %d,\n", strconv.Quote(k), start, count)
	}
	b.WriteString("\t}")
}

func lookupSource(group []spec) []byte {
	var b bytes.Buffer
	b.WriteString("// Code generated by genlight. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", group[0].pkg)
	b.WriteString("import \"github.com/rtb-0/rtbdict/iab/lightdict\"\n\n")
	if group[0].kind == "mapping" {
		writeMappingLookup(&b, group)
		return b.Bytes()
	}
	writeTaxonomyLookup(&b, group)
	return b.Bytes()
}

func writeTaxonomyLookup(b *bytes.Buffer, group []spec) {
	byID := append([]spec(nil), group...)
	sort.Slice(byID, func(i, j int) bool { return byID[i].id < byID[j].id })
	b.WriteString("// ByID returns the taxonomy with this id. An unknown id returns nil.\n")
	b.WriteString("func ByID(id string) *lightdict.Taxonomy {\n")
	b.WriteString("\tswitch id {\n")
	for _, s := range byID {
		fmt.Fprintf(b, "\tcase %s:\n\t\treturn %s\n", strconv.Quote(s.id), s.varName)
	}
	b.WriteString("\tdefault:\n\t\treturn nil\n\t}\n}\n\n")

	var withCattax []spec
	for _, s := range group {
		if s.cattax == 0 && s.family != "audience" {
			continue
		}
		withCattax = append(withCattax, s)
	}
	sort.Slice(withCattax, func(i, j int) bool { return withCattax[i].cattax < withCattax[j].cattax })
	b.WriteString("// ByCattax returns the taxonomy whose cattax is set to cattax. An unknown cattax returns nil.\n")
	b.WriteString("func ByCattax(cattax int) *lightdict.Taxonomy {\n")
	b.WriteString("\tswitch cattax {\n")
	for _, s := range withCattax {
		fmt.Fprintf(b, "\tcase %d:\n\t\treturn %s\n", s.cattax, s.varName)
	}
	b.WriteString("\tdefault:\n\t\treturn nil\n\t}\n}\n")
}

func writeMappingLookup(b *bytes.Buffer, group []spec) {
	byID := append([]spec(nil), group...)
	sort.Slice(byID, func(i, j int) bool { return byID[i].id < byID[j].id })
	b.WriteString("// ByID returns the mapping with this id. An unknown id returns nil.\n")
	b.WriteString("func ByID(id string) *lightdict.Mapping {\n")
	b.WriteString("\tswitch id {\n")
	for _, s := range byID {
		fmt.Fprintf(b, "\tcase %s:\n\t\treturn %s\n", strconv.Quote(s.id), s.varName)
	}
	b.WriteString("\tdefault:\n\t\treturn nil\n\t}\n}\n\n")

	type pair struct {
		from, to string
		varName  string
		unique   bool
	}
	index := map[string]int{}
	var pairs []pair
	for _, s := range group {
		key := s.from + "\x00" + s.to
		if i, ok := index[key]; ok {
			pairs[i].unique = false
			continue
		}
		index[key] = len(pairs)
		pairs = append(pairs, pair{from: s.from, to: s.to, varName: s.varName, unique: true})
	}
	fromOrder := []string{}
	byFrom := map[string][]pair{}
	for _, p := range pairs {
		if !p.unique {
			continue
		}
		if _, ok := byFrom[p.from]; !ok {
			fromOrder = append(fromOrder, p.from)
		}
		byFrom[p.from] = append(byFrom[p.from], p)
	}
	sort.Strings(fromOrder)
	b.WriteString("// ByTaxonomy returns the mapping from one taxonomy id to another.\n")
	b.WriteString("// An unknown pair, or a pair shared by more than one mapping, returns nil.\n")
	b.WriteString("func ByTaxonomy(fromID, toID string) *lightdict.Mapping {\n")
	b.WriteString("\tswitch fromID {\n")
	for _, from := range fromOrder {
		group := byFrom[from]
		sort.Slice(group, func(i, j int) bool { return group[i].to < group[j].to })
		fmt.Fprintf(b, "\tcase %s:\n", strconv.Quote(from))
		b.WriteString("\t\tswitch toID {\n")
		for _, p := range group {
			fmt.Fprintf(b, "\t\tcase %s:\n\t\t\treturn %s\n", strconv.Quote(p.to), p.varName)
		}
		b.WriteString("\t\t}\n")
	}
	b.WriteString("\t}\n\treturn nil\n}\n")
}

func writeGo(path string, src []byte) error {
	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, src)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, formatted, 0o644)
}
