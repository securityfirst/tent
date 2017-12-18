package tent

import (
	"net/http"
	"time"

	"github.com/cenkalti/log"
	"github.com/gin-gonic/gin"
	"gopkg.in/securityfirst/tent.v2/auth"
	"gopkg.in/securityfirst/tent.v2/repo"
)

const (
	pathInfo        = "/"
	pathTree        = "/api/tree"
	pathRepo        = "/api/repo"
	pathUpdate      = "/api/repo/update"
	pathCategory    = "/api/repo/category/:cat"
	pathSubcategory = "/api/repo/category/:cat/:sub"
	pathDifficulty  = "/api/repo/category/:cat/:sub/:diff"
	pathItem        = "/api/repo/category/:cat/:sub/:diff/item/:item"
	pathCheck       = "/api/repo/category/:cat/:sub/:diff/checks"
	pathAsset       = "/api/repo/asset"
	pathAssetId     = "/api/repo/asset/:asset"
	pathForm        = "/api/repo/form/:form"
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
	// Hook for github
	root.POST(pathUpdate, func(*gin.Context) {
		select {
		case hookCh <- struct{}{}:
			// starts an update
		default:
			// discard
		}
	})

	// Locale and Authorized handlers
	locale := root.Use(h.ParseLocale)
	authorized := locale.Use(func(c *gin.Context) {
		user := engine.GetUser(c)
		if user != nil {
			c.Set("user", user)
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		c.Abort()
	})

	locale.GET(pathInfo, h.Info)
	locale.GET(pathRepo, h.Root)

	locale.GET(pathCategory, h.SetCat, h.Show)
	authorized.PUT(pathCategory, h.ParseCat, h.Update)
	authorized.DELETE(pathCategory, h.ParseCat, h.CanDelete, h.Delete)
	authorized.POST(pathCategory, h.ParseCat, h.IsNew, h.Create)

	locale.GET(pathSubcategory, h.SetSub, h.Show)
	authorized.PUT(pathSubcategory, h.ParseSub, h.Update)
	authorized.DELETE(pathSubcategory, h.ParseSub, h.CanDelete, h.Delete)
	authorized.POST(pathSubcategory, h.ParseSub, h.IsNew, h.Create)

	locale.GET(pathDifficulty, h.SetDiff, h.Show)
	authorized.PUT(pathDifficulty, h.ParseDiff, h.Update)
	authorized.DELETE(pathDifficulty, h.ParseDiff, h.CanDelete, h.Delete)
	authorized.POST(pathDifficulty, h.ParseDiff, h.IsNew, h.Create)

	locale.GET(pathItem, h.SetItem, h.Show)
	authorized.PUT(pathItem, h.ParseItem, h.Update)
	authorized.DELETE(pathItem, h.ParseItem, h.CanDelete, h.Delete)
	authorized.POST(pathItem, h.ParseItem, h.IsNew, h.Create)

	locale.GET(pathCheck, h.SetCheck, h.ShowChecks)
	authorized.PUT(pathCheck, h.ParseCheck, h.UpdateChecks)

	locale.GET(pathAssetId, h.SetAsset, h.AssetShow)
	authorized.POST(pathAsset, h.ParseAsset, h.AssetCreate)

	locale.GET(pathForm, h.SetForm, h.Show)
	authorized.PUT(pathForm, h.ParseForm, h.Update)
	authorized.DELETE(pathForm, h.ParseForm, h.CanDelete, h.Delete)
	authorized.POST(pathForm, h.ParseForm, h.IsNew, h.Create)

	loop(o.repo.Pull, 10*time.Minute, hookCh)

	// Force first update
	log.Info("First repo update...")
	hookCh <- struct{}{}

}

func loop(action func(), every time.Duration, trigger <-chan struct{}) <-chan struct{} {
	t := time.NewTicker(every)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-t.C:
				action()
			case <-trigger:
				action()
			case <-stop:
				t.Stop()
				return
			}
		}
	}()
	return stop
}
