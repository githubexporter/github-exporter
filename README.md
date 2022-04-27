# Prometheus Jira Exporter

Exposes basic metrics for your projects from the Jira API, to a Prometheus compatible endpoint.

## Configuration

This exporter is setup to take input from environment variables. All variables are optional:

* `ORGS` If supplied, the exporter will enumerate all repositories for that organization. Expected in the format "org1, org2".
* `REPOS` If supplied, The repos you wish to monitor, expected in the format "user/repo1, user/repo2". Can be across different Jira users/orgs.
* `USERS` If supplied, the exporter will enumerate all repositories for that users. Expected in
the format "user1, user2".
* `JIRA_TOKEN` If supplied, enables the user to supply a jira authentication token that allows the API to be queried more often. Optional, but recommended.
* `JIRA_TOKEN_FILE` If supplied _instead of_ `JIRA_TOKEN`, enables the user to supply a path to a file containing a jira authentication token that allows the API to be queried more often. Optional, but recommended.
* `API_URL` Jira API URL. It should be https://<your-org>.atlassian.net/rest/api/3`
* `LISTEN_PORT` The port you wish to run the container on, the Dockerfile defaults this to `9171`
* `METRICS_PATH` the metrics URL path you wish to use, defaults to `/metrics`
* `LOG_LEVEL` The level of logging the exporter will run with, defaults to `debug`


## Install and deploy

Run manually from Docker Hub:
```
docker run -d --restart=always -p 9171:9171 jira-exporter
```

Build a docker image:
```
docker build -t <image-name> .
docker run -d --restart=always -p 9171:9171 <image-name>
```

## Docker compose

```
jira-exporter:
    tty: true
    stdin_open: true
    expose:
      - 9171
    ports:
      - 9171:9171
    image: jira-exporter:latest
    environment:
      - JIRA_TOKEN=<your jira api token>

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

## Version Release Procedure
Once a new pull request has been merged into `master` the following script should be executed locally. The script will trigger a new image build in docker hub with the new image having the tag `release-<version>`. The version is taken from the `VERSION` file and must follow semantic versioning. For more information see [semver.org](https://semver.org/).

Prior to running the following command ensure the number has been increased to desired version in `VERSION`: 

```bash
./release-version.sh
```
