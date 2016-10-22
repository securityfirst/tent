package component

import (
	"fmt"
	"path"
	"strings"
)

const (
	checksPath = "checks"
)

type Check struct {
	parent     *Subcategory
	Id         string `json:"-"`
	Hash       string `json:"hash"`
	Difficulty string `json:"difficulty"`
	Title      string `json:"title"`
	Text       string `json:"text"`
	NoCheck    bool   `json:"no_check"`
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
	return fmt.Sprintf("/%s/%s/%s/%s", i.parent.parent.Id, i.parent.Id, checksPath, i.Id)
}

func (i *Check) Contents() string {
	return getMeta(checkOrder, args{i.Title, i.Text, i.Difficulty, i.NoCheck})
}

func (i *Check) SetPath(filepath string) error {
	if p := strings.Split(filepath, "/"); len(p) != 5 || p[0] != "" || p[3] != checksPath {
		return ErrInvalid
	}
	i.Id = path.Base(filepath)
	return nil
}

func (i *Check) SetContents(contents string) error {
	if err := checkMeta(contents, checkOrder); err != nil {
		return err
	}
	return setMeta(contents, checkOrder, args{&i.Title, &i.Text, &i.Difficulty, &i.NoCheck})
}
