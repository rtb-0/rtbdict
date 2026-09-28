// Command gendict writes r0/dict and r0/lightdict from r0/datasets/r0-1.0.json.
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
)

func main() {
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	nodes, err := build(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	byR0, byCattax := indexes(nodes)
	kwNodes, kwEdges := keywordTrie(nodes)
	if err := writeGo(filepath.Join(root, "r0", "keyword", "phrases.go"), renderKeyword(kwNodes, kwEdges)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, out := range []struct {
		dir  string
		desc bool
	}{
		{"dict", true},
		{"lightdict", false},
	} {
		src := render(out.dir, out.desc, nodes, byR0, byCattax)
		path := filepath.Join(root, "r0", out.dir, "1.0.go")
		if err := writeGo(path, src); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
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

type rawNode struct {
	ID          int       `json:"id"`
	R0Code      string    `json:"r0_code"`
	Content10   string    `json:"c1_0_code"`
	Content31   string    `json:"c3_1_code"`
	AdProduct20 string    `json:"ap2_0_code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Keywords    []string  `json:"keywords"`
	Position    int       `json:"position"`
	Children    []rawNode `json:"children"`
}

type node struct {
	id                         int
	r0Code, name, description  string
	keywords                   []string
	position, parent           int
	children                   []int
	content10, content20       string
	content21, content22       string
	content30, content31       string
	adProduct10, adProduct11   string
	adProduct20                string
	content21All, content22All []string
}

func build(root string) ([]node, error) {
	body, err := os.ReadFile(filepath.Join(root, "r0", "datasets", "r0-1.0.json"))
	if err != nil {
		return nil, err
	}
	var doc struct {
		Nodes []rawNode `json:"nodes"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	content22 := taxonomyIDs(filepath.Join(root, "iab", "datasets", "content", "2.2.json"))
	content30 := taxonomyIDs(filepath.Join(root, "iab", "datasets", "content", "3.0.json"))
	adProduct10 := taxonomyIDs(filepath.Join(root, "iab", "datasets", "ad-product", "1.0.json"))
	c10ToC20 := mapping(filepath.Join(root, "iab", "datasets", "mappings", "content-1.0-to-content-2.0.json"))
	c10ToC21 := mapping(filepath.Join(root, "iab", "datasets", "mappings", "content-2.0-to-content-2.1.json"))
	ap20ToC21 := mapping(filepath.Join(root, "iab", "datasets", "mappings", "ad-product-2.0-to-content-2.1.json"))
	ap20ToAP11 := mapping(filepath.Join(root, "iab", "datasets", "mappings", "ad-product-2.0-to-ad-product-1.1.json"))

	var nodes []node
	var walk func([]rawNode, int)
	walk = func(raw []rawNode, parent int) {
		for _, r := range raw {
			n := node{
				id: r.ID, r0Code: r.R0Code, name: r.Name, description: r.Description,
				keywords: r.Keywords, position: r.Position, parent: parent,
				content10: r.Content10, content31: r.Content31, adProduct20: r.AdProduct20,
			}
			for _, child := range r.Children {
				n.children = append(n.children, child.ID)
			}
			if n.content10 != "" {
				n.content20 = first(c10ToC20[n.content10])
				n.content21All = c10ToC21[n.content10]
			} else if n.adProduct20 != "" {
				n.content21All = ap20ToC21[n.adProduct20]
			}
			n.content21 = first(n.content21All)
			for _, code := range n.content21All {
				if content22[code] {
					n.content22All = append(n.content22All, code)
				}
			}
			n.content22 = first(n.content22All)
			if n.content31 != "" && content30[n.content31] {
				n.content30 = n.content31
			}
			if n.adProduct20 != "" {
				n.adProduct11 = first(ap20ToAP11[n.adProduct20])
			}
			if n.adProduct11 != "" && adProduct10[n.adProduct11] {
				n.adProduct10 = n.adProduct11
			}
			nodes = append(nodes, n)
			walk(r.Children, r.ID)
		}
	}
	walk(doc.Nodes, 0)
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].id < nodes[j].id })
	for i, n := range nodes {
		if n.id != i+1 {
			return nil, fmt.Errorf("r0 id gap at %d", i+1)
		}
	}
	return nodes, nil
}

func first(codes []string) string {
	if len(codes) == 0 {
		return ""
	}
	return codes[0]
}

func taxonomyIDs(path string) map[string]bool {
	body, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var doc struct {
		Nodes []taxNode `json:"nodes"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		panic(err)
	}
	out := map[string]bool{}
	var walk func([]taxNode)
	walk = func(nodes []taxNode) {
		for _, n := range nodes {
			out[n.ID] = true
			walk(n.Children)
		}
	}
	walk(doc.Nodes)
	return out
}

type taxNode struct {
	ID       string    `json:"id"`
	Children []taxNode `json:"children"`
}

func mapping(path string) map[string][]string {
	body, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var doc struct {
		Entries []struct {
			From struct {
				ID string `json:"id"`
			} `json:"from"`
			To *struct {
				ID string `json:"id"`
			} `json:"to"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		panic(err)
	}
	out := map[string][]string{}
	for _, e := range doc.Entries {
		if e.From.ID == "" || e.To == nil || e.To.ID == "" {
			continue
		}
		list := out[e.From.ID]
		seen := false
		for _, id := range list {
			if id == e.To.ID {
				seen = true
				break
			}
		}
		if !seen {
			out[e.From.ID] = append(list, e.To.ID)
		}
	}
	return out
}

func indexes(nodes []node) (map[string]int, [10]map[string]int) {
	byR0 := make(map[string]int, len(nodes))
	owners := map[int]map[string][]int{}
	add := func(cattax int, code string, id int) {
		if code == "" {
			return
		}
		if owners[cattax] == nil {
			owners[cattax] = map[string][]int{}
		}
		owners[cattax][code] = append(owners[cattax][code], id)
	}
	for _, n := range nodes {
		if n.r0Code == "" {
			panic(fmt.Sprintf("node %d missing r0_code", n.id))
		}
		if _, ok := byR0[n.r0Code]; ok {
			panic("duplicate r0_code " + n.r0Code)
		}
		byR0[n.r0Code] = n.id
		add(1, n.content10, n.id)
		add(2, n.content20, n.id)
		for _, code := range n.content21All {
			add(5, code, n.id)
		}
		for _, code := range n.content22All {
			add(6, code, n.id)
		}
		add(7, n.content30, n.id)
		add(8, n.adProduct20, n.id)
		add(9, n.content31, n.id)
		add(3, n.adProduct10, n.id)
	}
	var byCattax [10]map[string]int
	for cattax, codes := range owners {
		byCattax[cattax] = map[string]int{}
		for code, ids := range codes {
			if len(ids) == 1 {
				byCattax[cattax][code] = ids[0]
			}
		}
	}
	return byR0, byCattax
}

func render(pkg string, withDescription bool, nodes []node, byR0 map[string]int, byCattax [10]map[string]int) []byte {
	var b bytes.Buffer
	b.WriteString("// Code generated by r0/cmd/gendict. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	b.WriteString("var nodes = []Node{\n")
	for _, n := range nodes {
		fmt.Fprintf(&b, "\t{id: %d, r0Code: %s, name: %s", n.id, strconv.Quote(n.r0Code), strconv.Quote(n.name))
		if withDescription && n.description != "" {
			fmt.Fprintf(&b, ", description: %s", strconv.Quote(n.description))
		}
		writeStrings(&b, "keywords", n.keywords)
		if n.position != 0 {
			fmt.Fprintf(&b, ", position: %d", n.position)
		}
		writeCode(&b, "content10", n.content10)
		writeCode(&b, "content20", n.content20)
		writeCode(&b, "content21", n.content21)
		writeCode(&b, "content22", n.content22)
		writeCode(&b, "content30", n.content30)
		writeCode(&b, "content31", n.content31)
		writeCode(&b, "adProduct10", n.adProduct10)
		writeCode(&b, "adProduct11", n.adProduct11)
		writeCode(&b, "adProduct20", n.adProduct20)
		if n.parent != 0 {
			fmt.Fprintf(&b, ", parent: %d", n.parent)
		}
		if len(n.children) > 0 {
			b.WriteString(", children: []int{")
			for i, id := range n.children {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(strconv.Itoa(id))
			}
			b.WriteString("}")
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n\n")
	writeMap(&b, "var byR0 = map[string]int{\n", byR0)
	b.WriteString("\nvar byCattax = [10]map[string]int{\n")
	for cattax, codes := range byCattax {
		if len(codes) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\t%d: map[string]int{\n", cattax)
		keys := make([]string, 0, len(codes))
		for code := range codes {
			keys = append(keys, code)
		}
		sort.Strings(keys)
		for _, code := range keys {
			fmt.Fprintf(&b, "\t\t%s: %d,\n", strconv.Quote(code), codes[code])
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n")
	return b.Bytes()
}

func writeCode(b *bytes.Buffer, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(b, ", %s: %s", name, strconv.Quote(value))
}

func writeStrings(b *bytes.Buffer, name string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Fprintf(b, ", %s: []string{", name)
	for i, v := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Quote(v))
	}
	b.WriteString("}")
}

func writeMap(b *bytes.Buffer, header string, values map[string]int) {
	b.WriteString(header)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(b, "\t%s: %d,\n", strconv.Quote(key), values[key])
	}
	b.WriteString("}\n")
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
