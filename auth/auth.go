package auth

import (
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"gopkg.in/securityfirst/tent.v1/models"
	"golang.org/x/oauth2"
)

const (
	sessionName = "github-auth"
	githubUser  = "github-email"
)

// NewEngine creates a new Engine using and adds the handle for authentication
func NewEngine(conf Config, root *gin.RouterGroup) *Engine {
	var e = Engine{config: conf.OAuth(root), encrypter: encrypter(conf.RandomString)}
	conf.Login.Redirect = e.config.AuthCodeURL(conf.RandomString, oauth2.AccessTypeOnline)
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
		if state != conf.RandomString {
			log.Printf("Invalid oauth state, expected %q, got %q", conf.RandomString, state)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid oauth state"})
			return
		}
		code := c.Query("code")
		token, err := e.config.Exchange(oauth2.NoContext, code)
		if err != nil {
			log.Printf("Cannot get Token: %s", err)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		user := models.User{Token: *token}
		u, _, err := github.NewClient(e.config.Client(oauth2.NoContext, &user.Token)).Users.Get("")
		if err != nil {
			log.Printf("Cannot get User: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		user.Name, user.Login = *u.Name, *u.Login
		if u.Email != nil {
			user.Email = *u.Email
		} else {
			user.Email = user.Login + "@tent.org"
		}
		tokenString, err := e.encrypt(&user)
		if err != nil {
			log.Printf("Cannot create Jwt token: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.SetCookie(githubUser, tokenString, 614880, "/", "", false, false)
		c.JSON(200, gin.H{"token": tokenString})
	})
	return &e
}

// Engine is e struct that eases Github OAuth and resource handling
type Engine struct {
	encrypter
	config *oauth2.Config
}

// GetUser returns the user connected
func (e *Engine) GetUser(c *gin.Context) *models.User {
	var data string
	if auth := c.Request.Header.Get("Authorization"); auth != "" {
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Invalid header %q", auth)
			return nil
		}
		data = parts[1]
	} else {
		cookie, err := c.Cookie(githubUser)
		if err != nil {
			return nil
		}
		data = cookie
	}
	u, err := e.decrypt(data)
	if err != nil {
		log.Printf("Decrypt %q error: %s", data, err)
		return nil
	}
	return u
}
