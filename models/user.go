package models

import (
	"time"

	"github.com/google/go-github/github"
)

// User is the model used for cookies
type User struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// AsAuthor returns a commit author for Github
func (u *User) AsAuthor() *github.CommitAuthor {
	now := time.Now()
	return &github.CommitAuthor{Name: &u.Name, Email: &u.Email, Date: &now}
}
