package component

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

type Subcategory struct {
	parent *Category
	Id     string `json:"-"`
	Name   string `json:"name"`
	Hash   string `json:"hash"`
	items  []*Item
	checks []*Check
}

func (s *Subcategory) HasChildren() bool {
	return len(s.items) != 0
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

func (s *Subcategory) Path() string {
	return fmt.Sprintf("/%s/%s/.metadata", s.parent.Id, s.Id)
}

func (s *Subcategory) Contents() string {
	return getMeta(categoryOrder, args{s.Name})
}

func (s *Subcategory) SetPath(filepath string) error {
	if p := strings.Split(filepath, "/"); len(p) != 4 || p[0] != "" || p[3] != suffixMeta {
		return ErrInvalid
	}
	s.Id = path.Base(path.Dir(filepath))
	return nil
}

func (s *Subcategory) SetContents(contents string) error {
	if err := checkMeta(contents, categoryOrder); err != nil {
		return err
	}
	return setMeta(contents, categoryOrder, args{&s.Name})
}
