package component

import (
	"bytes"
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
		"name":  s.Name,
		"items": s.Items(),
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

func (s *Subcategory) Add(items ...*Item) {
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

func (s *Subcategory) Path() string {
	return fmt.Sprintf("/%s/%s/.metadata", s.parent.Id, s.Id)
}

func (s *Subcategory) Contents() string {
	b := bytes.NewBuffer(nil)
	b.WriteString(prefixName)
	b.WriteString(s.Name)
	return b.String()
}

func (s *Subcategory) SetPath(filepath string) error {
	if p := strings.Split(filepath, "/"); len(p) != 4 || p[0] != "" || p[3] != suffixMeta {
		return ErrInvalid
	}
	s.Id = path.Base(path.Dir(filepath))
	return nil
}

func (s *Subcategory) SetContents(contents string) error {
	meta := strings.Split(strings.Trim(contents, "\n"), "\n")
	if len(meta) != 1 || !strings.HasPrefix(meta[0], prefixName) {
		return ErrInvalid
	}
	s.Name = meta[0][len(prefixName):]
	return nil
}
