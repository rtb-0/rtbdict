// Package keyword matches text against r0 category names and keywords.
// The trie is generated. Nodes and edges are integers, so the garbage collector does not scan them.
//
// Matching walks from each word start. Earlier words of a phrase must match exactly, then whitespace.
// The last word may be a prefix of the input word. The longest phrase wins.
package keyword

const foldMax = 256

type node struct {
	id        int
	phraseLen uint16
	edgeOff   uint32
	edgeCount uint16
}

type edge struct {
	b     byte
	child uint32
}

// Match returns the r0 id of the best phrase in text, or 0.
func Match(input string) int {
	if len(nodes) == 0 || input == "" {
		return 0
	}
	if len(input) > foldMax {
		return matchBytes(foldAlloc(input))
	}
	var buf [foldMax]byte
	n := foldInto(input, buf[:])
	return matchBytes(buf[:n])
}

func matchBytes(s []byte) int {
	start, end := 0, len(s)
	for start < end && isSpace(s[start]) {
		start++
	}
	for end > start && isSpace(s[end-1]) {
		end--
	}
	if start >= end {
		return 0
	}
	s = s[start:end]
	var bestID int
	var bestLen uint16
	for i := 0; i < len(s); {
		if (i == 0 || isSpace(s[i-1])) && !isSpace(s[i]) {
			id, plen := matchFrom(s, i)
			if id != 0 && (bestID == 0 || plen > bestLen) {
				bestID = id
				bestLen = plen
			}
		}
		i++
	}
	return bestID
}

func matchFrom(s []byte, pos int) (int, uint16) {
	nodeIdx := uint32(0)
	i := pos
	var bestID int
	var bestLen uint16
	for {
		n := &nodes[nodeIdx]
		if n.id != 0 && (bestID == 0 || n.phraseLen > bestLen) {
			bestID = n.id
			bestLen = n.phraseLen
		}
		if n.edgeCount == 0 || i >= len(s) {
			return bestID, bestLen
		}
		if isSpace(s[i]) {
			child, ok := findChild(n, ' ')
			if !ok {
				return bestID, bestLen
			}
			for i < len(s) && isSpace(s[i]) {
				i++
			}
			nodeIdx = child
			continue
		}
		child, ok := findChild(n, s[i])
		if !ok {
			return bestID, bestLen
		}
		i++
		nodeIdx = child
	}
}

func findChild(n *node, b byte) (uint32, bool) {
	edges := edges[n.edgeOff : n.edgeOff+uint32(n.edgeCount)]
	lo, hi := 0, len(edges)
	for lo < hi {
		mid := (lo + hi) >> 1
		eb := edges[mid].b
		if eb < b {
			lo = mid + 1
		} else if eb > b {
			hi = mid
		} else {
			return edges[mid].child, true
		}
	}
	return 0, false
}

func foldInto(s string, dst []byte) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		dst[i] = c
	}
	return len(s)
}

func foldAlloc(s string) []byte {
	b := make([]byte, len(s))
	foldInto(s, b)
	return b
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
}
