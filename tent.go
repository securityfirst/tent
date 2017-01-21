package tent

import (
	"net/http"
	"time"

	"github.com/cenkalti/log"
	"github.com/gin-gonic/gin"
	"github.com/securityfirst/tent/auth"
	"github.com/securityfirst/tent/repo"
)

const (
	pathInfo        = "/"
	pathTree        = "/api/tree"
	pathRepo        = "/api/repo"
	pathUpdate      = "/api/repo/update"
	pathCategory    = "/api/repo/category/:cat"
	pathSubcategory = "/api/repo/category/:cat/:sub"
	pathItem        = "/api/repo/category/:cat/:sub/item/:item"
	pathCheck       = "/api/repo/category/:cat/:sub/checks"
)

func New(r *repo.Repo) *Tent {
	return &Tent{repo: r}
}

type Tent struct {
	repo *repo.Repo
}

func (o *Tent) Register(root *gin.RouterGroup, c auth.Config) {
	var (
		engine = auth.NewEngine(c, root)
		hookCh = make(chan struct{})
		h      = o.repo.Handler()
	)

	o.repo.SetConf(c.OAuth(root))
	// Free handlers
	root.GET(pathTree, h.ParseLocale, h.Tree)

	// Authorized handlers
	authorized := root.Use(h.ParseLocale, func(c *gin.Context) {
		user := engine.GetUser(c)
		if user != nil {
			c.Set("user", user)
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		c.Abort()
	})

	authorized.GET(pathInfo, h.Info)
	authorized.GET(pathRepo, h.Root)
	// Hook for github
	authorized.POST(pathUpdate, func(*gin.Context) { hookCh <- struct{}{} })

	authorized.GET(pathCategory, h.SetCat, h.Show)
	authorized.PUT(pathCategory, h.ParseCat, h.Update)
	authorized.DELETE(pathCategory, h.ParseCat, h.CanDelete, h.Delete)
	authorized.POST(pathCategory, h.ParseCat, h.IsNew, h.Create)

	authorized.GET(pathSubcategory, h.SetSub, h.Show)
	authorized.PUT(pathSubcategory, h.ParseSub, h.Update)
	authorized.DELETE(pathSubcategory, h.ParseSub, h.CanDelete, h.Delete)
	authorized.POST(pathSubcategory, h.ParseSub, h.IsNew, h.Create)

	authorized.GET(pathItem, h.SetItem, h.Show)
	authorized.PUT(pathItem, h.ParseItem, h.Update)
	authorized.DELETE(pathItem, h.ParseItem, h.CanDelete, h.Delete)
	authorized.POST(pathItem, h.ParseItem, h.IsNew, h.Create)

	authorized.GET(pathCheck, h.SetCheck, h.Show)
	authorized.PUT(pathCheck, h.ParseCheck, h.Update)

	o.repo.StartSync(10*time.Minute, hookCh)
	// Force first update
	log.Info("First repo update...")
	hookCh <- struct{}{}

}
