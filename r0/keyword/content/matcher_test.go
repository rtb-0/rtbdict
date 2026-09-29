package content

import "testing"

func TestMatch(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"Pornography", 35},
		{"PORNOGRAPHY", 35},
		{"porn video", 35},
		{"porn videos", 35},
		{"  porn   video  ", 35},
		{"\tporn\tvideo\n", 35},
		{"adult industry news", 38},
		{"Adult Industry News", 38},
		{"sports", 27},
		{"the sports today", 27},
		{"sportss", 27},
		{"sporting event", 27},
		{"watch adult industry news today", 38},
		{"xporn", 0},
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
	copy(in[280:], []byte(" porn video"))
	if got := Match(string(in)); got != 35 {
		t.Fatalf("Match=%d want 35", got)
	}
}

func TestMatchAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_ = Match("PORN VIDEO")
		_ = Match("porn videos")
		_ = Match("adult industry news")
	})
	if allocs != 0 {
		t.Fatalf("allocs %v", allocs)
	}
}
