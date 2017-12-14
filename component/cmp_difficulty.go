package component

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Difficulty struct {
	parent    *Subcategory
	ID        string `json:"id"`
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	items     []*Item
	checklist *Checklist
}

func (d *Difficulty) Resource() Resource {
	return Resource{
		Slug: d.parent.Resource().Slug + "_" + d.ID,
		Content: []map[string]string{
			map[string]string{"name": d.Name},
		},
	}
}

func (d *Difficulty) HasChildren() bool {
	return len(d.items) != 0
}

func (d *Difficulty) Tree(html bool) interface{} {
	var items = make([]Item, len(d.items))
	for i, v := range d.items {
		items[i] = *v
		if html {
			items[i].Body = items[i].htmlBody
		}
	}
	return map[string]interface{}{
		"id":     d.ID,
		"name":   d.Name,
		"items":  items,
		"checks": d.checklist.Checks,
	}
}

func (d *Difficulty) SHA() string {
	return d.Hash
}

func (d *Difficulty) SetParent(s *Subcategory) {
	d.parent = s
}

func (d *Difficulty) MarshalJSON() ([]byte, error) {
	var m = map[string]interface{}{
		"name":   d.Name,
		"items":  d.ItemNames(),
		"checks": d.checklist.Checks,
	}
	if d.Hash != "" {
		m["hash"] = d.Hash
	}
	return json.Marshal(m)
}

func (d *Difficulty) Items() []Item {
	var dst = make([]Item, len(d.items))
	for i, v := range d.items {
		dst[i] = *v
	}
	return dst
}

func (d *Difficulty) ItemNames() []string {
	var r = make([]string, 0, len(d.items))
	for i := range d.items {
		r = append(r, d.items[i].ID)
	}
	return r
}

func (d *Difficulty) Checks() *Checklist {
	var dst []Check
	if d.checklist == nil {
		d.SetChecks(new(Checklist))
	}
	dst = make([]Check, len(d.checklist.Checks))
	for i, v := range d.checklist.Checks {
		dst[i] = v
	}
	return &Checklist{
		parent: d,
		Hash:   d.checklist.Hash,
		Checks: dst,
	}
}

func (d *Difficulty) AddChecks(c ...Check) {
	if d.checklist == nil {
		d.SetChecks(new(Checklist))
	}
	d.checklist.Add(c...)
}

func (d *Difficulty) SetChecks(c *Checklist) {
	d.checklist = c
	c.parent = d
}

func (d *Difficulty) AddItem(items ...*Item) error {
	for _, v := range items {
		if d.Item(v.ID) != nil {
			return fmt.Errorf("item %s exists in %s/%s", v.ID, d.parent.ID, d.ID)
		}
		v.parent = d
	}
	d.items = append(d.items, items...)
	return nil
}

func (d *Difficulty) Item(id string) *Item {
	for _, v := range d.items {
		if v.ID == id {
			return v
		}
	}
	return nil
}

func (d *Difficulty) basePath() string {
	return fmt.Sprintf("%s/%s", d.parent.basePath(), d.ID)
}

func (d *Difficulty) Path() string {
	return fmt.Sprintf("%s/.metadata%s", d.basePath(), fileExt)
}

var diffPath = regexp.MustCompile("contents(?:_[a-z]{2})?/[^/]+/[^/]+/([^/]+)/.metadata.md")

func (d *Difficulty) SetPath(filepath string) error {
	p := diffPath.FindStringSubmatch(filepath)
	if len(p) == 0 {
		return ErrContent
	}
	d.ID = p[1]
	return nil
}

func (d *Difficulty) order() []string { return []string{"Name"} }
func (d *Difficulty) pointers() args  { return args{&d.Name} }
func (d *Difficulty) values() args    { return args{d.Name} }

func (d *Difficulty) Contents() string { return getMeta(d) }

func (d *Difficulty) SetContents(contents string) error {
	if err := checkMeta(contents, d); err != nil {
		return err
	}
	if d.checklist == nil {
		d.checklist = new(Checklist)
	}
	return setMeta(contents, d)
}
