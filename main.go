package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klaidliadon/octo/auth"
	"github.com/klaidliadon/octo/repo"
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
	r, err := repo.New("klaidliadon", "octo-content", conf.OAuth())
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}
	var hookChan = make(chan struct{})
	r.StartSync(10*time.Minute, hookChan)
	hookChan <- struct{}{} // Force first update

	engine := auth.NewEngine(conf, gin.Default())

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

	authorized.GET("/api/repo/category/:cat", h.SetCat, h.Show)
	authorized.DELETE("/api/repo/category/:cat", h.ParseCat, h.CanDelete, h.Delete)
	authorized.PUT("/api/repo/category/:cat", h.ParseCat, h.Update)
	authorized.POST("/api/repo/category/:cat", h.ParseCat, h.IsNew, h.Create)

	authorized.GET("/api/repo/category/:cat/:sub", h.SetSub, h.Show)
	authorized.DELETE("/api/repo/category/:cat/:sub", h.ParseSub, h.CanDelete, h.Delete)
	authorized.PUT("/api/repo/category/:cat/:sub", h.ParseSub, h.Update)
	authorized.POST("/api/repo/category/:cat/:sub", h.ParseSub, h.IsNew, h.Create)

	authorized.GET("/api/repo/category/:cat/:sub/item/:item", h.SetItem, h.Show)
	authorized.DELETE("/api/repo/category/:cat/:sub/item/:item", h.ParseItem, h.CanDelete, h.Delete)
	authorized.PUT("/api/repo/category/:cat/:sub/item/:item", h.ParseItem, h.Update)
	authorized.POST("/api/repo/category/:cat/:sub/item/:item", h.ParseItem, h.IsNew, h.Create)
	engine.Run()
}
