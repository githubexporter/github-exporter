# GitHub Mixin

## Overview
Mixins are a collection of configurable, reusable Prometheus rules, alerts and/or Grafana dashboards for a particular system, usually created by experts in that system. By applying them to Prometheus and Grafana, you can quickly set up appropriate monitoring for your systems.

The GitHub mixin currently provides simple dashboards for visualizing GitHub metrics emitted by the exporter.

To use them, you need to have `jb`, `mixtool` and `jsonnetfmt` installed. If you have a working Go development environment, it's easiest to run the following:
```bash
$ go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
$ go get github.com/monitoring-mixins/mixtool/cmd/mixtool
$ go get github.com/google/go-jsonnet/cmd/jsonnetfmt
```

You can then build a directory `dashboard_out` with the JSON dashboard files for Grafana:
```bash
$ make all
```

For more advanced uses of mixins, see https://github.com/monitoring-mixins/docs.

## Dashboards
* GitHub Repository Stats - Graphs GitHub metrics for a given repository. Any repository monitored by the exporter can be selected on this dashboard.
* GitHub API Usage - GitHub enforces rate limiting on the API used by the exporter. This dashboard can be used to monitor if the exporter is running out of requests.

## Future Development
The mixin can be extended with recording and alerting rules for Prometheus. 
