package content_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/r0/dict/content"
	light "github.com/rtb-0/rtbdict/r0/lightdict/content"
)

func TestLookup(t *testing.T) {
	full := content.ByID(40)
	n := light.ByID(40)
	if n == nil || n.R0Code() != full.R0Code() || n.Name() != full.Name() {
		t.Fatalf("light %+v", n)
	}
	if !sameIAB(full, n) {
		t.Fatal("codes differ")
	}
	if light.ByR0Code("adl").R0Code() != "adl" {
		t.Fatal("adl")
	}
	if light.ByCattax(1, "IAB1-1").ID() != 40 {
		t.Fatal("cattax 1")
	}
	if light.ByKeyword("PORN VIDEO").ID() != 35 || light.ByKeyword("adult industry news").ID() != 38 {
		t.Fatal("keyword")
	}
	allocs := testing.AllocsPerRun(100, func() {
		n := light.ByID(35)
		_ = n.R0Code()
		_ = n.IAB()
		_ = n.IDs(9)
		_ = light.ByCattax(9, "Rm3SiT")
		_ = light.ByKeyword("PORN VIDEO")
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}

func sameIAB(full *content.Node, n *light.Node) bool {
	codes := n.IAB()
	if len(full.IAB()) != len(codes) {
		return false
	}
	var want [10][]string
	for i, code := range full.IAB() {
		if code.Cattax != codes[i].Cattax || code.ID != codes[i].ID {
			return false
		}
		want[code.Cattax] = append(want[code.Cattax], code.ID)
	}
	for cattax, ids := range want {
		got := n.IDs(cattax)
		if len(ids) != len(got) {
			return false
		}
		for i := range ids {
			if ids[i] != got[i] {
				return false
			}
		}
	}
	return n.IDs(-1) == nil && n.IDs(10) == nil
}

func TestChildren(t *testing.T) {
	root := light.ByID(1)
	var direct []int
	for child := range root.ChildrenNodes {
		direct = append(direct, child.ID())
	}
	if len(direct) != len(root.Children()) || direct[0] != 35 {
		t.Fatalf("direct %v", direct)
	}

	fam := light.ByID(9)
	var ids []int
	for id := range fam.AllChildren {
		ids = append(ids, id)
	}
	if len(ids) != 4 || ids[0] != 67 || ids[1] != 171 || ids[3] != 69 {
		t.Fatalf("all %v", ids)
	}

	var nodes []int
	for child := range fam.AllChildrenNodes {
		nodes = append(nodes, child.ID())
	}
	for i := range ids {
		if nodes[i] != ids[i] {
			t.Fatalf("order ids %v nodes %v", ids, nodes)
		}
	}

	n := 0
	for range light.ByID(35).AllChildren {
		n++
	}
	var none *light.Node
	for range none.ChildrenNodes {
		n++
	}
	if n != 0 {
		t.Fatalf("empty %d", n)
	}

	allocs := testing.AllocsPerRun(100, func() {
		for id := range root.AllChildren {
			childSink = id
		}
		for child := range root.ChildrenNodes {
			childSink = child.ID()
		}
		for child := range root.AllChildrenNodes {
			childSink = child.ID()
		}
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}

var childSink int
