package lightdict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/r0/dict"
	"github.com/rtb-0/rtbdict/r0/lightdict"
)

func TestLookup(t *testing.T) {
	full := dict.ByID(40)
	light := lightdict.ByID(40)
	if light == nil || light.R0Code() != full.R0Code() || light.Name() != full.Name() {
		t.Fatalf("light %+v", light)
	}
	if light.IABContent10() != full.IABContent10() || light.IABAdProduct20() != full.IABAdProduct20() {
		t.Fatal("codes differ")
	}
	if lightdict.ByR0Code("adl").IABAdProduct20() != "1001" {
		t.Fatal("adl")
	}
	if lightdict.ByCattax(1, "IAB1-1").ID() != 40 {
		t.Fatal("cattax 1")
	}
	if lightdict.ByKeyword("CLOUD STORAGE").ID() != 497 || lightdict.ByKeyword("skin care").ID() != 1281 {
		t.Fatal("keyword")
	}
	allocs := testing.AllocsPerRun(100, func() {
		n := lightdict.ByID(33)
		_ = n.R0Code()
		_ = n.IABContent31()
		_ = lightdict.ByCattax(9, "Rm3SiT")
		_ = lightdict.ByKeyword("CLOUD STORAGE")
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}

func TestChildren(t *testing.T) {
	root := lightdict.ByID(1)
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
	for i := range ids {
		if nodes[i] != ids[i] {
			t.Fatalf("order ids %v nodes %v", ids, nodes)
		}
	}

	n := 0
	for range lightdict.ByID(497).AllChildren {
		n++
	}
	var none *lightdict.Node
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
