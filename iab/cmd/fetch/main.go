// Command fetch downloads IAB Tech Lab taxonomy TSV files and writes JSON datasets.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	publisher    = "IAB Tech Lab"
	pageAd       = "https://iabtechlab.com/standards/ad-product-taxonomy/"
	pageAudience = "https://iabtechlab.com/standards/audience-taxonomy/"
	pageCont     = "https://iabtechlab.com/standards/content-taxonomy/"
	repoBlob     = "https://github.com/InteractiveAdvertisingBureau/Taxonomies/blob/main/"
	repoRaw      = "https://raw.githubusercontent.com/InteractiveAdvertisingBureau/Taxonomies/main/"
)

func main() {
	outDir := "iab/datasets"
	if len(os.Args) == 3 && os.Args[1] == "-out" {
		outDir = os.Args[2]
	} else if len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: fetch [-out dir]\n")
		os.Exit(2)
	}

	var failed bool
	for _, spec := range taxonomies {
		doc, warnings, err := buildTaxonomy(spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", spec.file, err)
			failed = true
			continue
		}
		if err := writeJSON(filepath.Join(outDir, spec.file), doc); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", spec.file, err)
			failed = true
			continue
		}
		fmt.Printf("%s nodes=%d\n", spec.file, countNodes(doc.Nodes))
		for _, w := range warnings {
			fmt.Printf("  warning: %s\n", w)
		}
	}
	for _, spec := range mappings {
		doc, err := buildMapping(spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", spec.file, err)
			failed = true
			continue
		}
		if err := writeJSON(filepath.Join(outDir, spec.file), doc); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", spec.file, err)
			failed = true
			continue
		}
		fmt.Printf("%s entries=%d\n", spec.file, len(doc.Entries))
	}
	if failed {
		os.Exit(1)
	}
}

type meta struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Version    string `json:"version,omitempty"`
	Cattax     *int   `json:"cattax,omitempty"`
	Deprecated bool   `json:"deprecated,omitempty"`
	From       string `json:"from,omitempty"`
	FromCattax *int   `json:"from_cattax,omitempty"`
	To         string `json:"to,omitempty"`
	ToCattax   *int   `json:"to_cattax,omitempty"`
	Publisher  string `json:"publisher"`
	Page       string `json:"page"`
	Source     string `json:"source"`
	Released   string `json:"released,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

type node struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Extension string `json:"extension,omitempty"`
	Children  []node `json:"children,omitempty"`
}

// flatNode is the row order from a TSV, before nodes are nested.
type flatNode struct {
	ID        string
	Parent    string
	Name      string
	Path      []string
	Extension string
}

type ref struct {
	ID   string   `json:"id,omitempty"`
	Name string   `json:"name,omitempty"`
	Path []string `json:"path,omitempty"`
}

type entry struct {
	From ref    `json:"from"`
	To   *ref   `json:"to,omitempty"`
	Note string `json:"note,omitempty"`
}

type taxonomyDoc struct {
	Meta  meta   `json:"meta"`
	Nodes []node `json:"nodes"`
}

type mappingDoc struct {
	Meta    meta    `json:"meta"`
	Entries []entry `json:"entries"`
}

type taxSpec struct {
	file       string
	upstream   string
	id         string
	name       string
	version    string
	cattax     int
	hasCattax  bool
	deprecated bool
	page       string
	released   string
	content10  bool
	// allowDup is an upstream id that appears twice. Both rows are kept.
	allowDup string
}

type colMap struct {
	header    string
	fromID    int
	fromName  int
	fromTiers []int
	toID      int
	toName    int
	toTiers   []int
	note      int
}

type mapSpec struct {
	file       string
	upstream   string
	id         string
	name       string
	from       string
	fromCattax int
	hasFromTax bool
	to         string
	toCattax   int
	hasToTax   bool
	page       string
	released   string
	notes      string
	cols       *colMap
	genre      bool
	// content20File is the mislabeled Content 2.0 to Content 2.1 TSV.
	content20File bool
}

func cattaxPtr(n int, ok bool) *int {
	if !ok {
		return nil
	}
	v := n
	return &v
}

func (s taxSpec) toMeta() meta {
	return meta{
		ID:         s.id,
		Kind:       "taxonomy",
		Name:       s.name,
		Version:    s.version,
		Cattax:     cattaxPtr(s.cattax, s.hasCattax),
		Deprecated: s.deprecated,
		Publisher:  publisher,
		Page:       s.page,
		Source:     blobURL(s.upstream),
		Released:   s.released,
	}
}

func (s mapSpec) toMeta() meta {
	return meta{
		ID:         s.id,
		Kind:       "mapping",
		Name:       s.name,
		Publisher:  publisher,
		Page:       s.page,
		Source:     blobURL(s.upstream),
		Released:   s.released,
		From:       s.from,
		FromCattax: cattaxPtr(s.fromCattax, s.hasFromTax),
		To:         s.to,
		ToCattax:   cattaxPtr(s.toCattax, s.hasToTax),
		Notes:      s.notes,
	}
}

func blobURL(rel string) string { return repoURL(repoBlob, rel) }
func rawURL(rel string) string  { return repoURL(repoRaw, rel) }

func repoURL(base, rel string) string {
	parts := strings.Split(rel, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return base + strings.Join(parts, "/")
}

func buildTaxonomy(spec taxSpec) (taxonomyDoc, []string, error) {
	body, err := fetch(rawURL(spec.upstream))
	if err != nil {
		return taxonomyDoc{}, nil, err
	}
	rows := parseTSV(body)
	var flats []flatNode
	var warnings []string
	if spec.content10 {
		flats, err = parseContent10(rows)
	} else {
		flats, warnings, err = parseTiered(rows, spec.allowDup)
	}
	if err != nil {
		return taxonomyDoc{}, nil, err
	}
	roots, nestWarnings, err := nest(flats)
	if err != nil {
		return taxonomyDoc{}, nil, err
	}
	return taxonomyDoc{Meta: spec.toMeta(), Nodes: roots}, append(warnings, nestWarnings...), nil
}

func parseContent10(rows [][]string) ([]flatNode, error) {
	header, data := splitHeader(rows, "IAB Code")
	if header == nil {
		return nil, fmt.Errorf("missing IAB Code header")
	}
	codeCol := indexOf(header, "IAB Code")
	tierCol := indexOf(header, "Tier")
	nameCol := indexOf(header, "IAB Category")
	if codeCol < 0 || tierCol < 0 || nameCol < 0 {
		return nil, fmt.Errorf("unexpected Content 1.0 columns: %q", header)
	}

	names := map[string]string{}
	var draft []flatNode
	var tiers []int
	for _, row := range data {
		if blank(row) {
			continue
		}
		id := cell(row, codeCol)
		name := cell(row, nameCol)
		tier, err := parseTierLabel(cell(row, tierCol))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		names[id] = name
		tiers = append(tiers, tier)
		draft = append(draft, flatNode{ID: id, Name: name})
	}
	for i := range draft {
		n := &draft[i]
		if parent, ok := content10Parent(n.ID); ok {
			n.Parent = parent
			n.Path = []string{names[parent], n.Name}
		} else {
			n.Path = []string{n.Name}
		}
		if len(n.Path) != tiers[i] {
			return nil, fmt.Errorf("%s: tier %d does not match path %v", n.ID, tiers[i], n.Path)
		}
	}
	if len(draft) == 0 {
		return nil, fmt.Errorf("no nodes")
	}
	return draft, nil
}

func content10Parent(code string) (string, bool) {
	i := strings.LastIndex(code, "-")
	if i <= 0 {
		return "", false
	}
	return code[:i], true
}

func parseTierLabel(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "Tier %d", &n); err != nil || n <= 0 {
		return 0, fmt.Errorf("bad tier %q", s)
	}
	return n, nil
}

func parseTiered(rows [][]string, allowDup string) ([]flatNode, []string, error) {
	header, data := headerWith(rows, "Unique ID")
	if header == nil {
		return nil, nil, fmt.Errorf("missing Unique ID header")
	}
	idCol := indexOf(header, "Unique ID")
	nameCol := indexOf(header, "Name")
	parentCol := indexOf(header, "Parent")
	if parentCol < 0 {
		parentCol = indexOf(header, "Parent ID")
	}
	var tierCols []int
	for _, label := range []string{"Tier 1", "Tier 2", "Tier 3", "Tier 4", "Tier 5", "Tier 6"} {
		if i := indexOf(header, label); i >= 0 {
			tierCols = append(tierCols, i)
		}
	}
	extCol := -1
	for i, h := range header {
		if strings.Contains(h, "Extension") || strings.Contains(h, "Note") {
			extCol = i
			break
		}
	}
	if idCol < 0 || parentCol < 0 || len(tierCols) == 0 {
		return nil, nil, fmt.Errorf("unexpected columns: %q", header)
	}

	var nodes []flatNode
	seen := map[string]int{}
	var warnings []string
	for _, row := range data {
		if blank(row) {
			continue
		}
		id := cell(row, idCol)
		if id == "" {
			return nil, nil, fmt.Errorf("empty id in row %q", row)
		}
		path := pathOf(row, tierCols)
		name := cell(row, nameCol)
		if name == "" && len(path) > 0 {
			name = path[len(path)-1]
		}
		if len(path) == 0 && name != "" {
			path = []string{name}
		}
		nodes = append(nodes, flatNode{
			ID:        id,
			Parent:    cell(row, parentCol),
			Name:      name,
			Path:      path,
			Extension: cell(row, extCol),
		})
		seen[id]++
	}
	var dups []string
	for id, n := range seen {
		if n > 1 && id != allowDup {
			dups = append(dups, id)
		}
	}
	if len(dups) > 0 {
		return nil, nil, fmt.Errorf("duplicate ids: %s", strings.Join(dups, ", "))
	}
	if allowDup != "" && seen[allowDup] == 2 {
		warnings = append(warnings, "upstream duplicate id "+allowDup+" kept as two nodes")
	} else if allowDup != "" && seen[allowDup] != 2 {
		return nil, nil, fmt.Errorf("expected id %s twice, found %d", allowDup, seen[allowDup])
	}
	if len(nodes) == 0 {
		return nil, nil, fmt.Errorf("no nodes")
	}
	return nodes, warnings, nil
}

func nest(flats []flatNode) ([]node, []string, error) {
	byID := map[string][]int{}
	byPath := map[string][]int{}
	for i, n := range flats {
		if n.ID == "" || n.Name == "" || len(n.Path) == 0 {
			return nil, nil, fmt.Errorf("bad node id=%q name=%q", n.ID, n.Name)
		}
		byID[n.ID] = append(byID[n.ID], i)
		byPath[pathKey(n.Path)] = append(byPath[pathKey(n.Path)], i)
	}

	parentOf := make([]int, len(flats))
	for i := range parentOf {
		parentOf[i] = -1
	}
	placedByPath := 0
	for i, n := range flats {
		if n.Parent != "" && n.Parent != n.ID {
			cands := byID[n.Parent]
			if len(cands) == 0 {
				return nil, nil, fmt.Errorf("%s parent %s is missing", n.ID, n.Parent)
			}
			if len(cands) != 1 {
				return nil, nil, fmt.Errorf("%s parent %s is ambiguous", n.ID, n.Parent)
			}
			parentOf[i] = cands[0]
			continue
		}
		if len(n.Path) <= 1 {
			continue
		}
		cands := byPath[pathKey(n.Path[:len(n.Path)-1])]
		if len(cands) != 1 {
			return nil, nil, fmt.Errorf("%s path %q has %d parents", n.ID, strings.Join(n.Path, " / "), len(cands))
		}
		parentOf[i] = cands[0]
		placedByPath++
	}
	for i := range flats {
		seen := map[int]bool{}
		for p := parentOf[i]; p >= 0; p = parentOf[p] {
			if p == i || seen[p] {
				return nil, nil, fmt.Errorf("cycle at %s", flats[i].ID)
			}
			seen[p] = true
		}
	}

	kids := make([][]int, len(flats))
	var roots []int
	for i, p := range parentOf {
		if p < 0 {
			roots = append(roots, i)
			continue
		}
		kids[p] = append(kids[p], i)
	}

	var build func(int) node
	build = func(i int) node {
		n := node{ID: flats[i].ID, Name: flats[i].Name, Extension: flats[i].Extension}
		for _, c := range kids[i] {
			n.Children = append(n.Children, build(c))
		}
		return n
	}
	out := make([]node, len(roots))
	total := 0
	var walk func(node) int
	walk = func(n node) int {
		c := 1
		for _, ch := range n.Children {
			c += walk(ch)
		}
		return c
	}
	for i, r := range roots {
		out[i] = build(r)
		total += walk(out[i])
	}
	if total != len(flats) {
		return nil, nil, fmt.Errorf("tree has %d nodes, source has %d", total, len(flats))
	}
	var warnings []string
	if placedByPath > 0 {
		warnings = append(warnings, fmt.Sprintf("placed %d nodes by path prefix", placedByPath))
	}
	return out, warnings, nil
}

func pathKey(path []string) string {
	return strings.Join(path, "\x1f")
}

func countNodes(nodes []node) int {
	n := 0
	for _, node := range nodes {
		n += 1 + countNodes(node.Children)
	}
	return n
}

func buildMapping(spec mapSpec) (mappingDoc, error) {
	body, err := fetch(rawURL(spec.upstream))
	if err != nil {
		return mappingDoc{}, err
	}
	rows := parseTSV(body)
	var entries []entry
	if spec.content20File {
		entries, err = parseContent20File(rows)
	} else if spec.genre {
		entries, err = parseByColumns(rows, spec.cols, true)
	} else {
		entries, err = parseByColumns(rows, spec.cols, false)
	}
	if err != nil {
		return mappingDoc{}, err
	}
	return mappingDoc{Meta: spec.toMeta(), Entries: entries}, nil
}

func parseByColumns(rows [][]string, cols *colMap, genre bool) ([]entry, error) {
	_, data := splitHeader(rows, cols.header)
	if data == nil {
		return nil, fmt.Errorf("missing header %q", cols.header)
	}
	var entries []entry
	dataRows := 0
	for _, row := range data {
		if blank(row) {
			continue
		}
		dataRows++
		from := ref{
			ID:   col(row, cols.fromID),
			Name: col(row, cols.fromName),
			Path: pathOf(row, cols.fromTiers),
		}
		if from.Name == "" && len(from.Path) > 0 {
			from.Name = from.Path[len(from.Path)-1]
		}
		to := &ref{
			ID:   col(row, cols.toID),
			Name: col(row, cols.toName),
			Path: pathOf(row, cols.toTiers),
		}
		if to.Name == "" && len(to.Path) > 0 {
			to.Name = to.Path[len(to.Path)-1]
		}
		e := entry{From: from, Note: col(row, cols.note)}
		if to.ID != "" || to.Name != "" {
			e.To = to
		}
		if genre && from.Name == "" {
			return nil, fmt.Errorf("genre row without a name: %q", row)
		}
		if !genre && from.ID == "" && from.Name == "" {
			return nil, fmt.Errorf("mapping row without a source: %q", row)
		}
		entries = append(entries, e)
	}
	if len(entries) != dataRows {
		return nil, fmt.Errorf("entries %d != data rows %d", len(entries), dataRows)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries")
	}
	return entries, nil
}

func parseContent20File(rows [][]string) ([]entry, error) {
	_, data := splitHeader(rows, "New Taxonomy ID")
	if data == nil {
		return nil, fmt.Errorf("missing New Taxonomy ID header")
	}
	var mapped []entry
	var unmapped []entry
	dataRows := 0
	for _, row := range data {
		if blank(row) {
			continue
		}
		dataRows++
		fromID := cell(row, 1)
		toID := cell(row, 0)
		if fromID == "" && toID == "" {
			return nil, fmt.Errorf("row without a mapping: %q", row)
		}
		mapped = append(mapped, entry{
			From: ref{ID: fromID, Name: cell(row, 2)},
			To:   &ref{ID: toID, Name: cell(row, 3)},
		})
		if id := cell(row, 7); id != "" {
			unmapped = append(unmapped, entry{
				From: ref{ID: id, Name: cell(row, 6)},
				Note: "NOT MAPPED (deprecated)",
			})
		}
	}
	if len(mapped) != dataRows {
		return nil, fmt.Errorf("mapped entries %d != data rows %d", len(mapped), dataRows)
	}
	if len(mapped) == 0 {
		return nil, fmt.Errorf("no entries")
	}
	return append(mapped, unmapped...), nil
}

func splitHeader(rows [][]string, first string) (header []string, data [][]string) {
	for i, row := range rows {
		if cell(row, 0) == first {
			return row, rows[i+1:]
		}
	}
	return nil, nil
}

func headerWith(rows [][]string, name string) (header []string, data [][]string) {
	for i, row := range rows {
		if indexOf(row, name) >= 0 {
			return row, rows[i+1:]
		}
	}
	return nil, nil
}

func indexOf(header []string, name string) int {
	for i, h := range header {
		if strings.TrimSpace(h) == name {
			return i
		}
	}
	return -1
}

func pathOf(row []string, cols []int) []string {
	var path []string
	for _, i := range cols {
		if v := cell(row, i); v != "" {
			path = append(path, v)
		}
	}
	return path
}

func col(row []string, i int) string {
	if i < 0 {
		return ""
	}
	return cell(row, i)
}

func cell(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func blank(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func parseTSV(body string) [][]string {
	body = strings.TrimPrefix(body, "\uFEFF")
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.TrimSuffix(body, "\n")
	lines := strings.Split(body, "\n")
	rows := make([][]string, len(lines))
	for i, line := range lines {
		rows[i] = strings.Split(line, "\t")
	}
	return rows
}

func fetch(raw string) (string, error) {
	resp, err := http.Get(raw)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", raw, resp.Status)
	}
	return string(b), nil
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

var taxonomies = []taxSpec{
	{file: "ad-product/1.0.json", upstream: "Ad Product Taxonomies/Ad Product Taxonomy 1.0.tsv", id: "iab.ad-product.1.0", name: "Ad Product Taxonomy", version: "1.0", cattax: 3, hasCattax: true, page: pageAd, released: "2022-07", allowDup: "51"},
	{file: "ad-product/1.1.json", upstream: "Ad Product Taxonomies/Ad Product Taxonomy 1.1.tsv", id: "iab.ad-product.1.1", name: "Ad Product Taxonomy", version: "1.1", page: pageAd, released: "2022-08", allowDup: "51"},
	{file: "ad-product/2.0.json", upstream: "Ad Product Taxonomies/Ad Product Taxonomy 2.0.tsv", id: "iab.ad-product.2.0", name: "Ad Product Taxonomy", version: "2.0", cattax: 8, hasCattax: true, page: pageAd, released: "2024-11"},
	{file: "audience/1.0.json", upstream: "Audience Taxonomies/Audience Taxonomy 1.0.tsv", id: "iab.audience.1.0", name: "Audience Taxonomy", version: "1.0", page: pageAudience, released: "2018-05"},
	{file: "audience/1.1.json", upstream: "Audience Taxonomies/Audience Taxonomy 1.1.tsv", id: "iab.audience.1.1", name: "Audience Taxonomy", version: "1.1", cattax: 4, hasCattax: true, page: pageAudience, released: "2020-10"},
	{file: "content/1.0.json", upstream: "Content Taxonomies/Content Taxonomy 1.0.tsv", id: "iab.content.1.0", name: "Content Taxonomy", version: "1.0", cattax: 1, hasCattax: true, deprecated: true, page: pageCont, content10: true},
	{file: "content/2.0.json", upstream: "Content Taxonomies/Content Taxonomy 2.0.tsv", id: "iab.content.2.0", name: "Content Taxonomy", version: "2.0", cattax: 2, hasCattax: true, page: pageCont, released: "2017-11"},
	{file: "content/2.1.json", upstream: "Content Taxonomies/Content Taxonomy 2.1.tsv", id: "iab.content.2.1", name: "Content Taxonomy", version: "2.1", cattax: 5, hasCattax: true, page: pageCont, released: "2020-10"},
	{file: "content/2.2.json", upstream: "Content Taxonomies/Content Taxonomy 2.2.tsv", id: "iab.content.2.2", name: "Content Taxonomy", version: "2.2", cattax: 6, hasCattax: true, page: pageCont, released: "2020-12"},
	{file: "content/3.0.json", upstream: "Content Taxonomies/Content Taxonomy 3.0.tsv", id: "iab.content.3.0", name: "Content Taxonomy", version: "3.0", cattax: 7, hasCattax: true, page: pageCont, released: "2022-06"},
	{file: "content/3.0-vectors.json", upstream: "Content Taxonomies/Content Taxonomy 3.0 Descriptive Vectors.tsv", id: "iab.content.3.0-vectors", name: "Content Taxonomy Descriptive Vectors", version: "3.0", page: pageCont, released: "2022-06"},
	{file: "content/3.1.json", upstream: "Content Taxonomies/Content Taxonomy 3.1.tsv", id: "iab.content.3.1", name: "Content Taxonomy", version: "3.1", cattax: 9, hasCattax: true, page: pageCont, released: "2024-12"},
}

var mappings = []mapSpec{
	{
		file: "mappings/ad-product-2.0-to-ad-product-1.1.json", upstream: "Taxonomy Mappings/Ad Product 2.0 to 1.1.tsv",
		id: "iab.map.ad-product-2.0.ad-product-1.1", name: "Ad Product 2.0 to Ad Product 1.1",
		from: "iab.ad-product.2.0", fromCattax: 8, hasFromTax: true,
		to: "iab.ad-product.1.1", page: pageAd, released: "2023-11",
		cols: &colMap{header: "AP2 Name", fromID: 1, fromName: 0, fromTiers: []int{3, 4, 5}, toID: 6, toName: 8, note: -1},
	},
	{
		file: "mappings/ad-product-2.0-to-content-1.0.json", upstream: "Taxonomy Mappings/Ad Product 2.0 to Content 1.0.tsv",
		id: "iab.map.ad-product-2.0.content-1.0", name: "Ad Product 2.0 to Content 1.0",
		from: "iab.ad-product.2.0", fromCattax: 8, hasFromTax: true,
		to: "iab.content.1.0", toCattax: 1, hasToTax: true,
		page: pageAd, released: "2024-12",
		cols: &colMap{header: "Name", fromID: 1, fromName: 0, fromTiers: []int{2, 3, 4}, toID: 6, toName: 5, toTiers: []int{7, 8}, note: -1},
	},
	{
		file: "mappings/ad-product-2.0-to-content-2.1.json", upstream: "Taxonomy Mappings/Ad Product 2.0 to Content 2.1.tsv",
		id: "iab.map.ad-product-2.0.content-2.1", name: "Ad Product 2.0 to Content 2.1",
		from: "iab.ad-product.2.0", fromCattax: 8, hasFromTax: true,
		to: "iab.content.2.1", toCattax: 5, hasToTax: true,
		page: pageAd,
		cols: &colMap{header: "Name", fromID: 1, fromName: 0, fromTiers: []int{2, 3, 4}, toID: 5, toName: 6, note: -1},
	},
	{
		file: "mappings/content-1.0-to-ad-product-2.0.json", upstream: "Taxonomy Mappings/Content 1.0 to Ad Product 2.0.tsv",
		id: "iab.map.content-1.0.ad-product-2.0", name: "Content 1.0 to Ad Product 2.0",
		from: "iab.content.1.0", fromCattax: 1, hasFromTax: true,
		to: "iab.ad-product.2.0", toCattax: 8, hasToTax: true,
		page: pageCont, released: "2023-11",
		cols: &colMap{header: "Content Taxonomy 1.0 Unique ID", fromID: 0, fromName: -1, fromTiers: []int{1, 2}, toID: 3, toName: -1, toTiers: []int{4, 5, 6}, note: -1},
	},
	{
		file: "mappings/content-1.0-to-content-2.0.json", upstream: "Taxonomy Mappings/Content 1.0 to Content 2.0.tsv",
		id: "iab.map.content-1.0.content-2.0", name: "Content 1.0 to Content 2.0",
		from: "iab.content.1.0", fromCattax: 1, hasFromTax: true,
		to: "iab.content.2.0", toCattax: 2, hasToTax: true,
		page: pageCont, released: "2023-11",
		cols: &colMap{header: "Content 1.0 Unique ID", fromID: 0, fromName: 1, toID: 3, toName: 5, toTiers: []int{6, 7, 8, 9}, note: 10},
	},
	{
		file: "mappings/content-2.0-to-content-2.1.json", upstream: "Taxonomy Mappings/Content 2.0 to Content 2.1.tsv",
		id: "iab.map.content-1.0.content-2.1", name: "Content 1.0 to Content 2.1",
		from: "iab.content.1.0", fromCattax: 1, hasFromTax: true,
		to: "iab.content.2.1", toCattax: 5, hasToTax: true,
		page: pageCont, released: "2023-11",
		notes:         "Upstream file is named Content 2.0 to Content 2.1. Rows map Content Taxonomy 1.0 codes (OLD RTB ID) to numeric IDs present in Content Taxonomy 2.0 and 2.1. Categories in the NOT MAPPED columns are a separate list and are emitted after the mappings with note \"NOT MAPPED (deprecated)\".",
		content20File: true,
	},
	{
		file: "mappings/content-2.1-to-ad-product-2.0.json", upstream: "Taxonomy Mappings/Content 2.1 to Ad Product 2.0.tsv",
		id: "iab.map.content-2.1.ad-product-2.0", name: "Content 2.1 to Ad Product 2.0",
		from: "iab.content.2.1", fromCattax: 5, hasFromTax: true,
		to: "iab.ad-product.2.0", toCattax: 8, hasToTax: true,
		page: pageCont,
		cols: &colMap{header: "Unique ID", fromID: 0, fromName: 2, fromTiers: []int{3, 4, 5, 6}, toID: 8, toName: 9, note: -1},
	},
	{
		file: "mappings/ctv-genre-to-content-3.1.json", upstream: "Taxonomy Mappings/CTV Genre Mapping.tsv",
		id: "iab.map.ctv-genre.content-3.1", name: "CTV Genre to Content 3.1",
		to: "iab.content.3.1", toCattax: 9, hasToTax: true,
		page: pageCont, released: "2024-12", genre: true,
		cols: &colMap{header: "CTV Genres", fromName: 0, fromID: -1, toID: 1, toName: 2, note: -1},
	},
	{
		file: "mappings/podcast-genre-to-content-3.1.json", upstream: "Taxonomy Mappings/Podcast Genre Mapping.tsv",
		id: "iab.map.podcast-genre.content-3.1", name: "Podcast Genre to Content 3.1",
		to: "iab.content.3.1", toCattax: 9, hasToTax: true,
		page: pageCont, released: "2024-12", genre: true,
		cols: &colMap{header: "Podcast Genres", fromName: 0, fromID: -1, toID: 1, toName: 2, note: -1},
	},
}
