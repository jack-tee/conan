/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// stateCmd represents the state command
var stateCmd = &cobra.Command{
	Use:    "state",
	Short:  "List the current state of each connector",
	Long:   `List the current state of each connector`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		listState(cmd, cmd.OutOrStdout(), args)
	},
}

func listState(cmd *cobra.Command, output io.Writer, args []string) {
	host, port = GetPersistentFlags(cmd)
	connectors := GetConnectorsMap(host, port)
	connectors = GetConnectorsDetails(host, port, connectors)
	templates.ExecuteTemplate(output, "StateListTemplate", connectors)
}

var saveCmd = &cobra.Command{
	Use:    "save",
	Short:  "save connector state to a file",
	Long:   `save connector state to a file`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		t := time.Now().UTC()
		filename := ""

		if len(args) > 0 {
			filename = args[0]
		} else {
			filename = fmt.Sprintf("./conan-state-%s", t.Format(time.RFC3339))
		}
		log.Debug(fmt.Sprintf("creating file %s", filename))

		f, err := os.Create(filename)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}
		listState(cmd, f, args)
	},
}

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:    "set",
	Short:  "set connector state based on an input file",
	Long:   `set connector state based on an input file`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 0 {
			host, port = GetPersistentFlags(cmd)
			connectors := GetConnectorsMap(host, port)
			connectors = GetConnectorsDetails(host, port, connectors)

			connectorStateMap := make(map[string]string)

			for _, connector := range connectors {
				connectorStateMap[connector.Name] = connector.Details.Connector.State
			}
			log.Debug("existing connector state", connectorStateMap)

			f, err := os.Open(args[0])
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			line := 1

			for scanner.Scan() {

				t := strings.TrimSpace(scanner.Text())

				if len(t) > 0 {
					connStatus := strings.Split(t, ",")

					if len(connStatus) == 2 {

						if existingState, ok := connectorStateMap[connStatus[0]]; ok {
							if existingState != "RUNNING" && existingState != "PAUSED" {
								log.Debug(fmt.Sprintf("skipping connector %s as existing state is %s", string(connStatus[0]), existingState))
							} else if existingState == connStatus[1] {
								log.Debug(fmt.Sprintf("skipping connector %s as already in desired state %s", string(connStatus[0]), existingState))
							} else {
								switch connStatus[1] {
								case "PAUSED":
									ExecuteConnectorOp(Pause, host, port, string(connStatus[0]))
									log.Info(fmt.Sprintf("setting connector state for %s to PAUSED\n", string(connStatus[0])))
								case "RUNNING":
									ExecuteConnectorOp(Resume, host, port, string(connStatus[0]))
									log.Info(fmt.Sprintf("setting connector state for %s to RUNNING\n", string(connStatus[0])))
								default:
									log.Warn(fmt.Sprintf("skipping connector state for %s because desired state is %s existing state is %s", string(connStatus[0]), string(connStatus[1]), existingState))
								}
							}
						}
					} else {
						log.Warn(fmt.Sprintf("line %d in state file could not be parsed [%s] expected [{connectorName},{state}]\n", line, t))
					}
				}
				line += 1
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.AddCommand(saveCmd)
	stateCmd.AddCommand(setCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
