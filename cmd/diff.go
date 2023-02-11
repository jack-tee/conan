/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type DiffResults struct {
	NewConnectors       []string
	ChangedConnectors   []DiffResult
	UnchangedConnectors []string
}

type DiffResult struct {
	ConnectorName string
	NewKeys       map[string]string
	MatchKeys     map[string]string
	MismatchKeys  map[string]MismatchVals
	RemovedKeys   map[string]string
}

type MismatchVals struct {
	Deployed string
	File     string
}

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:    "diff",
	Short:  "Compare connector config files with what is currently deployed",
	Long:   `Compare connector config files with what is currently deployed`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No args provided. Please provide paths to the configuration files to load e.g > conan diff /conf/conf.json /otherconf/*.json\n")
			return
		}

		files := make([]ConfigFile, 0)

		for _, path := range args {
			matches, err := filepath.Glob(path)

			cobra.CheckErr(err)

			if matches == nil {
				log.Warn("no files found for arg ", path)
			} else {
				log.Debug("for arg ", path, " found files ", matches)
				for _, file := range matches {
					configFile := ConfigFile{FileName: file}
					configFile.Read()
					files = append(files, configFile)
				}
			}
		}

		if len(files) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No configuration files found for provided paths.\n")
			return
		}

		var diffResults DiffResults

		// loop through connectors and compare their config to what is deployed
		for _, file := range files {
			log.Debug(file.ConnectorName)

			// get the currently deployed config for the connector
			resp := GetConnectorConfig(host, port, file.ConnectorName)
			log.Debug(resp)
			if _, exists := resp["error_code"]; exists {
				if message, exists := resp["message"]; exists {
					if strings.Contains(message, "not found") {
						log.Debug("the connector " + file.ConnectorName + " doesn't exist")
						diffResults.NewConnectors = append(diffResults.NewConnectors, file.ConnectorName)
					} else {
						log.Warn("there was an error getting the deployed connector config for " + file.ConnectorName + " - " + message)
					}
				} else {
					log.Warn("there was an error getting the deployed connector config for " + file.ConnectorName)
				}
				continue
			}

			newKeys := make(map[string]string)
			matchKeys := make(map[string]string)
			mismatchKeys := make(map[string]MismatchVals)
			removedKeys := make(map[string]string)

			for fileKey, fileVal := range file.Config {

				if deployedVal, exists := resp[fileKey]; exists {
					if fileVal == deployedVal {
						matchKeys[fileKey] = cleanseVal(fileKey, fileVal)
					} else {
						mismatchKeys[fileKey] = MismatchVals{Deployed: cleanseVal(fileKey, deployedVal), File: cleanseVal(fileKey, fileVal)}
					}

				} else {
					newKeys[fileKey] = cleanseVal(fileKey, fileVal)
				}
			}

			for deployedKey, deployedVal := range resp {
				if _, exists := file.Config[deployedKey]; !exists {
					removedKeys[deployedKey] = cleanseVal(deployedKey, deployedVal)
				}
			}
			result := DiffResult{
				file.ConnectorName,
				newKeys,
				matchKeys,
				mismatchKeys,
				removedKeys,
			}
			if len(result.NewKeys) == 0 && len(result.MismatchKeys) == 0 && len(result.RemovedKeys) == 0 {
				// the connector is unchanged
				diffResults.UnchangedConnectors = append(diffResults.UnchangedConnectors, result.ConnectorName)
			} else {
				diffResults.ChangedConnectors = append(diffResults.ChangedConnectors, result)
			}
		}
		log.Debug(diffResults)

		err := templates.ExecuteTemplate(cmd.OutOrStdout(), "DiffTemplate", diffResults)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Error rendering DiffTemplate template %e.\n", err)
			return
		}
	},
}

var keysToHide = []string{"connection.pass", "connection.user", "connection.url", "password"}

func cleanseVal(key string, val string) string {
	for _, substr := range keysToHide {
		if strings.Contains(key, substr) {
			return "***hidden***"
		}
	}
	return val
}
func init() {
	rootCmd.AddCommand(diffCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// diffCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// diffCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
