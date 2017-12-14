package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gopkg.in/securityfirst/tent.v1/models"

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

type encrypter string

func (e encrypter) encrypt(u *models.User) (string, error) {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(&u); err != nil {
		return "", err
	}
	return jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{githubUser: b.String()},
	).SignedString([]byte(e))
}

func (e encrypter) decrypt(auth string) (*models.User, error) {
	token, err := jwt.Parse(auth, func(_ *jwt.Token) (interface{}, error) {
		return []byte(e), nil
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
