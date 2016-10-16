package component

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

type Item struct {
	parent     *Subcategory
	Id         string `json:"-"`
	Hash       string `json:"hash"`
	Difficulty string `json:"difficulty"`
	Title      string `json:"title"`
	Body       string `json:"body"`
}

func (i *Item) SetParent(s *Subcategory) {
	i.parent = s
}

func (i *Item) HasChildren() bool {
	return false
}

func (i *Item) SHA() string {
	return i.Hash
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
	parts := strings.Split(strings.Trim(contents, "\n"), bodySeparator)
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
