package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/securityfirst/octo/auth"
	"github.com/securityfirst/octo/repo"
)

var conf = auth.Config{
	Port:      7000,
	OAuthHost: "http://127.0.0.1:2015",
	Login:     auth.HandleConf{"/auth/login", "/"},
	Logout:    auth.HandleConf{"/auth/logout", "/"},
	Callback:  auth.HandleConf{"/auth/callback", "/"},
}

func init() {
	flag.StringVar(&conf.Id, "id", os.Getenv("O_ID"), "Github Application ID")
	flag.StringVar(&conf.Secret, "secret", os.Getenv("O_SECRET"), "Github Application Secret")
	flag.Parse()
	if conf.Id == "" || conf.Secret == "" {
		flag.Usage()
		os.Exit(1)
	}
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {
	r, err := repo.New("securityfirst", "octo-content", conf.OAuth())
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}
	var hookChan = make(chan struct{})
	r.StartSync(10*time.Minute, hookChan)
	hookChan <- struct{}{} // Force first update

	engine := auth.NewEngine(conf, gin.Default())

	engine.GET("/api/tree", func(c *gin.Context) {
		c.JSON(http.StatusOK, r.Tree())
	})

	authorized := engine.Use(func(c *gin.Context) {
		user := engine.GetUser(c)
		if user != nil {
			c.Set("user", user)
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		c.Abort()
	})

	h := r.Handler()
	authorized.GET("/", h.Info)
	authorized.GET("/api/repo", h.Root)
	authorized.GET("/api/repo/update", func(*gin.Context) { hookChan <- struct{}{} })

	const (
		category    = "/api/repo/category/:cat"
		subcategory = "/api/repo/category/:cat/:sub"
		item        = "/api/repo/category/:cat/:sub/item/:item"
		check       = "/api/repo/category/:cat/:sub/check/:check"
	)

	authorized.GET(category, h.SetCat, h.Show)
	authorized.PUT(category, h.ParseCat, h.Update)
	authorized.DELETE(category, h.ParseCat, h.CanDelete, h.Delete)
	authorized.POST(category, h.ParseCat, h.IsNew, h.Create)

	authorized.GET(subcategory, h.SetSub, h.Show)
	authorized.PUT(subcategory, h.ParseSub, h.Update)
	authorized.DELETE(subcategory, h.ParseSub, h.CanDelete, h.Delete)
	authorized.POST(subcategory, h.ParseSub, h.IsNew, h.Create)

	authorized.GET(item, h.SetItem, h.Show)
	authorized.PUT(item, h.ParseItem, h.Update)
	authorized.DELETE(item, h.ParseItem, h.CanDelete, h.Delete)
	authorized.POST(item, h.ParseItem, h.IsNew, h.Create)

	authorized.GET(check, h.SetCheck, h.Show)
	authorized.PUT(check, h.ParseCheck, h.Update)
	authorized.DELETE(check, h.ParseCheck, h.CanDelete, h.Delete)
	authorized.POST(check, h.ParseCheck, h.IsNew, h.Create)

	engine.Run()
}
