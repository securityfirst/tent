package component

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

type Category struct {
	Id            string `json:"-"`
	Name          string `json:"name"`
	Hash          string `json:"hash"`
	subcategories []*Subcategory
}

func (c *Category) HasChildren() bool {
	return len(c.subcategories) != 0
}

func (c *Category) SHA() string {
	return c.Hash
}

func (c *Category) MarshalJSON() ([]byte, error) {
	var m = map[string]interface{}{
		"name":          c.Name,
		"subcategories": c.Subcategories(),
	}
	if c.Hash != "" {
		m["hash"] = c.Hash
	}
	return json.Marshal(m)
}

func (c *Category) Sub(id string) *Subcategory {
	for _, v := range c.subcategories {
		if v.Id == id {
			return v
		}
	}
	return nil
}

func (c *Category) Subcategories() []string {
	var r = make([]string, 0, len(c.subcategories))
	for _, v := range c.subcategories {
		r = append(r, v.Id)
	}
	return r
}

func (c *Category) Add(subs ...*Subcategory) {
	for _, v := range subs {
		v.parent = c
	}
	c.subcategories = append(c.subcategories, subs...)
}

func (c *Category) Path() string {
	return fmt.Sprintf("/%s/%s", c.Id, suffixMeta)
}

func (c *Category) Contents() string {
	return getMeta(categoryOrder, args{c.Name})
}

func (c *Category) SetPath(filepath string) error {
	if p := strings.Split(filepath, "/"); len(p) != 3 || p[0] != "" || p[2] != suffixMeta {
		return ErrInvalid
	}
	c.Id = path.Base(path.Dir(filepath))
	return nil
}

func (c *Category) SetContents(contents string) error {
	if err := checkMeta(contents, categoryOrder); err != nil {
		return err
	}
	return setMeta(contents, categoryOrder, args{&c.Name})
}
