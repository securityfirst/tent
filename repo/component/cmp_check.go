package component

import (
	"fmt"
	"regexp"
)

type Check struct {
	parent     *Subcategory
	Id         string  `json:"-"`
	Hash       string  `json:"hash"`
	Difficulty string  `json:"difficulty"`
	Title      string  `json:"title"`
	Text       string  `json:"text"`
	NoCheck    bool    `json:"no_check"`
	Order      float64 `json:"-`
}

func (i *Check) SetParent(s *Subcategory) {
	i.parent = s
}

func (i *Check) HasChildren() bool {
	return false
}

func (i *Check) SHA() string {
	return i.Hash
}

func (i *Check) Path() string {
	return fmt.Sprintf("%s/checks/%s", i.parent.basePath(), i.Id)
}

var checkPath = regexp.MustCompile("/contents(?:_[a-z]{2})?/[^/]+/[^/]+/checks/([^/]+)")

func (i *Check) SetPath(filepath string) error {
	p := checkPath.FindStringSubmatch(filepath)
	if len(p) == 0 {
		return ErrInvalid
	}
	i.Id = p[1]
	return nil
}

func (i *Check) Contents() string {
	return getMeta(checkOrder, args{i.Title, i.Text, i.Difficulty, i.NoCheck, i.Order})
}

func (i *Check) SetContents(contents string) error {
	if err := checkMeta(contents, checkOrder); err != nil {
		return err
	}
	return setMeta(contents, checkOrder, args{&i.Title, &i.Text, &i.Difficulty, &i.NoCheck, &i.Order})
}
