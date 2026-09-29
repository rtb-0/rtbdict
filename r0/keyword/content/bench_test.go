package content

import "testing"

var sink int

func BenchmarkMatchName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("Pornography")
	}
}

func BenchmarkMatchUpper(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("PORN VIDEO")
	}
}

func BenchmarkMatchLongest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("adult industry news")
	}
}

func BenchmarkMatchSentence(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("watch adult industry news today")
	}
}

func BenchmarkMatchMiss(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Match("no such phrase")
	}
}
