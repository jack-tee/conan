# Conan
A simple <span style="color:#3a9025">Con</span>nector M<span style="color:#3a9025">an</span>ager for Kafka Connect.

Mainly aimed at managing JDBC Source Connectors.

## List connectors

```
> conan list

CONNECTORS: 3
0 an-example-bulk-connector                                                        RUNNING
    0.0   select * from texample                                                       RUNNING  10.0.0.1:8083  
    
1 an-example-whitelist-connector                                                   RUNNING
    1.0   mydatabase:myschema.table1                                                   RUNNING  10.0.0.2:8083  
    1.1   mydatabase:myschema.table1                                                   RUNNING  10.0.0.1:8083  
    1.2   mydatabase:myschema.table1                                                   RUNNING  10.0.0.2:8083  
    
2 an-example-pubsub-sink-connector                                                 PAUSED
    2.0    .* -> mypubsubtopic                                                         PAUSED  10.0.0.3:8083
```

You can filter by connector name

```
> conan list sink

CONNECTORS: 1 
2 an-example-pubsub-sink-connector                                                 PAUSED
    2.0    .* -> mypubsubtopic                                                         PAUSED  10.0.0.3:8083
```

## Pause/Resume Connectors

You can pause or resume connectors

```
> conan resume sink

CONNECTORS: 1 
2 an-example-pubsub-sink-connector                                                 PAUSED
    2.0    .* -> mypubsubtopic                                                         PAUSED  10.0.0.3:8083

Enter a connectorId to resume it, enter all to resume all LISTED connectors or q to quit: all
Resume all LISTED connectors? Enter y to confirm: y
Connector 2 an-example-pubsub-sink-connector resumed.
```