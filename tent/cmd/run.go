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
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gopkg.in/securityfirst/tent.v2"
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
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%v", config.Server.Port),
			Handler: e,
		}
		r, err := newRepo()
		if err != nil {
			log.Fatalf("Repo error: %s", err)
		}

		o := tent.New(r)
		o.Register(e.Group(config.Server.Prefix), config.Config)

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		go srv.ListenAndServe()

		<-stop
		log.Println("Shutting down the server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
