package component

import (
	. "gopkg.in/check.v1"
)

func (CmpSuite) TestParseResource(c *C) {
	p := NewResourceParser()
	var err error

	catCmp := Category{Id: "cat", Locale: "en", Name: "Test Category"}
	catRes := Resource{
		Slug:    "cat",
		Content: []map[string]string{map[string]string{"name": "Categoria di prova"}},
	}
	err = p.Parse(&catCmp, &catRes, "it")
	c.Assert(err, IsNil)

	subCmp := Subcategory{Id: "sub", Name: "Test Sub", parent: &catCmp}
	subRes := Resource{
		Slug:    "cat___sub",
		Content: []map[string]string{map[string]string{"name": "Sub di prova"}},
	}
	err = p.Parse(&subCmp, &subRes, "it")
	c.Assert(err, IsNil)

	itemCmp := Item{Id: "item", Title: "Test item", Difficulty: "hard", parent: &subCmp, Body: "row1\n\nrow2\n\nrow3"}
	itemRes := Resource{
		Slug: "cat___sub___item",
		Content: []map[string]string{
			map[string]string{"title": "item di prova", "difficulty": "difficile"},
			map[string]string{"body": "riga1"},
			map[string]string{"body": "riga2"},
			map[string]string{"body": "riga3"},
		},
	}
	err = p.Parse(&itemCmp, &itemRes, "it")
	c.Assert(err, IsNil)

	legacyCmp := Item{Id: "item-legacy", Title: "Old Test item", Difficulty: "hard", parent: &subCmp, Body: "row1\n\nrow2\n\nrow3"}
	legacyRes := Resource{
		Slug: "cat___sub___item-legacy",
		Content: []map[string]string{
			map[string]string{"title": "item di prova legacy", "difficulty": "difficile", "body": "riga1\n\nriga2\n\nriga3"},
		},
	}
	err = p.Parse(&legacyCmp, &legacyRes, "it")
	c.Assert(err, IsNil)

	legacyRes.Content = append(legacyRes.Content, map[string]string{"body": "moar"})
	err = p.Parse(&legacyCmp, &legacyRes, "it")
	c.Assert(err, NotNil)
}
