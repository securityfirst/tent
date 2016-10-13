package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/klaidliadon/octo/auth"
	"github.com/klaidliadon/octo/repo"
)

type authChecker struct {
	engine *auth.Engine
}

func (a authChecker) User(c *gin.Context) {
	user := a.engine.GetUser(c)
	if user != nil {
		c.Set("user", user)
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	c.Abort()
}

type repoHandler struct {
	repo *repo.Repo
}

func (r repoHandler) CheckCat(c *gin.Context) {
	cat := r.repo.Category(c.Param("cat"))
	if cat != nil {
		c.Set("cat", cat)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
	c.Abort()
}

func (r repoHandler) CheckSub(c *gin.Context) {
	r.CheckCat(c)
	sub := c.MustGet("cat").(*repo.Category).Sub(c.Param("sub"))
	if sub != nil {
		c.Set("sub", sub)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "subcategory not found"})
	c.Abort()
}

func (r repoHandler) CheckItem(c *gin.Context) {
	r.CheckSub(c)
	item := c.MustGet("sub").(*repo.Subcategory).Item(c.Param("item"))
	if item != nil {
		c.Set("item", item)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
	c.Abort()
}

func (r repoHandler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user": c.MustGet("user"),
		"repo": r.repo,
	})
}

func (r repoHandler) Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"categories": r.repo.Categories(),
	})
}

func (r repoHandler) Category(c *gin.Context) {
	cat := c.MustGet("cat").(*repo.Category)
	c.JSON(http.StatusOK, gin.H{
		"name":          cat.Name,
		"subcategories": cat.Subcategories(),
	})
}

func (r repoHandler) Subcategory(c *gin.Context) {
	sub := c.MustGet("sub").(*repo.Subcategory)
	c.JSON(http.StatusOK, gin.H{
		"name":  sub.Name,
		"items": sub.Items(),
	})
}

func (r repoHandler) Item(c *gin.Context) {
	item := c.MustGet("item").(*repo.Item)
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
