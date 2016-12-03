package auth

import (
	"encoding/gob"
	"fmt"

	"github.com/securityfirst/octo/models"

	"golang.org/x/oauth2"
	lib "golang.org/x/oauth2/github"
)

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
	Id        string
	Secret    string
	OAuthHost string
	Host      string
	Login     HandleConf
	Logout    HandleConf
	Callback  HandleConf
}

// OAuth return the oauth2 configuration struct
func (c *Config) OAuth() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.Id,
		ClientSecret: c.Secret,
		RedirectURL:  fmt.Sprint(c.OAuthHost, c.Callback.Endpoint),
		Endpoint:     lib.Endpoint,
		Scopes:       []string{"user:email", "repo"},
	}
}
