package component

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Subcategory struct {
	parent       *Category
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Hash         string  `json:"hash"`
	Order        float64 `json:"-"`
	difficulties []*Difficulty
}

func (s *Subcategory) Resource() Resource {
	return Resource{
		Slug: s.parent.Resource().Slug + "_" + s.ID,
		Content: []map[string]string{
			map[string]string{"name": s.Name},
		},
	}
}

func (s *Subcategory) HasChildren() bool {
	return len(s.difficulties) != 0
}

func (s *Subcategory) Tree(html bool) interface{} {
	var difficulties = make([]interface{}, len(s.difficulties))
	for i, v := range s.difficulties {
		difficulties[i] = v.Tree(html)
	}
	return map[string]interface{}{
		"id":           s.ID,
		"name":         s.Name,
		"difficulties": difficulties,
	}
}

func (s *Subcategory) SHA() string {
	return s.Hash
}

func (s *Subcategory) SetParent(c *Category) {
	s.parent = c
}

func (s *Subcategory) MarshalJSON() ([]byte, error) {
	var m = map[string]interface{}{
		"name":         s.Name,
		"difficulties": s.DifficultyNames(),
	}
	if s.Hash != "" {
		m["hash"] = s.Hash
	}
	return json.Marshal(m)
}

func (s *Subcategory) Difficulties() []Difficulty {
	var dst = make([]Difficulty, len(s.difficulties))
	for i, v := range s.difficulties {
		dst[i] = *v
	}
	return dst
}

func (s *Subcategory) DifficultyNames() []string {
	var r = make([]string, 0, len(s.difficulties))
	for i := range s.difficulties {
		r = append(r, s.difficulties[i].ID)
	}
	return r
}

func (s *Subcategory) AddDifficulty(difficulties ...*Difficulty) error {
	for _, v := range difficulties {
		if s.Difficulty(v.ID) != nil {
			return fmt.Errorf("Difficulty %s exists in %s/%s", v.ID, s.parent.ID, s.ID)
		}
		v.parent = s
	}
	s.difficulties = append(s.difficulties, difficulties...)
	return nil
}

func (s *Subcategory) Difficulty(ID string) *Difficulty {
	for _, v := range s.difficulties {
		if v.ID == ID {
			return v
		}
	}
	return nil
}

func (s *Subcategory) basePath() string {
	return fmt.Sprintf("%s/%s", s.parent.basePath(), s.ID)
}

func (s *Subcategory) Path() string {
	return fmt.Sprintf("%s/.metadata%s", s.basePath(), fileExt)
}

var subPath = regexp.MustCompile("contents(?:_[a-z]{2})?/[^/]+/([^/]+)/.metadata.md")

func (s *Subcategory) SetPath(filepath string) error {
	p := subPath.FindStringSubmatch(filepath)
	if len(p) == 0 {
		return ErrContent
	}
	s.ID = p[1]
	return nil
}

func (*Subcategory) order() []string     { return []string{"Name", "Order"} }
func (*Subcategory) optionals() []string { return nil }
func (s *Subcategory) pointers() args    { return args{&s.Name, &s.Order} }
func (s *Subcategory) values() args      { return args{s.Name, s.Order} }

func (s *Subcategory) Contents() string { return getMeta(s) }

func (s *Subcategory) SetContents(contents string) error {
	return setMeta(contents, s)
}
