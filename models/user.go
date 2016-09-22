package models

import (
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type User struct {
	Login string
	Name  string
	Email string
	Token oauth2.Token
}

func (u *User) AsAuthor() *github.CommitAuthor {
	now := time.Now()
	return &github.CommitAuthor{Name: &u.Name, Email: &u.Email, Date: &now}
}
