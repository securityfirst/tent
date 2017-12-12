package component

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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

var (
	ErrContent = errors.New("Invalid content")
	ErrSlug    = errors.New("Invalid Slug")
)

const (
	bodySeparator = "\n\n"
	suffixMeta    = ".metadata"
	suffixChecks  = ".checks"
	fileExt       = ".md"
)

// A Component is en element of the resource tree
type Component interface {
	Path() string
	Contents() string
	SetPath(path string) error
	SetContents(contents string) error
	SHA() string
	HasChildren() bool
	Resource() Resource
}

type Resource struct {
	Slug    string
	Content []map[string]string
}

func newCmp(path string) (Component, error) {
	p := strings.Split(path, "/")
	switch l := len(p); l {
	case 2:
		if formPath.MatchString(path) {
			return new(Form), nil
		}
		if p[0] == "assets" && isImage(p[1]) {
			return new(Asset), nil
		}
	case 3:
		if strings.TrimSuffix(p[2], fileExt) == suffixMeta {
			return new(Category), nil
		}
	case 4:
		if strings.TrimSuffix(p[3], fileExt) == suffixMeta {
			return new(Subcategory), nil
		}
	case 5:
		switch strings.TrimSuffix(p[4], fileExt) {
		case suffixMeta:
			return new(Difficulty), nil
		case suffixChecks:
			return new(Checklist), nil
		default:
			if isMd(p[4]) {
				return new(Item), nil
			}
		}
	}
	return nil, ErrContent
}

func filterCat(name string) bool {
	return strings.HasSuffix(name, suffixMeta+fileExt)
}

func filterRes(name string) bool {
	if !strings.HasSuffix(name, fileExt) {
		return isImage(name)
	}
	return !strings.HasSuffix(name, suffixMeta+fileExt)
}

func isMd(name string) bool { return filepath.Ext(name) == fileExt }

func isImage(name string) bool {
	ext := filepath.Ext(name)
	for _, v := range []string{".jpg", ".jpeg", ".gif", ".png", ".bmp"} {
		if v == ext {
			return true
		}
	}
	return false
}

type parseError struct {
	file  string
	phase string
	err   interface{}
}

func (p parseError) Error() string { return fmt.Sprintf("[%s]%s - %v", p.phase, p.file, p.err) }

func strPtr(s string) *string { return &s }

func repoAddress(owner, name string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, name)
}

func uploadAddress(owner, name, file string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, name, file)
}
