package component

import (
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
	return fmt.Sprint(getMeta(itemOrder, args{i.Title, i.Difficulty}), bodySeparator, i.Body)
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
	if err := checkMeta(parts[0], itemOrder); err != nil {
		return err
	}
	i.Body = parts[1]
	return setMeta(
		parts[0],
		itemOrder,
		[]interface{}{&i.Title, &i.Difficulty},
	)
}
