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
	golog "log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	skipConfirm bool = false
)

type ConfigFile struct {
	FileName       string
	ConnectorName  string
	ConnectorClass string
	PluginClass    string
	Config         map[string]string
	ConfigBytes    []byte
	ValidationResp ValidationResponse
	LoadResp       *http.Response
	Error          error
}

func (cf *ConfigFile) FormattedStatus() string {
	return FormatStatus(cf.LoadResp.Status, cf.LoadResp.StatusCode)
}

func (cf *ConfigFile) Read() {
	configFile, _ := os.Open(cf.FileName)
	byteValue, _ := ioutil.ReadAll(configFile)

	var configObj map[string]interface{}
	err := json.Unmarshal([]byte(byteValue), &configObj)

	if err != nil {
		cf.Error = err
		return
	}

	connectorName := strings.Split(filepath.Base(cf.FileName), ".")[0]

	var conf = make(map[string]string)

	// check the configured name matches the filename
	configConnectorName := configObj["name"].(string)
	if configConnectorName != connectorName {
		log.Warnf("connector name [%s] does not match the name of the file [%s]", configConnectorName, cf.FileName)
	}
	connectorName = configConnectorName

	// if there is a config sub object use it
	if _, ok := configObj["config"]; ok {
		configObj = configObj["config"].(map[string]interface{})
	}

	cf.ConnectorName = connectorName

	for k, v := range configObj {

		switch t := v.(type) {
		case int:
			conf[k] = strconv.Itoa(t)
		case string:
			conf[k] = t
		case float64:
			conf[k] = strings.Trim(strings.Trim(fmt.Sprintf("%f", t), "0"), ".")
		default:
			log.Errorf("type of value for key %s is not understood", k)
		}

	}
	log.Debugf("Conf is: %+v", conf)
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
			fmt.Fprintf(cmd.OutOrStdout(), "No args provided. Please provide paths to the configuration files to load e.g > conan load /conf/conf.json /otherconf/*.json\n")
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

		if len(files) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No configuration files found for provided paths.\n")
			return
		}

		rhttp := retryablehttp.NewClient()

		// the retryablehttp client generates it's own logs that are not levelled
		// the following prevents these logs from being outputted if the debg flag is not set
		if !debug {
			rhttp.Logger = golog.New(ioutil.Discard, "", golog.LstdFlags)
		}
		rhttp.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, attempt int) {
			log.Debugf("Making request %d to %s", attempt, req.URL)
		}
		rhttp.ResponseLogHook = func(_ retryablehttp.Logger, resp *http.Response) {
			log.Debugf("received response from: %s status: %s", resp.Request.URL, resp.Status)
		}

		rhttp.RetryMax = 3
		rhttp.RetryWaitMin = time.Duration(5 * time.Second)

		// validate
		var allValid = true
		for i, file := range files {

			if file.Error != nil {
				allValid = false
				continue
			}

			files[i].ValidationResp = ValidateConfig(rhttp, host, port, file)
			if files[i].ValidationResp.ErrorCount > 0 {
				allValid = false
			}

		}

		if allValid && skipConfirm {
			fmt.Fprintf(cmd.OutOrStdout(), "All connectors are valid. Loading configs.\n")
			for i, file := range files {
				files[i].LoadResp = LoadConfig(rhttp, host, port, file)
			}
		} else if allValid {
			err := templates.ExecuteTemplate(cmd.OutOrStdout(), "ValidationTemplate", files)

			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Error rendering ValidationTemplate template %e.\n", err)
				return
			}
			fmt.Fprintf(cmd.OutOrStdout(), "All connectors are valid. Load connectors? y/N ")
			if AwaitUserConfirm() {
				fmt.Fprintf(cmd.OutOrStdout(), "Loading configs.\n")
				for i, file := range files {
					files[i].LoadResp = LoadConfig(rhttp, host, port, file)
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Skipped loading configs.\n")
				return
			}
		}

		err := templates.ExecuteTemplate(cmd.OutOrStdout(), "ValidationTemplate", files)

		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Error rendering ValidationTemplate template %e.\n", err)
			os.Exit(1)
		}
		if !allValid {
			fmt.Fprintf(cmd.OutOrStdout(), "Validation errors found, skipped loading configs.\n")
			os.Exit(1)
		}
	},
}

func LoadConfig(client *retryablehttp.Client, host string, port string, configFile ConfigFile) *http.Response {
	validateUrl := fmt.Sprintf("http://%s:%s/connectors/%s/config", host, port, configFile.ConnectorName)

	req, err := retryablehttp.NewRequest(http.MethodPut, validateUrl, bytes.NewBuffer(configFile.ConfigBytes))
	cobra.CheckErr(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	cobra.CheckErr(err)
	respBodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Debug(string(respBodyBytes))
	log.Debug("Put config for: ", configFile.ConnectorName, " got response status: ", resp.Status, " and code: ", resp.StatusCode)
	return resp

}

func ValidateConfig(client *retryablehttp.Client, host string, port string, configFile ConfigFile) ValidationResponse {

	validateUrl := fmt.Sprintf("http://%s:%s/connector-plugins/%s/config/validate", host, port, configFile.PluginClass)

	req, err := retryablehttp.NewRequest(http.MethodPut, validateUrl, bytes.NewBuffer(configFile.ConfigBytes))
	cobra.CheckErr(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
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

	loadCmd.Flags().BoolVarP(&skipConfirm, "skip-confirm", "f", false, "whether to prompt for confirmation when loading connectors")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
