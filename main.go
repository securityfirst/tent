package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/klaidliadon/octo/auth"
	"github.com/klaidliadon/octo/repo"
)

var conf = auth.Config{
	Port:      7000,
	OAuthHost: "http://127.0.0.1:2015",
	Login:     auth.HandleConf{"/github/login", ""},
	Logout:    auth.HandleConf{"/github/logout", "/"},
	Callback:  auth.HandleConf{"/github/callback", "/profile"},
}

var (
	hWebhook  = auth.HandleConf{"/api/webhook", ""}
	hLogin    = auth.HandleConf{"/", "/profile"}
	hRepoView = auth.HandleConf{"/api/repo/view/", ""}
	hRepoEdit = auth.HandleConf{"/api/repo/edit/", ""}
)

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

const sampleFile = "contents.md"

func main() {
	repo, err := repo.New("klaidliadon", "octo-content", conf.OAuth())
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}
	var hookChan = make(chan struct{})
	repo.StartSync(10*time.Minute, hookChan)
	hookChan <- struct{}{}

	var a = auth.New(conf, http.DefaultServeMux)
	a.AsNoone(hWebhook, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case hookChan <- struct{}{}:
			log.Println("Webook triggered: sending update")
		default: //Webook discarded: update pending
		}
	}))
	a.AsNoone(hLogin, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, tplLogin, conf.Login.Endpoint)
	}))
	a.AsSomeone(hLogin.Reverse(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := a.GetUser(r)
		fmt.Fprintf(w, tplProfile,
			u.Login, u.Email,
			repo,
			hRepoEdit.Endpoint+sampleFile,
			hRepoView.Endpoint+sampleFile,
			conf.Logout.Endpoint)
	}))
	a.AsSomeone(hRepoView, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := extractPath(r.URL.Path, hRepoView.Endpoint)
		content, err := repo.Get(path)
		if err != nil {
			handleError(w, err)
			return
		}
		fmt.Fprint(w, content)
	}))
	a.AsSomeone(hRepoEdit, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := extractPath(r.URL.Path, hRepoEdit.Endpoint)
		contents, err := repo.Get(file)
		if err != nil {
			handleError(w, err)
			return
		}
		u := a.GetUser(r)
		if r.Method == "GET" {
			hash, err := repo.Hash(file)
			if err != nil {
				handleError(w, err)
				return
			}
			fmt.Fprintf(w, tplEdit, u.Login, u.Email, file, contents, r.URL.Path, contents, hash)
			return
		}
		if err = repo.UpdateFile(file, r.FormValue("hash"), []byte(r.FormValue("contents")), u); err != nil {
			handleError(w, err)
			return
		}
		hookChan <- struct{}{}

		http.Redirect(w, r, hRepoView.Endpoint+file, http.StatusTemporaryRedirect)
	}))

	log.Printf("Running on :%v", conf.Port)
	a.Run()
}

func extractPath(path, strip string) string {
	return path[len(strip):]
}

func handleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	code := http.StatusInternalServerError
	if err == repo.ErrFileNotFound {
		code = http.StatusNotFound
	}
	http.Error(w, err.Error(), code)
}
