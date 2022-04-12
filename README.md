# prom_merge
merge promethues exporter metrics.

# how to use?
```
  -c, --config-path string      Path to the configuration file.
      --exporters-timeout int   HTTP client timeout for connecting to exporters. (ENV:MERGER_EXPORTERSTIMEOUT) (default 10)
  -h, --help                    help for prom_merge
      --listen-port int         Listen port for the HTTP server. (ENV:MERGER_PORT) (default 8080)
      --url strings             URL to scrape, Can be speficied multiple times. (ENV:MERGER_URLS,space-seperated)
  -v, --verbose                 Include debug messages to output. (ENV:MERGER_VERBOSE)
```

example:
```
./main --url="http://127.0.0.1:2112/metrics1" --url="http://127.0.0.1:2113/metrics2" --verbose

./main -c ./exporters.yaml -v
```

# alias host ip
exporters `--url="http://127.0.0.1:2112/metrics1"` alias name `instance` or others.

# yaml file
```
exporters:
- url: http://127.0.0.1:2112/metrics1
- url: http://127.0.0.1:2113/metrics2
host_alias: instance
```