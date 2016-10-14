package repo

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"gopkg.in/src-d/go-git.v3"

	"github.com/google/go-github/github"
	"github.com/klaidliadon/octo/models"
	"golang.org/x/oauth2"
)

var (
	ErrNotReady     = errors.New("Repository not ready")
	ErrFileNotFound = git.ErrFileNotFound
)

func New(owner, name string, conf *oauth2.Config) (*Repo, error) {
	r, err := git.NewRepository(repoAddress(owner, name), nil)
	if err != nil {
		return nil, err
	}
	return &Repo{
		repo:  r,
		name:  name,
		owner: owner,
		conf:  conf,
	}, nil
}

type Repo struct {
	sync.RWMutex
	shutdown   chan struct{}
	owner      string
	name       string
	conf       *oauth2.Config
	repo       *git.Repository
	commit     *git.Commit
	users      []string
	categories map[string]*Category
	ticker     *time.Ticker
}

func (r *Repo) client(u *models.User) *github.Client {
	return github.NewClient(r.conf.Client(oauth2.NoContext, &u.Token))
}

func (r *Repo) Handler() RepoHandler { return RepoHandler{r} }

func (r *Repo) StartSync(interval time.Duration, trigger <-chan struct{}) {
	r.ticker = time.NewTicker(interval)
	r.shutdown = make(chan struct{})
	go r.pullLoop(trigger)
}

func (r *Repo) StopSync() { close(r.shutdown) }

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

func (r *Repo) pullLoop(trigger <-chan struct{}) {
	for {
		select {
		case <-r.ticker.C:
			r.pull()
		case <-trigger:
			r.pull()
		case <-r.shutdown:
			r.ticker.Stop()
			return
		}
	}
}

func (r *Repo) pull() error {
	r.Lock()
	defer r.Unlock()
	if err := r.repo.PullDefault(); err != nil {
		return err
	}
	hash, err := r.repo.Remotes[git.DefaultRemoteName].Head()
	if err != nil {
		return err
	}
	if r.commit != nil && r.commit.Hash == hash {
		return nil
	}
	if r.commit != nil {
		log.Println("Changing commit from", r.commit.Hash, "to", hash)
	} else {
		log.Println("Checkout with", hash)
	}
	c, err := r.repo.Commit(hash)
	if err != nil {
		return err
	}
	r.commit = c
	r.categories, err = ParseTree(r.commit.Tree().Files())
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) file(c Component) (*git.File, error) {
	if r.commit == nil {
		return nil, ErrNotReady
	}
	r.RLock()
	defer r.RUnlock()
	return r.commit.File(c.Path()[1:])
}

func (r *Repo) Get(c Component) (string, error) {
	f, err := r.file(c)
	if err != nil {
		return "", err
	}
	return f.Contents()
}

func (r *Repo) Category(cat string) *Category {
	r.RLock()
	defer r.RUnlock()
	return r.categories[cat]
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

func (r *Repo) ComponentHash(c Component) (string, error) {
	f, err := r.file(c)
	if err != nil {
		return "", err
	}
	return f.Hash.String(), nil
}

func (r *Repo) Create(c Component, u *models.User) error {
	return r.request(c, actionCreate, u)
}

func (r *Repo) Delete(c Component, u *models.User) error {
	return r.request(c, actionDelete, u)
}

func (r *Repo) Update(c Component, u *models.User) error {
	return r.request(c, actionUpdate, u)
}

func (r *Repo) request(c Component, action int, u *models.User) (err error) {
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
		_, _, err = r.client(u).Repositories.DeleteFile(r.owner, r.name, file, commit)
	}
	if err == nil {
		go r.pull()
	}
	return err
}
