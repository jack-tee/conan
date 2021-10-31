package cmd

import (
	"fmt"
)

type TaskState struct {
	Id       int
	State    string
	WorkerId string `json:"worker_id"`
	Trace    string
	Config   map[string]string
}

func (t TaskState) Summary() string {
	if tables, ok := t.Config["tables"]; ok && tables != "" {
		return tables

	} else if query, ok := t.Config["query"]; ok && query != "" {
		return query

	} else if cpstopic, ok := t.Config["cps.topic"]; ok && cpstopic != "" {
		topicsregex := t.Config["topics.regex"]
		return fmt.Sprintf("%s -> %s", topicsregex, cpstopic)

	} else {
		return ""
	}
}
