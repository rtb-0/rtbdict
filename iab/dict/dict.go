//go:generate go run github.com/rtb-0/rtbdict/iab/cmd/gendict

// Package dict holds IAB taxonomy and mapping types shared by the generated dictionaries.
package dict

import "fmt"

// Meta describes a taxonomy or a mapping. Cattax, FromCattax, and ToCattax are 0
// when the source JSON omitted the field.
type Meta struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Version    string `json:"version,omitempty"`
	Cattax     int    `json:"cattax,omitempty"`
	Deprecated bool   `json:"deprecated,omitempty"`
	Publisher  string `json:"publisher"`
	Page       string `json:"page"`
	Source     string `json:"source"`
	Released   string `json:"released,omitempty"`
	From       string `json:"from,omitempty"`
	FromCattax int    `json:"from_cattax,omitempty"`
	To         string `json:"to,omitempty"`
	ToCattax   int    `json:"to_cattax,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// Node is one category in a taxonomy tree.
type Node struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Extension string  `json:"extension,omitempty"`
	Parent    *Node   `json:"-"`
	Children  []*Node `json:"children,omitempty"`
}

// Taxonomy is a category tree. Roots are the nodes with no parent.
type Taxonomy struct {
	Meta  Meta               `json:"meta"`
	Roots []*Node            `json:"roots,omitempty"`
	byID  map[string][]*Node `json:"-"`
}

// Nodes returns every node with id. Ad Product 1.0 and 1.1 use id "51" twice.
func (t *Taxonomy) Nodes(id string) []*Node {
	if t == nil {
		return nil
	}
	return t.byID[id]
}

// RawNode is one taxonomy row before the tree is linked. Parent is an index
// into the same slice, or -1 for a root.
type RawNode struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Extension string `json:"extension,omitempty"`
	Parent    int    `json:"parent"`
}

// NewTaxonomy links raw rows into a tree. Children stay in slice order.
func NewTaxonomy(meta Meta, raw []RawNode) (*Taxonomy, error) {
	nodes := make([]*Node, len(raw))
	for i, r := range raw {
		if r.ID == "" || r.Name == "" {
			return nil, fmt.Errorf("node %d missing id or name", i)
		}
		nodes[i] = &Node{ID: r.ID, Name: r.Name, Extension: r.Extension}
	}
	byID := make(map[string][]*Node, len(raw))
	var roots []*Node
	for i, r := range raw {
		n := nodes[i]
		byID[n.ID] = append(byID[n.ID], n)
		if r.Parent == -1 {
			roots = append(roots, n)
			continue
		}
		if r.Parent < 0 || r.Parent >= len(nodes) || r.Parent == i {
			return nil, fmt.Errorf("node %s bad parent %d", n.ID, r.Parent)
		}
		p := nodes[r.Parent]
		n.Parent = p
		p.Children = append(p.Children, n)
	}
	for i, n := range nodes {
		seen := map[*Node]bool{}
		for p := n.Parent; p != nil; p = p.Parent {
			if seen[p] {
				return nil, fmt.Errorf("cycle at %s", raw[i].ID)
			}
			seen[p] = true
		}
	}
	return &Taxonomy{Meta: meta, Roots: roots, byID: byID}, nil
}

// MustTaxonomy is NewTaxonomy for generated init functions.
func MustTaxonomy(meta Meta, raw []RawNode) *Taxonomy {
	t, err := NewTaxonomy(meta, raw)
	if err != nil {
		panic(err)
	}
	return t
}

// Ref is one side of a mapping entry.
type Ref struct {
	ID   string   `json:"id,omitempty"`
	Name string   `json:"name,omitempty"`
	Path []string `json:"path,omitempty"`
}

// Entry maps one category onto another. To is nil when the source row has no target.
type Entry struct {
	From Ref    `json:"from"`
	To   *Ref   `json:"to,omitempty"`
	Note string `json:"note,omitempty"`
}

// Mapping is a list of category pairs.
type Mapping struct {
	Meta    Meta               `json:"meta"`
	Entries []Entry            `json:"entries"`
	byFrom  map[string][]Entry `json:"-"`
}

// NewMapping indexes entries by From.ID, or by From.Name when the id is empty.
func NewMapping(meta Meta, entries []Entry) *Mapping {
	byFrom := make(map[string][]Entry, len(entries))
	for _, e := range entries {
		key := e.From.ID
		if key == "" {
			key = e.From.Name
		}
		if key == "" {
			continue
		}
		byFrom[key] = append(byFrom[key], e)
	}
	return &Mapping{Meta: meta, Entries: entries, byFrom: byFrom}
}

// From returns entries whose source id is id. For genre mappings the source has
// no id, and id is matched against From.Name.
func (m *Mapping) From(id string) []Entry {
	if m == nil {
		return nil
	}
	return m.byFrom[id]
}
