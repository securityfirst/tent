package component

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

const paragraphSep = "\n\n\n"

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
	parts := strings.Split(i.Body, paragraphSep)
	content := make([]map[string]string, len(parts)+1)
	content[0] = map[string]string{"title": i.Title, "difficulty": i.Difficulty}
	for i := range parts {
		content[i+1] = map[string]string{"body": parts[i]}
	}
	return Resource{Slug: i.parent.Resource().Slug + "_" + i.Id, Content: content}
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
		return ErrContent
	}
	i.Id = p[1]
	return nil
}

func (i *Item) order() []string { return []string{"Title", "Difficulty", "Order"} }
func (i *Item) pointers() args  { return args{&i.Title, &i.Difficulty, &i.Order} }
func (i *Item) values() args    { return args{i.Title, i.Difficulty, i.Order} }

func (i *Item) Contents() string {
	return fmt.Sprint(getMeta(i), bodySeparator, i.Body)
}

func (i *Item) SetContents(contents string) error {
	parts := strings.SplitN(strings.Trim(contents, "\n"), bodySeparator, 2)
	if len(parts) != 2 {
		return ErrContent
	}
	if err := checkMeta(parts[0], i); err != nil {
		return err
	}
	i.Body = parts[1]
	i.htmlBody = string(blackfriday.MarkdownCommon([]byte(i.Body)))
	return setMeta(parts[0], i)
}
