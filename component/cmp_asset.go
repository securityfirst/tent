package component

import (
	"fmt"
	"regexp"
)

type Asset struct {
	ID      string `json:"-"`
	Hash    string `json:"hash,omAssetpty"`
	Content string `json:"content"`
}

func (a *Asset) Resource() Resource {
	return Resource{}
}

func (*Asset) HasChildren() bool {
	return false
}

func (a *Asset) SHA() string {
	return a.Hash
}

func (i *Asset) Path() string {
	return fmt.Sprintf("assets/%s", i.ID)
}

var assetPath = regexp.MustCompile("assets/([^/]+)")

func (i *Asset) SetPath(filepath string) error {
	p := assetPath.FindStringSubmatch(filepath)
	if len(p) == 0 {
		return ErrContent
	}
	i.ID = p[1]
	return nil
}

func (a *Asset) Contents() string {
	return a.Content
}

func (a *Asset) SetContents(contents string) error {
	a.Content = contents
	return nil
}
