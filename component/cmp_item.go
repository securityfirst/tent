package component

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

type Item struct {
	parent     *Subcategory
	Id         string `json:"-"`
	Hash       string `json:"hash,omitempty"`
	Difficulty string `json:"difficulty"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	htmlBody   string
	Order      float64 `json:"-"`
}

func (i *Item) Resource() Resource {
	return Resource{
		Slug: i.parent.Resource().Slug + "_" + i.Id,
		Content: []map[string]string{
			map[string]string{
				"title":      i.Title,
				"body":       i.Body,
				"difficulty": i.Difficulty,
			},
		},
	}
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
	return fmt.Sprintf("%s/%s%s", i.parent.basePath(), i.Id, fileExt)
}

var itemPath = regexp.MustCompile("/contents(?:_[a-z]{2})?/[^/]+/[^/]+/([^/]+).md")

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
	parts := strings.SplitN(strings.Trim(contents, "\n"), bodySeparator, 2)
	if len(parts) != 2 {
		return ErrInvalid
	}
	if err := checkMeta(parts[0], itemOrder); err != nil {
		return err
	}
	i.Body = parts[1]
	i.htmlBody = string(blackfriday.MarkdownCommon([]byte(i.Body)))
	return setMeta(
		parts[0],
		itemOrder,
		[]interface{}{&i.Title, &i.Difficulty, &i.Order},
	)
}
