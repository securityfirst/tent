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
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/securityfirst/tent"
	"github.com/spf13/cobra"
)

// runCmd respresents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts CMS",
	Long:  `Starts the Tent CMS`,
	Run: func(cmd *cobra.Command, args []string) {
		if config.Id == "" || config.Secret == "" {
			flag.Usage()
			os.Exit(1)
		}
		e := gin.Default()
		r, err := newRepo()
		if err != nil {
			log.Fatalf("Repo error: %s", err)
		}

		o := tent.New(r)
		o.Register(e.Group("/v2"), config.Config)
		e.Run(fmt.Sprintf(":%v", config.Port))
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
