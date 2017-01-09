package component

import (
	"testing"

	. "gopkg.in/check.v1"
	git "gopkg.in/src-d/go-git.v3"
)

func TestAll(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&CmpSuite{})

type CmpSuite struct{}

func (*CmpSuite) TestParse(c *C) {
	r, err := git.NewRepository(repoAddress("securityfirst", "tent-content"), nil)
	c.Assert(err, IsNil)
	err = r.PullDefault()
	c.Assert(err, IsNil)
	hash, err := r.Remotes[git.DefaultRemoteName].Head()
	c.Assert(err, IsNil)
	commit, err := r.Commit(hash)
	c.Assert(err, IsNil)
	err = (&TreeParser{}).Parse(commit.Tree())
	c.Assert(err, IsNil)
}

func (*CmpSuite) TestCategory(c *C) {
	var v Category
	c.Assert(v.SetPath("/contents_en/path"), Equals, ErrInvalid)
	c.Assert(v.SetPath("/contents_en/path/.metadata"), IsNil)
	c.Assert(v.Id, Equals, "path")
	c.Assert(v.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(v.SetContents("Name:Test"), Equals, ErrInvalid)
	c.Assert(v.SetContents("Name:Test\nOrder:1"), IsNil)
	c.Assert(v.Name, Equals, "Test")
	c.Assert(v.Path(), Equals, "/contents_en/path/.metadata")
	c.Assert(v.Contents(), Equals, "Name:Test\nOrder:1")
}

func (*CmpSuite) TestSubcategory(c *C) {
	var (
		v Category
		s Subcategory
	)
	v.SetPath("/contents_en/path/.metadata")
	v.SetContents("Name:Test\nOrder:1")
	v.Add(&s)

	c.Assert(s.SetPath("/contents_en/path"), Equals, ErrInvalid)
	c.Assert(s.SetPath("/contents_en/path/sub"), Equals, ErrInvalid)
	c.Assert(s.SetPath("/contents_en/path/sub/.metadata"), IsNil)
	c.Assert(s.Id, Equals, "sub")
	c.Assert(s.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(s.SetContents("Name:SubTest"), Equals, ErrInvalid)
	c.Assert(s.SetContents("Name:SubTest\nOrder:1"), IsNil)
	c.Assert(s.Name, Equals, "SubTest")
	c.Assert(s.Path(), Equals, "/contents_en/path/sub/.metadata")
	c.Assert(s.Contents(), Equals, "Name:SubTest\nOrder:1")
}

func (*CmpSuite) TestItem(c *C) {
	var (
		v Category
		s Subcategory
		i Item
	)
	v.SetPath("/contents_en/path/.metadata")
	v.SetContents("Name:Test\nOrder:1")
	v.Add(&s)
	s.SetPath("/contents_en/path/sub/.metadata")
	s.SetContents("Name:SubTest\nOrder:1")
	s.AddItem(&i)

	c.Assert(i.SetPath("/contents_en/path"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/contents_en/path/sub"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/contents_en/path/sub/.metadata"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/contents_en/path/sub/filename"), IsNil)
	c.Assert(i.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\n---\nBody"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nDifficulty:easy\nOrder:1"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nDifficulty:easy\nOrder:1\n---\nBody"), IsNil)
	c.Assert(i.Path(), Equals, "/contents_en/path/sub/filename")
	c.Assert(i.Contents(), Equals, "Title:Title\nDifficulty:easy\nOrder:1\n---\nBody")
}

func (*CmpSuite) TestChecks(c *C) {
	var (
		v Category
		s Subcategory
		i Checklist
	)
	v.SetPath("/contents_en/path/.metadata")
	v.SetContents("Name:Test\nOrder:1")
	v.Add(&s)
	s.SetPath("/contents_en/path/sub/.metadata")
	s.SetContents("Name:SubTest\nOrder:1")
	s.SetChecks(&i)

	c.Assert(i.SetPath("/contents_en/path"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/contents_en/path/sub"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/contents_en/path/sub/.metadata"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/contents_en/path/sub/filename"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/contents_en/path/sub/.checks"), IsNil)
	c.Assert(i.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\n---\nBody"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nDifficulty:easy\nOrder:1"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nText:text\nDifficulty:easy\nNoCheck:true"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Text:text\nDifficulty:easy\nNoCheck:true"), IsNil)
	c.Assert(i.Path(), Equals, "/contents_en/path/sub/.checks")
	c.Assert(i.Contents(), Equals, "Text:text\nDifficulty:easy\nNoCheck:true")
	c.Assert(i.SetContents("Text:text\nDifficulty:easy\nNoCheck:true"+
		bodySeparator+"Text:text\nDifficulty:easy\nNoCheck:true"), IsNil)
	c.Assert(i.Contents(), Equals, "Text:text\nDifficulty:easy\nNoCheck:true"+
		bodySeparator+"Text:text\nDifficulty:easy\nNoCheck:true")
}

func (*CmpSuite) TestNew(c *C) {
	var (
		v   Component
		err error
	)

	v, err = newCmp("/contents_en/path")
	c.Assert(err, Equals, ErrInvalid)
	c.Assert(v, IsNil)

	v, err = newCmp("/contents_en/path/.metadata")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Category{Id: "path", Locale: "en"})

	v, err = newCmp("/contents_en/path/sub/.metadata")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Subcategory{Id: "sub"})

	v, err = newCmp("/contents_en/path/sub/item")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Item{Id: "item"})

	v, err = newCmp("/contents_en/path/sub/.checks")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Checklist{})
}
