package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/securityfirst/tent"
	"github.com/securityfirst/tent/auth"
	"github.com/securityfirst/tent/repo"
)

var conf = auth.Config{
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
	e := gin.Default()
	root := &e.RouterGroup
	r, err := repo.New("securityfirst", "tent-content", conf.OAuth(root))
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}

	o := tent.New(r)
	o.Register(root, conf)
	e.Run(":2015")
}
