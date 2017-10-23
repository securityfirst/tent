package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/google/go-github/github"
	"github.com/securityfirst/tent/component"
	"github.com/securityfirst/tent/models"
	"golang.org/x/oauth2"
)

var (
	ErrNotReady     = errors.New("Repository not ready")
	ErrFileNotFound = object.ErrFileNotFound
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

func New(owner, name, branch string) (*Repo, error) {
	address := repoAddress(owner, name)
	log.Printf("Using %q", address)
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: address})
	if err != nil {
		return nil, err
	}

	return &Repo{repo: r, name: name, owner: owner, branch: branch}, nil
}

type Repo struct {
	sync.RWMutex
	owner      string
	name       string
	branch     string
	conf       *oauth2.Config
	repo       *git.Repository
	commit     *object.Commit
	categories map[string][]*component.Category
	assets     []*component.Asset
	forms      []*component.Form
}

func (r *Repo) SetConf(c *oauth2.Config) { r.conf = c }

func (r *Repo) Tree(locale string, html bool) interface{} {
	r.RLock()
	defer r.RUnlock()

	var cats = make([]interface{}, 0, len(r.categories))
	for _, i := range r.Categories(locale) {
		cats = append(cats, r.Category(i, locale).Tree(html))
	}

	var ass = make([]string, len(r.assets))
	for i := range r.assets {
		ass[i] = r.assets[i].Id
	}

	var forms = make([]string, len(r.forms))
	for i := range r.forms {
		forms[i] = r.forms[i].Id
	}

	return map[string]interface{}{
		"categories": cats,
		"assets":     ass,
		"forms":      forms,
	}
}

func (r *Repo) client(u *models.User) *github.Client {
	return github.NewClient(r.conf.Client(oauth2.NoContext, &u.Token))
}

func (r *Repo) Handler() RepoHandler { return RepoHandler{r} }

func (r *Repo) All(locale string) []component.Component {
	var list []component.Component
	for _, cat := range r.categories[locale] {
		list = append(list, cat)
		for _, id := range cat.Subcategories() {
			sub := cat.Sub(id)
			list = append(list, sub)
			for _, id := range sub.ItemNames() {
				list = append(list, sub.Item(id))
			}
			if check := sub.Checks(); check.HasChildren() {
				list = append(list, check)
			}
		}
	}
	return list
}

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

	err := r.repo.Fetch(&git.FetchOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Println("Pull failed:", err)
		return
	}
	branch := plumbing.ReferenceName("refs/remotes/origin/" + r.branch)
	hash, err := r.repo.Reference(branch, false)
	if err != nil {
		log.Printf("Reference %q failed:", branch, err)
		return
	}
	if r.commit != nil && r.commit.Hash == hash.Hash() {
		return
	}
	if r.commit != nil {
		log.Println("Changing commit from", r.commit.Hash, "to", hash)
	} else {
		log.Println("Checkout with", hash)
	}
	r.commit, err = r.repo.CommitObject(hash.Hash())
	if err != nil {
		log.Println("Commit failed:", err)
		return
	}
	var parser component.Parser
	tree, err := r.commit.Tree()
	if err != nil {
		log.Println("Tree failed:", err)
		return
	}
	if err := parser.Parse(tree); err != nil {
		log.Println("Parsing failed:", err)
		return
	}
	r.categories = parser.Categories()
	r.assets = parser.Assets()
	r.forms = parser.Forms()
}

func (r *Repo) file(c component.Component) (*object.File, error) {
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

func (r *Repo) Form(id string, locale string) *component.Form {
	r.RLock()
	defer r.RUnlock()

	for _, f := range r.forms {
		if f.Id == id && f.Locale == locale {
			return f
		}
	}
	return nil
}

func (r *Repo) Category(cat, locale string) *component.Category {
	r.RLock()
	defer r.RUnlock()

	for _, c := range r.categories[locale] {
		if c.Id == cat {
			return c
		}
	}
	return nil
}

func (r *Repo) Categories(locale string) []string {
	r.RLock()
	defer r.RUnlock()
	var s []string
	for _, v := range r.categories[locale] {
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
