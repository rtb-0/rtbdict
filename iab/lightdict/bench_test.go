package lightdict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/iab/lightdict/ad-product"
	"github.com/rtb-0/rtbdict/iab/lightdict/content"
	"github.com/rtb-0/rtbdict/iab/lightdict/mappings"
)

var sinkStr string

func BenchmarkContent31Parent(b *testing.B) {
	tax := content.ByCattax(9)
	if tax == nil {
		b.Fatal("cattax 9 missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := tax.Indexes("151")
		sinkStr = tax.NodeID(tax.Parent(idx[0]))
	}
}

func BenchmarkContent31Children(b *testing.B) {
	tax := content.ByCattax(9)
	if tax == nil {
		b.Fatal("cattax 9 missing")
	}
	parent := tax.Parent(tax.Indexes("151")[0])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := tax.Children(parent)
		sinkStr = tax.NodeID(ch[0])
	}
}

func BenchmarkAdProduct51(b *testing.B) {
	tax := adproduct.ByID("iab.ad-product.1.0")
	if tax == nil {
		b.Fatal("ad product 1.0 missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := tax.Indexes("51")
		sinkStr = tax.NodeID(tax.Parent(idx[0]))
		sinkStr = tax.NodeID(tax.Parent(idx[1]))
	}
}

func BenchmarkMap1002(b *testing.B) {
	m := mappings.ByID("iab.map.ad-product-2.0.content-2.1")
	if m == nil {
		b.Fatal("mapping missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = m.Targets("1002")[0]
	}
}

func BenchmarkMapIAB185(b *testing.B) {
	m := mappings.ByID("iab.map.content-1.0.content-2.1")
	if m == nil {
		b.Fatal("content mapping missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := m.Targets("IAB18-5")
		sinkStr = got[len(got)-1]
	}
}
