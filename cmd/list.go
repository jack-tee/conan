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
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var taskFilter string
var stateFilter string

type ConnectorStatus struct {
	ConnectorId int
	Name        string
	Connector   ConnectorState
	Tasks       []TaskState
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:    "list",
	Short:  "List the connectors",
	Long:   `List the connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		_ = List(cmd, args)
	},
}

func List(cmd *cobra.Command, args []string) map[int]Connector {
	host, port = GetPersistentFlags(cmd)
	connectors := GetConnectorsMap(host, port)

	// filter connectors by Name
	if len(args) > 0 {
		filteredConnectors := make(map[int]Connector)
		for i, c := range connectors {
			if strings.Contains(strings.ToLower(c.Name), strings.ToLower(args[0])) {
				filteredConnectors[i] = c
			}
		}

		connectors = filteredConnectors
		log.Debug("connectors filtered by arg to ", connectors)
	}

	connectors = GetConnectorsDetails(host, port, connectors)

	if stateFilter != "" {
		filteredConnectors := make(map[int]Connector)
		for i, c := range connectors {
			if HasCaseInsensitivePrefix(c.Details.Connector.State, stateFilter) {
				filteredConnectors[i] = c
			} else {
				for _, t := range c.Details.Tasks {
					if HasCaseInsensitivePrefix(t.State, stateFilter) {
						filteredConnectors[i] = c
					}
				}
			}
		}

		connectors = filteredConnectors
		log.Debug("connectors filtered by state-filter to ", connectors)
	}

	if taskFilter != "" {
		filteredConnectors := make(map[int]Connector)
		for i, c := range connectors {
			for _, t := range c.Details.Tasks {
				if strings.Contains(strings.ToLower(t.Summary()), strings.ToLower(taskFilter)) {
					filteredConnectors[i] = c
				}
			}
		}

		connectors = filteredConnectors
		log.Debug("connectors filtered by task-filter to ", connectors)
	}

	templates.ExecuteTemplate(cmd.OutOrStdout(), "ListTemplate", connectors)

	return connectors
}

func init() {
	//fmt.Println("Running list.go init")
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&taskFilter, "task-filter", "t", "", "a substring to filter task summaries by")
	listCmd.Flags().StringVarP(&stateFilter, "state-filter", "s", "", "filter to connectors / tasks in this state")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
