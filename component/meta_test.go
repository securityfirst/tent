package component

import (
	"reflect"
	"strings"
	"testing"
)

func TestMeta(t *testing.T) {
	var testCases = []struct {
		meta   string
		result *FormInput
	}{
		{"[Type]: # (text_input)", nil},
		{"[Type]: # (text_input)\n[Name]: # (name)", nil},
		{"[Type]: # (text_input)\n[Label]: # (Name:)", nil},
		{"[Type]: # (text_input)\n[Name]: # (name)\n[Label]: # (Name:)", &FormInput{
			Type: "text_input", Name: "name", Label: "Name:",
		}},
		{"[Type]: # (text_input)\n[Name]: # (name)\n[Label]: # (Name:)\n[Value]: # (some)\n[Hint]: # (hint)\n[Lines]: # (10)", &FormInput{
			Type: "text_input", Name: "name", Label: "Name:", Value: []string{"some"}, Hint: "hint", Lines: 10,
		}},
		{"[Type]: # (text_input)\n[Name]: # (name)\n[Label]: # (Name:)\n[Value]: # (some)\n[Options]: # (a;b)\n[Hint]: # (hint)\n[Lines]: # (10)", &FormInput{
			Type: "text_input", Name: "name", Label: "Name:", Value: []string{"some"}, Options: []string{"a", "b"}, Hint: "hint", Lines: 10,
		}},
	}
	for _, tc := range testCases {
		f := FormInput{}
		if err := setMeta(tc.meta, &f); err != nil {
			if tc.result != nil {
				t.Error("expected ok, got error")
			}
			continue
		}
		if tc.result == nil {
			t.Error("expected error, got ok")
			continue
		}
		if !reflect.DeepEqual(&f, tc.result) {
			t.Error("expected", tc.result, " got", &f)
		}
		meta := getMeta(&f)
		if meta != tc.meta {
			t.Errorf("expected \n%q, got \n%q", tc.meta, meta)
		}
	}
}

func TestGetMeta(t *testing.T) {
	var base = FormInput{
		Type:    "a",
		Name:    "b",
		Label:   "c",
		Value:   []string{"d"},
		Options: []string{"e"},
		Hint:    "f",
		Lines:   1,
	}

	var testCases = []struct {
		input func(FormInput) FormInput
		lines int
	}{
		{func(f FormInput) FormInput {
			return f
		}, 7},
		{func(f FormInput) FormInput {
			f.Value = nil
			return f
		}, 6},
		{func(f FormInput) FormInput {
			f.Value = nil
			f.Options = nil
			return f
		}, 5},
		{func(f FormInput) FormInput {
			f.Value = nil
			f.Options = nil
			f.Hint = ""
			return f
		}, 4},
		{func(f FormInput) FormInput {
			f.Value = nil
			f.Options = nil
			f.Hint = ""
			f.Lines = 0
			return f
		}, 3},
		{func(f FormInput) FormInput {
			f.Hint = ""
			return f
		}, 6},
	}

	for _, tc := range testCases {
		f := tc.input(base)
		meta := strings.Split(getMeta(&f), "\n")
		if l := len(meta); l != tc.lines {
			t.Errorf("Expected %d lines, got %d\n%q", tc.lines, l, meta)
		}
	}

}
