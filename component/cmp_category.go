package component

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Category struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Hash          string  `json:"hash"`
	Locale        string  `json:"-"`
	Order         float64 `json:"-"`
	subcategories []*Subcategory
}

func (c *Category) Resource() Resource {
	return Resource{
		Slug: c.ID,
		Content: []map[string]string{
			map[string]string{"name": c.Name},
		},
	}
}

func (c *Category) HasChildren() bool {
	return len(c.subcategories) != 0
}

func (c *Category) SHA() string {
	return c.Hash
}

func (c *Category) Tree(html bool) interface{} {
	var subs = make([]interface{}, 0, len(c.subcategories))
	for i := range c.subcategories {
		subs = append(subs, c.subcategories[i].Tree(html))
	}
	return map[string]interface{}{
		"id":            c.ID,
		"name":          c.Name,
		"subcategories": subs,
	}
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

func (c *Category) Sub(ID string) *Subcategory {
	for _, v := range c.subcategories {
		if v.ID == ID {
			return v
		}
	}
	return nil
}

func (c *Category) Subcategories() []string {
	var r = make([]string, 0, len(c.subcategories))
	for _, v := range c.subcategories {
		r = append(r, v.ID)
	}
	return r
}

func (c *Category) Add(subs ...*Subcategory) {
	for _, v := range subs {
		v.parent = c
	}
	c.subcategories = append(c.subcategories, subs...)
}

func (c *Category) basePath() string {
	var loc string
	if c.Locale != "" {
		loc = "_" + c.Locale
	}
	return fmt.Sprintf("contents%s/%s", loc, c.ID)
}

func (c *Category) Path() string {
	return fmt.Sprintf("%s/%s%s", c.basePath(), suffixMeta, fileExt)
}

var catPath = regexp.MustCompile("contents_([a-z]{2})/([^/]+)/.metadata.md")

func (c *Category) SetPath(filepath string) error {
	p := catPath.FindStringSubmatch(filepath)
	if len(p) == 0 {
		return ErrContent
	}
	c.Locale = p[1]
	c.ID = p[2]
	return nil
}

func (c *Category) order() []string { return []string{"Name", "Order"} }
func (c *Category) pointers() args  { return args{&c.Name, &c.Order} }
func (c *Category) values() args    { return args{c.Name, c.Order} }

func (c *Category) Contents() string {
	return getMeta(c)
}

func (c *Category) SetContents(contents string) error {
	if err := checkMeta(contents, c); err != nil {
		return err
	}
	return setMeta(contents, c)
}
