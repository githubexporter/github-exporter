[![Build Status](https://travis-ci.org/infinityworks/github-exporter.svg?branch=master)](https://travis-ci.org/infinityworks/github-exporter)

# Prometheus GitHub Exporter

Exposes basic metrics for your repositories from the GitHub API, to a Prometheus compatible endpoint.

## Configuration

This exporter is setup to take configuration through environment variables. All variables are optional:

### Targets

All targets support arrays of values (comma delimited). e.g. "org1/repo1, org2/repo3".

* `ORGS` If supplied, the exporter will enumerate all repositories for any organizations are processed.
* `USERS` If supplied, the exporter will enumerate all repositories for that users in a similar fashion to organisations.
* `REPOS` If supplied, allows you to explicitly set the repos you wish to monitor Can be across different Github users/orgs.

### Authentication

Either through environmental variables or passed in through a token file (more secure).

* `GITHUB_TOKEN` If supplied, enables the user to supply a github authentication token that allows the API to be queried more often. Optional, but recommended.
* `GITHUB_TOKEN_FILE` If supplied _instead of_ `GITHUB_TOKEN`, enables the user to supply a path to a file containing a github authentication token that allows the API to be queried more often. Optional, but recommended.

### Optional Metrics

* `OPTIONAL_METRICS` allows you to specify additional collection of the following metrics per repository;

- Commits
- Releases
- Pulls

**Please be aware the above metrics can only be collected from the V3 API on a per repositry basis. As a result they are very expensive to capture and you may exceed the GitHub API rate limits if you're monitoring hundreds of repositories.**

In order to collect any of the above, populate the `OPTIONAL_METRICS` variable with a comma delimited set of metrics you wish to capture, e.g. `OPTIONAL_METRICS="commits, releases, pulls"`.

### Operating Configuration

Likely something most users will leave as the default, feel free to override as you see fit.

* `API_URL` Github API URL, shouldn't need to change this. Defaults to `https://api.github.com`
* `LISTEN_PORT` The port you wish to run the container on, the Dockerfile defaults this to `9171`
* `METRICS_PATH` the metrics URL path you wish to use, defaults to `/metrics`
* `LOG_LEVEL` The level of logging the exporter will run with, defaults to `debug`


## Install and deploy

Run manually from Docker Hub:
```
docker run -d --restart=always -p 9171:9171 -e REPOS="infinityworks/policies, infinityworks/prom-conf" infinityworks/github-exporter
```

Build a docker image:
```
docker build -t <image-name> .
docker run -d --restart=always -p 9171:9171 -e REPOS="infinityworks/policies, infinityworks/prom-conf" <image-name>
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

Metrics will be made available on port 9171 by default. An example of these metrics can be found in the `METRICS.md` markdown file in the root of this repository

## Tests

There is a set of blackbox behavioural tests which validate metrics endpoint in the `test` directory. 
Run as follows

```bash
make test
```

## Version Release Procedure
Once a new pull request has been merged into `master` the following script should be executed locally. The script will trigger a new image build in docker hub with the new image having the tag `release-<version>`. The version is taken from the `VERSION` file and must follow semantic versioning. For more information see [semver.org](https://semver.org/).

Prior to running the following command ensure the number has been increased to desired version in `VERSION`: 

```bash
./release-version.sh
```

## Metadata
[![](https://images.microbadger.com/badges/image/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own version badge on microbadger.com")
