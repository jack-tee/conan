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

// pauseCmd represents the pause command
var pauseCmd = &cobra.Command{
	Use:    "pause",
	Short:  "Pause connectors",
	Long:   `Pause connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {

		connectors := List(cmd, args)

		fmt.Fprintf(cmd.OutOrStdout(), "Enter a connectorId to pause it e.g 4, enter all to pause all LISTED connectors or q to quit:\n")
		connectorIdToPause := AwaitConnectorInput()

		if connectorIdToPause == -1 {
			fmt.Fprintf(cmd.OutOrStdout(), "Quitting\n")
			return

		} else if connectorIdToPause == -2 {
			fmt.Fprintf(cmd.OutOrStdout(), "Pausing all connectors\n")
			for id, connector := range connectors {
				fmt.Fprintf(cmd.OutOrStdout(), "Pausing connector %d %s\n", id, connector.Name)
				PauseConnector(host, port, connector.Name)
			}

		} else if connectorToPause, ok := connectors[connectorIdToPause]; !ok {
			fmt.Fprintf(cmd.OutOrStdout(), "ERROR. connectorId: [%d] not found in connectors. Exiting.\n", connectorIdToPause)

		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Pausing connector %d %s\n", connectorToPause.Id, connectorToPause.Name)
			PauseConnector(host, port, connectorToPause.Name)

		}
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pauseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pauseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func PauseConnector(host string, port string, connectorName string) {
	PutConnector("pause", host, port, connectorName)
}

func ResumeConnector(host string, port string, connectorName string) {
	PutConnector("resume", host, port, connectorName)
}

func PutConnector(mode string, host string, port string, connectorName string) {
	var emptyBody io.Reader = nil

	pauseUrl := fmt.Sprintf("http://%s:%s/connectors/%s/%s", host, port, connectorName, mode)
	log.Debug(mode, " connector with URL: ", pauseUrl)

	req, err := http.NewRequest(http.MethodPut, pauseUrl, emptyBody)
	cobra.CheckErr(err)

	resp, err := http.DefaultClient.Do(req)
	cobra.CheckErr(err)
	log.Debug("Got response status: ", resp.StatusCode)

}
