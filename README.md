# Prometheus GitHub Exporter

Exposes basic metrics for your repositories from the GitHub API, to a Prometheus compatible endpoint. 

## Configuration

This exporter is setup to take two parameters from environment variables:
`BIND_PORT` The port you wish to run the container on, defaults to 1234
`REPOS` The images you wish to monitor, expected in the format "user/repo1, user/repo2". Can be across different Github users/orgs.

## Install and deploy

Run manually from Docker Hub:
```
docker run -d --restart=always -p 9171:9171 -e IMAGES="infinityworks/ranch-eye, infinityworks/prom-conf" infinityworks/github-exporter
```

Build a docker image:
```
docker build -t <image-name> .
docker run -d --restart=always -p 9171:9171 -e IMAGES="infinityworks/ranch-eye, infinityworks/prom-conf" <image-name>
```

## Docker compose

```
github-exporter:
    tty: true
    stdin_open: true
    expose:
      - 1234:1234
    image: infinityworks/github-exporter
```

## Metrics

Metrics will be made available on port 9171 by default

```
# HELP github_forks Gauge of forks from the public API
# TYPE github_forks gauge
github_forks{repo="docker-hub-exporter",user="infinityworksltd"} 0.0
github_forks{repo="prometheus-rancher-exporter",user="infinityworksltd"} 9.0
# HELP github_stars Gauge of stars from the public API
# TYPE github_stars gauge
github_stars{repo="docker-hub-exporter",user="infinityworksltd"} 1.0
github_stars{repo="prometheus-rancher-exporter",user="infinityworksltd"} 6.0
# HELP github_open_issues Gauge of issues from the public API
# TYPE github_open_issues gauge
github_open_issues{repo="docker-hub-exporter",user="infinityworksltd"} 0.0
github_open_issues{repo="prometheus-rancher-exporter",user="infinityworksltd"} 2.0
# HELP github_watchers Gauge of watchers from the public API
# TYPE github_watchers gauge
github_watchers{repo="docker-hub-exporter",user="infinityworksltd"} 1.0
github_watchers{repo="prometheus-rancher-exporter",user="infinityworksltd"} 6.0
```

## Metadata
[![](https://images.microbadger.com/badges/image/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own version badge on microbadger.com")