package transifex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const KeyValueJson = "KEYVALUEJSON"

type Client struct {
	client  *http.Client
	Project Project
}

func NewClient(project Project, username, password string) Client {
	return Client{
		client:  &http.Client{Transport: auth{username, password}},
		Project: project,
	}
}

func (c Client) ListResources() ([]Resource, error) {
	resp, err := c.execute("GET", c.Project.list(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var resources []Resource
	_, err = unmarshal(resp, &resources, "Failed to list resources")
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (c Client) CreateResource(newResource UploadResourceRequest) (*Response, error) {
	data, err := json.Marshal(newResource)
	if err != nil {
		return nil, err
	}
	resp, err := c.execute("POST", c.Project.list(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var r Response
	_, err = unmarshal(resp, &r, fmt.Sprintf("Failed to create resource: %s\n", newResource.Slug))
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (c Client) UpdateResourceContent(slug, content string) (*Response, error) {
	data, err := json.Marshal(map[string]string{"slug": slug, "content": content})
	if err != nil {
		return nil, err
	}
	resp, err := c.execute("PUT", c.Project.resource(slug)+"/content/", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var r Response
	_, err = unmarshal(resp, &r, fmt.Sprintf("Error updating content of %s", slug))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c Client) ValidateConfiguration() error {
	msg := "Error occurred when checking credentials. Please check credentials and network connection"
	if _, err := c.SourceLanguage(); err != nil {
		return fmt.Errorf(msg)
	}
	return nil
}

func (c Client) UploadTranslationFile(slug, langCode, content string) (*Response, error) {
	data, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return nil, err
	}
	resp, err := c.execute("PUT", c.Project.resource(slug)+"/translation/"+langCode+"/", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var r Response
	_, err = unmarshal(resp, &r, fmt.Sprintf("Error adding %s translations for %s", langCode, slug))
	if err != nil {
		return nil, err
	}
	return nil, err
}

func (c Client) SourceLanguage() (string, error) {
	jsonData, err := c.getJson(c.Project.root(), "Error loading SourceLanguage")
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

func (c Client) Languages() ([]Language, error) {
	resp, err := c.execute("GET", fmt.Sprintf("%s/project/%s/languages", ApiUrl, c.Project), nil)

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
func (c Client) DownloadTranslations(slug string) (map[string]string, error) {
	sourceLang, err := c.SourceLanguage()
	if err != nil {
		return nil, err
	}
	fullLangs, err := c.Languages()
	if err != nil {
		return nil, err
	}
	langs := make([]string, len(fullLangs)+1)
	langs[0] = sourceLang
	for i, l := range fullLangs {
		langs[i+1] = l.LanguageCode
	}

	translations := make(map[string]string, len(langs))
	for _, lang := range langs {
		url := fmt.Sprintf("%s/project/%s/resource/%s/translation/%s", ApiUrl, c.Project, slug, lang)
		data, err := c.getJson(url, "Error downloing translations file")
		if err != nil {
			return nil, err
		}

		translations[lang] = data.(map[string]interface{})["content"].(string)
	}
	return translations, nil
}

func (c Client) getJson(url string, errMsg string) (interface{}, error) {
	resp, err := c.execute("GET", url, nil)
	if err != nil {
		return nil, err
	}

	jsonData, err := unmarshal(resp, nil, errMsg)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (c Client) execute(method string, url string, requestData io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, url, requestData)
	if err != nil {
		return nil, err
	}
	if requestData != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 400 {
		return nil, fmt.Errorf("Response Code: %v\nResponse Status: %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}
