package component

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

func splitSlug(s string) []string {
	s = strings.Replace(s, "_", "|", -1)
	s = strings.Replace(s, "|||", "|_|", -1)
	return strings.Split(s, "|")
}

func NewResourceParser() *ResourceParser {
	return &ResourceParser{index: make(map[[2]string]int), categories: make([]*Category, 0)}
}

type ResourceParser struct {
	buffer     bytes.Buffer
	index      map[[2]string]int
	categories []*Category
}

func (r *ResourceParser) add(c *Category) {
	r.index[[2]string{c.Id, c.Locale}] = len(r.categories)
	r.categories = append(r.categories, c)
}

func (r *ResourceParser) get(id, locale string) *Category {
	if len(r.index) == 0 {
		return nil
	}
	idx := r.index[[2]string{id, locale}]
	return r.categories[idx]
}

func (r *ResourceParser) Categories() map[string][]*Category {
	var res = make(map[string][]*Category)
	for _, cat := range r.categories {
		res[cat.Locale] = append(res[cat.Locale], cat)
	}
	return res
}

func (r *ResourceParser) Parse(cmp Component, res *Resource, locale string) error {
	switch v := cmp.(type) {
	case *Category:
		return r.parseCategory(v, res, locale)
	case *Subcategory:
		return r.parseSubcategory(v, res, locale)
	case *Item:
		return r.parseItem(v, res, locale)
	case *Checklist:
		return r.parseChecklist(v, res, locale)
	default:
		return errors.New("Invalid Component")
	}
}

func (r *ResourceParser) parseCategory(c *Category, res *Resource, locale string) error {
	if len(res.Content) != 1 {
		return ErrContent
	}
	r.add(&Category{
		Id:     c.Id,
		Order:  c.Order,
		Name:   res.Content[0]["name"],
		Locale: locale,
	})
	return nil
}

func (r *ResourceParser) parseSubcategory(s *Subcategory, res *Resource, locale string) error {
	if len(res.Content) != 1 {
		return ErrContent
	}
	cat := r.get(s.parent.Id, locale)
	if cat == nil {
		return fmt.Errorf("No cat %q (%s)", s.parent.Id, locale)
	}
	cat.Add(&Subcategory{
		Id:    s.Id,
		Order: s.Order,
		Name:  res.Content[0]["name"],
	})
	return nil
}

func (r *ResourceParser) parseItem(i *Item, res *Resource, locale string) error {
	if len(res.Content) == 0 {
		return ErrContent
	}
	cat := r.get(i.parent.parent.Id, locale)
	if cat == nil {
		return fmt.Errorf("No cat %q (%s)", i.parent.parent.Id, locale)
	}
	sub := cat.Sub(i.parent.Id)
	if sub == nil {
		return fmt.Errorf("No sub %q (%s)", i.parent.Id, locale)
	}
	item := &Item{
		Id:         i.Id,
		Difficulty: strings.TrimSpace(res.Content[0]["difficulty"]),
		Title:      strings.TrimSpace(res.Content[0]["title"]),
		Order:      i.Order,
	}
	r.buffer.Reset()
	// Old Version Compatibility
	if res.Content[0]["body"] != "" {
		if len(res.Content) != 1 {
			return fmt.Errorf("Invalid Legacy %q (%s)", i.parent.Id, locale)
		}
		r.buffer.WriteString(strings.TrimSpace(res.Content[0]["body"]))
	} else {
		for _, v := range res.Content[1:] {
			if r.buffer.Len() != 0 {
				r.buffer.WriteString(paragraphSep)
			}
			r.buffer.WriteString(strings.TrimSpace(v["body"]))
		}
	}
	item.Body = r.buffer.String()
	sub.AddItem(item)
	return nil
}

func (r *ResourceParser) parseChecklist(c *Checklist, res *Resource, locale string) error {
	for len(res.Content) > 0 && res.Content[0] == nil {
		res.Content = res.Content[1:]
	}
	if l, e := len(res.Content), len(c.Checks); l != e {
		return fmt.Errorf("%d checks, %s expected", l, e)
	}
	var checks Checklist
	for i, r := range res.Content {
		checks.Add(Check{
			Difficulty: strings.TrimSpace(r["difficulty"]),
			Text:       strings.TrimSpace(r["text"]),
			NoCheck:    c.Checks[i].NoCheck,
		})
	}
	cat := r.get(c.parent.parent.Id, locale)
	if cat == nil {
		return fmt.Errorf("No cat %q (%s)", c.parent.parent.Id, locale)
	}
	sub := cat.Sub(c.parent.Id)
	if sub == nil {
		return fmt.Errorf("No sub %q (%s)", c.parent.Id, locale)
	}
	sub.SetChecks(&checks)
	return nil
}
