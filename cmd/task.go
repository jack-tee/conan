package cmd

import (
	"bytes"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type TaskState struct {
	Id       int
	State    string
	WorkerId string `json:"worker_id"`
	Trace    string
	Config   map[string]string
}

func (t TaskState) Summary() string {
	tmpl := templates.Lookup(t.Config["connector.class"])
	if tmpl == nil {
		log.Debug("Template not found for ", t.Config["connector.class"])
		return ""
	}

	var output bytes.Buffer
	err := tmpl.ExecuteTemplate(&output, t.Config["connector.class"], t)
	cobra.CheckErr(err)
	return output.String()
}

func (t TaskState) FormattedState() string {
	return FormatState(t.State)
}
