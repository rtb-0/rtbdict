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
	if root.Children()[0] != 33 {
		t.Fatalf("children %v", root.Children())
	}
	child := dict.ByR0Code("adl.a")
	if child == nil || child.ID() != 33 || child.Parent() != 1 || child.IABContent31() != "Rm3SiT" {
		t.Fatalf("adl.a id=%d parent=%d c31=%q", child.ID(), child.Parent(), child.IABContent31())
	}
	if dict.ByCattax(9, "Rm3SiT") != child {
		t.Fatal("cattax 9")
	}
	if dict.ByCattax(7, "Rm3SiT").ID() != 33 || child.IABContent30() != "Rm3SiT" {
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
	clothing := dict.ByID(712)
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
	if dict.ByKeyword("skin care").ID() != 1282 || dict.ByKeyword("no such phrase") != nil {
		t.Fatal("keyword collision or miss")
	}
	allocs := testing.AllocsPerRun(100, func() {
		n := dict.ByR0Code("adl.a")
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
