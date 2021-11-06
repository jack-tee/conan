# Conan
A simple <span style="color:#3a9025">Con</span>nector M<span style="color:#3a9025">an</span>ager for Kafka Connect.

Mainly aimed at managing JDBC Source Connectors.

## List connectors

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

## Loading connectors
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
