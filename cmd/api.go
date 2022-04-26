package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Connector struct {
	Id      int
	Name    string
	Details ConnectorDetails
}

func (c Connector) PollInterval() string {
	if _, ok := c.Details.Config["poll.interval.ms"]; !ok {
		return ""
	}
	pollIntervalMs, err := strconv.Atoi(c.Details.Config["poll.interval.ms"])
	if err != nil {
		log.Debug("could not format poll.interval.ms")
		return ""
	}
	return FormatPollInterval((pollIntervalMs))
}

// GetConnectorsMap returns a map of connectorId -> connectorName
// The connectorId is based on the alphabetically sorted connectorNames
func GetConnectorsMap(host string, port string) map[int]Connector {

	url := fmt.Sprintf("http://%s:%s/connectors", host, port)
	log.Debug("getting connectors using URL: ", url)

	resp, err := http.Get(url)
	cobra.CheckErr(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	cobra.CheckErr(err)

	var connectors []string
	json.Unmarshal(bodyBytes, &connectors)

	sort.Strings(connectors)
	log.Debug("connectors found: ", connectors)

	connectorsMap := make(map[int]Connector)

	for i, connector := range connectors {
		connectorsMap[i] = Connector{Id: i, Name: connector}
	}

	log.Debug("connectorsMap: ", connectorsMap)

	return connectorsMap
}

func GetConnectorsDetails(host string, port string, connectors map[int]Connector) map[int]Connector {
	for connectorId, connector := range connectors {
		connector.Details = GetConnectorDetails(host, port, connectorId, connector.Name)
		connectors[connectorId] = connector
	}
	return connectors
}

type ConnectorDetails struct {
	Name      string
	Config    map[string]string
	Connector ConnectorState
	Tasks     []TaskState
}

type ConnectorState struct {
	State    string
	WorkerId string `json:"worker_id"`
}

func (c ConnectorState) FormattedState() string {
	return FormatState(c.State)
}

// GetConnectorDetails gets all connector and task statuses and config for a given connectorName
func GetConnectorDetails(host string, port string, connectorId int, connectorName string) ConnectorDetails {
	log.Debug("getting connector details for ", connectorId, " ", connectorName)

	cDetails := GetConnectorStatus(host, port, connectorName)
	cDetails.Config = GetConnectorConfig(host, port, connectorName)

	tasksMap := GetConnectorTasks(host, port, connectorName)

	for j, connectorTaskStatus := range cDetails.Tasks {
		cDetails.Tasks[j].Config = tasksMap[connectorTaskStatus.Id].Config
	}
	return cDetails

}

// GetConnectorStatus gets the connector status
func GetConnectorStatus(host string, port string, connectorName string) ConnectorDetails {
	statusUrl := fmt.Sprintf("http://%s:%s/connectors/%s/status", host, port, connectorName)
	log.Debug("getting connector status using URL: ", statusUrl)
	resp, err := http.Get(statusUrl)
	cobra.CheckErr(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	cobra.CheckErr(err)

	var status ConnectorDetails
	json.Unmarshal(bodyBytes, &status)
	return status
}

// GetConnectorConfig gets the connector config
func GetConnectorConfig(host string, port string, connectorName string) map[string]string {
	url := fmt.Sprintf("http://%s:%s/connectors/%s/config", host, port, connectorName)
	log.Debug("getting connector config using URL: ", url)
	resp, err := http.Get(url)
	cobra.CheckErr(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	cobra.CheckErr(err)

	var config map[string]string
	json.Unmarshal(bodyBytes, &config)
	return config
}

// Tasks

type TaskStatus struct {
	Id     TaskStatusId
	Config map[string]string
}

type TaskStatusId struct {
	Connector string
	TaskId    int `json:"task"`
}

type TaskStatusConfig struct {
	Tables      string
	Query       string
	PubsubTopic string `json:"cps.topic"`
	TopicsRegex string `json:"topics.regex"`
}

func GetConnectorTasks(host string, port string, connector string) map[int]TaskStatus {

	statusUrl := fmt.Sprintf("http://%s:%s/connectors/%s/tasks", host, port, connector)
	log.Debug("getting tasks using URL: ", statusUrl)

	resp, err := http.Get(statusUrl)
	cobra.CheckErr(err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	cobra.CheckErr(err)

	//fmt.Println(string(bodyBytes))

	var tasks []TaskStatus
	json.Unmarshal(bodyBytes, &tasks)
	//fmt.Println(tasks)

	tasksMap := make(map[int]TaskStatus)

	for _, taskStatus := range tasks {
		tasksMap[taskStatus.Id.TaskId] = taskStatus
	}

	return tasksMap

}
