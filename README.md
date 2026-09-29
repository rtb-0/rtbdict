# rtbdict

Go dictionaries for the r0 category tree and the IAB Tech Lab taxonomies. The module is `github.com/rtb-0/rtbdict` (Go 1.27) and has no dependencies. Generated tables are compiled into the binary.

`iab/dict` keeps the full taxonomy text and mapping rows. `iab/lightdict` and `r0/lightdict` keep the same lookups in compact tables. r0 ids, parents, and children are dense integers. IAB light tables are `int32` slices, so the garbage collector does not scan them.

## r0

Source: [`r0/datasets/content/1.0.json`](r0/datasets/content/1.0.json). 227 categories, 34 roots. Numeric ids run level by level from 1.

| Package                | Import                                          | Contents                                              |
| ---------------------- | ----------------------------------------------- | ----------------------------------------------------- |
| `r0/dict/content`      | `github.com/rtb-0/rtbdict/r0/dict/content`      | Categories, including descriptions and IAB names      |
| `r0/lightdict/content` | `github.com/rtb-0/rtbdict/r0/lightdict/content` | Same categories, no descriptions or IAB names         |
| `r0/keyword/content`   | `github.com/rtb-0/rtbdict/r0/keyword/content`   | Phrase index used by both dictionaries                |

```go
n := content.ByR0Code("adl.prn")       // r0/dict/content
n.ID()                                 // 35
n.Name()                               // Pornography
n.IAB()                                // cattax, id, and name
n.IDs(1)                               // ["IAB25-3"]

content.ByID(35)
content.ByCattax(1, "IAB25-3")         // Pornography
content.ByKeyword("porn video")        // 35, ASCII case is ignored

light := lightcontent.ByR0Code("adl.prn") // r0/lightdict/content
light.IAB()                            // cattax and id, no name
light.IDs(1)                           // ["IAB25-3"]
```

`ByID`, `ByR0Code`, `ByCattax`, and `ByKeyword` return `nil` when nothing matches. Each Content Taxonomy pair of cattax and id belongs to one category. `IAB` returns that category's codes in dataset order. In `r0/dict/content` each code has cattax, id, and name. In `r0/lightdict/content` each code has cattax and id. `IDs(cattax)` on either node returns the ids for that cattax, in the same order, as the stored slice. An unknown cattax returns `nil`. The codes are Content 1.0 (cattax 1), 2.0 (2), 2.1 (5), 2.2 (6), 3.0 (7), and 3.1 (9).

`Parent` is `0` for a root. `Children`, `Keywords`, `IAB`, and `IDs` return the stored slices.

### Keyword search

`ByKeyword` matches `name` and each `keywords` entry. ASCII letters are folded to lower case, and repeated whitespace collapses to one space. Matching starts at a word boundary. Earlier words of a phrase must match exactly. The last word may be a prefix of the input word (`synchronization` matches `synchronizations`). The longest phrase wins. If two categories share a phrase, the larger id stays in the index.

`Match` in `r0/keyword/content` returns the r0 id, or `0`. Input of at most 256 bytes is folded in a stack buffer and does not allocate.

## IAB

Source JSON is in [`iab/datasets`](iab/datasets): Ad Product, Content, Audience, and the official mappings. Refresh it from the IAB Tech Lab TSV files with:

```
go run ./iab/cmd/fetch
```

| Package              | Import                                                            |
| -------------------- | ----------------------------------------------------------------- |
| `iab/dict`           | `github.com/rtb-0/rtbdict/iab/dict`                               |
| `iab/dict/adproduct` | `github.com/rtb-0/rtbdict/iab/dict/adproduct`                     |
| `iab/dict/content`   | `github.com/rtb-0/rtbdict/iab/dict/content`                       |
| `iab/dict/audience`  | `github.com/rtb-0/rtbdict/iab/dict/audience`                      |
| `iab/dict/mappings`  | `github.com/rtb-0/rtbdict/iab/dict/mappings`                      |
| `iab/lightdict`      | `github.com/rtb-0/rtbdict/iab/lightdict`                          |
| light families       | `.../iab/lightdict/adproduct`, `content`, `audience`, `mappings`  |

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

`r0/cmd/gendict` writes `r0/dict/content/1.0.go`, `r0/lightdict/content/1.0.go`, and `r0/keyword/content/phrases.go`.

## Make

```bash
make test    # go test ./...
make bench   # go test -bench=. -benchmem
make lint    # go vet ./...
make tidy    # go mod tidy
```
