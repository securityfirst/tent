package transifex

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
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Deleted int `json:"deleted"`
}
