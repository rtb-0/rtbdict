package lightdict_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/r0/lightdict"
)

var sink string

func BenchmarkByID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = lightdict.ByID(33).R0Code()
	}
}

func BenchmarkByCattax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = lightdict.ByCattax(9, "Rm3SiT").IABContent31()
	}
}

func BenchmarkByKeyword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = lightdict.ByKeyword("CLOUD STORAGE").R0Code()
	}
}
