package models

import (
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type User struct {
	Login string       `json:"login"`
	Name  string       `json:"name"`
	Email string       `json:"email"`
	Token oauth2.Token `json:"-"`
}

func (u *User) AsAuthor() *github.CommitAuthor {
	now := time.Now()
	return &github.CommitAuthor{Name: &u.Name, Email: &u.Email, Date: &now}
}
