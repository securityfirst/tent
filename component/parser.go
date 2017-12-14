package component

import (
	"io"
	"sort"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Parser is an helper, creates a tree from the repo
type Parser struct {
	index      map[[2]string]int
	categories []*Category
	assets     []*Asset
	forms      []*Form
}

// Parse executes the parsing on a repo
func (p *Parser) Parse(t *object.Tree) error {
	p.index = make(map[[2]string]int)
	p.categories = make([]*Category, 0)
	if err := p.parse(t, filterCat); err != nil {
		return err
	}
	if err := p.parse(t, filterRes); err != nil {
		return err
	}
	sort.Sort(catSorter(p.categories))
	for i := range p.categories {
		sort.Sort(subSorter(p.categories[i].subcategories))
		for j := range p.categories[i].subcategories {
			for k := range p.categories[i].subcategories[j].difficulties {
				sort.Sort(itemSorter(p.categories[i].subcategories[j].difficulties[k].items))
			}
		}
	}
	return nil
}

func (p *Parser) parse(t *object.Tree, fn func(name string) bool) error {
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

func (p *Parser) parseFile(f *object.File) error {
	contents, err := f.Contents()
	if err != nil {
		return parseError{f.Name, "read", err}
	}
	cmp, err := newCmp(f.Name)
	if err != nil {
		return parseError{f.Name, "cmp", err}
	}
	if err := p.setPath(f.Name, cmp); err != nil {
		return err
	}
	if err := cmp.SetContents(strings.TrimSpace(contents)); err != nil {
		return parseError{f.Name, "contents", err}
	}
	return nil
}

func (p *Parser) setPath(name string, cmp Component) error {
	switch c := cmp.(type) {
	case *Asset:
		p.assets = append(p.assets, c)
		return nil
	case *Form:
		p.forms = append(p.forms, c)
		return nil
	}
	parts := strings.Split(name, "/")
	if err := cmp.SetPath(name); err != nil {
		return parseError{name, "path", err}
	}

	idx := [2]string{parts[1], name[9:11]}
	if cat, ok := cmp.(*Category); ok {
		p.index[idx] = len(p.categories)
		p.categories = append(p.categories, cat)
		return nil
	}
	cat := p.categories[p.index[idx]]
	if cat == nil {
		return parseError{name, "path", "Invalid cat"}
	}

	if sub, ok := cmp.(*Subcategory); ok {
		cat.Add(sub)
		return nil
	}
	sub := cat.Sub(parts[2])
	if sub == nil {
		return parseError{name, "path", "Invalid sub"}
	}

	if dif, ok := cmp.(*Difficulty); ok {
		sub.AddDifficulty(dif)
		return nil
	}
	dif := sub.Difficulty(parts[3])
	if dif == nil {
		return parseError{name, "path", "Invalid diff"}
	}

	switch c := cmp.(type) {
	case *Item:
		dif.AddItem(c)
	case *Checklist:
		dif.SetChecks(c)
	default:
		return parseError{name, "type", "Invalid Path"}
	}
	return nil
}

func (p *Parser) Categories() map[string][]*Category {
	var res = make(map[string][]*Category)
	for _, cat := range p.categories {
		res[cat.Locale] = append(res[cat.Locale], cat)
	}
	return res
}

func (p *Parser) Assets() []*Asset {
	return p.assets
}

func (p *Parser) Forms() []*Form {
	return p.forms
}
