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
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const ValidationTemplate = `CONNECTORS: {{ len . }}
{{ range $id, $file := . -}}
{{ printf "%-30s" $file.ConnectorName }} {{ printf "%-50s" $file.FileName }}
{{- if eq $file.ValidationResp.ErrorCount 0 -}} Valid {{- else -}} Invalid
{{- range $i, $field := $file.ValidationResp.Configs -}}
{{- if ne (len $field.Value.Errors) 0 }}
Error    Field: {{ $field.Value.Name }} - {{ $field.Value.Errors }}
{{- end -}}
{{- end }}
{{ end }}
{{ end }}
`

type ConfigFile struct {
	FileName       string
	ConnectorName  string
	ConnectorClass string
	PluginClass    string
	Config         map[string]string
	ConfigBytes    []byte
	ValidationResp ValidationResponse
	LoadResp       *http.Response
}

type LoadResponse struct {
}

func (cf *ConfigFile) Read() {
	configFile, _ := os.Open(cf.FileName)
	byteValue, _ := ioutil.ReadAll(configFile)

	var configObj map[string]interface{}
	json.Unmarshal([]byte(byteValue), &configObj)

	connectorName := strings.Split(filepath.Base(cf.FileName), ".")[0]

	var conf = make(map[string]string)

	if _, ok := configObj["config"]; ok {
		configConnectorName := configObj["name"].(string)
		if configConnectorName != connectorName {
			log.Warn("connector name [", configConnectorName, "] in file [", cf.FileName, "] does not match filename")
		}
		connectorName = configConnectorName

		configObj = configObj["config"].(map[string]interface{})
	}
	cf.ConnectorName = connectorName

	for k, v := range configObj {
		conf[k] = v.(string)
	}
	cf.Config = conf

	cf.ConnectorClass = cf.Config["connector.class"]
	classParts := strings.Split(cf.ConnectorClass, ".")
	cf.PluginClass = classParts[len(classParts)-1]

	cf.ConfigBytes, _ = json.Marshal(conf)

}

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:    "load",
	Short:  "Load connector config files into Kafka Connect",
	Long:   `Load connector config files into Kafka Connect`,
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No paths found.\n")
			return
		}

		// load the configs
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

		// validate
		valid := true
		for i, file := range files {
			files[i].ValidationResp = ValidateConfig(host, port, file)
			if files[i].ValidationResp.ErrorCount > 0 {
				valid = false
			}
		}

		t := template.Must(template.New("").Parse(ValidationTemplate))
		t.Execute(cmd.OutOrStdout(), files)

		if valid {
			for i, file := range files {
				files[i].LoadResp = LoadConfig(host, port, file)
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Validation errors found, skipping loading configs and exiting.\n")
			os.Exit(1)
		}

	},
}

func LoadConfig(host string, port string, configFile ConfigFile) *http.Response {
	validateUrl := fmt.Sprintf("http://%s:%s/connectors/%s/config", host, port, configFile.ConnectorName)

	req, err := http.NewRequest(http.MethodPut, validateUrl, bytes.NewBuffer(configFile.ConfigBytes))
	cobra.CheckErr(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	cobra.CheckErr(err)
	//respBodyBytes, _ := ioutil.ReadAll(resp.Body)

	log.Debug("Put config for: ", configFile.ConnectorName, " got response status: ", resp.StatusCode)
	return resp

}

func ValidateConfig(host string, port string, configFile ConfigFile) ValidationResponse {

	validateUrl := fmt.Sprintf("http://%s:%s/connector-plugins/%s/config/validate", host, port, configFile.PluginClass)

	req, err := http.NewRequest(http.MethodPut, validateUrl, bytes.NewBuffer(configFile.ConfigBytes))
	cobra.CheckErr(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	cobra.CheckErr(err)
	respBodyBytes, _ := ioutil.ReadAll(resp.Body)

	log.Debug("Got response status: ", resp.StatusCode)
	//fmt.Fprintf(cmd.OutOrStdout(), "response %s, StatusCode %d\n", string(respBodyBytes), resp.StatusCode)
	var validationResponse ValidationResponse

	json.Unmarshal(respBodyBytes, &validationResponse)

	// return ValidationResponse on a channel?

	return validationResponse
}

type ValidationResponse struct {
	ConnectorName string
	Name          string
	ErrorCount    int `json:"error_count"`
	Configs       []ValidationResponseField
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
