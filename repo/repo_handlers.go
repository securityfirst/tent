package repo

import (
	"errors"
	"log"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/klaidliadon/octo/models"
)

var (
	ErrExists   = errors.New("existing id")
	ErrNotFound = errors.New("not found")
)

type RepoHandler struct {
	repo *Repo
}

func (r *RepoHandler) err(c *gin.Context, status int, err error) {
	if re, ok := err.(*github.ErrorResponse); ok {
		status = re.Response.StatusCode
		err = errors.New(re.Message)
	}
	c.JSON(status, gin.H{"error": err.Error()})
	c.Abort()
}

func (r *RepoHandler) getCmp(c *gin.Context) Component {
	var cmp Component
	if v, ok := c.Get("cat"); ok {
		cmp = v.(*Category)
		if v, ok := c.Get("sub"); ok {
			cmp = v.(*Subcategory)
			if v, ok := c.Get("item"); ok {
				cmp = v.(*Item)
			}
		}
	}
	return cmp
}

func (r *RepoHandler) getUser(c *gin.Context) *models.User {
	return c.MustGet("user").(*models.User)
}

func (r *RepoHandler) IsNew(c *gin.Context) {
	var cmp Component
	switch t := r.getCmp(c).(type) {
	case *Category:
		if cat := r.repo.Category(t.Id); cat != nil {
			cmp = cat
		}
	case *Subcategory:
		if sub := t.parent.Sub(t.Id); sub != nil {
			cmp = sub
		}
	case *Item:
		if item := t.parent.Item(t.Id); item != nil {
			cmp = item
		}
	}
	if cmp != nil {
		r.err(c, http.StatusConflict, ErrExists)
	}
}

// SetCat loads the category using the url parameter
func (r *RepoHandler) SetCat(c *gin.Context) {
	cat := r.repo.Category(c.Param("cat"))
	if cat == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("cat", cat)
}

func (r *RepoHandler) ParseCat(c *gin.Context) {
	var cat Category
	if err := c.BindJSON(&cat); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	cat.Id = c.Param("cat")
	c.Set("cat", &cat)
}

func (r *RepoHandler) SetSub(c *gin.Context) {
	r.SetCat(c)
	sub := c.MustGet("cat").(*Category).Sub(c.Param("sub"))
	if sub == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("sub", sub)
}

func (r *RepoHandler) ParseSub(c *gin.Context) {
	r.SetCat(c)
	var sub Subcategory
	if err := c.BindJSON(&sub); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	sub.parent = c.MustGet("cat").(*Category)
	sub.Id = c.Param("sub")
	c.Set("sub", &sub)
}

func (r *RepoHandler) SetItem(c *gin.Context) {
	r.SetSub(c)
	item := c.MustGet("sub").(*Subcategory).Item(c.Param("item"))
	if item == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("item", item)
}

func (r *RepoHandler) ParseItem(c *gin.Context) {
	r.SetSub(c)
	var item Item
	if err := c.BindJSON(&item); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	item.parent = c.MustGet("sub").(*Subcategory)
	item.Id = c.Param("item")
	c.Set("item", &item)
}

func (r *RepoHandler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user": c.MustGet("user"),
		"repo": r.repo,
	})
}

func (r *RepoHandler) Root(c *gin.Context) {
	cats := r.repo.Categories()
	sort.Strings(cats)
	c.JSON(http.StatusOK, gin.H{
		"categories": cats,
	})
}

func (r *RepoHandler) Show(c *gin.Context) {
	cmp := r.getCmp(c)
	log.Println(cmp)
	hash, err := r.repo.ComponentHash(cmp)
	if err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	var out interface{}
	switch t := cmp.(type) {
	case *Category:
		v := *t
		v.Hash = hash
		out = &v
	case *Subcategory:
		v := *t
		v.Hash = hash
		out = &v
	case *Item:
		v := *t
		v.Hash = hash
		out = &v
	}
	c.JSON(http.StatusOK, out)
}

func (r *RepoHandler) Create(c *gin.Context) {
	if err := r.repo.Create(r.getCmp(c), r.getUser(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
	}
	c.Writer.WriteHeader(http.StatusCreated)
}

func (r *RepoHandler) Update(c *gin.Context) {
	cmp := r.getCmp(c)
	if err := r.repo.Update(cmp, r.getUser(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
