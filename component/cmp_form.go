package component

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Form struct {
	Id      string       `json:"-"`
	Name    string       `json:"name"`
	Hash    string       `json:"hash"`
	Locale  string       `json:"-"`
	Screens []FormScreen `json:"screens,omitempty"`
}

func (f *Form) Resource() Resource {
	return Resource{
		Slug: f.Id,
		Content: []map[string]string{
			map[string]string{"name": f.Name},
		},
	}
}

func (f *Form) HasChildren() bool { return false }

func (f *Form) SHA() string { return f.Hash }

func (f *Form) Path() string {
	var loc string
	if f.Locale != "" {
		loc = "_" + f.Locale
	}
	return fmt.Sprintf("/forms%s/%s%s", loc, f.Id, fileExt)
}

var formPath = regexp.MustCompile("/forms_([a-z]{2})/([^/]+).md")

func (f *Form) SetPath(filepath string) error {
	p := formPath.FindStringSubmatch(filepath)
	if len(p) == 0 {
		return ErrContent
	}
	f.Locale = p[1]
	f.Id = p[2]
	return nil
}

func (f *Form) order() []string { return []string{"Name"} }
func (f *Form) pointers() args  { return args{&f.Name} }
func (f *Form) values() args    { return args{f.Name} }

func (f *Form) Contents() string {
	b := bytes.NewBuffer(nil)
	fmt.Fprint(b, getMeta(f))
	for i := range f.Screens {
		fmt.Fprint(b, bodySeparator, getMeta(&f.Screens[i]))
		for _, v := range f.Screens[i].Items {
			fmt.Fprint(b, bodySeparator, getMeta(&v))
		}
	}
	return b.String()
}

func (f *Form) SetContents(contents string) error {
	parts := strings.Split(contents, bodySeparator)
	if err := setMeta(parts[0], f); err != nil {
		return err
	}
	screenIndex := -1
	for _, p := range parts[1:] {
		m := metaRow.FindStringSubmatch(strings.SplitN(p, "\n", 2)[0])
		if screenIndex < 0 && (len(m) != 3 || m[1] != "Type" || m[2] != "screen") {
			return ErrContent
		}
		switch m[2] {
		case "screen":
			var s FormScreen
			if err := setMeta(p, &s); err != nil {
				return err
			}
			f.Screens = append(f.Screens, s)
			screenIndex++
		default:
			var i FormInput
			if err := setMeta(p, &i); err != nil {
				return err
			}
			f.Screens[screenIndex].Items = append(f.Screens[screenIndex].Items, i)
		}

	}
	return nil
}

type FormScreen struct {
	Name  string      `json:"name"`
	Items []FormInput `json:"items,omitempty"`
}

func (f *FormScreen) order() []string { return []string{"Type", "Name"} }
func (f *FormScreen) pointers() args  { var s string; return args{&s, &f.Name} }
func (f *FormScreen) values() args    { return args{"screen", f.Name} }

type FormInput struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Label   string   `json:"label"`
	Value   []string `json:"value"`
	Options []string `json:"options,omitempty"`
	Hint    string   `json:"hint,omitempty"`
	Lines   int      `json:"lines,omitempty"`
}

func (f *FormInput) order() []string {
	return []string{"Type", "Name", "Label", "Value", "Options", "Hint", "Lines"}
}
func (f *FormInput) pointers() args {
	return args{&f.Type, &f.Name, &f.Label, &f.Value, &f.Options, &f.Hint, &f.Lines}
}
func (f *FormInput) values() args {
	return args{f.Type, f.Name, f.Label, f.Value, f.Options, f.Hint, f.Lines}
}
