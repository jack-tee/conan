/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:    "list",
	Short:  "List the connectors",
	Long:   `List the connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		host, port = GetPersistentFlags(cmd)
		connectors := GetConnectors(host, port)

		fmt.Println("CONNECTORS")
		for _, connector := range connectors {
			fmt.Println(connector)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func GetConnectors(host string, port string) []string {

	url := fmt.Sprintf("http://%s:%s/connectors", host, port)
	log.Debug("getting connectors using URL: ", url)

	resp, err := http.Get(url)
	cobra.CheckErr(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	cobra.CheckErr(err)

	var connectors []string
	json.Unmarshal(bodyBytes, &connectors)

	log.Debug("connectors found: ", connectors)
	return connectors
}
