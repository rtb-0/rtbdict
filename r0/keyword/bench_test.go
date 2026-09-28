package keyword

import "testing"

var sink int

func BenchmarkMatchName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("Cloud Storage")
	}
}

func BenchmarkMatchUpper(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("CLOUD STORAGE")
	}
}

func BenchmarkMatchLongest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("sports memorabilia and trading cards")
	}
}

func BenchmarkMatchSentence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("watch sports and sports memorabilia and trading cards")
	}
}

func BenchmarkMatchMiss(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("no such phrase")
	}
}
