package main

import (
	"encoding/json"
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

func init() {
	flag.StringVar(&conf.Id, "id", "", "Github Application ID")
	flag.StringVar(&conf.Secret, "secret", "", "Github Application Secret")
	flag.Parse()
	if conf.Id == "" || conf.Secret == "" {
		flag.Usage()
		os.Exit(1)
	}
	log.SetFlags(log.Ltime | log.Lshortfile)
}

func main() {
	const (
		repoOwner = "klaidliadon"
		repoName  = "octo-content"
	)
	r, err := repo.New(repoOwner, repoName, conf.OAuth())
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}

	var hookChan = make(chan struct{})
	r.StartSync(10*time.Minute, hookChan)
	hookChan <- struct{}{}
	var a = auth.New(conf, http.DefaultServeMux)

	handle := auth.HandleConf{"/", "/profile"}
	apiHandle := auth.HandleConf{"/api/profile", ""}
	repoView := auth.HandleConf{"/api/repo/view/", ""}
	repoEdit := auth.HandleConf{"/api/repo/edit/", ""}
	webHook := auth.HandleConf{"/api/webhook", ""}

	sampleFile := "contents.md"

	a.AsAny(webHook, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		select {
		case hookChan <- struct{}{}:
			log.Println("Webook triggered: sending update")
		default:
			log.Println("Webook discarded: update pending")
		}
	}))
	a.AsNoone(handle, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(rw, tplLogin, apiHandle.Endpoint, conf.Login.Endpoint)
	}))
	a.AsSomeone(handle.Reverse(), http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		u := a.GetUser(req)
		fmt.Fprintf(rw, tplProfile,
			u.Login, u.Email, repoEdit.Endpoint+sampleFile, repoView.Endpoint+sampleFile, conf.Logout.Endpoint)
	}))
	a.AsSomeone(apiHandle, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		u := a.GetUser(req)
		rw.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rw).Encode(u); err != nil {
			log.Printf("API profile: %s", err)
		}
	}))
	a.AsSomeone(repoView, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		path := extractPath(req.URL.Path, repoView.Endpoint)
		content, err := r.Get(path)
		if err != nil {
			handleError(rw, err)
			return
		}
		fmt.Fprint(rw, content)
	}))
	a.AsSomeone(repoEdit, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		file := extractPath(req.URL.Path, repoEdit.Endpoint)
		contents, err := r.Get(file)
		if err != nil {
			handleError(rw, err)
			return
		}
		u := a.GetUser(req)
		if req.Method == "GET" {
			hash, err := r.Hash(file)
			if err != nil {
				handleError(rw, err)
				return
			}
			fmt.Fprintf(rw, tplEdit, u.Login, u.Email, file, contents, req.URL.Path, contents, hash)
			return
		}
		if err = r.UpdateFile(file, req.FormValue("hash"), []byte(req.FormValue("contents")), u); err != nil {
			handleError(rw, err)
			return
		}
		hookChan <- struct{}{}

		http.Redirect(rw, req, repoView.Endpoint+file, http.StatusTemporaryRedirect)
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
