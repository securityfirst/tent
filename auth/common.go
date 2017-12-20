package auth

import (
	"fmt"
	"path"

	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	lib "golang.org/x/oauth2/github"
)

// HandleConf contains info about and Handler
type HandleConf struct {
	Endpoint string
	Redirect string
}

// Basic configuration for an Auth
type Config struct {
	ID        string
	Secret    string
	OAuthHost string
	Host      string
	State     string
	Login     HandleConf
	Logout    HandleConf
	Callback  HandleConf
}

// OAuth return the oauth2 configuration struct
func (c *Config) OAuth(root *gin.RouterGroup) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ID,
		ClientSecret: c.Secret,
		RedirectURL:  fmt.Sprint(c.OAuthHost, path.Clean(root.BasePath()+c.Callback.Endpoint)),
		Endpoint:     lib.Endpoint,
		Scopes:       []string{"user:email", "repo"},
	}
}
