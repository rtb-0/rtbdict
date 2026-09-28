package dict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/r0/dict"
)

var sink string

func BenchmarkByID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = dict.ByID(33).R0Code()
	}
}

func BenchmarkByCattax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = dict.ByCattax(9, "Rm3SiT").IABContent31()
	}
}

func BenchmarkByKeyword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = dict.ByKeyword("CLOUD STORAGE").R0Code()
	}
}
