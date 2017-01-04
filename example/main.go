package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/securityfirst/octo"
	"github.com/securityfirst/octo/auth"
	"github.com/securityfirst/octo/repo"
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
	r, err := repo.New("securityfirst", "octo-content", conf.OAuth())
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}

	o := octo.New(r)
	o.Register(&e.RouterGroup, conf)
	e.Run(":2015")
}
