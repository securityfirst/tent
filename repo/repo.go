package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"gopkg.in/src-d/go-git.v3"

	"github.com/google/go-github/github"
	"github.com/securityfirst/tent/component"
	"github.com/securityfirst/tent/models"
	"golang.org/x/oauth2"
)

var (
	ErrNotReady     = errors.New("Repository not ready")
	ErrFileNotFound = git.ErrFileNotFound
)

const (
	actionCreate = iota
	actionUpdate
	actionDelete
)

var commitMsg = map[int]string{
	actionCreate: "Create",
	actionUpdate: "Update",
	actionDelete: "Delete",
}

func New(owner, name string) (*Repo, error) {
	r, err := git.NewRepository(repoAddress(owner, name), nil)
	if err != nil {
		return nil, err
	}
	return &Repo{repo: r, name: name, owner: owner}, nil
}

type Repo struct {
	sync.RWMutex
	owner      string
	name       string
	conf       *oauth2.Config
	repo       *git.Repository
	commit     *git.Commit
	categories []*component.Category
	assets     []*component.Asset
}

func (r *Repo) SetConf(c *oauth2.Config) { r.conf = c }

func (r *Repo) Tree(locale string, html bool) interface{} {
	r.RLock()
	defer r.RUnlock()
	var s = make([]interface{}, 0, len(r.categories))
	for _, i := range r.Categories() {
		s = append(s, r.Category(i, locale).Tree(html))
	}
	return s
}

func (r *Repo) client(u *models.User) *github.Client {
	return github.NewClient(r.conf.Client(oauth2.NoContext, &u.Token))
}

func (r *Repo) Handler() RepoHandler { return RepoHandler{r} }

func (r *Repo) hash() string {
	if r.commit != nil {
		return r.commit.Hash.String()
	}
	return "n/a"
}

func (r *Repo) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"owner":  r.owner,
		"name":   r.name,
		"commit": r.hash(),
	})
}

func (r *Repo) String() string {
	return fmt.Sprintf("%s/%s %s", r.owner, r.name, r.hash())
}

func (r *Repo) Pull() {
	r.Lock()
	defer r.Unlock()
	if err := r.repo.PullDefault(); err != nil {
		log.Println("Pull failed:", err)
		return
	}
	hash, err := r.repo.Remotes[git.DefaultRemoteName].Head()
	if err != nil {
		log.Println("Head failed:", err)
		return
	}
	if r.commit != nil && r.commit.Hash == hash {
		return
	}
	if r.commit != nil {
		log.Println("Changing commit from", r.commit.Hash, "to", hash)
	} else {
		log.Println("Checkout with", hash)
	}
	r.commit, err = r.repo.Commit(hash)
	if err != nil {
		log.Println("Commit failed:", err)
		return
	}
	var parser component.Parser
	if err := parser.Parse(r.commit.Tree()); err != nil {
		log.Println("Parsing failed:", err)
		return
	}
	r.categories = parser.Categories
	r.assets = parser.Assets
}

func (r *Repo) file(c component.Component) (*git.File, error) {
	if r.commit == nil {
		return nil, ErrNotReady
	}
	r.RLock()
	defer r.RUnlock()
	return r.commit.File(c.Path()[1:])
}

func (r *Repo) Get(c component.Component) (string, error) {
	f, err := r.file(c)
	if err != nil {
		return "", err
	}
	return f.Contents()
}

func (r *Repo) Asset(id string) *component.Asset {
	r.RLock()
	defer r.RUnlock()

	for _, a := range r.assets {
		if a.Id == id {
			return a
		}
	}
	return nil
}

func (r *Repo) Category(cat, locale string) *component.Category {
	r.RLock()
	defer r.RUnlock()

	for _, c := range r.categories {
		if c.Id == cat && c.Locale == locale {
			return c
		}
	}
	return nil
}

func (r *Repo) Categories() []string {
	r.RLock()
	defer r.RUnlock()
	var s = make([]string, 0, len(r.categories))
	for _, v := range r.categories {
		s = append(s, v.Id)
	}
	return s
}

func (r *Repo) ComponentHash(c component.Component) (string, error) {
	f, err := r.file(c)
	if err != nil {
		return "", err
	}
	return f.Hash.String(), nil
}

func (r *Repo) Create(c component.Component, u *models.User) error {
	return r.request(c, actionCreate, u)
}

func (r *Repo) Delete(c component.Component, u *models.User) error {
	return r.request(c, actionDelete, u)
}

func (r *Repo) Update(c component.Component, u *models.User) error {
	return r.request(c, actionUpdate, u)
}

func (r *Repo) request(c component.Component, action int, u *models.User) (err error) {
	file := c.Path()[1:]
	msg := fmt.Sprintf("%s %s", commitMsg[action], file)
	commit := &github.RepositoryContentFileOptions{
		Message: &msg, Author: u.AsAuthor(),
	}
	switch action {
	case actionCreate:
		commit.Content = []byte(c.Contents())
		_, _, err = r.client(u).Repositories.CreateFile(r.owner, r.name, file, commit)
	case actionUpdate:
		commit.SHA = strPtr(c.SHA())
		commit.Content = []byte(c.Contents())
		_, _, err = r.client(u).Repositories.UpdateFile(r.owner, r.name, file, commit)
	case actionDelete:
		commit.SHA = strPtr(c.SHA())
		_, _, err = r.client(u).Repositories.DeleteFile(r.owner, r.name, file, commit)
	}
	if err == nil {
		go r.Pull()
	}
	return err
}

func strPtr(s string) *string { return &s }

func repoAddress(owner, name string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, name)
}

func uploadAddress(owner, name, file string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, name, file)
}
