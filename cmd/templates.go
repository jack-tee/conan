package cmd

const defaultTemplates = `
{{ define "ListTemplate" -}}
LIST: {{ len . }} Connectors
{{ range $id, $connector := . -}}
    {{ $connector.Id }} {{ printf "%-80s" $connector.Name }} {{ $connector.Details.Connector.State }}
    {{ range $task := $connector.Details.Tasks -}}
        {{- printf "%d.%-4d" $connector.Id $task.Id -}} 
        {{ printf "%-75.75s" $task.Summary }}
        {{- printf "  %s  %s  %s"  $task.State $task.WorkerId $task.Trace }}
    {{ end }}
{{ end }}
{{ end }}


{{ define "ValidationTemplate" -}}
VALIDATION: {{ len . }} Connectors
{{ range $id, $file := . -}}
{{ printf "%-50s" $file.FileName }} {{ printf "%-30s" $file.ConnectorName }}
{{- if and (eq $file.ValidationResp.ErrorCount 0) (eq $file.Error nil) -}} Config Valid. {{ if $file.LoadResp }} {{ $file.LoadResp.Status }} {{ end }}
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
{{ index .Config "topics" }} -> {{ index .Config "cps.topic" }}
{{ end }}
`
