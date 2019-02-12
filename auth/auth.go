package auth

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"github.com/securityfirst/tent/models"
)

const (
	sessionName = "github-auth"
	githubUser  = "github-email"
)

// NewEngine creates a new Engine using and adds the handle for authentication
func NewEngine(conf Config, root *gin.RouterGroup) *Engine {
	var e = Engine{
		config: conf.OAuth(root),
		cache:  make(map[string]models.User),
		state:  conf.State,
	}
	conf.Login.Redirect = e.config.AuthCodeURL(e.state, oauth2.AccessTypeOnline)
	conf.Callback.Redirect = path.Clean(root.BasePath() + conf.Callback.Redirect)
	conf.Logout.Redirect = path.Clean(root.BasePath() + conf.Logout.Redirect)
	root.GET(conf.Login.Endpoint, func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, conf.Login.Redirect)
	})
	root.GET(conf.Logout.Endpoint, func(c *gin.Context) {
		c.SetCookie(githubUser, "", -1, "/", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, conf.Logout.Redirect)
	})
	root.GET(conf.Callback.Endpoint, func(c *gin.Context) {
		if errString := c.Query("error"); errString != "" {
			log.Printf("Service responded with error: %s", errString)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errString})
			return
		}
		state := c.Query("state")
		if state != conf.State {
			log.Printf("Invalid oauth state, expected %q, got %q", conf.State, state)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid oauth state"})
			return
		}
		code := c.Query("code")
		token, err := e.config.Exchange(oauth2.NoContext, code)
		if err != nil || !token.Valid() {
			log.Printf("Cannot get Token: %s", err)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		c.SetCookie(githubUser, token.AccessToken, 614880, "/", "", false, false)
		c.JSON(200, token)
	})
	return &e
}

// Engine is e struct that eases Github OAuth and resource handling
type Engine struct {
	config *oauth2.Config
	state  string
	cache  map[string]models.User
}

func (e *Engine) EnsureUser(c *gin.Context) {
	var token string
	if auth := c.Request.Header.Get("Authorization"); auth != "" {
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization"})
			c.Abort()
			return
		}
		token = parts[1]
	} else {
		cookie, err := c.Cookie(githubUser)
		if err != nil && err != http.ErrNoCookie {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cookie"})
			c.Abort()
			return
		}
		token = cookie
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access denied"})
		c.Abort()
		return
	}
	if err := e.fetchUser(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		c.Abort()
		return
	}
	c.Set("token", token)
	c.Set("user", e.cache[token])
}

func (e *Engine) fetchUser(token string) error {
	if _, ok := e.cache[token]; ok {
		return nil
	}
	u, _, err := github.NewClient(e.config.Client(oauth2.NoContext, &oauth2.Token{AccessToken: token})).Users.Get(oauth2.NoContext, "")
	if err != nil {
		return fmt.Errorf("Cannot get User: %s", err)
	}
	var email string
	if u.Email != nil {
		email = *u.Email
	} else {
		email = *u.Login + "@tent.org"
	}
	e.cache[token] = models.User{Name: *u.Name, Login: *u.Login, Email: email}
	return nil
}
