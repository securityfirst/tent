package repo

import (
	"errors"
	"log"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/klaidliadon/octo/models"
	"github.com/klaidliadon/octo/repo/component"
)

var (
	ErrExists      = errors.New("existing id")
	ErrNotFound    = errors.New("not found")
	ErrHasChildren = errors.New("element has children")
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

func (r *RepoHandler) cmp(c *gin.Context) component.Component {
	var cmp component.Component
	if v, ok := c.Get("cat"); ok {
		cmp = v.(*component.Category)
		if v, ok := c.Get("sub"); ok {
			cmp = v.(*component.Subcategory)
			if v, ok := c.Get("item"); ok {
				cmp = v.(*component.Item)
			}
			if v, ok := c.Get("check"); ok {
				cmp = v.(*component.Check)
			}
		}
	}
	return cmp
}

func (r *RepoHandler) user(c *gin.Context) *models.User {
	return c.MustGet("user").(*models.User)
}

func (r *RepoHandler) cat(c *gin.Context) *component.Category {
	return c.MustGet("cat").(*component.Category)
}

func (r *RepoHandler) sub(c *gin.Context) *component.Subcategory {
	return c.MustGet("sub").(*component.Subcategory)
}

func (r *RepoHandler) item(c *gin.Context) *component.Item {
	return c.MustGet("item").(*component.Item)
}

func (r *RepoHandler) check(c *gin.Context) *component.Check {
	return c.MustGet("check").(*component.Check)
}

func (r *RepoHandler) IsNew(c *gin.Context) {
	var cmp component.Component
	switch t := r.cmp(c).(type) {
	case *component.Category:
		if cat := r.repo.Category(t.Id); cat != nil {
			cmp = cat
		}
	case *component.Subcategory:
		if sub := r.cat(c).Sub(t.Id); sub != nil {
			cmp = sub
		}
	case *component.Item:
		if item := r.sub(c).Item(t.Id); item != nil {
			cmp = item
		}
	case *component.Check:
		if check := r.sub(c).Check(t.Id); check != nil {
			cmp = check
		}
	}
	if cmp != nil {
		r.err(c, http.StatusConflict, ErrExists)
	}
}

func (r *RepoHandler) CanDelete(c *gin.Context) {
	var cmp component.Component
	switch t := r.cmp(c).(type) {
	case *component.Category:
		if cat := r.repo.Category(t.Id); cat != nil {
			cmp = cat
		}
	case *component.Subcategory:
		if sub := r.cat(c).Sub(t.Id); sub != nil {
			cmp = sub
		}
	case *component.Item:
		if item := r.sub(c).Item(t.Id); item != nil {
			cmp = item
		}
	case *component.Check:
		if item := r.sub(c).Check(t.Id); item != nil {
			cmp = item
		}
	}
	if cmp == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	if cmp.HasChildren() {
		r.err(c, http.StatusForbidden, ErrHasChildren)
		return
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
	var cat component.Category
	if err := c.BindJSON(&cat); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	cat.Id = c.Param("cat")
	c.Set("cat", &cat)
}

func (r *RepoHandler) SetSub(c *gin.Context) {
	r.SetCat(c)
	sub := r.cat(c).Sub(c.Param("sub"))
	if sub == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("sub", sub)
}

func (r *RepoHandler) ParseSub(c *gin.Context) {
	r.SetCat(c)
	var sub component.Subcategory
	if err := c.BindJSON(&sub); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	sub.SetParent(r.cat(c))
	sub.Id = c.Param("sub")
	log.Println(sub)
	c.Set("sub", &sub)
}

func (r *RepoHandler) SetItem(c *gin.Context) {
	r.SetSub(c)
	item := r.sub(c).Item(c.Param("item"))
	if item == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("item", item)
}

func (r *RepoHandler) ParseItem(c *gin.Context) {
	r.SetSub(c)
	var item component.Item
	if err := c.BindJSON(&item); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	item.SetParent(r.sub(c))
	item.Id = c.Param("item")
	c.Set("item", &item)
}

func (r *RepoHandler) SetCheck(c *gin.Context) {
	r.SetSub(c)
	check := r.sub(c).Check(c.Param("check"))
	if check == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("check", check)
}

func (r *RepoHandler) ParseCheck(c *gin.Context) {
	r.SetSub(c)
	var check component.Check
	if err := c.BindJSON(&check); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	check.SetParent(r.sub(c))
	check.Id = c.Param("check")
	c.Set("check", &check)
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
	cmp := r.cmp(c)
	log.Println(cmp)
	hash, err := r.repo.ComponentHash(cmp)
	if err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	var out interface{}
	switch t := cmp.(type) {
	case *component.Category:
		v := *t
		v.Hash = hash
		out = &v
	case *component.Subcategory:
		v := *t
		v.Hash = hash
		out = &v
	case *component.Item:
		v := *t
		v.Hash = hash
		out = &v
	case *component.Check:
		v := *t
		v.Hash = hash
		out = &v
	}
	c.JSON(http.StatusOK, out)
}

func (r *RepoHandler) Create(c *gin.Context) {
	if err := r.repo.Create(r.cmp(c), r.user(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
	}
	c.Writer.WriteHeader(http.StatusCreated)
}

func (r *RepoHandler) Update(c *gin.Context) {
	if err := r.repo.Update(r.cmp(c), r.user(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (r *RepoHandler) Delete(c *gin.Context) {
	if err := r.repo.Delete(r.cmp(c), r.user(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
