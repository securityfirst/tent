package component

import . "gopkg.in/check.v1"

var _ = Suite(&CategorySuite{})

type CategorySuite struct{}

func (*CategorySuite) TestCategory(c *C) {
	var v Category
	c.Assert(v.SetPath("/path"), Equals, ErrInvalid)
	c.Assert(v.SetPath("/path/.metadata"), IsNil)
	c.Assert(v.Id, Equals, "path")
	c.Assert(v.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(v.SetContents("Name:Test"), IsNil)
	c.Assert(v.Name, Equals, "Test")
	c.Assert(v.Path(), Equals, "/path/.metadata")
	c.Assert(v.Contents(), Equals, "Name:Test")
}

func (*CategorySuite) TestSubcategory(c *C) {
	var (
		v Category
		s Subcategory
	)
	v.SetPath("/path/.metadata")
	v.SetContents("Name:Test")
	v.Add(&s)

	c.Assert(s.SetPath("/path"), Equals, ErrInvalid)
	c.Assert(s.SetPath("/path/sub"), Equals, ErrInvalid)
	c.Assert(s.SetPath("/path/sub/.metadata"), IsNil)
	c.Assert(s.Id, Equals, "sub")
	c.Assert(s.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(s.SetContents("Name:SubTest"), IsNil)
	c.Assert(s.Name, Equals, "SubTest")
	c.Assert(s.Path(), Equals, "/path/sub/.metadata")
	c.Assert(s.Contents(), Equals, "Name:SubTest")
}

func (*CategorySuite) TestItem(c *C) {
	var (
		v Category
		s Subcategory
		i Item
	)
	v.SetPath("/path/.metadata")
	v.SetContents("Name:Test")
	v.Add(&s)
	s.SetPath("/path/sub/.metadata")
	s.SetContents("Name:SubTest")
	s.Add(&i)

	c.Assert(i.SetPath("/path"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/path/sub"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/path/sub/.metadata"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/path/sub/filename"), IsNil)
	c.Assert(i.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\n---\nBody"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nDifficulty:easy"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nDifficulty:easy\n---\nBody"), IsNil)
	c.Assert(i.Path(), Equals, "/path/sub/filename")
	c.Assert(i.Contents(), Equals, "Title:Title\nDifficulty:easy\n---\nBody")
}

func (*CategorySuite) TestNew(c *C) {
	var (
		v   Component
		err error
	)

	v, err = New("/path")
	c.Assert(err, Equals, ErrInvalid)
	c.Assert(v, IsNil)

	v, err = New("/path/.metadata")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Category{Id: "path"})

	v, err = New("/path/sub/.metadata")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Subcategory{Id: "sub"})

	v, err = New("/path/sub/item")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Item{Id: "item"})
}
