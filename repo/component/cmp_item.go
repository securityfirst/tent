package component

import (
	"fmt"
	"regexp"
	"strings"
)

type Item struct {
	parent     *Subcategory
	Id         string  `json:"-"`
	Hash       string  `json:"hash"`
	Difficulty string  `json:"difficulty"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	Order      float64 `json:"-`
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
	return i.parent.basePath() + "/" + i.Id
}

var itemPath = regexp.MustCompile("/contents(?:_[a-z]{2})?/[^/]+/[^/]+/([^/]+)")

func (i *Item) SetPath(filepath string) error {
	p := itemPath.FindStringSubmatch(filepath)
	if len(p) == 0 || p[1] == suffixMeta {
		return ErrInvalid
	}
	i.Id = p[1]
	return nil
}

func (i *Item) Contents() string {
	return fmt.Sprint(getMeta(itemOrder, args{i.Title, i.Difficulty, i.Order}), bodySeparator, i.Body)
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
		[]interface{}{&i.Title, &i.Difficulty, &i.Order},
	)
}
