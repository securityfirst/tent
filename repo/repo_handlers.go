package repo

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RepoHandler struct {
	repo *Repo
}

func (r *RepoHandler) CheckCat(c *gin.Context) {
	cat := r.repo.Category(c.Param("cat"))
	if cat != nil {
		c.Set("cat", cat)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
	c.Abort()
}

func (r *RepoHandler) CheckSub(c *gin.Context) {
	r.CheckCat(c)
	sub := c.MustGet("cat").(*Category).Sub(c.Param("sub"))
	if sub != nil {
		c.Set("sub", sub)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "subcategory not found"})
	c.Abort()
}

func (r *RepoHandler) CheckItem(c *gin.Context) {
	r.CheckSub(c)
	item := c.MustGet("sub").(*Subcategory).Item(c.Param("item"))
	if item != nil {
		c.Set("item", item)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
	c.Abort()
}

func (r *RepoHandler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user": c.MustGet("user"),
		"repo": r.repo,
	})
}

func (r *RepoHandler) Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"categories": r.repo.Categories(),
	})
}

func (r *RepoHandler) Category(c *gin.Context) {
	cat := c.MustGet("cat").(*Category)
	c.JSON(http.StatusOK, gin.H{
		"name":          cat.Name,
		"subcategories": cat.Subcategories(),
	})
}

func (r *RepoHandler) Subcategory(c *gin.Context) {
	sub := c.MustGet("sub").(*Subcategory)
	c.JSON(http.StatusOK, gin.H{
		"name":  sub.Name,
		"items": sub.Items(),
	})
}

func (r *RepoHandler) Item(c *gin.Context) {
	item := c.MustGet("item").(*Item)
	hash, err := r.repo.ComponentHash(item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"title":      item.Title,
		"difficulty": item.Difficulty,
		"hash":       hash,
		"body":       item.Body,
	})
}
