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
	return &ResourceParser{
		categories: make(map[string][]*Category),
		forms:      make(map[string][]*Form),
	}
}

type ResourceParser struct {
	buffer     bytes.Buffer
	categories map[string][]*Category
	forms      map[string][]*Form
}

func (r *ResourceParser) Categories() map[string][]*Category { return r.categories }

func (r *ResourceParser) Parse(cmp Component, res *Resource, locale string) error {
	switch v := cmp.(type) {
	case *Form:
		return r.parseForm(v, res, locale)
	case *Category:
		return r.parseCategory(v, res, locale)
	case *Subcategory:
		return r.parseSubcategory(v, res, locale)
	case *Difficulty:
		return r.parseDifficulty(v, res, locale)
	case *Item:
		return r.parseItem(v, res, locale)
	case *Checklist:
		return r.parseChecklist(v, res, locale)
	default:
		return errors.New("Invalid Component")
	}
}

func (r *ResourceParser) parseForm(f *Form, res *Resource, locale string) error {
	var newForm = Form{
		ID:      f.ID,
		Name:    res.Content[0]["form"],
		Locale:  locale,
		Screens: make([]FormScreen, len(f.Screens)),
	}
	m := res.Content[1:]
	for i := range newForm.Screens {
		screen := &newForm.Screens[i]
		screen.Items = make([]FormInput, len(f.Screens[i].Items))
		if f.Screens[i].Name != "" {
			if len(m) == 0 {
				return fmt.Errorf("No more at screen %d/%d", i+1, len(f.Screens))
			}
			if name := m[0]["screen"]; name != "" {
				screen.Name = name
				m = m[1:]
			} else {
				return fmt.Errorf("Expected screen %d, got item", i)
			}
		}
		for j := range screen.Items {
			item := screen.Items[j]
			if item.Label == "" && item.Hint == "" && item.Options == nil {
				continue
			}
			if s := m[0]["screen"]; s != "" {
				return fmt.Errorf("Expected item %d/%d, got screen %q", i, j, s)
			}
			item.Label, item.Hint = m[0]["label"], m[0]["hint"]
			item.Options = strings.Split(m[0]["options"], ";")
			m = m[1:]
		}
	}
	r.forms[newForm.Locale] = append(r.forms[newForm.Locale], &newForm)
	return nil
}

func (r *ResourceParser) getCategory(cat *Category, locale string) *Category {
	for _, c := range r.categories[locale] {
		if c.ID == cat.ID {
			return c
		}
	}
	c := Category{ID: cat.ID, Order: cat.Order, Locale: locale}
	r.categories[locale] = append(r.categories[locale], &c)
	return &c
}

func (r *ResourceParser) parseCategory(c *Category, res *Resource, locale string) error {
	if len(res.Content) != 1 {
		return ErrContent
	}
	if cat := r.getCategory(c, locale); cat != nil {
		cat.Name = res.Content[0]["name"]
		return nil
	}
	r.categories[locale] = append(r.categories[locale], &Category{
		ID:     c.ID,
		Order:  c.Order,
		Name:   res.Content[0]["name"],
		Locale: locale,
	})
	return nil
}

func (r *ResourceParser) getSubcategory(sub *Subcategory, locale string) *Subcategory {
	cat := r.getCategory(sub.parent, locale)
	for _, s := range cat.subcategories {
		if s.ID == sub.ID {
			return s
		}
	}
	s := Subcategory{ID: sub.ID, Order: sub.Order}
	cat.Add(&s)
	return &s
}

func (r *ResourceParser) parseSubcategory(s *Subcategory, res *Resource, locale string) error {
	if len(res.Content) != 1 {
		return ErrContent
	}
	sub := r.getSubcategory(s, locale)
	sub.Name = res.Content[0]["name"]
	return nil
}

func (r *ResourceParser) getDifficulty(diff *Difficulty, locale string) *Difficulty {
	sub := r.getSubcategory(diff.parent, locale)
	for _, d := range sub.difficulties {
		if d.ID == diff.ID {
			return d
		}
	}
	d := Difficulty{ID: diff.ID}
	sub.AddDifficulty(&d)
	return &d
}

func (r *ResourceParser) parseDifficulty(d *Difficulty, res *Resource, locale string) error {
	if len(res.Content) != 1 {
		return ErrContent
	}
	diff := r.getDifficulty(d, locale)
	diff.Descr = res.Content[0]["description"]
	return nil
}

func (r *ResourceParser) parseItem(i *Item, res *Resource, locale string) error {
	if len(res.Content) == 0 {
		return ErrContent
	}
	item := &Item{
		ID:    i.ID,
		Title: strings.TrimSpace(res.Content[0]["title"]),
		Order: i.Order,
	}
	r.buffer.Reset()
	// Old Verion Compatibility
	if res.Content[0]["body"] != "" {
		if len(res.Content) != 1 {
			return fmt.Errorf("Invalid Legacy %q (%s)", i.parent.ID, locale)
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
	r.getDifficulty(i.parent, locale).AddItem(item)
	return nil
}

func (r *ResourceParser) parseChecklist(c *Checklist, res *Resource, locale string) error {
	for len(res.Content) > 0 && res.Content[0] == nil {
		res.Content = res.Content[1:]
	}
	if l, e := len(res.Content), len(c.Checks); l != e {
		return fmt.Errorf("%d checks, %d expected", l, e)
	}

	var checks Checklist
	for i, r := range res.Content {
		checks.Add(Check{
			Text:    strings.TrimSpace(r["text"]),
			NoCheck: c.Checks[i].NoCheck,
		})
	}
	r.getDifficulty(c.parent, locale).SetChecks(&checks)
	return nil
}
