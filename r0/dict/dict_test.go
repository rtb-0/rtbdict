package dict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/r0/dict"
)

func TestLookup(t *testing.T) {
	root := dict.ByID(1)
	if root == nil || root.R0Code() != "adl" || root.IABAdProduct20() != "1001" || root.Parent() != 0 {
		t.Fatalf("root %+v", root)
	}
	if root.IABContent31() != "Rm3SiT" || root.Children()[0] != 33 {
		t.Fatalf("root c31=%q children %v", root.IABContent31(), root.Children())
	}
	porn := dict.ByR0Code("adl.p")
	if porn == nil || porn.ID() != 33 || porn.Parent() != 1 || porn.IABContent10() != "IAB25-3" {
		t.Fatalf("adl.p id=%d parent=%d c10=%q", porn.ID(), porn.Parent(), porn.IABContent10())
	}
	if dict.ByCattax(9, "Rm3SiT") != root {
		t.Fatal("cattax 9")
	}
	if dict.ByCattax(7, "Rm3SiT") != root || root.IABContent30() != "Rm3SiT" {
		t.Fatal("content 3.0")
	}
	books := dict.ByCattax(1, "IAB1-1")
	if books == nil || books.ID() != 40 || books.Description() == "" {
		t.Fatalf("IAB1-1 %+v", books)
	}
	arts := dict.ByID(99)
	if arts.IABContent10() != "IAB1" || arts.IABContent20() != "1014" {
		t.Fatalf("arts c10=%q c20=%q", arts.IABContent10(), arts.IABContent20())
	}
	clothing := dict.ByID(715)
	if clothing.IABContent10() != "IAB18-5" || clothing.IABContent21() != "566" || clothing.IABContent22() != "566" {
		t.Fatalf("clothing c21=%q c22=%q", clothing.IABContent21(), clothing.IABContent22())
	}
	if dict.ByCattax(5, "576") != clothing || dict.ByCattax(6, "576") != clothing {
		t.Fatal("IAB18-5 extra target")
	}
	leaf := dict.ByID(497)
	if leaf.Name() != "Cloud Storage" || len(leaf.Keywords()) == 0 || leaf.Children() != nil {
		t.Fatalf("leaf name=%q keywords=%v children=%v", leaf.Name(), leaf.Keywords(), leaf.Children())
	}
	if dict.ByID(0) != nil || dict.ByR0Code("missing") != nil || dict.ByCattax(1, "missing") != nil {
		t.Fatal("missing lookup")
	}
	if dict.ByKeyword("Cloud Storage").ID() != 497 || dict.ByKeyword("CLOUD STORAGE").ID() != 497 {
		t.Fatal("name case")
	}
	if dict.ByKeyword("hosted file synchronization").ID() != 497 || dict.ByKeyword("hosted file synchronizations").ID() != 497 {
		t.Fatal("keyword prefix")
	}
	if dict.ByKeyword("skin care").ID() != 1281 || dict.ByKeyword("no such phrase") != nil {
		t.Fatal("keyword collision or miss")
	}
	allocs := testing.AllocsPerRun(100, func() {
		n := dict.ByR0Code("adl.p")
		_ = n.ID()
		_ = n.IABContent31()
		_ = dict.ByCattax(8, "1001").IABAdProduct20()
		_ = root.Children()
		_ = dict.ByKeyword("CLOUD STORAGE").R0Code()
		_ = dict.ByKeyword("hosted file synchronization").ID()
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}

func TestChildren(t *testing.T) {
	root := dict.ByID(1)
	var direct []int
	for child := range root.ChildrenNodes {
		direct = append(direct, child.ID())
	}
	if len(direct) != len(root.Children()) || direct[0] != 33 {
		t.Fatalf("direct %v", direct)
	}

	var ids []int
	for id := range root.AllChildren {
		ids = append(ids, id)
		if len(ids) == 6 {
			break
		}
	}
	if len(ids) != 6 || ids[0] != 33 || ids[1] != 540 || ids[5] != 34 {
		t.Fatalf("all %v", ids)
	}

	var nodes []int
	for child := range root.AllChildrenNodes {
		nodes = append(nodes, child.ID())
		if len(nodes) == len(ids) {
			break
		}
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
	for range dict.ByID(497).AllChildren {
		n++
	}
	for range dict.ByID(497).ChildrenNodes {
		n++
	}
	var none *dict.Node
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
