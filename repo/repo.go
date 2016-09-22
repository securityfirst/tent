package repo

import (
	"errors"
	"fmt"
	"log"
	"path"
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
	shutdown chan struct{}
	owner    string
	name     string
	conf     *oauth2.Config
	repo     *git.Repository
	commit   *git.Commit
	ticker   *time.Ticker
}

func (r *Repo) StartSync(interval time.Duration, trigger <-chan struct{}) {
	r.ticker = time.NewTicker(interval)
	r.shutdown = make(chan struct{})
	go r.updateLoop(trigger)
}

func (r *Repo) StopSync() {
	close(r.shutdown)
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

func (r *Repo) Hash(path string) (string, error) {
	f, err := r.file(path)
	if err != nil {
		return "", err
	}
	return f.Hash.String(), nil
}

func (r *Repo) UpdateFile(file, sha string, data []byte, u *models.User) error {
	var client = github.NewClient(r.conf.Client(oauth2.NoContext, &u.Token))
	_, _, err := client.Repositories.UpdateFile(r.owner, r.name, file, newCommit(file, sha, data, u))
	return err
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
