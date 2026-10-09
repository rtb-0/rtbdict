package content_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/r0/dict/content"
)

func TestLookup(t *testing.T) {
	root := content.ByID(1)
	if root == nil || root.R0Code() != "adl" || root.Parent() != 0 || root.Description() == "" {
		t.Fatalf("root %+v", root)
	}
	if root.Children()[0] != 35 || !hasIAB(root, 6, "Rm3SiT", "Adult & Explicit Sexual Content") {
		t.Fatalf("root children %v iab %v", root.Children(), root.IAB())
	}
	if !hasIAB(root, 7, "Rm3SiT", "Adult & Explicit Sexual Content") || !hasIAB(root, 9, "Rm3SiT", "Adult & Explicit Sexual Content") {
		t.Fatalf("root iab %v", root.IAB())
	}
	porn := content.ByR0Code("adl.prn")
	if porn == nil || porn.ID() != 35 || porn.Parent() != 1 || !hasIAB(porn, 1, "IAB25-3", "Pornography") {
		t.Fatalf("adl.prn id=%d parent=%d iab=%v", porn.ID(), porn.Parent(), porn.IAB())
	}
	if content.ByCattax(6, "Rm3SiT") != root || content.ByCattax(7, "Rm3SiT") != root || content.ByCattax(9, "Rm3SiT") != root {
		t.Fatal("Rm3SiT")
	}
	if content.ByCattax(1, "IAB25-3") != porn {
		t.Fatal("IAB25-3")
	}
	if len(porn.IDs(1)) != 1 || porn.IDs(1)[0] != "IAB25-3" || len(root.IDs(6)) != 1 || root.IDs(6)[0] != "Rm3SiT" {
		t.Fatalf("ids porn=%v root=%v", porn.IDs(1), root.IDs(6))
	}
	if porn.IDs(-1) != nil || porn.IDs(10) != nil {
		t.Fatal("unknown cattax")
	}
	books := content.ByCattax(1, "IAB1-1")
	if books == nil || books.R0Code() != "art.bl" || books.ID() != 40 || books.Description() == "" {
		t.Fatalf("IAB1-1 %+v", books)
	}
	if len(porn.Keywords()) == 0 || porn.Children() != nil {
		t.Fatalf("leaf keywords=%v children=%v", porn.Keywords(), porn.Children())
	}
	if content.ByID(0) != nil || content.ByR0Code("missing") != nil || content.ByCattax(1, "missing") != nil {
		t.Fatal("missing lookup")
	}
	if content.ByKeyword("Pornography").ID() != 35 || content.ByKeyword("PORN VIDEO").ID() != 35 {
		t.Fatal("name case")
	}
	if content.ByKeyword("porn videos").ID() != 35 || content.ByKeyword("  porn   video  ").ID() != 35 {
		t.Fatal("keyword prefix")
	}
	if content.ByKeyword("adult industry news").ID() != 38 || content.ByKeyword("no such phrase") != nil {
		t.Fatal("keyword longest or miss")
	}
	allocs := testing.AllocsPerRun(100, func() {
		n := content.ByR0Code("adl.prn")
		_ = n.ID()
		_ = n.IAB()
		_ = n.IDs(1)
		_ = content.ByCattax(9, "Rm3SiT").R0Code()
		_ = root.Children()
		_ = content.ByKeyword("PORN VIDEO").R0Code()
		_ = content.ByKeyword("porn videos").ID()
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}

func hasIAB(n *content.Node, cattax int, id, name string) bool {
	for _, code := range n.IAB() {
		if code.Cattax == cattax && code.ID == id && code.Name == name {
			return true
		}
	}
	return false
}

func TestChildren(t *testing.T) {
	root := content.ByID(1)
	var direct []int
	for child := range root.ChildrenNodes {
		direct = append(direct, child.ID())
	}
	if len(direct) != len(root.Children()) || direct[0] != 35 {
		t.Fatalf("direct %v", direct)
	}

	fam := content.ByID(9)
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
	if len(nodes) != len(ids) {
		t.Fatalf("nodes %v", nodes)
	}
	for i := range ids {
		if nodes[i] != ids[i] {
			t.Fatalf("order ids %v nodes %v", ids, nodes)
		}
	}

	n := 0
	for range content.ByID(35).AllChildren {
		n++
	}
	for range content.ByID(35).ChildrenNodes {
		n++
	}
	var none *content.Node
	for range none.AllChildrenNodes {
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

func TestHasParent(t *testing.T) {
	baby := content.ByID(171)
	if !baby.HasParent(67) || !baby.HasParent(9) {
		t.Fatal("ancestors")
	}
	if baby.HasParent(171) || baby.HasParent(1) || baby.HasParent(0) {
		t.Fatal("self, unrelated, or zero")
	}
	if content.ByID(9).HasParent(9) {
		t.Fatal("root")
	}
	var none *content.Node
	if none.HasParent(9) || none.HasParentOneOf(9) {
		t.Fatal("nil")
	}
	if !baby.HasParentOneOf(1, 67) || baby.HasParentOneOf(1, 2) || baby.HasParentOneOf() || baby.HasParentOneOf(171) || !baby.HasParentOneOf(171, 67) {
		t.Fatal("one of")
	}
	if !content.ByID(39).HasParentOneOf(2) {
		t.Fatal("one of")
	}
	allocs := testing.AllocsPerRun(100, func() {
		if !baby.HasParent(9) {
			panic("parent")
		}
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}
