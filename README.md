# Conan
A simple <span style="color:#3a9025">Con</span>nector M<span style="color:#3a9025">an</span>ager for Kafka Connect.

Mainly aimed at managing JDBC Source Connectors.

## List Connectors

The list command provides a summary of each task, returning the `tables` or `query` config parameter to make it easy to identify each task.
```
> conan list

CONNECTORS: 3
0 an-example-bulk-connector                                     RUNNING
    0.0   select * from texample                                    RUNNING  10.0.0.1:8083  
    
1 an-example-whitelist-connector                                RUNNING
    1.0   mydatabase:myschema.table1                                RUNNING  10.0.0.2:8083  
    1.1   mydatabase:myschema.table2                                RUNNING  10.0.0.1:8083  
    1.2   mydatabase:myschema.table3                                RUNNING  10.0.0.2:8083  
    
2 an-example-pubsub-sink-connector                              PAUSED
    2.0    .* -> mypubsubtopic                                      PAUSED  10.0.0.3:8083
```

You can filter by connector name

```
> conan list sink

CONNECTORS: 1 
2 an-example-pubsub-sink-connector                              PAUSED
    2.0    .* -> mypubsubtopic                                      PAUSED  10.0.0.3:8083
```

Or by the task summary to make it easier to identify connectors that read from certain tables

```
> conan list -t example

CONNECTORS: 1
0 an-example-bulk-connector                                     RUNNING
    0.0   select * from texample                                    RUNNING  10.0.0.1:8083

```

### Top level args
You can specify a host and port to connect to, by default it uses localhost:8083, there is also a debug mode to see debug output

```
> conan -H myhost -p 4042 list -d

INFO[0000] debug logs enabled                           
DEBU[0000] getting connectors using URL: http://myhost:4042/connectors 
DEBU[0000] connectors found: [an-example-bulk-connector an-example-whitelist-connector an-example-pubsub-sink-connector]
DEBU[0000] getting connector details for 0 an-example-bulk-connector 
DEBU[0000] getting connector status using URL: http://myhost:4042/connectors/an-example-bulk-connector/status
...


CONNECTORS: 3
0 an-example-bulk-connector                                     RUNNING
    0.0   select * from texample                                    RUNNING  10.0.0.1:8083
...

```


## Pause/Resume/Delete Connectors

You can pause or resume connectors

```
> conan resume sink

CONNECTORS: 1 
2 an-example-pubsub-sink-connector                              PAUSED
    2.0    .* -> mypubsubtopic                                      PAUSED  10.0.0.3:8083

Enter a connectorId to resume it, enter all to resume all LISTED connectors or q to quit: all
Resume all LISTED connectors? Enter y to confirm: y
Connector 2 an-example-pubsub-sink-connector resumed.


... sometime later ...


> conan pause pubsub

CONNECTORS: 1 
2 an-example-pubsub-sink-connector                              RUNNING
    2.0    .* -> mypubsubtopic                                      RUNNING  10.0.0.3:8083

Enter a connectorId to pause it, enter all to pause all LISTED connectors or q to quit: 2
Connector 2 an-example-pubsub-sink-connector paused.
```

Or delete them

```
conan delete pubsub

CONNECTORS: 1 
2 an-example-pubsub-sink-connector                              PAUSED
    2.0    .* -> mypubsubtopic                                      PAUSED  10.0.0.3:8083

Enter a connectorId to delete it, enter all to delete all LISTED connectors or q to quit: 2
Connector 2 an-example-pubsub-sink-connector deleted.
```

## Loading Connectors
You can load and update connectors

This command validates all connector configuration first and will then load the connectors if there are no errors reported.

```
> conan load connectors/*.json /another/path/myconnector.json

CONNECTORS: 3
my-bulk-connector              config/bulk.json                        Valid
my-sink-connector              config/sink.json                        Valid
my-other-connector             /another/path/myconnector.json          Valid

```
If there are errors they will be reported and no connectors will be loaded

```
CONNECTORS: 3
my-sink-connector              config/sink.json                        Valid
my-bulk-connector              config/bulk.json                        Invalid
Error    Field: name - [Missing required configuration "name" which has no default value.]
Error    Field: incrementing.column.name - [Query mode must be specified]
Error    Field: timestamp.column.name - [Query mode must be specified]
Error    Field: timestamp.initial - [Query mode must be specified]
Error    Field: validate.non.null - [Query mode must be specified]

my-other-connector             /another/path/myconnector.json          Valid

Validation errors found, skipping the loading of configs and exiting.
```

## Comparing Connector Config (diff)

The `diff` command can be used to show differences between connector json files and what is currently deployed to Kafka Connect.

This can be used before loading connectors to check what changes will be applied.

```
> diff my-connectors/*.json

Unchanged Connectors: 2
    my-existing-connector-1
    my-existing-connector-2

New Connectors: 2
    my-new-connector-1
    my-new-connector-2

Changed Connectors: 3
    my-example-connector
      + query.suffix: LIMIT 20000
      ~ poll.interval.ms: 86400000 -> 900000
      - numeric.mapping: best_fit

    my-custom-query-connector
      ~ query: 
          - SELECT * FROM (SELECT * FROM example WHERE status IN ('Created'))
          + SELECT * FROM (SELECT * FROM example WHERE status IN ('Created', 'Removed'))
```
Coloured output helps make the above command more readable.


## Saving and Setting Connector State

Imagine the scenario where you have many connectors running and you need to pause a chunk of them for whatever reason and then want to return to the previous state. Conan can save the current state of all connectors using

```
> conan state save 
```

or

```
> conan state save mystatefile
```

You'd then pause whichever connectors you need to, e.g

```
> conan pause db1

...

Connector 2 db1-table1-connector paused.
Connector 3 db1-table2-connector paused.
Connector 4 db1-table3-connector paused.

```

Then set the connector state back to your saved state file

```
> conan state set mystatefile

setting connector state for db1-table1-connector to RUNNING
setting connector state for db1-table2-connector to RUNNING
setting connector state for db1-table3-connector to RUNNING

```


## Templated Output
It is possible to override or add to the console output for most commands.

A common use case for this is to make the task summary more relevant to each connector.

For example if you were using a `org.apache.kafka.connect.file.FileStreamSinkConnector` by default you would get output from the `list` command like

```
> conan list

CONNECTORS: 1
0 my-file-sink                                                  RUNNING
    0.0                                                             RUNNING  10.0.0.1:8083
```
It tells you it's running but not much else, we can modify what it tells us about a task using templates.

So if we create a templates/task_summary.tmpl file that looks something like:
```
{{ define "org.apache.kafka.connect.file.FileStreamSinkConnector" }}
{{ index .Config "topic" }} -> {{ index .Config "file" }}
{{ end }}
```
Now we can pass this file as the `--templatesPath` argument which allows a glob path for where to look for template files.

> Note the default value for templatesPath is `templates/*.tmpl` so you can omit this argument if you place your template files in this directory relative to the path where conan is run from.

```
conan list --templatesPath templates/task_summary.tmpl -d

...
DEBU[0000] Loading templates using path templates/task_summary.tmpl 
DEBU[0000] Templates loaded.; defined templates are: "org.apache.kafka.connect.file.FileStreamSinkConnector", "ListTemplate", "ValidationTemplate", "io.confluent.connect.jdbc.JdbcSourceConnector"
...


CONNECTORS: 1
0 my-file-sink                                                  RUNNING
    0.0 my.topic -> /tmp/myfile.txt                                 RUNNING  10.0.0.1:8083
```
This can then also be used to filter tasks, with the above template in place you can find the task that is writing to a specific file
```
> conan list -t myfile

CONNECTORS: 1
0 my-file-sink                                                  RUNNING
    0.0 my.topic -> /tmp/myfile.txt                                 RUNNING  10.0.0.1:8083
```
