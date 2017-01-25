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
	r, err := git.NewRepository(repoAddress("klaidliadon", "tent-content"), nil)
	c.Assert(err, IsNil)
	err = r.PullDefault()
	c.Assert(err, IsNil)
	hash, err := r.Remotes[git.DefaultRemoteName].Head()
	c.Assert(err, IsNil)
	commit, err := r.Commit(hash)
	c.Assert(err, IsNil)
	var t TreeParser
	err = t.Parse(commit.Tree())
	c.Assert(err, IsNil)
	c.Logf("%v Assets", len(t.Assets))
	for _, cat := range t.Categories {
		c.Log(cat.Name)
		for _, sub := range cat.subcategories {
			c.Logf("\t%q, items:%v checks:%v", sub.Name, len(sub.items), len(sub.checklist.Checks))
		}
	}
}
