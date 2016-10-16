package component

import (
	"bytes"
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
	b := bytes.NewBuffer(nil)
	b.WriteString(prefixName)
	b.WriteString(c.Name)
	return b.String()
}

func (c *Category) SetPath(filepath string) error {
	if p := strings.Split(filepath, "/"); len(p) != 3 || p[0] != "" || p[2] != suffixMeta {
		return ErrInvalid
	}
	c.Id = path.Base(path.Dir(filepath))
	return nil
}

func (c *Category) SetContents(contents string) error {
	meta := strings.Split(strings.Trim(contents, "\n"), "\n")
	if len(meta) != 1 || !strings.HasPrefix(meta[0], prefixName) {
		return ErrInvalid
	}
	c.Name = meta[0][len(prefixName):]
	return nil
}
