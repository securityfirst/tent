package transifex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	KeyValueJson string = "KEYVALUEJSON"
)

type TransifexAPI struct {
	client   *http.Client
	ApiUrl   string
	Project  string
	username string
	password string
}

func NewTransifexAPI(project, username, password string) TransifexAPI {
	return TransifexAPI{
		client:   http.DefaultClient,
		ApiUrl:   "https://www.transifex.com/api/2",
		Project:  project,
		username: username,
		password: password,
	}
}

func (t TransifexAPI) ListResources() ([]Resource, error) {
	resp, err := t.execRequest("GET", t.resourcesUrl(true), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var resources []Resource

	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, err
	}

	return resources, nil
}

func (t TransifexAPI) CreateResource(newResource UploadResourceRequest) (*ResourceResp, error) {
	data, err := json.Marshal(newResource)
	if err != nil {
		return nil, err
	}
	resp, err := t.execRequest("POST", t.resourcesUrl(false), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	checkData, err := t.checkValidJsonResponse(resp, fmt.Sprintf("Failed to create resource: %s\n", newResource.Slug))
	if err != nil {
		return nil, err
	}

	d := checkData.([]interface{})
	return &ResourceResp{
		int(d[0].(float64)),
		int(d[1].(float64)),
		int(d[2].(float64)),
	}, nil
}

func (t TransifexAPI) UpdateResourceContent(slug, content string) (*ResourceResp, error) {
	data, err := json.Marshal(map[string]string{"slug": slug, "content": content})
	if err != nil {
		return nil, err
	}

	resp, err := t.execRequest("PUT", t.resourceUrl(slug, true)+"content/", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	checkData, err := t.checkValidJsonResponse(resp, fmt.Sprintf("Error updating content of %s", slug))
	if err != nil {
		return nil, err
	}
	d := checkData.(map[string]interface{})
	return &ResourceResp{
		int(d["strings_added"].(float64)),
		int(d["strings_updated"].(float64)),
		int(d["strings_delete"].(float64)),
	}, nil
}

func (t TransifexAPI) ValidateConfiguration() error {
	msg := "Error occurred when checking credentials. Please check credentials and network connection"
	if _, err := t.SourceLanguage(); err != nil {
		return fmt.Errorf(msg)
	}
	return nil
}

func (t TransifexAPI) UploadTranslationFile(slug, langCode, content string) error {
	data, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}

	resp, err := t.execRequest("PUT", t.resourceUrl(slug, true)+"translation/"+langCode+"/", bytes.NewReader(data))
	if err != nil {
		return err
	}

	checkData, err := t.checkValidJsonResponse(resp, fmt.Sprintf("Error adding %s translations for %s", langCode, slug))
	dataMap := checkData.(map[string]interface{})
	log.Printf(`Update %s %s Translation summary:

Strings Added: %v
Strings updated: %v
Strings deleted: %v

`, slug, langCode, dataMap["strings_added"], dataMap["strings_updated"], dataMap["strings_delete"])
	return err
}

func (t TransifexAPI) SourceLanguage() (string, error) {
	jsonData, err := t.getJson(t.ApiUrl+"/project/"+t.Project, "Error loading SourceLanguage")
	if err != nil {
		return "", err
	}
	lang, has := jsonData.(map[string]interface{})["source_language_code"]
	if !has {
		return "", fmt.Errorf("An error occurred while reading response. Expected a 'source_language_code' json field:\n%s", jsonData)
	}
	sourceLang := lang.(string)
	if strings.TrimSpace(sourceLang) == "" {
		return "", fmt.Errorf("No source language found.  This is probably a bug")
	}
	return sourceLang, nil
}

func (t TransifexAPI) Languages() ([]Language, error) {
	resp, err := t.execRequest("GET", fmt.Sprintf("%s/project/%s/languages", t.ApiUrl, t.Project), nil)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var jsonData []Language
	if err = json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	return jsonData, nil
}
func (t TransifexAPI) DownloadTranslations(slug string) (map[string]string, error) {
	sourceLang, err := t.SourceLanguage()
	if err != nil {
		return nil, err
	}
	fullLangs, langErr := t.Languages()
	if langErr != nil {
		return nil, langErr
	}
	langs := make([]string, len(fullLangs)+1)
	langs[0] = sourceLang
	for i, l := range fullLangs {
		langs[i+1] = l.LanguageCode
	}

	translations := make(map[string]string, len(langs))
	for _, lang := range langs {
		url := fmt.Sprintf("%s/project/%s/resource/%s/translation/%s", t.ApiUrl, t.Project, slug, lang)
		data, err2 := t.getJson(url, "Error downloing translations file")
		if err2 != nil {
			return nil, err2
		}

		translations[lang] = data.(map[string]interface{})["content"].(string)
	}
	return translations, nil
}

func (t TransifexAPI) getJson(url string, errMsg string) (interface{}, error) {
	resp, err := t.execRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	jsonData, err := t.checkValidJsonResponse(resp, errMsg)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (t TransifexAPI) execRequest(method string, url string, requestData io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, url, requestData)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(t.username, t.password)
	if requestData != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	resp, finalErr := t.client.Do(request)
	if resp.StatusCode > 400 {
		return nil, fmt.Errorf("Response Code: %v\nResponse Status: %s", resp.StatusCode, resp.Status)
	}

	return resp, finalErr
}

func (t TransifexAPI) resourcesUrl(endSlash bool) string {
	url := fmt.Sprintf("%s/project/%s/resources", t.ApiUrl, t.Project)
	if endSlash {
		return url + "/"
	}
	return url
}

func (t TransifexAPI) resourceUrl(slug string, endSlash bool) string {
	url := fmt.Sprintf("%s/project/%s/resource/%s", t.ApiUrl, t.Project, slug)
	if endSlash {
		return url + "/"
	}
	return url
}

func (t TransifexAPI) checkValidJsonResponse(resp *http.Response, errorMsg string) (interface{}, error) {
	responseData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var jsonData interface{}
	if err := json.Unmarshal(responseData, &jsonData); err != nil {
		return nil, fmt.Errorf(errorMsg + "\n\nError:\n" + string(responseData))
	}

	return jsonData, nil
}
