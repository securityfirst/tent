package auth

import (
	"log"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	"github.com/securityfirst/tent/models"
	"golang.org/x/oauth2"
)

const (
	sessionName = "github-auth"
	githubUser  = "github-email"
)

// NewEngine creates a new Engine using and adds the handle for authentication
func NewEngine(c Config, root *gin.RouterGroup) *Engine {
	var e = Engine{sessions: sessions.NewCookieStore([]byte(c.RandomString), nil)}
	config := c.OAuth(root)
	c.Login.Redirect = config.AuthCodeURL(c.RandomString, oauth2.AccessTypeOnline)
	c.Callback.Redirect = path.Clean(root.BasePath() + c.Callback.Redirect)
	c.Logout.Redirect = path.Clean(root.BasePath() + c.Logout.Redirect)
	root.GET(c.Login.Endpoint, func(g *gin.Context) {
		g.Redirect(http.StatusTemporaryRedirect, c.Login.Redirect)
	})
	root.GET(c.Logout.Endpoint, func(g *gin.Context) {
		if err := e.destoySession(g); err != nil {
			log.Println("Error while invalidating the session:", err)
		}
		g.Redirect(http.StatusTemporaryRedirect, c.Logout.Redirect)
	})
	root.GET(c.Callback.Endpoint, func(g *gin.Context) {
		if errString := g.Query("error"); errString != "" {
			log.Printf("Service responded with error: %s", errString)
			return
		}
		state := g.Query("state")
		if state != c.RandomString {
			log.Printf("Invalid oauth state, expected %q, got %q", c.RandomString, state)
			g.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		code := g.Query("code")
		token, err := config.Exchange(oauth2.NoContext, code)
		if err != nil {
			log.Printf("Cannot get Token: %s", err)
			g.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		user := models.User{Token: *token}
		u, _, err := github.NewClient(e.config.Client(oauth2.NoContext, &user.Token)).Users.Get("")
		if err != nil {
			log.Printf("Cannot get User: %s", err)
			g.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		user.Name, user.Email, user.Login = *u.Name, *u.Email, *u.Login
		if err := e.createSession(g, &user); err != nil {
			log.Printf("Cannot save session: %s", err)
			g.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		g.Redirect(http.StatusTemporaryRedirect, c.Callback.Redirect)
	})
	return &e
}

// Engine is e struct that eases Github OAuth and resource handling
type Engine struct {
	root     *gin.RouterGroup
	config   *oauth2.Config
	port     int
	sessions *sessions.CookieStore
}

func (e *Engine) createSession(c *gin.Context, u *models.User) error {
	s, err := e.sessions.New(c.Request, sessionName)
	if err != nil {
		return err
	}
	s.Values[githubUser] = u
	return s.Save(c.Request, c.Writer)
}

func (e *Engine) destoySession(c *gin.Context) error {
	s, err := e.sessions.Get(c.Request, sessionName)
	if err != nil {
		return err
	}
	s.Options.MaxAge = -1
	return s.Save(c.Request, c.Writer)
}

// GetUser returns the user connected
func (e *Engine) GetUser(c *gin.Context) *models.User {
	s, err := e.sessions.Get(c.Request, sessionName)
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
