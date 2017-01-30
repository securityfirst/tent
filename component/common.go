package component

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
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
	bodySeparator = "\n\n"
	suffixMeta    = ".metadata"
	suffixChecks  = ".checks"
	fileExt       = ".md"
)

var (
	categoryOrder = []string{"Name", "Order"}
	itemOrder     = []string{"Title", "Difficulty", "Order"}
	checkOrder    = []string{"Text", "Difficulty", "NoCheck"}
)

// A Component is en element of the resource tree
type Component interface {
	Path() string
	Contents() string
	SetPath(path string) error
	SetContents(contents string) error
	SHA() string
	HasChildren() bool
}

func newCmp(path string) (Component, error) {
	p := strings.Split(path, "/")
	switch l := len(p); l {
	case 3:
		if isImage(p[2]) {
			return new(Asset), nil
		}
		return nil, ErrInvalid
	case 4:
		return new(Category), nil
	case 5:
		if !IsMd(p[4]) {
			return nil, ErrInvalid
		}
		switch p[4][:len(p[4])-len(fileExt)] {
		case suffixMeta:
			return new(Subcategory), nil
		case suffixChecks:
			return new(Checklist), nil
		default:
			return new(Item), nil
		}
	default:
		return nil, ErrInvalid
	}
}

// TreeParser is an helper, creates a tree from the repo
type TreeParser struct {
	index      map[string]int
	Categories []*Category
	Assets     []*Asset
}

// Parse executes the parsing on a repo
func (t *TreeParser) parse(tree *git.Tree, fn func(name string) bool) error {
	for iter := tree.Files(); ; {
		f, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if !fn(f.Name) {
			continue
		}
		if err := t.parseFile(f); err != nil {
			return err
		}
	}
	return nil
}

func (t *TreeParser) filterCat(name string) bool {
	return strings.HasSuffix(name, suffixMeta+fileExt)
}

func (t *TreeParser) filterRes(name string) bool {
	if !strings.HasSuffix(name, fileExt) {
		return isImage(name)
	}
	return !strings.HasSuffix(name, suffixMeta+fileExt)
}

func IsMd(name string) bool { return filepath.Ext(name) == fileExt }

func isImage(name string) bool {
	ext := filepath.Ext(name)
	for _, v := range []string{".jpg", ".jpeg", ".gif", ".png", ".bmp"} {
		if v == ext {
			return true
		}
	}
	return false
}

// Parse executes the parsing on a repo
func (t *TreeParser) Parse(tree *git.Tree) error {
	t.index = make(map[string]int)
	t.Categories = make([]*Category, 0)
	if err := t.parse(tree, t.filterCat); err != nil {
		return err
	}
	if err := t.parse(tree, t.filterRes); err != nil {
		return err
	}
	sort.Sort(catSorter(t.Categories))
	for i := range t.Categories {
		sort.Sort(subSorter(t.Categories[i].subcategories))
		for j := range t.Categories[i].subcategories {
			sort.Sort(itemSorter(t.Categories[i].subcategories[j].items))
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
	if err := cmp.SetPath("/" + f.Name); err != nil {
		return parseError{f.Name, "path", err}
	}
	switch c := cmp.(type) {
	case *Category:
		t.index[p[1]] = len(t.Categories)
		t.Categories = append(t.Categories, c)
	case *Subcategory:
		t.Categories[t.index[p[1]]].Add(c)
	case *Item:
		t.Categories[t.index[p[1]]].Sub(p[2]).AddItem(c)
	case *Checklist:
		t.Categories[t.index[p[1]]].Sub(p[2]).SetChecks(c)
	case *Asset:
		t.Assets = append(t.Assets, c)
	default:
		return parseError{f.Name, "type", "Invalid Path"}
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
