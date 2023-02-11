package cmd

const defaultTemplates = `
{{ define "ListTemplate" -}}
LIST: {{ len . }} Connectors
{{ range $id, $connector := . -}}
    {{ printf "%-3d %-78s" $connector.Id $connector.Name }} {{ printf "%-11s" $connector.Details.Connector.FormattedState }} {{ $connector.PollInterval }}
    {{ range $task := $connector.Details.Tasks -}}
        {{- printf "%3d.%-2d" $connector.Id $task.Id -}} 
        {{ printf "%-75.75s" $task.Summary }}
        {{- printf " %8s %s  %s"  $task.FormattedState $task.WorkerId $task.Trace }}
    {{ end }}
{{ end }}
Total: {{ len . }} Connectors
{{ end }}

{{ define "StateListTemplate" -}}
{{ range $id, $connector := . -}}
    {{ printf "%s" $connector.Name }},{{ $connector.Details.Connector.State }}
{{ end }}
{{ end }}

{{ define "ValidationTemplate" -}}
VALIDATION: {{ len . }} Connectors
{{ range $id, $file := . -}}
{{ printf "%-50s" $file.FileName }} {{ printf "%-30s" $file.ConnectorName }}
{{- if and (eq $file.ValidationResp.ErrorCount 0) (eq $file.Error nil) -}} Config Valid. {{ if $file.LoadResp }}- {{ $file.FormattedStatus }} {{ end }}
{{- else -}} Config Invalid.
{{- if ne $file.Error nil }}
    File Error {{ $file.Error }}
{{- else -}}
    {{- range $i, $field := $file.ValidationResp.Configs -}}
    {{- if ne (len $field.Value.Errors) 0 }}
        Config Error    Field: {{ $field.Value.Name }} - {{ $field.Value.Errors }}
    {{- end -}}
    {{- end }}
{{- end -}}
{{- end }}
{{ end }}
{{ end }}

{{ define "DiffTemplate" -}}
Unchanged Connectors: {{ len .UnchangedConnectors }}
{{- range $id, $name := .UnchangedConnectors }}
    {{ $name }}
{{- end }}

New Connectors: {{ len .NewConnectors }}
{{- range $id, $name := .NewConnectors }}
    {{ Green $name }}
{{- end }}

Changed Connectors: {{ len .ChangedConnectors }}
{{- range $id, $diff := .ChangedConnectors }}
    {{ $diff.ConnectorName }}
    {{- range $key, $val := $diff.NewKeys }}
      {{ Green (printf "+ %s: %s" $key $val) }}
    {{- end }}
    {{- range $key, $val := $diff.MismatchKeys }}
        {{- if gt (len $val.Deployed ) 40 }}
      {{ Yellow "~" }} {{ $key }}: 
          {{ Red (printf "- %s" $val.Deployed) }}
          {{ Green (printf "+ %s" $val.File) }}
        {{- else }}
      {{ Yellow "~" }} {{ $key }}: {{ Red $val.Deployed }} -> {{ Green  $val.File }}
        {{- end }}
    {{- end }}
    {{- range $key, $val := $diff.RemovedKeys }}
      {{ Red (printf "- %s: %s" $key $val) }}
    {{- end }}
{{ end }}
Unchanged: {{ len .UnchangedConnectors }}, New: {{ len .NewConnectors }}, Changed: {{ len .ChangedConnectors }}
{{ end }}


{{ define "io.confluent.connect.jdbc.JdbcSourceConnector" }}
{{- if ne (index .Config "tables") "" -}}
{{- index .Config "tables" -}}
{{- else if ne (index .Config "query") "" -}}
{{- index .Config "query" -}}
{{- else -}}
n/a
{{- end -}}
{{ end }}


{{ define "com.google.pubsub.kafka.sink.CloudPubSubSinkConnector" }}
{{- or (index .Config "topics") (index .Config "topics.regex") }} -> {{ index .Config "cps.topic" -}}
{{ end }}

{{ define "com.google.pubsub.kafka.source.CloudPubSubSourceConnector" }}
{{- index .Config "cps.subscription" }} -> {{ index .Config "kafka.topic" -}}
{{ end }}

{{ define "org.apache.kafka.connect.mirror.MirrorSourceConnector" }}
{{- index .Config "topics" }} with prefix {{ index .Config "source.cluster.alias" -}}
{{ end }}
`
