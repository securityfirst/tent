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
	repository, err := repo.New("klaidliadon", "octo-content", conf.OAuth())
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}
	var hookChan = make(chan struct{})
	repository.StartSync(10*time.Minute, hookChan)
	hookChan <- struct{}{}

	var engine = auth.NewEngine(conf, gin.Default())

	rcheck := repoChecker{repository}
	aCheck := authChecker{engine}

	engine.GET("/", aCheck.User, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user": engine.GetUser(c),
			"repo": repository,
		})
	})
	engine.GET("/api/repo", aCheck.User, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"categories": repository.Categories(),
		})
	})
	engine.GET("/api/repo/category/:cat", aCheck.User, rcheck.Category, func(c *gin.Context) {
		cat := c.MustGet("cat").(*repo.Category)
		c.JSON(http.StatusOK, gin.H{
			"name":          cat.Name,
			"subcategories": cat.Subcategories(),
		})
	})
	engine.GET("/api/repo/category/:cat/:sub", aCheck.User, rcheck.Sub, func(c *gin.Context) {
		sub := c.MustGet("sub").(*repo.Subcategory)
		c.JSON(http.StatusOK, gin.H{
			"name":  sub.Name,
			"items": sub.Items(),
		})
	})
	engine.GET("/api/repo/category/:cat/:sub/item/:item", aCheck.User, rcheck.Item, func(c *gin.Context) {
		item := c.MustGet("item").(*repo.Item)
		hash, err := repository.ComponentHash(item)
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
	})
	engine.Run()
}
