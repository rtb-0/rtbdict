package content_test

import (
	"testing"

	"github.com/rtb-0/rtbdict/r0/lightdict/content"
)

var sink string

func BenchmarkByID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = content.ByID(35).R0Code()
	}
}

func BenchmarkByCattax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = content.ByCattax(9, "Rm3SiT").R0Code()
	}
}

func BenchmarkByKeyword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = content.ByKeyword("PORN VIDEO").R0Code()
	}
}

func BenchmarkChildrenNodes(b *testing.B) {
	root := content.ByID(1)
	for i := 0; i < b.N; i++ {
		for child := range root.ChildrenNodes {
			sinkID = child.ID()
		}
	}
}

func BenchmarkAllChildren(b *testing.B) {
	root := content.ByID(1)
	for i := 0; i < b.N; i++ {
		for id := range root.AllChildren {
			sinkID = id
		}
	}
}

func BenchmarkAllChildrenNodes(b *testing.B) {
	root := content.ByID(1)
	for i := 0; i < b.N; i++ {
		for child := range root.AllChildrenNodes {
			sinkID = child.ID()
		}
	}
}

var sinkID int
