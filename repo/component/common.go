package component

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/src-d/go-git.v3"
)

const (
	actionCreate = iota
	actionUpdate
	actionDelete
)

var commitMsg = map[int]string{
	actionCreate: "Create",
	actionUpdate: "Update",
	actionDelete: "Delete",
}

var ErrInvalid = errors.New("Invalid content")

const (
	bodySeparator = "\n---\n"
	suffixMeta    = ".metadata"
)

var (
	categoryOrder = []string{"Name:", "Order:"}
	itemOrder     = []string{"Title:", "Difficulty:", "Order:"}
	checkOrder    = []string{"Title:", "Text:", "Difficulty:", "NoCheck:", "Order:"}
)

type Component interface {
	Path() string
	Contents() string
	SetPath(path string) error
	SetContents(contents string) error
	SHA() string
	HasChildren() bool
}

func newCmp(path string) (Component, error) {
	var c Component
	p := strings.Split(path, "/")
	switch l := len(p); l {
	case 4:
		c = &Category{}
	case 5:
		if p[4] == suffixMeta {
			c = &Subcategory{}
		} else {
			c = &Item{}
		}
	case 6:
		c = &Check{}
	default:
		return nil, ErrInvalid
	}
	return c, c.SetPath(path)
}

type TreeParser struct {
	index      map[string]int
	Categories []*Category
}

func (t *TreeParser) Parse(iter *git.FileIter) error {
	t.index = make(map[string]int)
	t.Categories = make([]*Category, 0)
	for {
		f, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if f.Name == "LICENSE" || strings.ToLower(f.Name) == "readme.md" {
			continue
		}
		if err := t.parseFile(f); err != nil {
			return err
		}
	}
	sort.Sort(catSorter(t.Categories))
	for i := range t.Categories {
		sort.Sort(subSorter(t.Categories[i].subcategories))
		for j := range t.Categories[i].subcategories {
			sort.Sort(itemSorter(t.Categories[i].subcategories[j].items))
			sort.Sort(checkSorter(t.Categories[i].subcategories[j].checks))
		}
	}
	return nil
}

type parseError struct {
	file  string
	phase string
	err   interface{}
}

func (p parseError) Error() string { return fmt.Sprintf("[%s]%s - %v", p.phase, p.file, p.err) }

func (t *TreeParser) parseFile(f *git.File) error {
	contents, err := f.Contents()
	if err != nil {
		return parseError{f.Name, "read", err}
	}
	cmp, err := newCmp("/" + f.Name)
	if err != nil {
		return parseError{f.Name, "cmp", err}
	}
	p := strings.Split(f.Name, "/")
	switch c := cmp.(type) {
	case *Category:
		t.index[p[1]] = len(t.Categories)
		t.Categories = append(t.Categories, c)
	case *Subcategory:
		t.Categories[t.index[p[1]]].Add(c)
	case *Item:
		t.Categories[t.index[p[1]]].Sub(p[2]).AddItem(c)
	case *Check:
		t.Categories[t.index[p[1]]].Sub(p[2]).AddCheck(c)
	default:
		return parseError{f.Name, "type", "Invalid Path"}
	}
	if err := cmp.SetPath("/" + f.Name); err != nil {
		return parseError{f.Name, "path", err}
	}
	if err := cmp.SetContents(contents); err != nil {
		return parseError{f.Name, "contents", err}
	}
	return nil
}

func strPtr(s string) *string { return &s }

func repoAddress(owner, name string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, name)
}

func uploadAddress(owner, name, file string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, name, file)
}

func checkMeta(meta string, order []string) error {
	rows := strings.Split(meta, "\n")
	if len(rows) != len(order) {
		return ErrInvalid
	}
	for i := range order {
		if !strings.HasPrefix(rows[i], order[i]) {
			return ErrInvalid
		}
	}
	return nil
}

type args []interface{}

func setMeta(meta string, order []string, pointers args) error {
	rows := strings.Split(strings.TrimSpace(meta), "\n")
	if len(rows) != len(order) {
		return ErrInvalid
	}
	for i, p := range pointers {
		v := rows[i][len(order[i]):]
		switch pointer := p.(type) {
		case *string:
			*pointer = v
		case *bool:
			*pointer = v == "true"
		case *int:
			n, err := strconv.Atoi(v)
			if err != nil {
				return ErrInvalid
			}
			*pointer = n
		case *float64:
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return ErrInvalid
			}
			*pointer = n
		default:
			panic(fmt.Sprintf("Unknown type: %T", pointer))
		}
	}
	return nil
}

func getMeta(order []string, values args) string {
	b := bytes.NewBuffer(nil)
	for i := range values {
		if i > 0 {
			b.WriteRune('\n')
		}
		b.WriteString(order[i])
		fmt.Fprint(b, values[i])
	}
	return b.String()
}
