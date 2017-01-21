package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"path"
	"strings"

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
func NewEngine(conf Config, root *gin.RouterGroup) *Engine {
	var e = Engine{
		sessions: sessions.NewCookieStore([]byte(conf.RandomString), nil),
		config:   conf.OAuth(root),
		secret:   conf.RandomString,
	}
	conf.Login.Redirect = e.config.AuthCodeURL(e.secret, oauth2.AccessTypeOnline)
	conf.Callback.Redirect = path.Clean(root.BasePath() + conf.Callback.Redirect)
	conf.Logout.Redirect = path.Clean(root.BasePath() + conf.Logout.Redirect)
	root.GET(conf.Login.Endpoint, func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, conf.Login.Redirect)
	})
	root.GET(conf.Logout.Endpoint, func(c *gin.Context) {
		if err := e.destoySession(c); err != nil {
			log.Println("Error while invalidating the session:", err)
		}
		c.Redirect(http.StatusTemporaryRedirect, conf.Logout.Redirect)
	})
	root.GET(conf.Callback.Endpoint, func(c *gin.Context) {
		if errString := c.Query("error"); errString != "" {
			log.Printf("Service responded with error: %s", errString)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errString})
			return
		}
		state := c.Query("state")
		if state != e.secret {
			log.Printf("Invalid oauth state, expected %q, got %q", e.secret, state)
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
		user.Name, user.Email, user.Login = *u.Name, *u.Email, *u.Login
		if err := e.createSession(c, &user); err != nil {
			log.Printf("Cannot save session: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		tokenString, err := e.createJwt(c, &user)
		if err != nil {
			log.Printf("Cannot create Jwt token: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.JSON(200, gin.H{"token": tokenString})
	})
	return &e
}

// Engine is e struct that eases Github OAuth and resource handling
type Engine struct {
	root     *gin.RouterGroup
	config   *oauth2.Config
	secret   string
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

func (e *Engine) createJwt(c *gin.Context, u *models.User) (string, error) {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(&u); err != nil {
		return "", err
	}
	return jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{githubUser: b.String()},
	).SignedString([]byte(e.secret))
}

// GetUser returns the user connected
func (e *Engine) GetUser(c *gin.Context) *models.User {
	if auth := c.Request.Header.Get("Authorization"); auth != "" {
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Invalid header %q", auth)
			return nil
		}
		user, err := e.parseJwt(parts[1])
		if err != nil {
			log.Printf("Auth %q error: %s", parts[1], err)
			return nil
		}
		return user
	}
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

func (e *Engine) parseJwt(auth string) (*models.User, error) {
	token, err := jwt.Parse(auth, func(_ *jwt.Token) (interface{}, error) {
		return []byte(e.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("Invalid Token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}
	jsonStr, ok := claims[githubUser].(string)
	if !ok {
		return nil, err
	}
	var user models.User
	if err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil

}
