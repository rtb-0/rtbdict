//go:generate go run github.com/rtb-0/rtbdict/r0/cmd/gendict

// Package lightdict is the r0 category dictionary without descriptions.
package lightdict

import "github.com/rtb-0/rtbdict/r0/keyword"

// Node is one r0 category. It has the same codes as dict.Node and no description.
type Node struct {
	id          int
	r0Code      string
	name        string
	keywords    []string
	position    int
	content10   string
	content20   string
	content21   string
	content22   string
	content30   string
	content31   string
	adProduct10 string
	adProduct11 string
	adProduct20 string
	parent      int
	children    []int
}

// ID returns the r0 numeric id. Ids are assigned level by level, starting at 1.
func (n *Node) ID() int {
	if n == nil {
		return 0
	}
	return n.id
}

// R0Code returns the hierarchical r0 code.
func (n *Node) R0Code() string {
	if n == nil {
		return ""
	}
	return n.r0Code
}

// Name returns the category name.
func (n *Node) Name() string {
	if n == nil {
		return ""
	}
	return n.name
}

// Keywords returns the category keywords. The result is the stored slice.
func (n *Node) Keywords() []string {
	if n == nil {
		return nil
	}
	return n.keywords
}

// Position returns the sibling position, or 0 when the dataset omitted it.
func (n *Node) Position() int {
	if n == nil {
		return 0
	}
	return n.position
}

// IABContent10 returns the Content Taxonomy 1.0 code, cattax 1.
func (n *Node) IABContent10() string {
	if n == nil {
		return ""
	}
	return n.content10
}

// IABContent20 returns the Content Taxonomy 2.0 code, cattax 2.
func (n *Node) IABContent20() string {
	if n == nil {
		return ""
	}
	return n.content20
}

// IABContent21 returns the Content Taxonomy 2.1 code, cattax 5.
func (n *Node) IABContent21() string {
	if n == nil {
		return ""
	}
	return n.content21
}

// IABContent22 returns the Content Taxonomy 2.2 code, cattax 6.
func (n *Node) IABContent22() string {
	if n == nil {
		return ""
	}
	return n.content22
}

// IABContent30 returns the Content Taxonomy 3.0 code, cattax 7.
func (n *Node) IABContent30() string {
	if n == nil {
		return ""
	}
	return n.content30
}

// IABContent31 returns the Content Taxonomy 3.1 code, cattax 9.
func (n *Node) IABContent31() string {
	if n == nil {
		return ""
	}
	return n.content31
}

// IABAdProduct10 returns the Ad Product Taxonomy 1.0 code, cattax 3.
func (n *Node) IABAdProduct10() string {
	if n == nil {
		return ""
	}
	return n.adProduct10
}

// IABAdProduct11 returns the Ad Product Taxonomy 1.1 code. Version 1.1 has no cattax.
func (n *Node) IABAdProduct11() string {
	if n == nil {
		return ""
	}
	return n.adProduct11
}

// IABAdProduct20 returns the Ad Product Taxonomy 2.0 code, cattax 8.
func (n *Node) IABAdProduct20() string {
	if n == nil {
		return ""
	}
	return n.adProduct20
}

// Parent returns the parent r0 id, or 0 for a root.
func (n *Node) Parent() int {
	if n == nil {
		return 0
	}
	return n.parent
}

// Children returns child r0 ids. The result is the stored slice.
func (n *Node) Children() []int {
	if n == nil {
		return nil
	}
	return n.children
}

// ChildrenNodes iterates the direct child categories, in Children order.
func (n *Node) ChildrenNodes(yield func(*Node) bool) {
	if n == nil {
		return
	}
	for _, id := range n.children {
		if !yield(ByID(id)) {
			return
		}
	}
}

// AllChildren iterates the id of every descendant.
// A child comes before that child's subtree, then the next sibling.
func (n *Node) AllChildren(yield func(int) bool) {
	if n == nil {
		return
	}
	n.walkIDs(yield)
}

func (n *Node) walkIDs(yield func(int) bool) bool {
	for _, id := range n.children {
		if !yield(id) || !ByID(id).walkIDs(yield) {
			return false
		}
	}
	return true
}

// AllChildrenNodes iterates every descendant category, in AllChildren order.
func (n *Node) AllChildrenNodes(yield func(*Node) bool) {
	if n == nil {
		return
	}
	n.walkNodes(yield)
}

func (n *Node) walkNodes(yield func(*Node) bool) bool {
	for _, id := range n.children {
		child := ByID(id)
		if !yield(child) || !child.walkNodes(yield) {
			return false
		}
	}
	return true
}

// ByID returns the category with this r0 id. An unknown id returns nil.
func ByID(id int) *Node {
	if id < 1 || id > len(nodes) {
		return nil
	}
	return &nodes[id-1]
}

// ByKeyword returns the category whose name or keyword best matches text.
// Matching ignores ASCII case. An unknown phrase returns nil.
func ByKeyword(text string) *Node {
	return ByID(keyword.Match(text))
}

// ByR0Code returns the category with this r0 code. An unknown code returns nil.
func ByR0Code(code string) *Node {
	id, ok := byR0[code]
	if !ok {
		return nil
	}
	return ByID(id)
}

// ByCattax returns the category identified by an IAB cattax and category code.
// A code shared by more than one r0 category returns nil.
// Ad Product 1.1 has no cattax; use IABAdProduct11 on a category found another way.
func ByCattax(cattax int, code string) *Node {
	if cattax < 0 || cattax >= len(byCattax) || code == "" || byCattax[cattax] == nil {
		return nil
	}
	id, ok := byCattax[cattax][code]
	if !ok {
		return nil
	}
	return ByID(id)
}
