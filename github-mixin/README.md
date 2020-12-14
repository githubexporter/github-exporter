# GitHub Mixin

## Overview
The GitHub Mixin is a set of configurable, reusable, and extensible dashboards based on the metrics exported by the GitHub Exporter. The mixin creates suitable dashboard descriptions for Grafana. 

To use them, you need to have `jb`, `mixtool` and `jsonnetfmt` installed. If you have a working Go development environment, it's easiest to run the following:
```bash
$ go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
$ go get github.com/monitoring-mixins/mixtool/cmd/mixtool
$ go get github.com/google/go-jsonnet/cmd/jsonnetfmt
```

You can then build a directory `dashboard_out` with the JSON dashboard files for Grafana:
```bash
$ make build
```

For more advanced uses of mixins, see https://github.com/monitoring-mixins/docs.

The mixin can be extended with recording and alerting rules for Prometheus. 