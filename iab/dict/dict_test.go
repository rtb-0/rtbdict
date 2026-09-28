package dict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/iab/dict/ad-product"
	"github.com/rtb-0/rtbdict/iab/dict/content"
	"github.com/rtb-0/rtbdict/iab/dict/mappings"
)

func TestContent31Tree(t *testing.T) {
	tax, ok := content.ByCattax(9)
	if !ok {
		t.Fatal("cattax 9 missing")
	}
	byID, ok := content.ByID("iab.content.3.1")
	if !ok || byID != tax {
		t.Fatal("ByID and ByCattax disagree")
	}
	nodes := tax.Nodes("151")
	if len(nodes) != 1 {
		t.Fatalf("151 count %d", len(nodes))
	}
	if nodes[0].Parent == nil || nodes[0].Parent.ID != "150" || nodes[0].Parent.Name != "Attractions" {
		t.Fatalf("151 parent %+v", nodes[0].Parent)
	}
}

func TestAdProductDuplicate51(t *testing.T) {
	tax, ok := adproduct.ByID("iab.ad-product.1.0")
	if !ok {
		t.Fatal("ad product 1.0 missing")
	}
	if tax.Meta.Cattax != 3 {
		t.Fatalf("cattax %d", tax.Meta.Cattax)
	}
	nodes := tax.Nodes("51")
	if len(nodes) != 2 {
		t.Fatalf("51 count %d", len(nodes))
	}
	parents := map[string]string{}
	for _, n := range nodes {
		if n.Parent == nil {
			t.Fatal("51 has no parent")
		}
		parents[n.Parent.ID] = n.Name
	}
	if parents["27"] != "Ticket Services" || parents["54"] != "Coupe" {
		t.Fatalf("parents %v", parents)
	}
}

func TestAlcoholMapping(t *testing.T) {
	m, ok := mappings.ByID("iab.map.ad-product-2.0.content-2.1")
	if !ok {
		t.Fatal("mapping missing")
	}
	same, ok := mappings.ByTaxonomy("iab.ad-product.2.0", "iab.content.2.1")
	if !ok || same != m {
		t.Fatal("ByTaxonomy mismatch")
	}
	entries := m.From("1002")
	if len(entries) != 1 || entries[0].To == nil || entries[0].To.ID != "211" {
		t.Fatalf("1002 -> %+v", entries)
	}
	genre, ok := mappings.ByID("iab.map.ctv-genre.content-3.1")
	if !ok {
		t.Fatal("ctv mapping missing")
	}
	got := genre.From("Action")
	if len(got) != 1 || got[0].To == nil || got[0].To.ID != "325" {
		t.Fatalf("Action -> %+v", got)
	}
}
