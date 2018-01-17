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
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/securityfirst/tent.v2/auth"
	"gopkg.in/securityfirst/tent.v2/repo"
	"gopkg.in/securityfirst/tent.v2/transifex"
)

var cfgFile string

var config struct {
	Server struct {
		Port   int
		Prefix string
	}
	Github struct {
		Handler, Project, Branch string
	}
	Transifex struct {
		Project        transifex.Project
		Language       string
		Username       string
		Password       string
		RequestPerHour int
		Filter         string
	}
	PauseOnError bool
	Root         string
	auth.Config
}

func newRepo() (*repo.Repo, error) {
	return repo.New(config.Github.Handler, config.Github.Project, config.Github.Branch)
}

var RootCmd = &cobra.Command{
	Use:   "tent",
	Short: "Repo based CMS",
	Long:  `Tent is a Content Managment System that store data in a Github Repository.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tent.yaml)")
	log.SetFlags(log.Ltime | log.Lshortfile)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".tent") // name of config file (without extension)
	viper.AddConfigPath(".")     // adding current directory as first search path
	viper.AddConfigPath("$HOME") // adding home directory as second
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return
	}
	log.Println("Using config file:", viper.ConfigFileUsed())
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("Error:", err)
	}
	if config.Transifex.Language == "" {
		config.Transifex.Language = "en"
	}
	if config.Transifex.RequestPerHour == 0 {
		config.Transifex.RequestPerHour = 6000
	}
}
