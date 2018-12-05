---
title: Service-Broker Run
menu:
  product_service-broker_0.1.0:
    identifier: service-broker-run
    name: Service-Broker Run
    parent: reference
product_name: service-broker
menu_name: product_service-broker_0.1.0
section_menu_id: reference
---
## service-broker run

Launch AppsCode Service Broker server

### Synopsis

Launch AppsCode Service Broker server

```
service-broker run [flags]
```

### Options

```
      --async                   Indicates whether the broker is handling the requests asynchronously.
      --catalog-names strings   List of catalogs those can be run by this service-broker, comma separated.
      --catalog-path string     The path to the catalog. (default "/etc/config/catalogs")
  -h, --help                    help for run
      --insecure                use --insecure to use HTTP vs HTTPS.
      --kube-config string      specify the kube config path to be used.
      --port int                use '--port' option to specify the port for broker to listen on. (default 8080)
      --storage-class string    name of the storage-class for database storage. (default "standard")
      --tlsCert string          base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used. If '--tlsCert' is not used, then TLS will not be used.
      --tlsKey string           base-64 encoded PEM block to use as the private key matching the TLS certificate. If '--tlsKey' is used, then '--tlsCert' must also be used.
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send analytical events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [service-broker](/docs/reference/service-broker.md)	 - 

