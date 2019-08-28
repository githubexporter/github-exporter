[![Build Status](https://travis-ci.org/infinityworks/github-exporter.svg?branch=master)](https://travis-ci.org/infinityworks/github-exporter)

# Prometheus GitHub Exporter

Exposes basic metrics for your repositories from the GitHub API, to a Prometheus compatible endpoint.

## Configuration

This exporter is setup to take input from environment variables:

### Required
* `ORGS` If supplied, the exporter will enumerate all repositories for that organization. Expected in the format "org1, org2".
* `REPOS` If supplied, The repos you wish to monitor, expected in the format "user/repo1, user/repo2". Can be across different Github users/orgs.
* `USERS` If supplied, the exporter will enumerate all repositories for that users. Expected in
the format "user1, user2".

At least one of those 3 options should be provided.

### Optional
* `GITHUB_TOKEN` If supplied, enables the user to supply a github authentication token that allows the API to be queried more often. Optional, but recommended.
* `GITHUB_TOKEN_FILE` If supplied _instead of_ `GITHUB_TOKEN`, enables the user to supply a path to a file containing a github authentication token that allows the API to be queried more often. Optional, but recommended.
* `API_URL` Github API URL, shouldn't need to change this. Defaults to `https://api.github.com`
* `LISTEN_PORT` The port you wish to run the container on, the Dockerfile defaults this to `9171`
* `METRICS_PATH` the metrics URL path you wish to use, defaults to `/metrics`
* `LOG_LEVEL` The level of logging the exporter will run with, defaults to `debug`


## Install and deploy

Run manually from Docker Hub:
```
docker run -d --restart=always -p 9171:9171 -e REPOS="infinityworks/ranch-eye, infinityworks/prom-conf" infinityworks/github-exporter
```

Build a docker image:
```
docker build -t <image-name> .
docker run -d --restart=always -p 9171:9171 -e REPOS="infinityworks/ranch-eye, infinityworks/prom-conf" <image-name>
```

## Docker compose

```
github-exporter:
    tty: true
    stdin_open: true
    expose:
      - 9171
    ports:
      - 9171:9171
    image: infinityworks/github-exporter:latest
    environment:
      - REPOS=<REPOS you want to monitor>
      - GITHUB_TOKEN=<your github api token>

```

## Metrics

Metrics will be made available on port 9171 by default
An example of these metrics can be found in the `METRICS.md` markdown file in the root of this repository

## Tests

There is a set of blackbox behavioural tests which validate metrics endpoint in the `test` directory. 
Run as follows

```bash
make test
```

## Metadata
[![](https://images.microbadger.com/badges/image/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own version badge on microbadger.com")
