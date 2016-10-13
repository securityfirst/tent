package main

import (
	"flag"
	"log"
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
	hookChan <- struct{}{}

	var engine = auth.NewEngine(conf, gin.Default())

	repo := repoHandler{r}

	var authorized = engine.Use(authChecker{engine}.User)

	authorized.GET("/", repo.Info)
	authorized.GET("/api/repo", repo.Root)
	authorized.GET("/api/repo/category/:cat", repo.CheckCat, repo.Category)
	authorized.GET("/api/repo/category/:cat/:sub", repo.CheckSub, repo.Subcategory)
	engine.GET("/api/repo/category/:cat/:sub/item/:item", repo.CheckItem, repo.Item)
	engine.Run()
}
