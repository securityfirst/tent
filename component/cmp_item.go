package component

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

const paragraphSep = "\n\n"

type Item struct {
	parent   *Difficulty
	ID       string `json:"id"`
	Hash     string `json:"hash,omitempty"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	htmlBody string
	Order    float64 `json:"-"`
}

func (i *Item) Resource() Resource {
	parts := strings.Split(i.Body, paragraphSep)
	content := make([]map[string]string, len(parts)+1)
	content[0] = map[string]string{"title": i.Title}
	for i := range parts {
		content[i+1] = map[string]string{"body": parts[i]}
	}
	return Resource{Slug: i.parent.Resource().Slug + "_" + i.ID, Content: content}
}

func (i *Item) SetParent(d *Difficulty) {
	i.parent = d
}

func (i *Item) HasChildren() bool {
	return false
}

func (i *Item) SHA() string {
	return i.Hash
}

func (i *Item) Path() string {
	return fmt.Sprintf("%s/%s%s", i.parent.basePath(), i.ID, fileExt)
}

var itemPath = regexp.MustCompile("contents(?:_[a-z]{2})?/[^/]+/[^/]+/[^/]+/([^/]+).md")

func (i *Item) SetPath(filepath string) error {
	p := itemPath.FindStringSubmatch(filepath)
	if len(p) == 0 || p[1] == suffixMeta {
		return ErrContent
	}
	i.ID = p[1]
	return nil
}

func (*Item) order() []string     { return []string{"Title", "Order"} }
func (*Item) optionals() []string { return nil }
func (i *Item) pointers() args    { return args{&i.Title, &i.Order} }
func (i *Item) values() args      { return args{i.Title, i.Order} }

func (i *Item) Contents() string {
	return fmt.Sprint(getMeta(i), bodySeparator, i.Body)
}

func (i *Item) SetContents(contents string) error {
	parts := strings.SplitN(strings.Trim(contents, "\n"), bodySeparator, 2)
	if len(parts) != 2 {
		return ErrContent
	}
	if err := setMeta(parts[0], i); err != nil {
		return err
	}
	i.Body = parts[1]
	i.htmlBody = string(blackfriday.MarkdownCommon([]byte(i.Body)))
	return nil
}
