// Package content is the r0 content dictionary, including descriptions.
package content

import (
	"github.com/rtb-0/rtbdict"
	keyword "github.com/rtb-0/rtbdict/r0/keyword/content"
)

// IAB is one Content Taxonomy code placed on an r0 category.
type IAB struct {
	Cattax int
	ID     string
	Name   string
}

// Node is one r0 content category.
// ids holds the Content Taxonomy ids for cattax 0 through 9.
type Node struct {
	id          int
	r0Code      string
	name        string
	description string
	keywords    []string
	position    int
	iab         []IAB
	ids         [10][]string
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

// Description returns the category description.
func (n *Node) Description() string {
	if n == nil {
		return ""
	}
	return n.description
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

// IAB returns the Content Taxonomy codes placed on this category.
// The result is the stored slice, in dataset order.
func (n *Node) IAB() []IAB {
	if n == nil {
		return nil
	}
	return n.iab
}

// IDs returns the Content Taxonomy ids for cattax, in dataset order.
// The result is the stored slice. An unknown cattax returns nil.
func (n *Node) IDs(cattax int) []string {
	if n == nil || cattax < 0 || cattax >= len(n.ids) {
		return nil
	}
	return n.ids[cattax]
}

// Parent returns the parent r0 id, or 0 for a root.
func (n *Node) Parent() int {
	if n == nil {
		return 0
	}
	return n.parent
}

// HasParent reports whether id is this category's parent or a further ancestor.
// The category itself does not count. A nil node returns false.
func (n *Node) HasParent[T rtbdict.ID](id T) bool {
	if n == nil || id == 0 || int(id) == n.id {
		return false
	}
	for p := n.parent; p != 0; p = ByID(p).Parent() {
		if p == int(id) {
			return true
		}
	}
	return false
}

// HasParentOneOf reports whether any id is this category's parent or a further ancestor.
func (n *Node) HasParentOneOf[T rtbdict.ID](ids ...T) bool {
	if n == nil {
		return false
	}
	var mask uint64
	for _, id := range ids {
		if id == 0 || int(id) == n.id {
			continue
		}
		mask |= uint64(1) << (uint(id) & 63)
	}
	if mask == 0 {
		return false
	}
	for p := n.parent; p != 0; p = ByID(p).Parent() {
		if mask&(uint64(1)<<(uint(p)&63)) == 0 {
			continue
		}
		for _, id := range ids {
			if p == int(id) {
				return true
			}
		}
	}
	return false
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

// ByCattax returns the category that carries this Content Taxonomy cattax and id.
// An unknown pair returns nil.
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
