package cmd

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/securityfirst/tent/utils"

	"github.com/securityfirst/tent/component"

	"github.com/securityfirst/tent/transifex"
	"github.com/spf13/cobra"
)

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

func downloadRun(cmd *cobra.Command, args []string) {
	r, err := newRepo()
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}
	r.Pull()

	client := transifex.NewClient(config.Transifex.Project, config.Transifex.Username, config.Transifex.Password)
	client.RateLimit(time.Hour, config.Transifex.RequestPerHour)

	parser := component.NewResourceParser()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	var count int

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	go func() {
		defer func() { quit <- nil }()
		for _, cmp := range r.All(config.Transifex.Language) {
			count++
			var (
				resource     = cmp.Resource()
				translations = (map[string]string)(nil)
				cachePath    = filepath.Join(config.Root, resource.Slug)
			)
			if _, err := os.Stat(cachePath); os.IsNotExist(err) {
				translations, err = client.DownloadTranslations(resource.Slug)
				if err != nil {
					log.Printf("%s: %s", resource.Slug, red(err))
					continue
				}
				f, err := os.OpenFile(cachePath, os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					log.Printf("%s: %s", resource.Slug, red(err))
					continue
				}
				defer f.Close()
				if err := json.NewEncoder(f).Encode(translations); err != nil {
					log.Printf("%s: %s", resource.Slug, red(err))
					continue
				}
			} else if err == nil {
				f, err := os.Open(cachePath)
				if err != nil {
					log.Printf("%s: %s", resource.Slug, red(err))
					continue
				}
				defer f.Close()
				if err = json.NewDecoder(f).Decode(&translations); err != nil {
					log.Printf("%s: %s", resource.Slug, red(err))
					continue
				}
			} else {
				log.Printf("%s: %s", resource.Slug, red(err))
				continue
			}
			for _, t := range args {
				target, ok := translations[t]
				if !ok {
					log.Printf("%s: %s not found", resource.Slug, t)
					continue
				}
				var m []map[string]string
				if err := json.NewDecoder(strings.NewReader(target)).Decode(&m); err != nil {
					log.Printf("%s (%s) %s\n%s", resource.Slug, t, err, target)
					continue
				}
				resource.Content = m
				if err := parser.Parse(cmp, &resource, t[:2]); err != nil {
					log.Printf("%s (%s) %s", resource.Slug, t, err)
					continue
				}
				if target != translations["en"] {
					log.Printf("%s %s - %s", green("translated"), t, resource.Slug)
				} else {
					log.Printf("%s %s - %s", red("not translated"), t, resource.Slug)
				}
			}
		}
	}()

	<-quit

	var printCmp = func(cmp component.Component) {
		if err := utils.WriteCmp(config.Root, cmp); err != nil {
			log.Println(cmp.Path(), red(err))
			return
		}
		log.Println(cmp.Path(), green("ok"))
	}

	log.Printf("\n\n***** Saving %d files *****\n\n", count)
	for lang, cats := range parser.Categories() {
		log.Printf("Language [%s]", lang)
		for _, cat := range cats {
			printCmp(cat)
			for _, s := range cat.Subcategories() {
				sub := cat.Sub(s)
				printCmp(sub)
				if check := sub.Checks(); check.HasChildren() {
					printCmp(check)
				}
				for _, i := range sub.ItemNames() {
					printCmp(sub.Item(i))
				}
			}
		}
	}

}
