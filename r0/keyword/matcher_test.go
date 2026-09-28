package keyword

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"Cloud Storage", 497},
		{"CLOUD STORAGE", 497},
		{"hosted file synchronization", 497},
		{"hosted file synchronizations", 497},
		{"remote file retention", 497},
		{"  Cloud   Storage  ", 497},
		{"\tcloud\tstorage\n", 497},
		{"skin care", 1282},
		{"Skin Care", 1282},
		{"sports", 27},
		{"the sports today", 27},
		{"sportss", 27},
		{"sports memorabilia and trading cards", 911},
		{"sports  memorabilia   and trading cards", 911},
		{"SPORTS MEMORABILIA AND TRADING CARDS", 911},
		{"watch sports and sports memorabilia and trading cards", 911},
		{"sporting goods", 455},
		{"sporting goods stores", 1030},
		{"xsports", 0},
		{"sporting", 0},
		{"", 0},
		{"   ", 0},
		{"no such phrase", 0},
	}
	for _, tc := range cases {
		if got := Match(tc.in); got != tc.want {
			t.Errorf("Match(%q)=%d want %d", tc.in, got, tc.want)
		}
	}
}

func TestMatchLongInput(t *testing.T) {
	in := make([]byte, 300)
	for i := range in {
		in[i] = 'x'
	}
	copy(in[280:], []byte(" cloud storage"))
	if got := Match(string(in)); got != 497 {
		t.Fatalf("Match=%d want 497", got)
	}
}

func TestMatchAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_ = Match("CLOUD STORAGE")
		_ = Match("hosted file synchronization")
		_ = Match("sports memorabilia and trading cards")
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}
