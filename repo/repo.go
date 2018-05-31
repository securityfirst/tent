package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/securityfirst/tent.v2/component"
	"gopkg.in/securityfirst/tent.v2/models"
)

var logger = log.New(os.Stdout, "[repo]", log.Ltime|log.Lshortfile)

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

func Local(dir, branch string) (*Repo, error) {
	logger.Printf("Using %q", dir)
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: fmt.Sprintf("file://%s", dir)})
	if err != nil {
		return nil, err
	}
	if branch == "" {
		branch = "master"
	}
	return &Repo{repo: r, name: path.Base(dir), branch: branch}, nil
}

func New(owner, name, branch string) (*Repo, error) {
	address := repoAddress(owner, name)
	logger.Printf("Using %q", address)
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: address})
	if err != nil {
		return nil, err
	}
	if branch == "" {
		branch = "master"
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
		ass[i] = r.assets[i].ID
	}

	var forms = make([]*component.Form, 0)
	for i := range r.forms {
		if r.forms[i].Locale != locale {
			continue
		}
		forms = append(forms, r.forms[i])
	}

	return map[string]interface{}{
		"categories": cats,
		"assets":     ass,
		"forms":      forms,
	}
}

func (r *Repo) client(token string) *github.Client {
	return github.NewClient(r.conf.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token}))
}

func (r *Repo) Handler() RepoHandler { return RepoHandler{r} }

func (r *Repo) Locale() []string {
	var locale = make([]string, 0, len(r.categories))
	for k := range r.categories {
		locale = append(locale, k)
	}
	return locale
}

func (r *Repo) All(locale string) []component.Component {
	var list []component.Component
	for _, cat := range r.categories[locale] {
		list = append(list, cat)
		for _, id := range cat.Subcategories() {
			sub := cat.Sub(id)
			list = append(list, sub)
			for _, diff := range sub.Difficulties() {
				for _, id := range diff.ItemNames() {
					list = append(list, diff.Item(id))
				}
				if check := diff.Checks(); check.HasChildren() {
					list = append(list, check)
				}
			}
		}
	}
	for _, form := range r.forms {
		if form.Locale != locale {
			continue
		}
		list = append(list, form)
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
		logger.Println("Pull failed:", err)
		return
	}
	branch := plumbing.ReferenceName("refs/remotes/origin/" + r.branch)
	hash, err := r.repo.Reference(branch, false)
	if err != nil {
		logger.Printf("Reference %q failed:", branch, err)
		return
	}
	if r.commit != nil && r.commit.Hash == hash.Hash() {
		return
	}
	if r.commit != nil {
		logger.Println("Changing commit from", r.commit.Hash, "to", hash)
	} else {
		logger.Println("Checkout with", hash)
	}
	r.commit, err = r.repo.CommitObject(hash.Hash())
	if err != nil {
		logger.Println("Commit failed:", err)
		return
	}
	var parser component.Parser
	tree, err := r.commit.Tree()
	if err != nil {
		logger.Println("Tree failed:", err)
		return
	}
	if err := parser.Parse(tree); err != nil {
		logger.Println("Parsing failed:", err)
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
	return r.commit.File(c.Path()[:])
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
		if a.ID == id {
			return a
		}
	}
	return nil
}

func (r *Repo) Forms(locale string) []string {
	r.RLock()
	defer r.RUnlock()
	var s []string
	for _, v := range r.forms {
		if v.Locale == locale {
			s = append(s, v.ID)
		}
	}
	return s
}

func (r *Repo) Form(id string, locale string) *component.Form {
	r.RLock()
	defer r.RUnlock()

	for _, f := range r.forms {
		if f.ID == id && f.Locale == locale {
			return f
		}
	}
	return nil
}

func (r *Repo) Category(cat, locale string) *component.Category {
	r.RLock()
	defer r.RUnlock()

	for _, c := range r.categories[locale] {
		if c.ID == cat {
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
		s = append(s, v.ID)
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

func (r *Repo) Create(c component.Component, u models.User, token string) error {
	return r.request(c, actionCreate, u, token)
}

func (r *Repo) Delete(c component.Component, u models.User, token string) error {
	return r.request(c, actionDelete, u, token)
}

func (r *Repo) Update(c component.Component, u models.User, token string) error {
	return r.request(c, actionUpdate, u, token)
}

func (r *Repo) request(c component.Component, action int, u models.User, token string) (err error) {
	file := c.Path()
	msg := fmt.Sprintf("%s %s", commitMsg[action], file)
	commit := &github.RepositoryContentFileOptions{
		Message: &msg, Author: u.AsAuthor(),
	}
	switch action {
	case actionCreate:
		commit.Content = []byte(c.Contents())
		_, _, err = r.client(token).Repositories.CreateFile(r.owner, r.name, file, commit)
	case actionUpdate:
		commit.SHA = strPtr(c.SHA())
		commit.Content = []byte(c.Contents())
		_, _, err = r.client(token).Repositories.UpdateFile(r.owner, r.name, file, commit)
	case actionDelete:
		commit.SHA = strPtr(c.SHA())
		_, _, err = r.client(token).Repositories.DeleteFile(r.owner, r.name, file, commit)
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
