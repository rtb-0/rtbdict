# rtbdict

Go dictionaries for the r0 category tree and the IAB Tech Lab taxonomies. The module is `github.com/rtb-0/rtbdict` (Go 1.27) and has no dependencies. Generated tables are compiled into the binary.

`iab/dict` keeps the full taxonomy text and mapping rows. `iab/lightdict` and `r0/lightdict` keep the same lookups in compact tables. r0 ids, parents, and children are dense integers. IAB light tables are `int32` slices, so the garbage collector does not scan them.

## r0

Source: [`r0/datasets/r0-1.0.json`](r0/datasets/r0-1.0.json). 1471 categories, 32 roots. Numeric ids run level by level from 1.

| Package        | Import                                  | Contents                                      |
| -------------- | --------------------------------------- | --------------------------------------------- |
| `r0/dict`      | `github.com/rtb-0/rtbdict/r0/dict`      | Categories, including descriptions            |
| `r0/lightdict` | `github.com/rtb-0/rtbdict/r0/lightdict` | Same categories, no descriptions              |
| `r0/keyword`   | `github.com/rtb-0/rtbdict/r0/keyword`   | Shared phrase index used by both dictionaries |

```go
n := dict.ByR0Code("tec.c7")
n.ID()            // 497
n.Name()          // Cloud Storage
n.IABContent10()  // Content Taxonomy 1.0 code, when this node has one

dict.ByID(497)
dict.ByCattax(1, "IAB1-1")          // Books and Literature
dict.ByKeyword("cloud storage")     // 497, ASCII case is ignored
```

`ByID`, `ByR0Code`, `ByCattax`, and `ByKeyword` return `nil` when nothing matches. `ByCattax` also returns `nil` when the same code belongs to more than one r0 category. Ad Product 1.1 has no cattax; read it with `IABAdProduct11` after finding the category another way.

Direct codes on a node are Content 1.0 (`c1_0_code`, cattax 1), Content 3.1 (`c3_1_code`, cattax 9), and Ad Product 2.0 (`ap2_0_code`, cattax 8). The other IAB codes are derived from that node's own direct codes through the official mappings. One node is not given another node's direct code.

| Cattax | Taxonomy       | Method           |
| ------ | -------------- | ---------------- |
| 1      | Content 1.0    | `IABContent10`   |
| 2      | Content 2.0    | `IABContent20`   |
| 3      | Ad Product 1.0 | `IABAdProduct10` |
| 5      | Content 2.1    | `IABContent21`   |
| 6      | Content 2.2    | `IABContent22`   |
| 7      | Content 3.0    | `IABContent30`   |
| 8      | Ad Product 2.0 | `IABAdProduct20` |
| 9      | Content 3.1    | `IABContent31`   |
| —      | Ad Product 1.1 | `IABAdProduct11` |

`Parent` is `0` for a root. `Children` and `Keywords` return the stored slices.

### Keyword search

`ByKeyword` matches `name` and each `keywords` entry. ASCII letters are folded to lower case, and repeated whitespace collapses to one space. Matching starts at a word boundary. Earlier words of a phrase must match exactly. The last word may be a prefix of the input word (`synchronization` matches `synchronizations`). The longest phrase wins. If two categories share a phrase, the larger id stays in the index.

`keyword.Match` returns the r0 id, or `0`. Input of at most 256 bytes is folded in a stack buffer and does not allocate.

## IAB

Source JSON is in [`iab/datasets`](iab/datasets): Ad Product, Content, Audience, and the official mappings. Refresh it from the IAB Tech Lab TSV files with:

```
go run ./iab/cmd/fetch
```

| Package              | Import                                                            |
| -------------------- | ----------------------------------------------------------------- |
| `iab/dict`           | `github.com/rtb-0/rtbdict/iab/dict`                               |
| `iab/dict/adproduct` | `github.com/rtb-0/rtbdict/iab/dict/ad-product`                    |
| `iab/dict/content`   | `github.com/rtb-0/rtbdict/iab/dict/content`                       |
| `iab/dict/audience`  | `github.com/rtb-0/rtbdict/iab/dict/audience`                      |
| `iab/dict/mappings`  | `github.com/rtb-0/rtbdict/iab/dict/mappings`                      |
| `iab/lightdict`      | `github.com/rtb-0/rtbdict/iab/lightdict`                          |
| light families       | `.../iab/lightdict/ad-product`, `content`, `audience`, `mappings` |

`iab/dict` lookups return `(value, bool)`. `iab/lightdict` lookups return a pointer, or `nil`.

```go
tax, ok := content.ByCattax(9)          // iab/dict/content, Content 3.1
nodes := tax.Nodes("Rm3SiT")

light := lightcontent.ByCattax(9)       // iab/lightdict/content
idx := light.Indexes("Rm3SiT")
light.Children(idx[0])

m := lightmappings.ByTaxonomy("iab.content.1.0", "iab.content.2.0")
m.Targets("IAB1")
```

Ad Product id `51` is two categories (parents 27 and 54). `Nodes` and `Indexes` return both. Audience 1.0 has no cattax. In `iab/lightdict/audience`, `ByCattax(0)` returns that taxonomy. Audience 1.1 is cattax 4.

Light mappings store targets in one string table. `Targets` returns a subslice. CTV and podcast mappings are keyed by genre name. Empty mapping targets are omitted.

## Generate

Datasets are the source. Do not edit generated `.go` files.

```
go generate ./iab/dict/
go generate ./iab/lightdict/
go generate ./r0/dict/
```

`r0/cmd/gendict` writes `r0/dict/1.0.go`, `r0/lightdict/1.0.go`, and `r0/keyword/phrases.go`.

## Make

```bash
make test    # go test ./...
make bench   # go test -bench=. -benchmem
make lint    # go vet ./...
make tidy    # go mod tidy
```
