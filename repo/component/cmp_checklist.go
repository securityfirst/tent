package component

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Checklist struct {
	parent *Subcategory
	Hash   string  `json:"hash"`
	Checks []Check `json:"checks"`
}

type Check struct {
	Difficulty string `json:"difficulty"`
	Text       string `json:"text"`
	NoCheck    bool   `json:"no_check"`
}

func (c *Checklist) SetParent(s *Subcategory) {
	c.parent = s
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

var checklistPath = regexp.MustCompile("/contents(?:_[a-z]{2})?/[^/]+/[^/]+/.checks.md")

func (*Checklist) SetPath(filepath string) error {
	p := checklistPath.FindString(filepath)
	if len(p) == 0 {
		return ErrInvalid
	}
	return nil
}

func (c *Checklist) Contents() string {
	b := bytes.NewBuffer(nil)
	for i, v := range c.Checks {
		if i != 0 {
			fmt.Fprint(b, bodySeparator)
		}
		fmt.Fprint(b, getMeta(checkOrder, args{v.Text, v.Difficulty, v.NoCheck}))
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
		if err := checkMeta(v, checkOrder); err != nil {
			return err
		}
		setMeta(v, checkOrder, args{&checks[i].Text, &checks[i].Difficulty, &checks[i].NoCheck})
	}
	c.Checks = checks
	return nil
}

func (c *Checklist) Add(v ...Check) { c.Checks = append(c.Checks, v...) }
