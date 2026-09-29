package main

// Content 1.0 categories with no Content 2.0 target. Each becomes its own
// node under the nearest Content 3.1 category.
var content10Parent = map[string]struct {
	parent string
	reason string
}{
	"IAB25-3": {"Rm3SiT", "Pornography is adult sexual content."},
	"IAB25-5": {"HxqYV1", "Hate Content is the same category as Hate Speech and Acts of Aggression."},
	"IAB25-4": {"j9PaO9", "Profane Content is the same category as Obscenity and Profanity."},
	"IAB25-2": {"XtODT3", "Extreme graphic violence is a harmful act."},
	"IAB25-1": {"6i4dB6", "Unmoderated UGC sits with spam and harmful content."},
	"IAB25-7": {"6i4dB6", "Incentivized content sits with spam and harmful content."},
	"IAB25-6": {"v9i3On", "Under-construction pages have no finer 3.1 topic."},
	"IAB25":   {"v9i3On", "Non-standard content is the sensitive-topics bucket."},
	"IAB26":   {"XtODT3", "Illegal content is a crime and harmful-acts topic."},
	"IAB26-1": {"XtODT3", "Illegal content is a crime and harmful-acts topic."},
	"IAB26-2": {"mm3UXx", "Warez is online piracy."},
	"IAB26-4": {"mm3UXx", "Copyright infringement is online piracy."},
	"IAB26-3": {"6i4dB6", "Spyware and malware are harmful content."},
	"IAB24":   {"v9i3On", "Uncategorized has no finer 3.1 topic."},
	"IAB14-8": {"Z7rJBM", "Ethnic-specific content is a debated sensitive social issue."},
}

// Categories removed as movie or TV genres and continued as shared genres.
var genreParent = map[string]struct {
	parent string
	reason string
}{
	"329": {"641", "Animation movies continued as the shared Animation & Anime genre."},
	"330": {"646", "Comedy movies continued as the shared Comedy genre."},
	"333": {"647", "Drama movies continued as the shared Drama genre."},
	"334": {"645", "Family and children movies continued as the shared Family/Children genre."},
	"327": {"652", "Science fiction movies continued as the shared Science Fiction genre."},
	"328": {"643", "Special interest movies continued as the shared Special Interest genre."},
	"379": {"386", "News and Politics was dissolved; Politics is the surviving former child."},
	"382": {"1020", "International news is a scope of the news purpose."},
	"384": {"1020", "Local news is a scope of the news purpose."},
	"385": {"1020", "National news is a scope of the news purpose."},
}

// sameCategory lists Content 1.0 categories that name the same thing as the
// node they land on. The code is stored on that node. A different 1.0
// category that already occupies the code stays its own child.
var sameCategory = map[string]bool{
	"IAB25-5": true,
	"IAB25-4": true,
	"IAB22-2": true,
}

type contentDoc struct {
	Meta  contentMeta   `json:"meta"`
	Nodes []contentNode `json:"nodes"`
}

type contentMeta struct {
	ID            string      `json:"id"`
	Kind          string      `json:"kind"`
	Name          string      `json:"name"`
	Version       string      `json:"version"`
	Base          string      `json:"base"`
	CategoryCount int         `json:"category_count"`
	RootCount     int         `json:"root_count"`
	Placements    []placement `json:"placements"`
}

type contentNode struct {
	ID       int           `json:"id"`
	R0       string        `json:"r0_code"`
	C10      nullString    `json:"c1_0_code"`
	C20      nullString    `json:"c2_0_code"`
	C21      nullString    `json:"c2_1_code"`
	C22      nullString    `json:"c2_2_code"`
	C30      nullString    `json:"c3_0_code"`
	C31      nullString    `json:"c3_1_code"`
	Name     string        `json:"name"`
	Children []contentNode `json:"children,omitempty"`
}

type nullString string

func (s nullString) MarshalJSON() ([]byte, error) {
	if s == "" {
		return []byte("null"), nil
	}
	return jsonMarshalString(string(s))
}

func buildContent(root string) (contentDoc, error) {
	c31, err := readTax(root, "iab/datasets/content/3.1.json")
	if err != nil {
		return contentDoc{}, err
	}
	c30, err := readTax(root, "iab/datasets/content/3.0.json")
	if err != nil {
		return contentDoc{}, err
	}
	c22, err := readTax(root, "iab/datasets/content/2.2.json")
	if err != nil {
		return contentDoc{}, err
	}
	c21, err := readTax(root, "iab/datasets/content/2.1.json")
	if err != nil {
		return contentDoc{}, err
	}
	c20, err := readTax(root, "iab/datasets/content/2.0.json")
	if err != nil {
		return contentDoc{}, err
	}
	c10, err := readTax(root, "iab/datasets/content/1.0.json")
	if err != nil {
		return contentDoc{}, err
	}
	vectors, err := readTax(root, "iab/datasets/content/3.0-vectors.json")
	if err != nil {
		return contentDoc{}, err
	}
	mapping, err := readMap(root, "iab/datasets/mappings/content-1.0-to-content-2.0.json")
	if err != nil {
		return contentDoc{}, err
	}

	in31 := taxIDs(c31.Nodes)
	in30 := taxIDs(c30.Nodes)
	in22 := taxIDs(c22.Nodes)
	in21 := taxIDs(c21.Nodes)
	in20 := taxIDs(c20.Nodes)

	byID := map[string]*node{}
	var roots []*node
	var placements []placement

	var take func([]taxNode, *node, bool)
	take = func(raw []taxNode, parent *node, isRoot bool) {
		for _, rawNode := range raw {
			n := byID[rawNode.ID]
			if n == nil {
				n = &node{name: rawNode.Name, iab: rawNode.ID}
				byID[rawNode.ID] = n
				if parent != nil {
					parent.add(n)
				} else if isRoot {
					roots = append(roots, n)
				}
			}
			take(rawNode.Children, n, false)
		}
	}
	take(c31.Nodes, nil, true)
	take(vectors.Nodes, nil, true)

	var placeOld func([]taxNode, string) error
	placeOld = func(raw []taxNode, parentID string) error {
		for _, rawNode := range raw {
			if byID[rawNode.ID] == nil {
				n := &node{name: rawNode.Name, iab: rawNode.ID}
				pid := parentID
				var reason string
				if over, ok := genreParent[rawNode.ID]; ok {
					pid = over.parent
					reason = over.reason
				}
				parent := byID[pid]
				if parent == nil {
					return errf("content %s %s has no parent %s", rawNode.ID, rawNode.Name, pid)
				}
				parent.add(n)
				byID[rawNode.ID] = n
				if reason != "" {
					placements = append(placements, placement{
						Category: rawNode.Name,
						Code:     rawNode.ID,
						Parent:   parent.name,
						Reason:   reason,
					})
				}
			}
			if err := placeOld(rawNode.Children, rawNode.ID); err != nil {
				return err
			}
		}
		return nil
	}
	if err := placeOld(c22.Nodes, ""); err != nil {
		return contentDoc{}, err
	}

	for id, n := range byID {
		if in31[id] {
			n.c31 = id
		}
		if in30[id] {
			n.c30 = id
		}
		if in22[id] {
			n.c22 = id
		}
		if in21[id] {
			n.c21 = id
		}
		if in20[id] {
			n.c20 = id
		}
	}

	c10Names := map[string]string{}
	walkTax(c10.Nodes, "", func(n taxNode, _ string) { c10Names[n.ID] = n.Name })
	c10Target := map[string]string{}
	for _, e := range mapping.Entries {
		if e.To != nil && e.To.ID != "" {
			c10Target[e.From.ID] = e.To.ID
		}
	}

	var place10 func([]taxNode) error
	place10 = func(raw []taxNode) error {
		for _, rawNode := range raw {
			target := c10Target[rawNode.ID]
			switch {
			case target != "":
				host := byID[target]
				if host == nil {
					return errf("content 1.0 %s maps to missing %s", rawNode.ID, target)
				}
				placeC10(host, rawNode.ID, rawNode.Name, c10Names, &placements)
			case rawNode.ID == "IAB7":
				parent, reason, err := healthParent(byID, c10Target, c10.Nodes)
				if err != nil {
					return err
				}
				parent.add(&node{name: rawNode.Name, c10: rawNode.ID})
				placements = append(placements, placement{
					Category: rawNode.Name, Code: rawNode.ID, Parent: parent.name, Reason: reason,
				})
			default:
				over, ok := content10Parent[rawNode.ID]
				if !ok {
					return errf("content 1.0 %s has no placement", rawNode.ID)
				}
				parent := byID[over.parent]
				if parent == nil {
					return errf("content 1.0 parent %s missing for %s", over.parent, rawNode.ID)
				}
				if sameCategory[rawNode.ID] {
					placeC10(parent, rawNode.ID, rawNode.Name, c10Names, &placements)
				} else {
					parent.add(&node{name: rawNode.Name, c10: rawNode.ID})
					placements = append(placements, placement{
						Category: rawNode.Name, Code: rawNode.ID, Parent: parent.name, Reason: over.reason,
					})
				}
			}
			if err := place10(rawNode.Children); err != nil {
				return err
			}
		}
		return nil
	}
	if err := place10(c10.Nodes); err != nil {
		return contentDoc{}, err
	}

	assignIDs(roots)
	assignR0(roots)

	doc := contentDoc{
		Meta: contentMeta{
			ID:            "r0.content.1.0",
			Kind:          "taxonomy",
			Name:          "Content",
			Version:       "1.0",
			Base:          "iab.content.3.1",
			CategoryCount: countNodes(roots),
			RootCount:     len(roots),
			Placements:    placements,
		},
		Nodes: contentTree(roots),
	}
	return doc, nil
}

func placeC10(host *node, id, name string, names map[string]string, placements *[]placement) {
	if host.c10 == "" {
		host.c10 = id
		if sameCategory[id] {
			*placements = append(*placements, placement{
				Category: name,
				Code:     id,
				Parent:   host.name,
				Reason:   "Same category as the parent; the code is stored on that node.",
			})
		}
		return
	}
	if sameCategory[id] && !sameCategory[host.c10] {
		old := host.c10
		host.c10 = id
		host.add(&node{name: names[old], c10: old})
		*placements = append(*placements,
			placement{
				Category: name,
				Code:     id,
				Parent:   host.name,
				Reason:   "Same category as the parent; the code is stored on that node.",
			},
			placement{
				Category: names[old],
				Code:     old,
				Parent:   host.name,
				Reason:   "Different category that shares this node's official mapping.",
			},
		)
		return
	}
	host.add(&node{name: name, c10: id})
	*placements = append(*placements, placement{
		Category: name,
		Code:     id,
		Parent:   host.name,
		Reason:   "Another Content 1.0 category already holds this node's code.",
	})
}

func healthParent(byID map[string]*node, target map[string]string, roots []taxNode) (*node, string, error) {
	var iab7 *taxNode
	var find func([]taxNode) *taxNode
	find = func(ns []taxNode) *taxNode {
		for i := range ns {
			if ns[i].ID == "IAB7" {
				return &ns[i]
			}
			if n := find(ns[i].Children); n != nil {
				return n
			}
		}
		return nil
	}
	iab7 = find(roots)
	if iab7 == nil {
		return nil, "", errf("IAB7 missing")
	}
	medical, healthy := 0, 0
	var walk func([]taxNode)
	walk = func(ns []taxNode) {
		for _, n := range ns {
			host := byID[target[n.ID]]
			for p := host; p != nil; p = p.parent {
				switch p.iab {
				case "286":
					medical++
				case "223":
					healthy++
				}
			}
			walk(n.Children)
		}
	}
	walk(iab7.Children)
	id := "286"
	reason := "Most Health & Fitness children map under Medical Health."
	if healthy > medical {
		id = "223"
		reason = "Most Health & Fitness children map under Healthy Living."
	}
	parent := byID[id]
	if parent == nil {
		return nil, "", errf("health parent %s missing", id)
	}
	return parent, reason, nil
}

func contentTree(roots []*node) []contentNode {
	out := make([]contentNode, len(roots))
	for i, n := range roots {
		out[i] = contentNode{
			ID: n.id, R0: n.r0, Name: n.name,
			C10: nullString(n.c10), C20: nullString(n.c20), C21: nullString(n.c21),
			C22: nullString(n.c22), C30: nullString(n.c30), C31: nullString(n.c31),
			Children: contentTree(n.children),
		}
	}
	return out
}
