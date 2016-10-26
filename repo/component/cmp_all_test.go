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
	r, err := git.NewRepository(repoAddress("secirutyfirst", "octo-content"), nil)
	c.Assert(err, IsNil)
	err = r.PullDefault()
	c.Assert(err, IsNil)
	hash, err := r.Remotes[git.DefaultRemoteName].Head()
	c.Assert(err, IsNil)
	commit, err := r.Commit(hash)
	c.Assert(err, IsNil)
	err = (&TreeParser{}).Parse(commit.Tree().Files())
	c.Assert(err, IsNil)
}

func (*CmpSuite) TestCategory(c *C) {
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

func (*CmpSuite) TestSubcategory(c *C) {
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

func (*CmpSuite) TestItem(c *C) {
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
	s.AddItem(&i)

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

func (*CmpSuite) TestCheck(c *C) {
	var (
		v Category
		s Subcategory
		i Check
	)
	v.SetPath("/path/.metadata")
	v.SetContents("Name:Test")
	v.Add(&s)
	s.SetPath("/path/sub/.metadata")
	s.SetContents("Name:SubTest")
	s.AddCheck(&i)

	c.Assert(i.SetPath("/path"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/path/sub"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/path/sub/.metadata"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/path/sub/filename"), Equals, ErrInvalid)
	c.Assert(i.SetPath("/path/sub/checks/filename"), IsNil)
	c.Assert(i.SetContents("contents"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\n---\nBody"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nDifficulty:easy"), Equals, ErrInvalid)
	c.Assert(i.SetContents("Title:Title\nText:text\nDifficulty:easy\nNoCheck:true"), IsNil)
	c.Assert(i.Path(), Equals, "/path/sub/checks/filename")
	c.Assert(i.Contents(), Equals, "Title:Title\nText:text\nDifficulty:easy\nNoCheck:true")
}

func (*CmpSuite) TestNew(c *C) {
	var (
		v   Component
		err error
	)

	v, err = newCmp("/path")
	c.Assert(err, Equals, ErrInvalid)
	c.Assert(v, IsNil)

	v, err = newCmp("/path/.metadata")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Category{Id: "path"})

	v, err = newCmp("/path/sub/.metadata")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Subcategory{Id: "sub"})

	v, err = newCmp("/path/sub/item")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Item{Id: "item"})

	v, err = newCmp("/path/sub/checks/check")
	c.Assert(err, IsNil)
	c.Assert(v, DeepEquals, &Check{Id: "check"})
}
