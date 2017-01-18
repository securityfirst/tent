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
	OAuthHost:    "http://127.0.0.1:8080",
	Login:        auth.HandleConf{"/auth/login", "/"},
	Logout:       auth.HandleConf{"/auth/logout", "/"},
	Callback:     auth.HandleConf{"/auth/callback", "/"},
	RandomString: "horse battery staple",
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
	r, err := repo.New("klaidliadon", "tent-content")
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}

	o := tent.New(r)
	o.Register(e.Group("/v2"), conf)
	e.Run(":8080")
}
