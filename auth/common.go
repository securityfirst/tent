package auth

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/klaidliadon/octo/models"

	"golang.org/x/oauth2"
	lib "golang.org/x/oauth2/github"
)

func init() {
	gob.Register(&models.User{})
}

// AccessFn is function that allows/denies access to a resource
type AccessFn func(*Auth, *http.Request) bool

var accessAlways = AccessFn(func(*Auth, *http.Request) bool { return true })

// HandleConf contains info about and Handler
type HandleConf struct {
	Endpoint string
	Redirect string
}

// Reverse returns and Handle with fields swapped
func (h HandleConf) Reverse() HandleConf {
	return HandleConf{h.Redirect, h.Endpoint}
}

// CondHandleConf is a HandleConf with a condition
type CondHandleConf struct {
	HandleConf
	Access AccessFn
}

// Reverse returns and Handle with fields swapped
func (c CondHandleConf) Reverse() CondHandleConf {
	return CondHandleConf{c.HandleConf.Reverse(), c.Access}
}

// Basic configuration for an Auth
type Config struct {
	Id        string
	Secret    string
	Port      int
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
