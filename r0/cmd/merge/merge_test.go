package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestMergeFiles(t *testing.T) {
	root := mustRoot(t)
	content, err := buildContent(root)
	if err != nil {
		t.Fatal(err)
	}
	ad, err := buildAdProduct(root)
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(root, "r0", "datasets", "content", "1.0.json"), content)
	assertFile(t, filepath.Join(root, "r0", "datasets", "adproduct", "1.0.json"), ad)

	checkContent(t, root, content)
	checkAd(t, root, ad)
}

func checkContent(t *testing.T, root string, doc contentDoc) {
	t.Helper()
	versions := []struct {
		file, field string
	}{
		{"iab/datasets/content/1.0.json", "c1_0_code"},
		{"iab/datasets/content/2.0.json", "c2_0_code"},
		{"iab/datasets/content/2.1.json", "c2_1_code"},
		{"iab/datasets/content/2.2.json", "c2_2_code"},
		{"iab/datasets/content/3.0.json", "c3_0_code"},
		{"iab/datasets/content/3.1.json", "c3_1_code"},
	}
	for _, v := range versions {
		tax, err := readTax(root, v.file)
		if err != nil {
			t.Fatal(err)
		}
		if err := sameCounts(taxCounts(tax.Nodes), contentCounts(doc.Nodes, v.field)); err != nil {
			t.Errorf("%s: %v", v.field, err)
		}
	}
	checkShape(t, contentIDs(doc.Nodes))
	names := map[string]bool{}
	var collect func([]contentNode)
	collect = func(ns []contentNode) {
		for _, n := range ns {
			names[n.Name] = true
			collect(n.Children)
		}
	}
	collect(doc.Nodes)
	vectors, err := readTax(root, "iab/datasets/content/3.0-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, rel := range []string{
		"iab/datasets/content/2.2.json",
		"iab/datasets/content/3.0.json",
		"iab/datasets/content/3.1.json",
	} {
		tax, err := readTax(root, rel)
		if err != nil {
			t.Fatal(err)
		}
		for id := range taxIDs(tax.Nodes) {
			known[id] = true
		}
	}
	walkTax(vectors.Nodes, "", func(n taxNode, _ string) {
		if !known[n.ID] && !names[n.Name] {
			t.Errorf("vector category %s %s is missing", n.ID, n.Name)
		}
	})
}

func checkAd(t *testing.T, root string, doc adDoc) {
	t.Helper()
	versions := []struct {
		file, field string
	}{
		{"iab/datasets/adproduct/1.0.json", "ap1_0_code"},
		{"iab/datasets/adproduct/1.1.json", "ap1_1_code"},
		{"iab/datasets/adproduct/2.0.json", "ap2_0_code"},
	}
	for _, v := range versions {
		tax, err := readTax(root, v.file)
		if err != nil {
			t.Fatal(err)
		}
		if err := sameCounts(taxCounts(tax.Nodes), adCounts(doc.Nodes, v.field)); err != nil {
			t.Errorf("%s: %v", v.field, err)
		}
	}
	checkShape(t, adIDs(doc.Nodes))
	var invented []string
	for _, n := range doc.Nodes {
		if n.AP20 == "" {
			invented = append(invented, n.Name)
		}
	}
	if len(invented) != 1 || invented[0] != "Hobbies and Interests" {
		t.Errorf("new roots = %v", invented)
	}
}

type shaped struct {
	id       int
	r0       string
	parent   string
	parentID int
	depth    int
}

func checkShape(t *testing.T, nodes []shaped) {
	t.Helper()
	rootRe := regexp.MustCompile(`^[a-z]{1,3}[0-9]*$`)
	childRe := regexp.MustCompile(`^[a-z][0-9]*$`)
	seenCode := map[string]bool{}
	seenID := map[int]bool{}
	byDepth := map[int][]int{}
	for _, n := range nodes {
		if seenCode[n.r0] {
			t.Errorf("duplicate r0_code %s", n.r0)
		}
		seenCode[n.r0] = true
		if seenID[n.id] {
			t.Errorf("duplicate id %d", n.id)
		}
		seenID[n.id] = true
		byDepth[n.depth] = append(byDepth[n.depth], n.id)
		if n.parent == "" {
			if !rootRe.MatchString(n.r0) {
				t.Errorf("root code %s", n.r0)
			}
			continue
		}
		if len(n.r0) <= len(n.parent)+1 || n.r0[:len(n.parent)+1] != n.parent+"." {
			t.Errorf("code %s is not under %s", n.r0, n.parent)
			continue
		}
		if !childRe.MatchString(n.r0[len(n.parent)+1:]) {
			t.Errorf("child segment %s", n.r0)
		}
		if n.id <= n.parentID {
			t.Errorf("id %d is not after parent %d", n.id, n.parentID)
		}
	}
	if len(seenID) != len(nodes) {
		t.Fatalf("id count %d nodes %d", len(seenID), len(nodes))
	}
	prevMax := 0
	for depth := 0; ; depth++ {
		ids := byDepth[depth]
		if len(ids) == 0 {
			break
		}
		minID, maxID := ids[0], ids[0]
		for _, id := range ids[1:] {
			if id < minID {
				minID = id
			}
			if id > maxID {
				maxID = id
			}
		}
		if minID != prevMax+1 || maxID-minID+1 != len(ids) {
			t.Errorf("depth %d ids %d..%d count %d, previous max %d", depth, minID, maxID, len(ids), prevMax)
		}
		prevMax = maxID
	}
	if prevMax != len(nodes) {
		t.Errorf("last id %d, nodes %d", prevMax, len(nodes))
	}
}

func taxCounts(nodes []taxNode) map[string]int {
	out := map[string]int{}
	walkTax(nodes, "", func(n taxNode, _ string) { out[n.ID]++ })
	return out
}

func contentCounts(nodes []contentNode, field string) map[string]int {
	out := map[string]int{}
	var walk func([]contentNode)
	walk = func(ns []contentNode) {
		for _, n := range ns {
			if v := contentField(n, field); v != "" {
				out[v]++
			}
			walk(n.Children)
		}
	}
	walk(nodes)
	return out
}

func contentField(n contentNode, field string) string {
	switch field {
	case "c1_0_code":
		return string(n.C10)
	case "c2_0_code":
		return string(n.C20)
	case "c2_1_code":
		return string(n.C21)
	case "c2_2_code":
		return string(n.C22)
	case "c3_0_code":
		return string(n.C30)
	case "c3_1_code":
		return string(n.C31)
	default:
		return ""
	}
}

func adCounts(nodes []adNode, field string) map[string]int {
	out := map[string]int{}
	var walk func([]adNode)
	walk = func(ns []adNode) {
		for _, n := range ns {
			var v string
			switch field {
			case "ap1_0_code":
				v = string(n.AP10)
			case "ap1_1_code":
				v = string(n.AP11)
			case "ap2_0_code":
				v = string(n.AP20)
			}
			if v != "" {
				out[v]++
			}
			walk(n.Children)
		}
	}
	walk(nodes)
	return out
}

func contentIDs(nodes []contentNode) []shaped {
	var out []shaped
	var walk func([]contentNode, string, int, int)
	walk = func(ns []contentNode, parent string, parentID, depth int) {
		for _, n := range ns {
			out = append(out, shaped{id: n.ID, r0: n.R0, parent: parent, parentID: parentID, depth: depth})
			walk(n.Children, n.R0, n.ID, depth+1)
		}
	}
	walk(nodes, "", 0, 0)
	return out
}

func adIDs(nodes []adNode) []shaped {
	var out []shaped
	var walk func([]adNode, string, int, int)
	walk = func(ns []adNode, parent string, parentID, depth int) {
		for _, n := range ns {
			out = append(out, shaped{id: n.ID, r0: n.R0, parent: parent, parentID: parentID, depth: depth})
			walk(n.Children, n.R0, n.ID, depth+1)
		}
	}
	walk(nodes, "", 0, 0)
	return out
}

func sameCounts(want, got map[string]int) error {
	for id, n := range want {
		if got[id] != n {
			return errf("%s want %d got %d", id, n, got[id])
		}
	}
	for id, n := range got {
		if want[id] != n {
			return errf("%s extra %d", id, n)
		}
	}
	return nil
}

func assertFile(t *testing.T, path string, v any) {
	t.Helper()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(body, buf.Bytes()) {
		t.Fatalf("%s does not match the builder output", path)
	}
}

func mustRoot(t *testing.T) string {
	t.Helper()
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}
