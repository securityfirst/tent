package component

import (
	"io"
	"sort"
	"strings"

	git "gopkg.in/src-d/go-git.v3"
)

// Parser is an helper, creates a tree from the repo
type Parser struct {
	index      map[string]int
	Categories []*Category
	Assets     []*Asset
}

// Parse executes the parsing on a repo
func (p *Parser) Parse(t *git.Tree) error {
	p.index = make(map[string]int)
	p.Categories = make([]*Category, 0)
	if err := p.parse(t, filterCat); err != nil {
		return err
	}
	if err := p.parse(t, filterRes); err != nil {
		return err
	}
	sort.Sort(catSorter(p.Categories))
	for i := range p.Categories {
		sort.Sort(subSorter(p.Categories[i].subcategories))
		for j := range p.Categories[i].subcategories {
			sort.Sort(itemSorter(p.Categories[i].subcategories[j].items))
		}
	}
	return nil
}

func (p *Parser) parse(t *git.Tree, fn func(name string) bool) error {
	for iter := t.Files(); ; {
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
		if err := p.parseFile(f); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) parseFile(f *git.File) error {
	contents, err := f.Contents()
	if err != nil {
		return parseError{f.Name, "read", err}
	}
	cmp, err := newCmp("/" + f.Name)
	if err != nil {
		return parseError{f.Name, "cmp", err}
	}
	parts := strings.Split(f.Name, "/")
	if err := cmp.SetPath("/" + f.Name); err != nil {
		return parseError{f.Name, "path", err}
	}
	switch c := cmp.(type) {
	case *Category:
		p.index[parts[1]] = len(p.Categories)
		p.Categories = append(p.Categories, c)
	case *Subcategory:
		p.Categories[p.index[parts[1]]].Add(c)
	case *Item:
		p.Categories[p.index[parts[1]]].Sub(parts[2]).AddItem(c)
	case *Checklist:
		p.Categories[p.index[parts[1]]].Sub(parts[2]).SetChecks(c)
	case *Asset:
		p.Assets = append(p.Assets, c)
	default:
		return parseError{f.Name, "type", "Invalid Path"}
	}
	if err := cmp.SetContents(strings.TrimSpace(contents)); err != nil {
		return parseError{f.Name, "contents", err}
	}
	return nil
}
