package component

import (
	"errors"
	"fmt"
	"io"
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
	bodySeparator    = "\n---\n"
	prefixName       = "Name:"
	prefixTitle      = "Title:"
	prefixDifficulty = "Difficulty:"
	suffixMeta       = ".metadata"
)

type Component interface {
	Path() string
	Contents() string
	SetPath(path string) error
	SetContents(contents string) error
	SHA() string
	HasChildren() bool
}

func New(path string) (Component, error) {
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

func ParseTree(iter *git.FileIter) (map[string]*Category, error) {
	m := make(map[string]*Category)
	for {
		f, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if err := parseFile(m, f); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func parseFile(m map[string]*Category, f *git.File) error {
	contents, err := f.Contents()
	if err != nil {
		return err
	}
	cmp, err := New("/" + f.Name)
	if err != nil {
		return err
	}
	p := strings.Split(f.Name, "/")
	switch t := cmp.(type) {
	case *Category:
		m[p[0]] = t
	case *Subcategory:
		m[p[0]].Add(t)
	case *Item:
		m[p[0]].Sub(p[1]).Add(t)
	default:
		return fmt.Errorf("%s - Invalid Path", f.Name)
	}
	if err := cmp.SetPath("/" + f.Name); err != nil {
		return fmt.Errorf("%s - Path: %s", f.Name, err)
	}
	if err := cmp.SetContents(contents); err != nil {
		return fmt.Errorf("%s - Content: %s", f.Name, err)
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
