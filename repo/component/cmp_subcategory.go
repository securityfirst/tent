package component

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Subcategory struct {
	parent *Category
	Id     string  `json:"-"`
	Name   string  `json:"name"`
	Hash   string  `json:"hash"`
	Order  float64 `json:"-"`
	items  []*Item
	checks []*Check
}

func (s *Subcategory) HasChildren() bool {
	return len(s.items) != 0
}

func (s *Subcategory) Tree() interface{} {
	return map[string]interface{}{
		"name":   s.Name,
		"items":  s.items,
		"checks": s.checks,
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
		"items":  s.Items(),
		"checks": s.Checks(),
	}
	if s.Hash != "" {
		m["hash"] = s.Hash
	}
	return json.Marshal(m)
}

func (s *Subcategory) Items() []string {
	var r = make([]string, 0, len(s.items))
	for _, v := range s.items {
		r = append(r, v.Id)
	}
	return r
}

func (s *Subcategory) Checks() []string {
	var r = make([]string, 0, len(s.checks))
	for _, v := range s.checks {
		r = append(r, v.Id)
	}
	return r
}

func (s *Subcategory) AddItem(items ...*Item) {
	for _, v := range items {
		v.parent = s
	}
	s.items = append(s.items, items...)
}

func (s *Subcategory) Item(id string) *Item {
	for _, v := range s.items {
		if v.Id == id {
			return v
		}
	}
	return nil
}

func (s *Subcategory) AddCheck(checks ...*Check) {
	for _, v := range checks {
		v.parent = s
	}
	s.checks = append(s.checks, checks...)
}

func (s *Subcategory) Check(id string) *Check {
	for _, v := range s.checks {
		if v.Id == id {
			return v
		}
	}
	return nil
}

func (s *Subcategory) basePath() string {
	return s.parent.basePath() + "/" + s.Id
}

func (s *Subcategory) Path() string {
	return fmt.Sprintf("%s/.metadata", s.basePath())
}

var subPath = regexp.MustCompile("/contents(?:_[a-z]{2})?/[^/]+/([^/]+)/.metadata")

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
	return setMeta(contents, categoryOrder, args{&s.Name, &s.Order})
}
