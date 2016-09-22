package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	"github.com/klaidliadon/octo/models"
	"golang.org/x/oauth2"
)

const (
	sessionName = "github-auth"
	githubUser  = "github-email"
)

// random string for oauth2 API calls to protect against CSRF
const randomString = "horse battery staple"

// New creates a new Auth using the given configuration
func New(c Config, mux *http.ServeMux) *Auth {
	var a = Auth{
		config:   c.OAuth(),
		mux:      mux,
		port:     c.Port,
		sessions: sessions.NewCookieStore([]byte(randomString), nil),
	}
	c.Login.Redirect = a.config.AuthCodeURL(randomString, oauth2.AccessTypeOnline)
	a.AsNoone(c.Login, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, c.Login.Redirect, http.StatusTemporaryRedirect)
	}))
	a.AsSomeone(c.Logout, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := a.destoySession(w, r); err != nil {
			log.Println("Error while invalidating the session:", err)
		}
		http.Redirect(w, r, c.Logout.Redirect, http.StatusFound)
	}))
	a.AsNoone(c.Callback, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errString := r.FormValue("error"); errString != "" {
			log.Printf("Service responded with error: %s", errString)
			return
		}
		state := r.FormValue("state")
		if state != randomString {
			log.Printf("Invalid oauth state, expected %q, got %q", randomString, state)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		code := r.FormValue("code")
		token, err := a.config.Exchange(oauth2.NoContext, code)
		if err != nil {
			log.Printf("Cannot get Token: %s", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		user := models.User{Token: *token}
		defer http.Redirect(w, r, c.Callback.Redirect, http.StatusTemporaryRedirect)

		var client = github.NewClient(a.config.Client(oauth2.NoContext, &user.Token))
		u, _, err := client.Users.Get("")
		if err != nil {
			log.Printf("Cannot get User: %s", err)
			return
		}
		user.Name, user.Email, user.Login = *u.Name, *u.Email, *u.Login
		if err := a.createSession(w, r, &user); err != nil {
			log.Printf("Cannot save session: %s", err)
			return
		}
	}))
	return &a
}

// Auth is a struct that eases Github OAuth and resource handling
type Auth struct {
	config   *oauth2.Config
	mux      *http.ServeMux
	port     int
	sessions *sessions.CookieStore
}

// Run starts the configuration address
func (a *Auth) Run() {
	http.ListenAndServe(fmt.Sprint(":", a.port), a.mux)
}

func (a *Auth) createSession(w http.ResponseWriter, r *http.Request, u *models.User) error {
	s, err := a.sessions.New(r, sessionName)
	if err != nil {
		return err
	}
	s.Values[githubUser] = u
	return s.Save(r, w)
}

func (a *Auth) destoySession(w http.ResponseWriter, r *http.Request) error {
	s, err := a.sessions.Get(r, sessionName)
	if err != nil {
		return err
	}
	s.Options.MaxAge = -1
	return s.Save(r, w)
}

func (a *Auth) WithCondition(c CondHandleConf, h http.Handler) {
	a.mux.Handle(c.Endpoint, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case c.Access(a, r):
			h.ServeHTTP(w, r)
		case c.Redirect != "":
			http.Redirect(w, r, c.Redirect, http.StatusFound)
		default:
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "Access denied")
		}
	}))
}

func (a *Auth) AsAny(c HandleConf, h http.Handler) {
	a.WithCondition(CondHandleConf{c, accessAlways}, h)
}

func (a *Auth) AsSomeone(c HandleConf, h http.Handler) {
	a.WithCondition(CondHandleConf{c, (*Auth).connected}, h)
}

func (a *Auth) AsNoone(c HandleConf, h http.Handler) {
	a.WithCondition(CondHandleConf{c, (*Auth).disconnected}, h)
}

func (a *Auth) GetUser(r *http.Request) *models.User {
	s, err := a.sessions.Get(r, sessionName)
	if err != nil {
		return nil
	}
	v := s.Values[githubUser]
	if v == nil {
		return nil
	}
	u, ok := v.(*models.User)
	if !ok {
		return nil
	}
	return u
}

func (a *Auth) connected(r *http.Request) bool    { return a.GetUser(r) != nil }
func (a *Auth) disconnected(r *http.Request) bool { return a.GetUser(r) == nil }
