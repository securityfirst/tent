package component

import (
	"testing"

	. "gopkg.in/check.v1"
	git "gopkg.in/src-d/go-git.v3"
)

func TestAll(t *testing.T) {
	TestingT(t)
}

func TestParse(t *testing.T) {
	r, err := git.NewRepository(repoAddress("klaidliadon", "octo-content"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.PullDefault(); err != nil {
		t.Fatal(err)
	}
	hash, err := r.Remotes[git.DefaultRemoteName].Head()
	if err != nil {
		t.Fatal(err)
	}
	c, err := r.Commit(hash)
	if err != nil {
		t.Fatal(err)
	}
	m, err := ParseTree(c.Tree().Files())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m)
}
