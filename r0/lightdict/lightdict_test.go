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
	if lightdict.ByKeyword("CLOUD STORAGE").ID() != 497 || lightdict.ByKeyword("skin care").ID() != 1282 {
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
