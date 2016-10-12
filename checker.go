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

func (a *authChecker) User(c *gin.Context) {
	u := a.engine.GetUser(c)
	if u != nil {
		return
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
	c.Abort()
}

type repoChecker struct {
	repo *repo.Repo
}

func (r repoChecker) Category(c *gin.Context) {
	cat := r.repo.Category(c.Param("cat"))
	if cat != nil {
		c.Set("cat", cat)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
	c.Abort()
}

func (r repoChecker) Sub(c *gin.Context) {
	r.Category(c)
	sub := c.MustGet("cat").(*repo.Category).Sub(c.Param("sub"))
	if sub != nil {
		c.Set("sub", sub)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "subcategory not found"})
	c.Abort()
}

func (r repoChecker) Item(c *gin.Context) {
	r.Sub(c)
	item := c.MustGet("sub").(*repo.Subcategory).Item(c.Param("item"))
	if item != nil {
		c.Set("item", item)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
	c.Abort()
}
