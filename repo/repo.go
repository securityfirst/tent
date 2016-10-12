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
	categories map[string]*Category
	ticker     *time.Ticker
}

func (r *Repo) StartSync(interval time.Duration, trigger <-chan struct{}) {
	r.ticker = time.NewTicker(interval)
	r.shutdown = make(chan struct{})
	go r.updateLoop(trigger)
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

func (r *Repo) updateLoop(trigger <-chan struct{}) {
	for {
		select {
		case <-r.ticker.C:
			r.update()
		case <-trigger:
			r.update()
		case <-r.shutdown:
			r.ticker.Stop()
			return
		}
	}
}

func (r *Repo) update() error {
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
		log.Println("No changes")
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

func (r *Repo) file(path string) (*git.File, error) {
	if r.commit == nil {
		return nil, ErrNotReady
	}
	r.RLock()
	defer r.RUnlock()
	return r.commit.File(path)
}

func (r *Repo) Get(path string) (string, error) {
	f, err := r.file(path)
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
	f, err := r.file(c.Path()[1:])
	if err != nil {
		return "", err
	}
	return f.Hash.String(), nil
}

func (r *Repo) Update(c Component, sha string, u *models.User) error {
	client := github.NewClient(r.conf.Client(oauth2.NoContext, &u.Token))
	file := c.Path()[1:]
	commit := newCommit(file, sha, []byte(c.Contents()), u)
	_, _, err := client.Repositories.UpdateFile(r.owner, r.name, file, commit)
	return err
}
