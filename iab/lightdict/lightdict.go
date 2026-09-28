//go:generate go run github.com/rtb-0/rtbdict/iab/cmd/genlight

// Package lightdict holds compact IAB taxonomies and mappings.
// Hierarchy and id slots are int32 tables, which the garbage collector does not scan.
// Each dictionary keeps one string table and one map. Lookups return indexes or subslices.
package lightdict

// Taxonomy is one category tree addressed by node index.
// Cattax is 0 when the source JSON omitted it.
// childOf has one extra slot: children of node i are childAt[childOf[i]:childOf[i+1]].
// at stores start<<16|count into slots. A repeated id, such as Ad Product "51", occupies consecutive slots.
type Taxonomy struct {
	ID      string            `json:"id"`
	Version string            `json:"version,omitempty"`
	Cattax  int               `json:"cattax,omitempty"`
	ids     []string          `json:"-"`
	parent  []int32           `json:"-"`
	childAt []int32           `json:"-"`
	childOf []int32           `json:"-"`
	at      map[string]uint32 `json:"-"`
	slots   []int32           `json:"-"`
}

// NewTaxonomy stores the given tables. The slices and the map are kept as passed.
func NewTaxonomy(id, version string, cattax int, ids []string, parent, childAt, childOf []int32, at map[string]uint32, slots []int32) *Taxonomy {
	n := len(ids)
	if len(parent) != n || len(childOf) != n+1 || len(slots) != n {
		panic("lightdict: taxonomy shape")
	}
	if childOf[0] != 0 || int(childOf[n]) != len(childAt) {
		panic("lightdict: childOf bounds")
	}
	for i := 0; i < n; i++ {
		if childOf[i] > childOf[i+1] || int(childOf[i+1]) > len(childAt) {
			panic("lightdict: childOf order")
		}
		p := parent[i]
		if p < -1 || int(p) >= n || int(p) == i {
			panic("lightdict: parent")
		}
	}
	covered := 0
	for _, packed := range at {
		start, count := span(packed)
		if count == 0 || start+count > len(slots) {
			panic("lightdict: id index")
		}
		covered += count
	}
	if covered != len(slots) {
		panic("lightdict: slot coverage")
	}
	return &Taxonomy{
		ID: id, Version: version, Cattax: cattax,
		ids: ids, parent: parent, childAt: childAt, childOf: childOf,
		at: at, slots: slots,
	}
}

// Indexes returns the node indexes of id. The result is a subslice of the slot table.
// Ad Product 1.0 and 1.1 return two indexes for "51". An unknown id returns nil.
func (t *Taxonomy) Indexes(id string) []int32 {
	if t == nil {
		return nil
	}
	packed, ok := t.at[id]
	if !ok {
		return nil
	}
	start, count := span(packed)
	return t.slots[start : start+count]
}

// Parent returns the parent index of node i, or -1 when i is a root.
func (t *Taxonomy) Parent(i int32) int32 {
	if t == nil {
		return -1
	}
	return t.parent[i]
}

// Children returns the child indexes of node i. The result is a subslice of the child table.
func (t *Taxonomy) Children(i int32) []int32 {
	if t == nil {
		return nil
	}
	start := t.childOf[i]
	end := t.childOf[i+1]
	return t.childAt[start:end]
}

// NodeID returns the id stored for node i.
func (t *Taxonomy) NodeID(i int32) string {
	if t == nil {
		return ""
	}
	return t.ids[i]
}

// Mapping is a key to one or more target ids.
// Targets for one key are adjacent in to. at stores start<<16|count.
// Empty targets from the source are omitted. Genre mappings use the genre name as the key.
type Mapping struct {
	ID string            `json:"id"`
	at map[string]uint32 `json:"-"`
	to []string          `json:"-"`
}

// NewMapping stores the given tables. The slice and the map are kept as passed.
func NewMapping(id string, to []string, at map[string]uint32) *Mapping {
	covered := 0
	for _, packed := range at {
		start, count := span(packed)
		if count == 0 || start+count > len(to) {
			panic("lightdict: mapping span")
		}
		covered += count
	}
	if covered != len(to) {
		panic("lightdict: mapping coverage")
	}
	return &Mapping{ID: id, to: to, at: at}
}

// Targets returns the ids mapped from key. The result is a subslice of the target table.
// A missing key returns nil.
func (m *Mapping) Targets(key string) []string {
	if m == nil {
		return nil
	}
	packed, ok := m.at[key]
	if !ok {
		return nil
	}
	start, count := span(packed)
	return m.to[start : start+count]
}

func span(packed uint32) (start, count int) {
	return int(packed >> 16), int(packed & 0xffff)
}
