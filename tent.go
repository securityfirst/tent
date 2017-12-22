package tent

import (
	"time"

	"log"

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
	pathAssetID     = "/api/repo/asset/:asset"
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
	root.POST(pathUpdate, func(*gin.Context) { // Hook for github
		select {
		case hookCh <- struct{}{}: // starts an update
		default: // discard
		}
	})
	locale := root.Use(h.ParseLocale)
	locale.GET(pathTree, h.Tree)
	locale.GET(pathInfo, h.Info)
	locale.GET(pathRepo, h.Root)
	locale.GET(pathCategory, h.SetCat, h.Show)
	locale.GET(pathSubcategory, h.SetSub, h.Show)
	locale.GET(pathDifficulty, h.SetDiff, h.Show)
	locale.GET(pathItem, h.SetItem, h.Show)
	locale.GET(pathCheck, h.SetCheck, h.ShowChecks)
	locale.GET(pathAssetID, h.SetAsset, h.AssetShow)
	locale.GET(pathForm, h.SetForm, h.Show)

	// Locale and Authorized handlers
	authorized := root.Use(engine.EnsureUser, h.ParseLocale)

	authorized.PUT(pathCategory, h.ParseCat, h.Update)
	authorized.DELETE(pathCategory, h.ParseCat, h.CanDelete, h.Delete)
	authorized.POST(pathCategory, h.ParseCat, h.IsNew, h.Create)

	authorized.PUT(pathSubcategory, h.ParseSub, h.Update)
	authorized.DELETE(pathSubcategory, h.ParseSub, h.CanDelete, h.Delete)
	authorized.POST(pathSubcategory, h.ParseSub, h.IsNew, h.Create)

	authorized.PUT(pathDifficulty, h.ParseDiff, h.Update)
	authorized.DELETE(pathDifficulty, h.ParseDiff, h.CanDelete, h.Delete)
	authorized.POST(pathDifficulty, h.ParseDiff, h.IsNew, h.Create)

	authorized.PUT(pathItem, h.ParseItem, h.Update)
	authorized.DELETE(pathItem, h.ParseItem, h.CanDelete, h.Delete)
	authorized.POST(pathItem, h.ParseItem, h.IsNew, h.Create)

	authorized.PUT(pathCheck, h.ParseCheck, h.UpdateChecks)

	authorized.POST(pathAsset, h.ParseAsset, h.AssetCreate)

	authorized.PUT(pathForm, h.ParseForm, h.Update)
	authorized.DELETE(pathForm, h.ParseForm, h.CanDelete, h.Delete)
	authorized.POST(pathForm, h.ParseForm, h.IsNew, h.Create)

	loop(o.repo.Pull, 10*time.Minute, hookCh)

	// Force first update
	log.Println("First repo update...")
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
