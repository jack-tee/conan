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
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const connectorStateTemplate = `{{ .Name }}  {{ .Connector.State }}
{{ range $task := .Tasks }}    Task {{ printf "%2d" $task.Id }}  {{ $task.State }}      {{ $task.WorkerId }}
{{ end }}`

type ConnectorStatus struct {
	Name      string
	Connector ConnectorState
	Tasks     []TaskState
}

type ConnectorState struct {
	State    string
	WorkerId string `json:"worker_id"`
}

type TaskState struct {
	Id       int
	State    string
	WorkerId string `json:"worker_id"`
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:    "list",
	Short:  "List the connectors",
	Long:   `List the connectors.`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		host, port = GetPersistentFlags(cmd)
		connectors := GetConnectors(host, port)
		t := template.Must(template.New("").Parse(connectorStateTemplate))

		fmt.Println("CONNECTORS")

		for _, connector := range connectors {
			status := GetConnectorStatus(host, port, connector)
			t.Execute(os.Stdout, status)
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

func GetConnectorStatus(host string, port string, connector string) ConnectorStatus {
	statusUrl := fmt.Sprintf("http://%s:%s/connectors/%s/status", host, port, connector)
	log.Debug("getting connector status using URL: ", statusUrl)
	resp, err := http.Get(statusUrl)
	cobra.CheckErr(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	cobra.CheckErr(err)

	var status ConnectorStatus
	json.Unmarshal(bodyBytes, &status)
	return status
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
