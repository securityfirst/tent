package transifex

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const ApiUrl = "https://www.transifex.com/api/2"

type BaseResource struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	I18nType string `json:"i18n_type"`
	Priority string `json:"priority"`
	Category string `json:"category"`
}

type Resource struct {
	BaseResource
	SourceLanguage string `json:"source_language_code"`
}

type UploadResourceRequest struct {
	BaseResource
	Content             string `json:"content"`
	Accept_translations string `json:"accept_translations"`
}

type Language struct {
	Coordinators []string `json:"coordinators"`
	LanguageCode string   `json:"language_code"`
	Translators  []string `json:"translators"`
	Reviewers    []string `json:"reviewers"`
}

type Response struct {
	Added   int `json:"strings_added"`
	Updated int `json:"strings_updated"`
	Deleted int `json:"strings_delete"`
}

func (r *Response) UnmarshalJSON(raw []byte) error {
	var dst interface{}
	if err := json.Unmarshal(raw, &dst); err != nil {
		return err
	}
	switch v := dst.(type) {
	case []interface{}:
		r.Added = int(v[0].(float64))
		r.Updated = int(v[1].(float64))
		r.Deleted = int(v[2].(float64))
	case map[string]interface{}:
		r.Added = int(v["strings_added"].(float64))
		r.Updated = int(v["strings_updated"].(float64))
		r.Deleted = int(v["strings_delete"].(float64))
	default:
		return fmt.Errorf("Unkwown type %T", v)
	}
	return nil
}

type Project string

func (p Project) root() string              { return fmt.Sprintf("%s/project/%s", ApiUrl, p) }
func (p Project) list() string              { return fmt.Sprintf("%s/resources", p.root()) }
func (p Project) resource(id string) string { return fmt.Sprintf("%s/resource/%s", p.root(), id) }

func unmarshal(resp *http.Response, dst interface{}, errPrefix string) (interface{}, error) {
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &dst); err != nil {
		return nil, fmt.Errorf(errPrefix + "\n\nError:\n" + string(raw))
	}
	return dst, nil
}

type auth struct{ Username, Password string }

func (a auth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(a.Username, a.Password)
	return http.DefaultTransport.RoundTrip(req)
}
