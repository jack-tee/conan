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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load connector config files into Kafka Connect",
	Long:  `Load connector config files into Kafka Connect`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No paths found.\n")
			return
		}

		files := make([]string, 0)

		for _, path := range args {
			matches, err := filepath.Glob(path)
			if err != nil || matches == nil {
				log.Warn("no files found for arg ", path)
			}
			log.Debug("for arg ", path, " found files ", matches)
			files = append(files, matches...)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "loading files %s\n", files)

		for _, file := range files {
			configFile, _ := os.Open(file)
			byteValue, _ := ioutil.ReadAll(configFile)

			var configObj map[string]interface{}
			json.Unmarshal([]byte(byteValue), &configObj)

			connectorName := strings.Split(filepath.Base(file), ".")[0]

			var conf = make(map[string]string)

			if _, ok := configObj["config"]; ok {
				configConnectorName := configObj["name"].(string)
				if configConnectorName != connectorName {
					log.Warn("connector name [", configConnectorName, "] in file [", file, "] does not match filename")
				}
				connectorName = configConnectorName

				configObj = configObj["config"].(map[string]interface{})
			}
			for k, v := range configObj {
				conf[k] = v.(string)
			}

			log.Debug("in ", file, " found ", conf)
			fmt.Fprintf(cmd.OutOrStdout(), "loading %s as %s\n", file, connectorName)

			//fmt.Println(conf)

			connectorClass := conf["connector.class"]
			classParts := strings.Split(connectorClass, ".")
			pluginClass := classParts[len(classParts)-1]

			fmt.Fprintf(cmd.OutOrStdout(), "class %s, plugin %s\n", connectorClass, pluginClass)

			// validate
			confJson, _ := json.Marshal(conf)

			validateUrl := fmt.Sprintf("http://%s:%s/connector-plugins/%s/config/validate", host, port, pluginClass)

			req, err := http.NewRequest(http.MethodPut, validateUrl, bytes.NewBuffer(confJson))
			cobra.CheckErr(err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			cobra.CheckErr(err)
			respBodyBytes, _ := ioutil.ReadAll(resp.Body)

			log.Debug("Got response status: ", resp.StatusCode)
			//fmt.Fprintf(cmd.OutOrStdout(), "response %s, StatusCode %d\n", string(respBodyBytes), resp.StatusCode)
			var validationResponse ValidationResponse

			json.Unmarshal(respBodyBytes, &validationResponse)

			fmt.Println(validationResponse)
			// return ValidationResponse on a channel?

			if validationResponse.ErrorCount == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "config for %s is valid\n", validationResponse.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "config for %s is invalid, skipping loading\n", validationResponse.Name)
				continue
			}
			fmt.Println("loading connector")
			// check response error_count = 0 means it is valid
			// log any errors

			// put to the config endpoint
			loadUrl := fmt.Sprintf("http://%s:%s/connectors/%s/config", host, port, connectorName)

			req, err = http.NewRequest(http.MethodPut, loadUrl, bytes.NewBuffer(confJson))
			cobra.CheckErr(err)

			req.Header.Set("Content-Type", "application/json")

			resp, err = http.DefaultClient.Do(req)
			cobra.CheckErr(err)
			//respBodyBytes, _ = ioutil.ReadAll(resp.Body)

			log.Debug("Got response status: ", resp.StatusCode)
			fmt.Println("loaded connector ", resp.StatusCode)

		}

	},
}

type ValidationResponse struct {
	Name       string
	ErrorCount int `json:"error_count"`
	Configs    []ValidationResponseField
}

type ValidationResponseField struct {
	Value ValidationResponseFieldValue
}

type ValidationResponseFieldValue struct {
	Name   string
	Errors []string
}

func init() {
	rootCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ValidateConfig(host string, port string, connectorName string) {

}
