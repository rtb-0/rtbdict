// Command merge writes r0/datasets/content/1.0.json and r0/datasets/adproduct/1.0.json.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	content, err := buildContent(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ad, err := buildAdProduct(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	contentPath := filepath.Join(root, "r0", "datasets", "content", "1.0.json")
	adPath := filepath.Join(root, "r0", "datasets", "adproduct", "1.0.json")
	if err := writeDoc(contentPath, content); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := writeDoc(adPath, ad); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("content nodes=%d roots=%d placements=%d\n", content.Meta.CategoryCount, content.Meta.RootCount, len(content.Meta.Placements))
	fmt.Printf("adproduct nodes=%d roots=%d placements=%d\n", ad.Meta.CategoryCount, ad.Meta.RootCount, len(ad.Meta.Placements))
	fmt.Printf("adproduct new roots:")
	for _, n := range ad.Nodes {
		if n.AP20 == "" {
			fmt.Printf(" %s", n.Name)
		}
	}
	fmt.Println()
}
