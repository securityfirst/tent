package auth

import (
	"encoding/gob"
	"fmt"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/securityfirst/tent/models"

	"golang.org/x/oauth2"
	lib "golang.org/x/oauth2/github"
)

// register in gob models saved with cookies
func init() {
	gob.Register(&models.User{})
}

// HandleConf contains info about and Handler
type HandleConf struct {
	Endpoint string
	Redirect string
}

// Basic configuration for an Auth
type Config struct {
	Id           string
	Secret       string
	OAuthHost    string
	Host         string
	RandomString string
	Login        HandleConf
	Logout       HandleConf
	Callback     HandleConf
}

// OAuth return the oauth2 configuration struct
func (c *Config) OAuth(root *gin.RouterGroup) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.Id,
		ClientSecret: c.Secret,
		RedirectURL:  fmt.Sprint(c.OAuthHost, path.Clean(root.BasePath()+c.Callback.Endpoint)),
		Endpoint:     lib.Endpoint,
		Scopes:       []string{"user:email", "repo"},
	}
}
