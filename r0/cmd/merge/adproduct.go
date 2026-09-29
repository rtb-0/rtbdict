package main

// adParent is the Ad Product 2.0 path that holds an unmapped Ad Product 1.1 root.
var adParent = map[string]struct {
	path   []string
	reason string
}{
	"1":   {[]string{"Computer Software"}, "Ad Product 2.0 folded applications into verticals; the Apps umbrella sits with software."},
	"27":  {[]string{"Media"}, "Arts and Entertainment is the media umbrella; its events already map to Events and Performances."},
	"113": {[]string{"Cosmetic Services"}, "Beauty services are cosmetic services."},
	"676": {[]string{"Durable Goods"}, "Hardware supplies are durable goods."},
	"739": {[]string{"Family and Parenting"}, "Life events are family occasions."},
	"762": {[]string{"Pet Ownership"}, "Pet services are the service side of pet ownership."},
	"775": {[]string{"Fitness Activities"}, "Recreation and fitness activities are fitness activities."},
	"783": {[]string{"Computer Software"}, "Software is computer software."},
	"860": {[]string{"Business and Industrial", "Business Services", "Information Technology Services"}, "Web services are information technology services."},
}

// appParent sends each unmapped app type to the vertical it names.
var appParent = map[string]struct {
	path   []string
	reason string
}{
	"2":  {[]string{"Vehicles"}, "Auto and vehicle apps are the vehicles vertical."},
	"4":  {[]string{"Business and Industrial"}, "Business apps are the business vertical."},
	"5":  {[]string{"Education and Careers"}, "Education apps are the education vertical."},
	"6":  {[]string{"Media"}, "Entertainment apps are the media vertical."},
	"8":  {[]string{"Food and Beverage Services"}, "Food and drink apps are the food vertical."},
	"12": {[]string{"Media"}, "Magazine and newspaper apps are the media vertical."},
	"14": {[]string{"Media"}, "Music apps are the media vertical."},
	"15": {[]string{"Travel and Tourism"}, "Navigation apps are the travel vertical."},
	"16": {[]string{"Media"}, "News apps are the media vertical."},
	"21": {[]string{"Retail"}, "Shopping apps are the retail vertical."},
	"22": {[]string{"Media"}, "Social networking apps are the media vertical."},
	"23": {[]string{"Sporting Goods"}, "Sports apps are the sporting-goods vertical."},
	"24": {[]string{"Travel and Tourism"}, "Travel apps are the travel vertical."},
}

var hobbyParent = map[string]struct {
	path   []string
	reason string
	root   bool
}{
	"717": {path: []string{"Religion and Spirituality", "Astrology"}, reason: "Psychics and astrology are the astrology category."},
	"718": {path: []string{"Education and Careers"}, reason: "Workshops and classes are instruction."},
	"714": {root: true, reason: "Arts, crafts, and musical instruments have no Ad Product 2.0 vertical."},
}

type adDoc struct {
	Meta  adMeta   `json:"meta"`
	Nodes []adNode `json:"nodes"`
}

type adMeta struct {
	ID            string      `json:"id"`
	Kind          string      `json:"kind"`
	Name          string      `json:"name"`
	Version       string      `json:"version"`
	Base          string      `json:"base"`
	CategoryCount int         `json:"category_count"`
	RootCount     int         `json:"root_count"`
	Placements    []placement `json:"placements"`
}

type adNode struct {
	ID       int        `json:"id"`
	R0       string     `json:"r0_code"`
	AP10     nullString `json:"ap1_0_code"`
	AP11     nullString `json:"ap1_1_code"`
	AP20     nullString `json:"ap2_0_code"`
	Name     string     `json:"name"`
	Children []adNode   `json:"children,omitempty"`
}

type occ struct {
	id       string
	name     string
	path     []string
	parent   *occ
	children []*occ
}

func buildAdProduct(root string) (adDoc, error) {
	ap20, err := readTax(root, "iab/datasets/adproduct/2.0.json")
	if err != nil {
		return adDoc{}, err
	}
	ap11, err := readTax(root, "iab/datasets/adproduct/1.1.json")
	if err != nil {
		return adDoc{}, err
	}
	ap10, err := readTax(root, "iab/datasets/adproduct/1.0.json")
	if err != nil {
		return adDoc{}, err
	}
	mapping, err := readMap(root, "iab/datasets/mappings/ad-product-2.0-to-ad-product-1.1.json")
	if err != nil {
		return adDoc{}, err
	}

	in10 := map[string]int{}
	walkTax(ap10.Nodes, "", func(n taxNode, _ string) { in10[n.ID]++ })

	by20 := map[string]*node{}
	var roots []*node
	var take func([]taxNode, *node)
	take = func(raw []taxNode, parent *node) {
		for _, rawNode := range raw {
			n := &node{name: rawNode.Name, iab: rawNode.ID, ap20: rawNode.ID}
			by20[rawNode.ID] = n
			if parent == nil {
				roots = append(roots, n)
			} else {
				parent.add(n)
			}
			take(rawNode.Children, n)
		}
	}
	take(ap20.Nodes, nil)

	byPath := map[string]*node{}
	byName := map[string][]*node{}
	var index func([]*node, []string)
	index = func(ns []*node, path []string) {
		for _, n := range ns {
			p := append(append([]string{}, path...), n.name)
			byPath[normPath(p)] = n
			byName[normName(n.name)] = append(byName[normName(n.name)], n)
			index(n.children, p)
		}
	}
	index(roots, nil)

	var occs []*occ
	byOcc := map[string][]*occ{}
	var walkOcc func([]taxNode, *occ, []string)
	walkOcc = func(raw []taxNode, parent *occ, path []string) {
		for _, rawNode := range raw {
			p := append(append([]string{}, path...), rawNode.Name)
			o := &occ{id: rawNode.ID, name: rawNode.Name, path: p, parent: parent}
			if parent != nil {
				parent.children = append(parent.children, o)
			}
			occs = append(occs, o)
			byOcc[o.id] = append(byOcc[o.id], o)
			walkOcc(rawNode.Children, o, p)
		}
	}
	walkOcc(ap11.Nodes, nil, nil)

	claims := map[*occ][]*node{}
	for _, e := range mapping.Entries {
		if e.To == nil || e.To.ID == "" {
			continue
		}
		from, err := resolve20(e.From, byPath, byName)
		if err != nil {
			return adDoc{}, err
		}
		o, err := resolve11(e.To, byOcc)
		if err != nil {
			return adDoc{}, err
		}
		claims[o] = append(claims[o], from)
	}

	var placements []placement
	made := map[*occ]*node{}
	for _, o := range occs {
		targets := uniqueNodes(claims[o])
		if len(targets) == 0 {
			continue
		}
		host := bestHost(o, targets)
		if len(targets) > 1 {
			placements = append(placements, placement{
				Category: o.name,
				Code:     o.id,
				Parent:   host.name,
				Reason:   "Several Ad Product 2.0 categories map here; the closest path keeps the code.",
			})
		}
		made[o] = putAdCode(host, o, in10[o.id] > 0)
	}

	for _, o := range occs {
		if _, ok := made[o]; ok {
			continue
		}
		n := &node{name: o.name, ap11: o.id}
		if in10[o.id] > 0 {
			n.ap10 = o.id
		}
		made[o] = n
		parent, reason, isRoot, err := adPlace(o, roots, made)
		if err != nil {
			return adDoc{}, err
		}
		if isRoot {
			roots = append(roots, n)
			placements = append(placements, placement{
				Category: o.name, Code: o.id, Parent: "", Reason: reason,
			})
			continue
		}
		parent.add(n)
		if reason != "" {
			placements = append(placements, placement{
				Category: o.name, Code: o.id, Parent: parent.name, Reason: reason,
			})
		}
	}

	assignIDs(roots)
	assignR0(roots)
	return adDoc{
		Meta: adMeta{
			ID:            "r0.adproduct.1.0",
			Kind:          "taxonomy",
			Name:          "Ad Product",
			Version:       "1.0",
			Base:          "iab.ad-product.2.0",
			CategoryCount: countNodes(roots),
			RootCount:     len(roots),
			Placements:    placements,
		},
		Nodes: adTree(roots),
	}, nil
}

func adPlace(o *occ, roots []*node, made map[*occ]*node) (*node, string, bool, error) {
	if o.parent != nil && o.parent.id == "1" {
		if over, ok := appParent[o.id]; ok {
			parent, err := findPath(roots, over.path)
			return parent, over.reason, false, err
		}
	}
	if over, ok := hobbyParent[o.id]; ok {
		if over.root {
			return nil, over.reason, true, nil
		}
		parent, err := findPath(roots, over.path)
		return parent, over.reason, false, err
	}
	if o.parent == nil {
		over, ok := adParent[o.id]
		if !ok {
			return nil, "", false, errf("ad product root %s %s has no placement", o.id, o.name)
		}
		parent, err := findPath(roots, over.path)
		return parent, over.reason, false, err
	}
	if host := made[o.parent]; host != nil {
		return host, "", false, nil
	}
	return nil, "", false, errf("ad product %s %s parent %s is not placed yet", o.id, o.name, o.parent.id)
}

func putAdCode(host *node, o *occ, in10 bool) *node {
	if host.ap11 == "" || host.ap11 == o.id {
		host.ap11 = o.id
		if in10 && host.ap10 == "" {
			host.ap10 = o.id
		}
		return host
	}
	extra := &node{name: o.name, ap11: o.id}
	if in10 {
		extra.ap10 = o.id
	}
	host.add(extra)
	return extra
}

func bestHost(o *occ, targets []*node) *node {
	best := targets[0]
	bestScore := pathScore(o.path, nodePath(best))
	for _, t := range targets[1:] {
		score := pathScore(o.path, nodePath(t))
		if score > bestScore || (score == bestScore && t.iab < best.iab) {
			best = t
			bestScore = score
		}
	}
	return best
}

func nodePath(n *node) []string {
	var path []string
	for p := n; p != nil; p = p.parent {
		path = append(path, p.name)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func pathScore(a, b []string) int {
	set := map[string]bool{}
	for _, s := range a {
		set[normName(s)] = true
	}
	n := 0
	for _, s := range b {
		if set[normName(s)] {
			n++
		}
	}
	if len(a) > 0 && len(b) > 0 && normName(a[len(a)-1]) == normName(b[len(b)-1]) {
		n += 2
	}
	return n
}

func uniqueNodes(ns []*node) []*node {
	seen := map[*node]bool{}
	var out []*node
	for _, n := range ns {
		if n != nil && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

func resolve20(from mapRef, byPath map[string]*node, byName map[string][]*node) (*node, error) {
	if n := byPath[normPath(from.Path)]; n != nil {
		return n, nil
	}
	hits := byName[normName(from.Name)]
	if len(hits) == 1 {
		return hits[0], nil
	}
	return nil, errf("ad product 2.0 path not resolved: %s", from.Name)
}

func resolve11(to *mapRef, byOcc map[string][]*occ) (*occ, error) {
	id := to.ID
	if id == "10" && normName(to.Name) == normName("Health and Fitness Apps") {
		id = "0"
	}
	cands := byOcc[id]
	if len(cands) == 0 {
		return nil, errf("ad product 1.1 id %s (%s) not found", to.ID, to.Name)
	}
	if len(cands) == 1 {
		return cands[0], nil
	}
	for _, o := range cands {
		if normName(o.name) == normName(to.Name) {
			return o, nil
		}
	}
	return nil, errf("ad product 1.1 id %s name %s is ambiguous", to.ID, to.Name)
}

func adTree(roots []*node) []adNode {
	out := make([]adNode, len(roots))
	for i, n := range roots {
		out[i] = adNode{
			ID: n.id, R0: n.r0, Name: n.name,
			AP10: nullString(n.ap10), AP11: nullString(n.ap11), AP20: nullString(n.ap20),
			Children: adTree(n.children),
		}
	}
	return out
}
