package repo

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/securityfirst/tent/component"
	"github.com/securityfirst/tent/models"
)

var (
	ErrExists      = errors.New("existing id")
	ErrNotFound    = errors.New("not found")
	ErrHasChildren = errors.New("element has children")
	ErrLanguage    = errors.New("invalid language")
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
	cat, ok := c.Get("cat")
	if !ok {
		return nil
	}
	sub, ok := c.Get("sub")
	if !ok {
		return cat.(component.Component)
	}
	if item, ok := c.Get("item"); ok {
		return item.(component.Component)
	}
	if check, ok := c.Get("checks"); ok {
		return check.(component.Component)
	}
	return sub.(component.Component)
}

func (r *RepoHandler) user(c *gin.Context) *models.User {
	return c.MustGet("user").(*models.User)
}

func (r *RepoHandler) asset(c *gin.Context) *component.Asset {
	return c.MustGet("asset").(*component.Asset)
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

func (r *RepoHandler) checks(c *gin.Context) *component.Checklist {
	return c.MustGet("checks").(*component.Checklist)
}

func (r *RepoHandler) locale(c *gin.Context) string {
	return c.MustGet("locale").(string)
}

func (r *RepoHandler) ParseLocale(c *gin.Context) {
	s := c.Request.Header.Get("X-Tent-Language")
	if s == "" {
		s = "en"
	}
	if len(s) != 2 {
		r.err(c, http.StatusBadRequest, ErrLanguage)
		return
	}
	c.Set("locale", s)
}

func (r *RepoHandler) IsNew(c *gin.Context) {
	var cmp component.Component
	switch t := r.cmp(c).(type) {
	case *component.Category:
		if cat := r.repo.Category(t.Id, r.locale(c)); cat != nil {
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
	}
	if cmp != nil {
		r.err(c, http.StatusConflict, ErrExists)
	}
}

func (r *RepoHandler) CanDelete(c *gin.Context) {
	var cmp component.Component
	switch t := r.cmp(c).(type) {
	case *component.Category:
		if cat := r.repo.Category(t.Id, r.locale(c)); cat != nil {
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
	cat := r.repo.Category(c.Param("cat"), r.locale(c))
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
	cat.Id, cat.Locale = c.Param("cat"), r.locale(c)
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
	c.Set("checks", r.sub(c).Checks())
}

func (r *RepoHandler) ParseCheck(c *gin.Context) {
	r.SetSub(c)
	var check component.Checklist
	if err := c.BindJSON(&check); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	check.SetParent(r.sub(c))
	c.Set("checks", &check)
}

func (r *RepoHandler) SetAsset(c *gin.Context) {
	c.Set("asset", r.repo.Asset(c.Param("asset")))
}

func (r *RepoHandler) ParseAsset(c *gin.Context) {
	file := c.Request.Header.Get("file")
	if file == "" {
		file = "upload.jpg"
	}
	b := bytes.NewBuffer(nil)
	io.Copy(b, c.Request.Body)
	c.Set("asset", &component.Asset{
		Id:      RandStringBytesMaskImprSrc(10) + filepath.Ext(file),
		Content: b.String(),
	})
}

func (r *RepoHandler) Info(c *gin.Context) {
	u := r.user(c)
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"name":  u.Name,
			"login": u.Login,
			"email": u.Email,
		},
		"repo": r.repo,
	})
}

func (r *RepoHandler) Root(c *gin.Context) {
	cats := r.repo.Categories(r.locale(c))
	sort.Strings(cats)
	c.JSON(http.StatusOK, gin.H{
		"categories": cats,
	})
}

func (r *RepoHandler) Show(c *gin.Context) {
	cmp := r.cmp(c)
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
	case *component.Checklist:
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

func (r *RepoHandler) AssetShow(c *gin.Context) {
	c.Writer.WriteHeader(200)
	c.Writer.WriteString(r.asset(c).Contents())
}

func (r *RepoHandler) AssetCreate(c *gin.Context) {
	asset := r.asset(c)
	if err := r.repo.Create(asset, r.user(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(201, gin.H{"id": asset.Id})
}

func (r *RepoHandler) Tree(c *gin.Context) {
	c.JSON(http.StatusOK, r.repo.Tree(r.locale(c), c.Query("content") == "html"))
}
