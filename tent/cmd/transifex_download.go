package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/securityfirst/tent/utils"

	"github.com/securityfirst/tent/component"

	"github.com/securityfirst/tent/transifex"
	"github.com/spf13/cobra"
)

var mute = true

// downloadCmd respresents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Uploads latest contents to Transifex",
	Long:  `Downloads the lastest version of Transifex contents and uploads them to Tent.`,
	Run:   downloadRun,
}

func init() {
	transifexCmd.AddCommand(downloadCmd)
}

var (
	client transifex.Client
	parser = component.NewResourceParser()
)

func downloadRun(cmd *cobra.Command, args []string) {
	client = transifex.NewClient(config.Transifex.Project, config.Transifex.Username, config.Transifex.Password)
	client.RateLimit(time.Hour, config.Transifex.RequestPerHour)
	difficulties, err := downloadDifficulties(args)
	if err != nil {
		log.Fatalf("difficulties: %s", err)
	}

	r, err := newRepo()
	r.Pull()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go func() {
		defer func() { quit <- nil }()
		for _, cmp := range r.All(config.Transifex.Language) {
			if err := parseComponent(cmp, args); err != nil {
				log.Println(cmp.Path(), err)
			}
		}
	}()
	<-quit
	saveResults(parser, difficulties)

}

func parseComponent(cmp component.Component, langs []string) error {
	defer func() {
		if recover() != nil {
			log.Println(cmp)
		}
	}()
	resource := cmp.Resource()
	translations, err := downloadTranslation(filepath.Join(config.Root, resource.Slug+".json"), resource)
	if err != nil {
		return err
	}
	for _, t := range langs {
		target, ok := translations[t]
		if !ok || target == translations["en"] {
			continue
		}
		if err := json.NewDecoder(strings.NewReader(target)).Decode(&resource.Content); err != nil {
			return fmt.Errorf("%s: %s", t, err)
		}
		if err := parser.Parse(cmp, &resource, t[:2]); err != nil {
			return fmt.Errorf("%s: %s", t, err)
		}
	}
	return nil
}

func downloadTranslation(cachePath string, resource component.Resource) (map[string]string, error) {
	var translations = (map[string]string)(nil)
	if _, err := os.Stat(cachePath); err != nil {
		translations, err = client.DownloadTranslations(resource.Slug)
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(cachePath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err := json.NewEncoder(f).Encode(translations); err != nil {
			return nil, err
		}
	}
	f, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&translations); err != nil {
		return nil, err
	}
	return translations, nil
}

func downloadDifficulties(langs []string) (map[string]map[string]string, error) {
	resource := component.Resource{Slug: "difficultiesjson"}
	translations, err := downloadTranslation(filepath.Join(config.Root, "difficulties.json"), resource)
	if err != nil {
		return nil, err
	}
	difficulties := make(map[string]map[string]string)
	for _, t := range langs {
		var m map[string]string
		if err := json.NewDecoder(bytes.NewBufferString(translations[t])).Decode(&m); err != nil {
			log.Fatalf("Difficulty error: %s", err)
		}
		difficulties[t[:2]] = m
	}
	return difficulties, nil
}

func saveResults(parser *component.ResourceParser, difficulties map[string]map[string]string) {
	for lang, cats := range parser.Categories() {
		for _, cat := range cats {
			utils.WriteCmp(config.Root, cat)
			for _, s := range cat.Subcategories() {
				sub := cat.Sub(s)
				utils.WriteCmp(config.Root, sub)
				for _, d := range sub.DifficultyNames() {
					diff := sub.Difficulty(d)
					if l, ok := difficulties[lang]; ok {
						slug := strings.Join([]string{cat.ID, sub.ID, diff.ID}, "___")
						if desc := l[slug]; desc != "" {
							diff.Descr = desc
						}
					}
					if check := diff.Checks(); check.HasChildren() {
						utils.WriteCmp(config.Root, check)
					}
					for _, i := range diff.ItemNames() {
						utils.WriteCmp(config.Root, diff.Item(i))
					}
				}
			}
		}
	}
}
