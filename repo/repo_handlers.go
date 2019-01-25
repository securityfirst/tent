package repo

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/securityfirst/tent/component"
	"github.com/securityfirst/tent/models"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
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
	asset, ok := c.Get("asset")
	if ok {
		return asset.(*component.Asset)
	}
	form, ok := c.Get("form")
	if ok {
		return form.(*component.Form)
	}
	cat, ok := c.Get("cat")
	if !ok {
		return nil
	}
	sub, ok := c.Get("sub")
	if !ok {
		return cat.(component.Component)
	}
	diff, ok := c.Get("diff")
	if !ok {
		return sub.(component.Component)
	}
	if item, ok := c.Get("item"); ok {
		return item.(component.Component)
	}
	if check, ok := c.Get("checks"); ok {
		return check.(component.Component)
	}
	return diff.(component.Component)
}

func (r *RepoHandler) token(c *gin.Context) string {
	return c.MustGet("token").(string)
}

func (r *RepoHandler) user(c *gin.Context) models.User {
	return c.MustGet("user").(models.User)
}

func (r *RepoHandler) asset(c *gin.Context) *component.Asset {
	return c.MustGet("asset").(*component.Asset)
}

func (r *RepoHandler) form(c *gin.Context) *component.Form {
	return c.MustGet("form").(*component.Form)
}

func (r *RepoHandler) cat(c *gin.Context) *component.Category {
	return c.MustGet("cat").(*component.Category)
}

func (r *RepoHandler) sub(c *gin.Context) *component.Subcategory {
	return c.MustGet("sub").(*component.Subcategory)
}

func (r *RepoHandler) diff(c *gin.Context) *component.Difficulty {
	return c.MustGet("diff").(*component.Difficulty)
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
		if cat := r.repo.Category(t.ID, r.locale(c)); cat != nil {
			cmp = cat
		}
	case *component.Subcategory:
		if sub := r.cat(c).Sub(t.ID); sub != nil {
			cmp = sub
		}
	case *component.Difficulty:
		if item := r.sub(c).Difficulty(t.ID); item != nil {
			cmp = item
		}
	case *component.Item:
		if item := r.diff(c).Item(t.ID); item != nil {
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
		if cat := r.repo.Category(t.ID, r.locale(c)); cat != nil {
			cmp = cat
		}
	case *component.Subcategory:
		if sub := r.cat(c).Sub(t.ID); sub != nil {
			cmp = sub
		}
	case *component.Difficulty:
		if diff := r.sub(c).Difficulty(t.ID); diff != nil {
			cmp = diff
		}
	case *component.Item:
		if item := r.diff(c).Item(t.ID); item != nil {
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
	cat.ID, cat.Locale = c.Param("cat"), r.locale(c)
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
	sub.ID = c.Param("sub")
	log.Println(sub)
	c.Set("sub", &sub)
}

func (r *RepoHandler) SetDiff(c *gin.Context) {
	r.SetSub(c)
	diff := r.sub(c).Difficulty(c.Param("diff"))
	if diff == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("diff", diff)
}

func (r *RepoHandler) ParseDiff(c *gin.Context) {
	r.SetSub(c)
	var diff component.Difficulty
	if err := c.BindJSON(&diff); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	diff.SetParent(r.sub(c))
	diff.ID = c.Param("diff")
	c.Set("diff", &diff)
}

func (r *RepoHandler) SetItem(c *gin.Context) {
	r.SetDiff(c)
	item := r.diff(c).Item(c.Param("item"))
	if item == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("item", item)
}

func (r *RepoHandler) ParseItem(c *gin.Context) {
	r.SetDiff(c)
	var item component.Item
	if err := c.BindJSON(&item); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	item.SetParent(r.diff(c))
	item.ID = c.Param("item")
	c.Set("item", &item)
}

func (r *RepoHandler) SetCheck(c *gin.Context) {
	r.SetDiff(c)
	diff := r.diff(c)
	if diff.Checks() == nil {
		diff.SetChecks(&component.Checklist{Checks: []component.Check{}})
	}
	c.Set("checks", diff.Checks())
}

func (r *RepoHandler) ParseCheck(c *gin.Context) {
	r.SetDiff(c)
	var check component.Checklist
	if err := c.BindJSON(&check); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	check.SetParent(r.diff(c))
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
		ID:      RandStringBytesMaskImprSrc(10) + filepath.Ext(file),
		Content: b.String(),
	})
}

// SetForm loads the form using the url parameter
func (r *RepoHandler) SetForm(c *gin.Context) {
	form := r.repo.Form(c.Param("form"), r.locale(c))
	if form == nil {
		r.err(c, http.StatusNotFound, ErrNotFound)
		return
	}
	c.Set("form", form)
}

func (r *RepoHandler) ParseForm(c *gin.Context) {
	var form component.Form
	if err := c.BindJSON(&form); err != nil {
		r.err(c, http.StatusBadRequest, err)
		return
	}
	form.ID, form.Locale = c.Param("form"), r.locale(c)
	c.Set("form", &form)
}

func (r *RepoHandler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user": r.user(c),
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

func (r *RepoHandler) ShowChecks(c *gin.Context) {
	cmp := r.cmp(c)
	hash, err := r.repo.ComponentHash(cmp)
	if err != nil && err != object.ErrFileNotFound {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	cmp.(*component.Checklist).Hash = hash
	c.JSON(http.StatusOK, cmp)
}

func (r *RepoHandler) UpdateChecks(c *gin.Context) {
	hash, err := r.repo.ComponentHash(r.cmp(c))
	if err != nil && err != object.ErrFileNotFound {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	if hash != "" {
		r.Update(c)
	} else {
		r.Create(c)
	}
}

func (r *RepoHandler) Show(c *gin.Context) {
	cmp := r.cmp(c)
	hash, err := r.repo.ComponentHash(cmp)
	if err != nil {
		if _, ok := cmp.(*component.Checklist); !ok || err != object.ErrFileNotFound {
			r.err(c, http.StatusInternalServerError, err)
			return
		}
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
	case *component.Difficulty:
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
	case *component.Form:
		v := *t
		v.Hash = hash
		out = &v
	}
	c.JSON(http.StatusOK, out)
}

func (r *RepoHandler) Create(c *gin.Context) {
	if err := r.repo.Create(r.cmp(c), r.user(c), r.token(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusCreated)
}

func (r *RepoHandler) Update(c *gin.Context) {
	if err := r.repo.Update(r.cmp(c), r.user(c), r.token(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (r *RepoHandler) Delete(c *gin.Context) {
	if err := r.repo.Delete(r.cmp(c), r.user(c), r.token(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (r *RepoHandler) AssetShow(c *gin.Context) {
	a := r.asset(c)
	var ct string
	switch path.Ext(a.ID) {
	case ".png":
		ct = "image/png"
	case ".jpg", "jpeg":
		ct = "image/jpeg"
	default:
		ct = "application/octet-stream"
	}
	c.Writer.Header().Set("content-type", ct)
	c.Writer.WriteString(r.asset(c).Contents())
}

func (r *RepoHandler) AssetCreate(c *gin.Context) {
	asset := r.asset(c)
	if err := r.repo.Create(asset, r.user(c), r.token(c)); err != nil {
		r.err(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(201, gin.H{"id": asset.ID})
}

func (r *RepoHandler) Tree(c *gin.Context) {
	c.JSON(http.StatusOK, r.repo.Tree(r.locale(c), c.Query("content") == "html"))
}
