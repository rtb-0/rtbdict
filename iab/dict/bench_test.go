package dict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/iab/dict/ad-product"
	"github.com/rtb-0/rtbdict/iab/dict/content"
	"github.com/rtb-0/rtbdict/iab/dict/mappings"
)

var sinkStr string

func BenchmarkContent31Parent(b *testing.B) {
	tax, ok := content.ByCattax(9)
	if !ok {
		b.Fatal("cattax 9 missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodes := tax.Nodes("151")
		sinkStr = nodes[0].Parent.ID
	}
}

func BenchmarkContent31Children(b *testing.B) {
	tax, ok := content.ByCattax(9)
	if !ok {
		b.Fatal("cattax 9 missing")
	}
	parent := tax.Nodes("151")[0].Parent
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := parent.Children
		sinkStr = ch[0].ID
	}
}

func BenchmarkAdProduct51(b *testing.B) {
	tax, ok := adproduct.ByID("iab.ad-product.1.0")
	if !ok {
		b.Fatal("ad product 1.0 missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodes := tax.Nodes("51")
		sinkStr = nodes[0].Parent.ID
		sinkStr = nodes[1].Parent.ID
	}
}

func BenchmarkMap1002(b *testing.B) {
	m, ok := mappings.ByID("iab.map.ad-product-2.0.content-2.1")
	if !ok {
		b.Fatal("mapping missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStr = m.From("1002")[0].To.ID
	}
}

func BenchmarkMapIAB185(b *testing.B) {
	m, ok := mappings.ByID("iab.map.content-1.0.content-2.1")
	if !ok {
		b.Fatal("content mapping missing")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := m.From("IAB18-5")
		sinkStr = got[len(got)-1].To.ID
	}
}
