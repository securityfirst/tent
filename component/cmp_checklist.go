package component

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Checklist struct {
	parent *Difficulty
	Hash   string  `json:"hash"`
	Checks []Check `json:"checks"`
}

func (c *Checklist) Resource() Resource {
	var content = make([]map[string]string, 0, len(c.Checks))
	for _, c := range c.Checks {
		content = append(content, map[string]string{
			"text": c.Text,
		})
	}
	return Resource{
		Slug:    c.parent.Resource().Slug + "_" + "_checks",
		Content: content,
	}
}

type Check struct {
	Text    string `json:"text"`
	NoCheck bool   `json:"no_check"`
}

func (c *Check) order() []string { return []string{"Text", "NoCheck"} }
func (c *Check) pointers() args  { return args{&c.Text, &c.NoCheck} }
func (c *Check) values() args    { return args{c.Text, c.NoCheck} }

func (c *Checklist) SetParent(d *Difficulty) {
	c.parent = d
}

func (c *Checklist) HasChildren() bool {
	return len(c.Checks) != 0
}

func (c *Checklist) SHA() string {
	return c.Hash
}

func (c *Checklist) Path() string {
	return fmt.Sprintf("%s/.checks%s", c.parent.basePath(), fileExt)
}

var checklistPath = regexp.MustCompile("contents(?:_[a-z]{2})?/[^/]+/[^/]+/[^/]+/.checks.md")

func (*Checklist) SetPath(filepath string) error {
	p := checklistPath.FindString(filepath)
	if len(p) == 0 {
		return ErrContent
	}
	return nil
}

func (c *Checklist) Contents() string {
	b := bytes.NewBuffer(nil)
	for i := range c.Checks {
		if i != 0 {
			fmt.Fprint(b, bodySeparator)
		}
		fmt.Fprint(b, getMeta(&c.Checks[i]))
	}
	return b.String()
}

func (c *Checklist) SetContents(contents string) error {
	if contents == "" {
		if c.Checks != nil {
			c.Checks = c.Checks[:0]
		}
		return nil
	}
	parts := strings.Split(contents, bodySeparator)
	var checks = make([]Check, len(parts))
	for i, v := range parts {
		if err := checkMeta(v, &checks[i]); err != nil {
			return err
		}
		setMeta(v, &checks[i])
	}
	c.Checks = checks
	return nil
}

func (c *Checklist) Add(v ...Check) { c.Checks = append(c.Checks, v...) }
