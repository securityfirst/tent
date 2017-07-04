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
	var t Parser
	err = t.Parse(commit.Tree())
	c.Assert(err, IsNil)
	c.Logf("%v Assets", len(t.Assets()))
	for _, cat := range t.Categories()["en"] {
		c.Log(cat.Name)
		for _, sub := range cat.subcategories {
			c.Logf("\t%q, items:%v checks:%v", sub.Name, len(sub.items), len(sub.checklist.Checks))
		}
	}
}

func (CmpSuite) TestForm(c *C) {
	var f Form

	path := `/forms_en/formid.md`
	contents := `[Name]: # (Form Name)

[Type]: # (screen)
[Name]: # (Screen 1)

[Type]: # (text_input)
[Name]: # (text_1)
[Label]: # (Label1)
[Value]: # (value1)
[Options]: # ()
[Hint]: # (hint1)
[Lines]: # (1)

[Type]: # (screen)
[Name]: # (Screen 2)

[Type]: # (single_choice)
[Name]: # (choice_1)
[Label]: # (Label2)
[Value]: # ()
[Options]: # (option1|option2|option3)
[Hint]: # (hint2)
[Lines]: # (0)`

	c.Assert(f.SetPath(path), IsNil)
	c.Assert(f.SetContents(contents), IsNil)

	c.Assert(f.Path(), Equals, path)
	c.Assert(f.Contents(), Equals, contents)

	c.Logf("%v", f)
}
