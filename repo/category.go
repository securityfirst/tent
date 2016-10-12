package repo

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strings"
)

var ErrInvalid = errors.New("Invalid content")

const (
	bodySeparator    = "\n---\n"
	prefixName       = "Name:"
	prefixTitle      = "Title:"
	prefixDifficulty = "Difficulty:"
	suffixMeta       = ".metadata"
)

func NewComponent(path string) (Component, error) {
	var c Component
	p := strings.Split(path, "/")
	switch l := len(p); l {
	case 3:
		c = &Category{}
	case 4:
		if p[3] == suffixMeta {
			c = &Subcategory{}
		} else {
			c = &Item{}
		}
	default:
		return nil, ErrInvalid
	}
	return c, c.SetPath(path)
}

type Component interface {
	Path() string
	Contents() string
	SetPath(path string) error
	SetContents(contents string) error
}

type Category struct {
	Id            string
	Name          string
	subcategories []*Subcategory
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
	meta := strings.Split(contents, "\n")
	if len(meta) != 1 || !strings.HasPrefix(meta[0], prefixName) {
		return ErrInvalid
	}
	c.Name = meta[0][len(prefixName):]
	return nil
}

type Subcategory struct {
	parent *Category
	Id     string
	Name   string
	items  []*Item
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
	meta := strings.Split(contents, "\n")
	if len(meta) != 1 || !strings.HasPrefix(meta[0], prefixName) {
		return ErrInvalid
	}
	s.Name = meta[0][len(prefixName):]
	return nil
}

type Item struct {
	parent     *Subcategory
	Id         string
	Difficulty string
	Title      string
	Body       string
}

func (i *Item) Path() string {
	return fmt.Sprintf("/%s/%s/%s", i.parent.parent.Id, i.parent.Id, i.Id)
}

func (i *Item) Contents() string {
	b := bytes.NewBuffer(nil)
	b.WriteString(prefixTitle)
	b.WriteString(i.Title)
	b.WriteRune('\n')
	b.WriteString(prefixDifficulty)
	b.WriteString(i.Difficulty)
	b.WriteString(bodySeparator)
	b.WriteString(i.Body)
	return b.String()
}

func (i *Item) SetPath(filepath string) error {
	if p := strings.Split(filepath, "/"); len(p) != 4 || p[0] != "" || p[3] == suffixMeta {
		return ErrInvalid
	}
	i.Id = path.Base(filepath)
	return nil
}

func (i *Item) SetContents(contents string) error {
	parts := strings.Split(contents, bodySeparator)
	if len(parts) != 2 {
		return ErrInvalid
	}
	meta := strings.Split(parts[0], "\n")
	if len(meta) != 2 || !strings.HasPrefix(meta[0], prefixTitle) || !strings.HasPrefix(meta[1], prefixDifficulty) {
		return ErrInvalid
	}
	i.Title = meta[0][len(prefixTitle):]
	i.Difficulty = meta[1][len(prefixDifficulty):]
	i.Body = parts[1]
	return nil
}
