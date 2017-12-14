// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"gopkg.in/securityfirst/tent.v2/transifex"
	"github.com/spf13/cobra"
)

// uploadCmd respresents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads latest contents to Transifex",
	Long:  `Downloads the lastest version of Tent contents and uploads them to Transifex.`,
	Run:   uploadRun,
}

func init() {
	transifexCmd.AddCommand(uploadCmd)
}

func uploadRun(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		args = []string{config.Transifex.Language}
	}
	r, err := newRepo()
	if err != nil {
		log.Fatalf("Repo error: %s", err)
	}
	r.Pull()

	locale := r.Locale()
	for _, a := range args {
		var found bool
		for _, l := range locale {
			if l == a {
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("Locale not found: %s", a)
		}
	}

	client := transifex.NewClient(config.Transifex.Project, config.Transifex.Username, config.Transifex.Password)
	client.RateLimit(time.Hour, config.Transifex.RequestPerHour)

	resources, err := client.ListResources()
	if err != nil {
		log.Fatalf("Resource list: %s", err)
	}

	var existing = make(map[string]struct{}, len(resources))
	for _, r := range resources {
		existing[r.Slug] = struct{}{}
	}

	var buffer = bytes.NewBuffer(nil)

	for _, a := range args {
		for _, cmp := range r.All(a) {
			buffer.Reset()
			resource := cmp.Resource()
			json.NewEncoder(buffer).Encode(resource.Content)
			var (
				resp *transifex.Response
				err  error
			)
			if a == config.Transifex.Language {
				if _, ok := existing[resource.Slug]; !ok {
					resp, err = client.CreateResource(transifex.UploadResourceRequest{
						BaseResource: transifex.BaseResource{
							Slug:     resource.Slug,
							Name:     resource.Slug + ".json",
							I18nType: transifex.KeyValueJson,
						},
						Content: buffer.String(),
					})
				} else {
					resp, err = client.UpdateResourceContent(resource.Slug, buffer.String())
				}
			} else {
				resp, err = client.UploadTranslationFile(resource.Slug, a, buffer.String())
			}

			if err != nil {
				log.Printf("Error: %s (%s)", err, cmp.Path())
			} else {
				log.Printf("%+v (%s)", resp, cmp.Path())
			}
		}
	}
}
