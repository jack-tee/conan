package cmd

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func GetPersistentFlags(cmd *cobra.Command) (string, string) {
	host, err := cmd.Root().PersistentFlags().GetString("host")
	cobra.CheckErr(err)

	port, err := cmd.Root().PersistentFlags().GetString("port")
	cobra.CheckErr(err)

	return host, port
}

func HasCaseInsensitivePrefix(s string, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}

func AwaitUserConfirm() bool {
	input := "n"
	fmt.Scanln(&input)
	log.Debug(input, " input\n")

	if strings.Contains(strings.ToLower(input), "y") {
		log.Debug("confirmed")
		return true
	}
	return false

}

// AwaitConnectorInput prompts the user for input
// A connector can be selected by entering a connector id e.g
// > 4
// Which will result in a return value of 4
func AwaitConnectorInput() (bool, bool, []int) {

	selected := "-1"
	fmt.Scanln(&selected)
	log.Debug(selected, " selected\n")

	if strings.Contains(selected, "q") {
		log.Debug("found 'q' in input so exiting")
		return true, false, nil
	}

	if strings.Contains(selected, "all") {
		log.Debug("found 'all' in input so exiting")
		return false, true, nil
	}

	var connectorIdsSelected []int

	for _, c := range strings.Split(selected, ",") {
		c_int, err := strconv.Atoi(c)
		if err != nil {
			panic("could not parse connector id from selected")
		}
		connectorIdsSelected = append(connectorIdsSelected, c_int)
	}

	log.Debugf("parsed user input from [%s] connectorId: %v", selected, connectorIdsSelected)
	return false, false, connectorIdsSelected
}

// AwaitConnectorTaskInput prompts the user for input
// A connector can be selected by entering a connector id e.g
// > 4
// Which will result in a return value of 4, -1
// Or a connector and task can be selected by entering e.g
// > 5.2
// Which will result in a return value of 5, 2
func AwaitConnectorTaskInput() (int, int) {

	selected := "-1"
	fmt.Scanln(&selected)
	log.Debug(selected, " selected\n")

	var connectorIdSelected int = -1
	var taskIdSelected int = -1
	var err error

	if strings.Contains(selected, ".") {
		parts := strings.Split(selected, ".")

		connectorIdSelected, err = strconv.Atoi(parts[0])
		if err != nil {
			panic("could not parse connector id from selected[0]")
		}
		taskIdSelected, err = strconv.Atoi(parts[1])
		if err != nil {
			panic("could not parse task id from selected[1]")
		}

	} else {

		connectorIdSelected, err = strconv.Atoi(selected)
		if err != nil {
			panic("could not parse connector id from selected")
		}
	}
	log.Debugf("parsed user input from [%s] connectorId: %d taskId: %d", selected, connectorIdSelected, taskIdSelected)
	return connectorIdSelected, taskIdSelected
}

func FormatPollInterval(pollIntervalMs int) string {
	if pollIntervalMs < 1000 {
		return fmt.Sprintf("%dms", pollIntervalMs)
	}
	hours := pollIntervalMs / (1000 * 60 * 60)
	minutes := (pollIntervalMs % (1000 * 60 * 60)) / (1000 * 60)
	seconds := (pollIntervalMs % 60000) / 1000
	if hours == 0 {
		if minutes == 0 {
			return fmt.Sprintf("%ds", seconds)
		}
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[38m"
var White = "\033[97m"

func FormatState(state string) string {
	color := ""
	switch state {
	case "RUNNING":
		color = Green
	case "PAUSED":
		color = Yellow
	case "FAILED":
		color = Red
	case "UNASSIGNED":
		color = Gray
	}
	return color + state + Reset
}

func FormatStatus(status string, statusCode int) string {
	color := ""
	switch {
	case statusCode < 200:
		color = Gray
	case statusCode < 300:
		color = Green
	case statusCode < 400:
		color = Yellow
	default:
		color = Red
	}
	return color + status + Reset
}
