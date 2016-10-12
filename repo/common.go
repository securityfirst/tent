package repo

import (
	"fmt"
	"io"
	"path"
	"strings"

	"gopkg.in/src-d/go-git.v3"

	"github.com/google/go-github/github"
	"github.com/klaidliadon/octo/models"
)

func repoAddress(owner, name string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, name)
}

func uploadAddress(owner, name, file string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, name, file)
}

func newCommit(file, sha string, data []byte, u *models.User) *github.RepositoryContentFileOptions {
	author := u.AsAuthor()
	msg := fmt.Sprintf("Updated %s", path.Base(file))
	return &github.RepositoryContentFileOptions{
		Message:   &msg,
		Content:   data,
		SHA:       &sha,
		Author:    author,
		Committer: author,
	}
}

func ParseTree(iter *git.FileIter) (map[string]*Category, error) {
	m := make(map[string]*Category)
	for {
		f, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		contents, err := f.Contents()
		if err != nil {
			return nil, err
		}
		cmp, err := NewComponent("/" + f.Name)
		if err != nil {
			return nil, err
		}
		p := strings.Split(f.Name, "/")
		switch t := cmp.(type) {
		case *Category:
			m[p[0]] = t
		case *Subcategory:
			m[p[0]].Add(t)
		case *Item:
			m[p[0]].Sub(p[1]).Add(t)
		default:
			return nil, fmt.Errorf("%s - Invalid Path", f.Name)
		}
		if err := cmp.SetPath("/" + f.Name); err != nil {
			return nil, fmt.Errorf("%s - Path: %s", f.Name, err)
		}
		if err := cmp.SetContents(contents); err != nil {
			return nil, fmt.Errorf("%s - Content: %s", f.Name, err)
		}
	}
	return m, nil
}
