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
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type ConnectorMode string

const (
	Pause  ConnectorMode = "pause"
	Resume ConnectorMode = "resume"
)

var pauseCmd = &cobra.Command{
	Use:    "pause",
	Short:  "Pause connectors",
	Long:   `Pause connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {

		connectors := List(cmd, args)

		executeConnectorOperation(cmd, connectors, Pause)

	},
}

var resumeCmd = &cobra.Command{
	Use:    "resume",
	Short:  "Resume connectors",
	Long:   `Resume connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {

		connectors := List(cmd, args)

		executeConnectorOperation(cmd, connectors, Resume)
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(resumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pauseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pauseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func executeConnectorOperation(cmd *cobra.Command, connectors map[int]Connector, mode ConnectorMode) {
	fmt.Fprintf(cmd.OutOrStdout(), "Enter a connectorId to %s it e.g 4, enter all to %s all LISTED connectors or q to quit:\n", mode, mode)

	connectorIdSelected := AwaitConnectorInput()

	if connectorIdSelected == -1 {
		fmt.Fprintf(cmd.OutOrStdout(), "Quitting.\n")
		return

	} else if connectorIdSelected == -2 {
		fmt.Fprintf(cmd.OutOrStdout(), "%s all LISTED connectors? Enter y to confirm:\n", mode)

		if AwaitUserConfirm() {
			for id, connector := range connectors {
				PutConnector(mode, host, port, connector.Name)
				fmt.Fprintf(cmd.OutOrStdout(), "Connector %d %s %sd.\n", id, connector.Name, mode)
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Quitting.\n")
		}

	} else if connectorSelected, ok := connectors[connectorIdSelected]; !ok {
		fmt.Fprintf(cmd.OutOrStdout(), "ERROR. connectorId: [%d] not found in connectors. Exiting.\n", connectorIdSelected)

	} else {
		PutConnector(mode, host, port, connectorSelected.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Connector %d %s %sd.\n", connectorSelected.Id, connectorSelected.Name, mode)
	}
}

func PutConnector(mode ConnectorMode, host string, port string, connectorName string) {
	var emptyBody io.Reader = nil

	pauseUrl := fmt.Sprintf("http://%s:%s/connectors/%s/%s", host, port, connectorName, mode)
	log.Debug(mode, " connector with URL: ", pauseUrl)

	req, err := http.NewRequest(http.MethodPut, pauseUrl, emptyBody)
	cobra.CheckErr(err)

	resp, err := http.DefaultClient.Do(req)
	cobra.CheckErr(err)
	log.Debug("Got response status: ", resp.StatusCode)

}
