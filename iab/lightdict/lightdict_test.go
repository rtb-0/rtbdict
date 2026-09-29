package lightdict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/iab/lightdict/adproduct"
	"github.com/rtb-0/rtbdict/iab/lightdict/audience"
	"github.com/rtb-0/rtbdict/iab/lightdict/content"
	"github.com/rtb-0/rtbdict/iab/lightdict/mappings"
)

func TestAudienceCattax0(t *testing.T) {
	tax := audience.ByCattax(0)
	if tax == nil || tax.ID != "iab.audience.1.0" {
		t.Fatalf("cattax 0 -> %v", tax)
	}
	if content.ByCattax(0) != nil {
		t.Fatal("content cattax 0")
	}
}

func TestContent31Tree(t *testing.T) {
	tax := content.ByCattax(9)
	if tax == nil {
		t.Fatal("cattax 9 missing")
	}
	byID := content.ByID("iab.content.3.1")
	if byID == nil || byID != tax {
		t.Fatal("ByID and ByCattax disagree")
	}
	if tax.Cattax != 9 || tax.Version != "3.1" {
		t.Fatalf("meta %+v", tax)
	}
	idx := tax.Indexes("151")
	if len(idx) != 1 {
		t.Fatalf("151 count %d", len(idx))
	}
	parent := tax.Parent(idx[0])
	if parent < 0 || tax.NodeID(parent) != "150" {
		t.Fatalf("151 parent %d %q", parent, tax.NodeID(parent))
	}
	children := tax.Children(parent)
	found := false
	for _, c := range children {
		if c == idx[0] {
			found = true
		}
	}
	if !found {
		t.Fatal("150 children do not include 151")
	}
	allocs := testing.AllocsPerRun(100, func() {
		got := tax.Indexes("151")
		p := tax.Parent(got[0])
		_ = tax.NodeID(p)
		_ = tax.Children(p)
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}

func TestAdProductDuplicate51(t *testing.T) {
	tax := adproduct.ByID("iab.ad-product.1.0")
	if tax == nil {
		t.Fatal("ad product 1.0 missing")
	}
	if tax.Cattax != 3 {
		t.Fatalf("cattax %d", tax.Cattax)
	}
	idx := tax.Indexes("51")
	if len(idx) != 2 {
		t.Fatalf("51 count %d", len(idx))
	}
	parents := map[string]bool{}
	for _, i := range idx {
		p := tax.Parent(i)
		if p < 0 {
			t.Fatal("51 has no parent")
		}
		parents[tax.NodeID(p)] = true
	}
	if !parents["27"] || !parents["54"] {
		t.Fatalf("parents %v", parents)
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = tax.Indexes("51")
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}

func TestMappings(t *testing.T) {
	m := mappings.ByID("iab.map.ad-product-2.0.content-2.1")
	if m == nil {
		t.Fatal("mapping missing")
	}
	same := mappings.ByTaxonomy("iab.ad-product.2.0", "iab.content.2.1")
	if same == nil || same != m {
		t.Fatal("ByTaxonomy mismatch")
	}
	got := m.Targets("1002")
	if len(got) != 1 || got[0] != "211" {
		t.Fatalf("1002 -> %v", got)
	}
	wide := mappings.ByID("iab.map.content-1.0.content-2.1")
	if wide == nil {
		t.Fatal("content mapping missing")
	}
	clothing := wide.Targets("IAB18-5")
	want := []string{"566", "582", "576", "575"}
	if len(clothing) != len(want) {
		t.Fatalf("IAB18-5 -> %v", clothing)
	}
	for i := range want {
		if clothing[i] != want[i] {
			t.Fatalf("IAB18-5 -> %v", clothing)
		}
	}
	genre := mappings.ByID("iab.map.ctv-genre.content-3.1")
	if genre == nil {
		t.Fatal("ctv mapping missing")
	}
	action := genre.Targets("Action")
	if len(action) != 1 || action[0] != "325" {
		t.Fatalf("Action -> %v", action)
	}
	if mappings.ByTaxonomy("", "iab.content.3.1") != nil {
		t.Fatal("ctv and podcast share an empty source")
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = m.Targets("1002")
		_ = wide.Targets("IAB18-5")
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}
