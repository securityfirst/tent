package component

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Subcategory struct {
	parent    *Category
	Id        string  `json:"-"`
	Name      string  `json:"name"`
	Hash      string  `json:"hash"`
	Order     float64 `json:"-"`
	items     []*Item
	checklist *Checklist
}

func (s *Subcategory) HasChildren() bool {
	return len(s.items) != 0
}

func (s *Subcategory) Tree(html bool) interface{} {
	var items = make([]Item, len(s.items))
	for i, v := range s.items {
		items[i] = *v
		if html {
			items[i].Body = items[i].htmlBody
		}
	}
	return map[string]interface{}{
		"name":   s.Name,
		"items":  items,
		"checks": s.checklist.Checks,
	}
}

func (s *Subcategory) SHA() string {
	return s.Hash
}

func (s *Subcategory) SetParent(c *Category) {
	s.parent = c
}

func (s *Subcategory) MarshalJSON() ([]byte, error) {
	var m = map[string]interface{}{
		"name":   s.Name,
		"items":  s.ItemNames(),
		"checks": s.checklist.Checks,
	}
	if s.Hash != "" {
		m["hash"] = s.Hash
	}
	return json.Marshal(m)
}

func (s *Subcategory) Items() []Item {
	var dst = make([]Item, len(s.items))
	for i, v := range s.items {
		dst[i] = *v
	}
	return dst
}

func (s *Subcategory) ItemNames() []string {
	var r = make([]string, 0, len(s.items))
	for i := range s.items {
		r = append(r, s.items[i].Id)
	}
	return r
}

func (s *Subcategory) Checks() *Checklist {
	var dst []Check
	if s.checklist == nil {
		s.SetChecks(new(Checklist))
	}
	dst = make([]Check, len(s.checklist.Checks))
	for i, v := range s.checklist.Checks {
		dst[i] = v
	}
	return &Checklist{
		parent: s,
		Hash:   s.checklist.Hash,
		Checks: dst,
	}
}

func (s *Subcategory) AddChecks(c ...Check) {
	if s.checklist == nil {
		s.SetChecks(new(Checklist))
	}
	s.checklist.Add(c...)
}

func (s *Subcategory) SetChecks(c *Checklist) {
	s.checklist = c
	c.parent = s
}

func (s *Subcategory) AddItem(items ...*Item) error {
	for _, v := range items {
		if s.Item(v.Id) != nil {
			return fmt.Errorf("item %s exists in %s/%s", v.Id, s.parent.Id, s.Id)
		}
		v.parent = s
	}
	s.items = append(s.items, items...)
	return nil
}

func (s *Subcategory) Item(id string) *Item {
	for _, v := range s.items {
		if v.Id == id {
			return v
		}
	}
	return nil
}

func (s *Subcategory) basePath() string {
	return fmt.Sprintf("%s/%s", s.parent.basePath(), s.Id)
}

func (s *Subcategory) Path() string {
	return fmt.Sprintf("%s/.metadata%s", s.basePath(), fileExt)
}

var subPath = regexp.MustCompile("/contents(?:_[a-z]{2})?/[^/]+/([^/]+)/.metadata.md")

func (s *Subcategory) SetPath(filepath string) error {
	p := subPath.FindStringSubmatch(filepath)
	if len(p) == 0 {
		return ErrInvalid
	}
	s.Id = p[1]
	return nil
}

func (s *Subcategory) Contents() string {
	return getMeta(categoryOrder, args{s.Name, s.Order})
}

func (s *Subcategory) SetContents(contents string) error {
	if err := checkMeta(contents, categoryOrder); err != nil {
		return err
	}
	if s.checklist == nil {
		s.checklist = new(Checklist)
	}
	return setMeta(contents, categoryOrder, args{&s.Name, &s.Order})
}
