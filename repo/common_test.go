package repo

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestAll(t *testing.T) {
	TestingT(t)
}

func TestParse(t *testing.T) {
	r, err := New("klaidliadon", "octo-content", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.update(); err != nil {
		t.Fatal(err)
	}
	m, err := ParseTree(r.commit.Tree().Files())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m)
}
