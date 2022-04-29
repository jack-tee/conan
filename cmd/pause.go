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

type Operation struct {
	Mode       string
	HttpMethod string
	Endpoint   string
}

var (
	Pause   Operation = Operation{"pause", http.MethodPut, "pause"}
	Resume  Operation = Operation{"resume", http.MethodPut, "resume"}
	Delete  Operation = Operation{"delete", http.MethodDelete, ""}
	Restart Operation = Operation{"restart", http.MethodPost, "restart"}
)

var pauseCmd = &cobra.Command{
	Use:    "pause",
	Short:  "Pause connectors",
	Long:   `Pause connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		opCommand(cmd, Pause, args)
	},
}

var resumeCmd = &cobra.Command{
	Use:    "resume",
	Short:  "Resume connectors",
	Long:   `Resume connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		opCommand(cmd, Resume, args)
	},
}

var deleteCmd = &cobra.Command{
	Use:    "delete",
	Short:  "Delete connectors",
	Long:   `Delete connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		opCommand(cmd, Delete, args)
	},
}

var restartCmd = &cobra.Command{
	Use:    "restart",
	Short:  "Restart connectors",
	Long:   `Restart connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		opCommand(cmd, Restart, args)
	},
}

func opCommand(cmd *cobra.Command, op Operation, args []string) {
	connectors := List(cmd, args)
	executeConnectorOperation(cmd, connectors, op)
}

func init() {
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(restartCmd)

	pauseCmd.Flags().StringVarP(&taskFilter, "task-filter", "t", "", "a substring to filter task summaries by")
	resumeCmd.Flags().StringVarP(&taskFilter, "task-filter", "t", "", "a substring to filter task summaries by")
	deleteCmd.Flags().StringVarP(&taskFilter, "task-filter", "t", "", "a substring to filter task summaries by")
	restartCmd.Flags().StringVarP(&taskFilter, "task-filter", "t", "", "a substring to filter task summaries by")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pauseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pauseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func executeConnectorOperation(cmd *cobra.Command, connectors map[int]Connector, op Operation) {
	fmt.Fprintf(cmd.OutOrStdout(), "Enter a connectorId to %s it e.g 4, enter all to %s all LISTED connectors or q to quit:\n", op.Mode, op.Mode)

	connectorIdSelected := AwaitConnectorInput()

	if connectorIdSelected == -1 {
		fmt.Fprintf(cmd.OutOrStdout(), "Quitting.\n")
		return

	} else if connectorIdSelected == -2 {
		fmt.Fprintf(cmd.OutOrStdout(), "%s all LISTED connectors? Enter y to confirm:\n", op.Mode)

		if AwaitUserConfirm() {
			for id, connector := range connectors {
				ExecuteOp(op, host, port, connector.Name)
				fmt.Fprintf(cmd.OutOrStdout(), "Connector %d %s %sd.\n", id, connector.Name, op.Mode)
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Quitting.\n")
			return
		}

	} else if connectorSelected, ok := connectors[connectorIdSelected]; !ok {
		fmt.Fprintf(cmd.OutOrStdout(), "ERROR. connectorId: [%d] not found in connectors. Exiting.\n", connectorIdSelected)

	} else {
		ExecuteOp(op, host, port, connectorSelected.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Connector %d %s %sd.\n", connectorSelected.Id, connectorSelected.Name, op.Mode)
	}
}

func ExecuteOp(op Operation, host string, port string, connectorName string) {
	var emptyBody io.Reader = nil

	opUrl := fmt.Sprintf("http://%s:%s/connectors/%s/%s", host, port, connectorName, op.Endpoint)
	log.Debug(op.Mode, " connector with URL: ", opUrl)

	req, err := http.NewRequest(op.HttpMethod, opUrl, emptyBody)
	cobra.CheckErr(err)

	resp, err := http.DefaultClient.Do(req)
	cobra.CheckErr(err)
	log.Debug("Got response status: ", resp.StatusCode)

}
