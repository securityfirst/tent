package component

import (
	"io"
	"sort"
	"strings"

	git "gopkg.in/src-d/go-git.v3"
)

// TreeParser is an helper, creates a tree from the repo
type TreeParser struct {
	index      map[string]int
	Categories []*Category
	Assets     []*Asset
}

// Parse executes the parsing on a repo
func (t *TreeParser) Parse(tree *git.Tree) error {
	t.index = make(map[string]int)
	t.Categories = make([]*Category, 0)
	if err := t.parse(tree, filterCat); err != nil {
		return err
	}
	if err := t.parse(tree, filterRes); err != nil {
		return err
	}
	sort.Sort(catSorter(t.Categories))
	for i := range t.Categories {
		sort.Sort(subSorter(t.Categories[i].subcategories))
		for j := range t.Categories[i].subcategories {
			sort.Sort(itemSorter(t.Categories[i].subcategories[j].items))
		}
	}
	return nil
}

func (t *TreeParser) parse(tree *git.Tree, fn func(name string) bool) error {
	for iter := tree.Files(); ; {
		f, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if !fn(f.Name) {
			continue
		}
		if err := t.parseFile(f); err != nil {
			return err
		}
	}
	return nil
}

func (t *TreeParser) parseFile(f *git.File) error {
	contents, err := f.Contents()
	if err != nil {
		return parseError{f.Name, "read", err}
	}
	cmp, err := newCmp("/" + f.Name)
	if err != nil {
		return parseError{f.Name, "cmp", err}
	}
	p := strings.Split(f.Name, "/")
	if err := cmp.SetPath("/" + f.Name); err != nil {
		return parseError{f.Name, "path", err}
	}
	switch c := cmp.(type) {
	case *Category:
		t.index[p[1]] = len(t.Categories)
		t.Categories = append(t.Categories, c)
	case *Subcategory:
		t.Categories[t.index[p[1]]].Add(c)
	case *Item:
		t.Categories[t.index[p[1]]].Sub(p[2]).AddItem(c)
	case *Checklist:
		t.Categories[t.index[p[1]]].Sub(p[2]).SetChecks(c)
	case *Asset:
		t.Assets = append(t.Assets, c)
	default:
		return parseError{f.Name, "type", "Invalid Path"}
	}
	if err := cmp.SetContents(contents); err != nil {
		return parseError{f.Name, "contents", err}
	}
	return nil
}
