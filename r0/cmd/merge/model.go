package main

import (
	"strings"
	"unicode"
)

// node is one category in a merged tree. Version codes stay empty until the
// category actually exists in that IAB version.
type node struct {
	id       int
	r0       string
	name     string
	iab      string
	c10      string
	c20      string
	c21      string
	c22      string
	c30      string
	c31      string
	ap10     string
	ap11     string
	ap20     string
	parent   *node
	children []*node
}

func (n *node) add(child *node) {
	child.parent = n
	n.children = append(n.children, child)
}

type placement struct {
	Category string `json:"category"`
	Code     string `json:"code,omitempty"`
	Parent   string `json:"parent"`
	Reason   string `json:"reason"`
}

type taxNode struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Children []taxNode `json:"children"`
}

type taxDoc struct {
	Nodes []taxNode `json:"nodes"`
}

type mapRef struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Path []string `json:"path"`
}

type mapEntry struct {
	From mapRef  `json:"from"`
	To   *mapRef `json:"to"`
}

type mapDoc struct {
	Entries []mapEntry `json:"entries"`
}

func walkTax(nodes []taxNode, parent string, visit func(n taxNode, parent string)) {
	for _, n := range nodes {
		visit(n, parent)
		walkTax(n.Children, n.ID, visit)
	}
}

func taxIDs(nodes []taxNode) map[string]bool {
	out := map[string]bool{}
	walkTax(nodes, "", func(n taxNode, _ string) { out[n.ID] = true })
	return out
}

func normName(s string) string {
	s = strings.ReplaceAll(s, "&", "and")
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func normPath(path []string) string {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = normName(p)
	}
	return strings.Join(parts, "\x1f")
}

var stopWords = map[string]bool{
	"and": true, "the": true, "of": true, "or": true,
	"a": true, "for": true, "to": true, "in": true,
}

// words returns lowercase a-z tokens, skipping the r0 stop list.
func words(name string) []string {
	var b strings.Builder
	var raw []string
	flush := func() {
		if b.Len() == 0 {
			return
		}
		raw = append(raw, b.String())
		b.Reset()
	}
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(unicode.ToLower(r))
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case unicode.IsLetter(r):
			// Keep the token open so "Māori" stays one word.
		default:
			flush()
		}
	}
	flush()
	var out []string
	for _, w := range raw {
		if w != "" && !stopWords[w] {
			out = append(out, w)
		}
	}
	return out
}

func stem(name string, n int) string {
	ws := words(name)
	if len(ws) == 0 {
		return "x"
	}
	s := ws[0]
	if len(s) > n {
		s = s[:n]
	}
	if s == "" {
		return "x"
	}
	return s
}

func assignIDs(roots []*node) {
	id := 1
	level := roots
	for len(level) > 0 {
		var next []*node
		for _, n := range level {
			n.id = id
			id++
			next = append(next, n.children...)
		}
		level = next
	}
}

func assignR0(roots []*node) {
	used := map[string]bool{}
	for _, r := range roots {
		base := stem(r.name, 3)
		code := base
		for i := 1; used[code]; i++ {
			code = base + itoa(i)
		}
		r.r0 = code
		used[code] = true
		assignChildR0(r, used)
	}
}

func assignChildR0(parent *node, used map[string]bool) {
	seen := map[string]int{}
	for _, c := range parent.children {
		letter := stem(c.name, 1)
		n := seen[letter]
		seen[letter] = n + 1
		seg := letter
		if n > 0 {
			seg = letter + itoa(n)
		}
		code := parent.r0 + "." + seg
		for used[code] {
			n++
			seg = letter + itoa(n)
			code = parent.r0 + "." + seg
			seen[letter] = n + 1
		}
		c.r0 = code
		used[code] = true
		assignChildR0(c, used)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func findPath(roots []*node, path []string) (*node, error) {
	var found *node
	var walk func([]*node, int)
	walk = func(ns []*node, depth int) {
		for _, n := range ns {
			if normName(n.name) == normName(path[depth]) {
				if depth == len(path)-1 {
					found = n
					return
				}
				walk(n.children, depth+1)
				if found != nil {
					return
				}
			}
		}
	}
	walk(roots, 0)
	if found == nil {
		return nil, errf("path not found: %s", strings.Join(path, " / "))
	}
	return found, nil
}

func countNodes(roots []*node) int {
	n := 0
	var walk func([]*node)
	walk = func(ns []*node) {
		for _, node := range ns {
			n++
			walk(node.children)
		}
	}
	walk(roots)
	return n
}
