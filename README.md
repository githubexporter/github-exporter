# Prometheus GitHub Exporter

Exposes basic metrics for your repositories from the GitHub API, to a Prometheus compatible endpoint.

## Configuration

This exporter is configured via environment variables. All variables are optional unless otherwise stated. Below is a list of supported configuration values:

| Variable                     | Description                                                                        | Default                  |
|------------------------------|------------------------------------------------------------------------------------|--------------------------|
| `ORGS`                       | Comma-separated list of GitHub organizations to monitor (e.g. `org1,org2`).        |                          |
| `REPOS`                      | Comma-separated list of repositories to monitor (e.g. `user/repo1,user/repo2`).    |                          |
| `USERS`                      | Comma-separated list of GitHub users to monitor (e.g. `user1,user2`).              |                          |
| `GITHUB_TOKEN`               | GitHub personal access token for API authentication.                               |                          |
| `GITHUB_TOKEN_FILE`          | Path to a file containing a GitHub personal access token.                          |                          |
| `GITHUB_APP`                 | Set to `true` to authenticate as a GitHub App.                                     | `false`                  |
| `GITHUB_APP_ID`              | The App ID of the GitHub App. Required if `GITHUB_APP` is `true`.                  |                          |
| `GITHUB_APP_INSTALLATION_ID` | The Installation ID of the GitHub App. Required if `GITHUB_APP` is `true`.         |                          |
| `GITHUB_APP_KEY_PATH`        | Path to the GitHub App private key file. Required if `GITHUB_APP` is `true`.       |                          |
| `GITHUB_RATE_LIMIT_ENABLED`  | Whether to fetch GitHub API rate limit metrics (`true` or `false`).                | `true`                   |
| `GITHUB_RESULTS_PER_PAGE`    | Number of results to request per page from the GitHub API (max 100).               | `100`                    |
| `API_URL`                    | GitHub API URL. You should not need to change this unless using GitHub Enterprise. | `https://api.github.com` |
| `LISTEN_PORT`                | The port the exporter will listen on.                                              | `9171`                   |
| `METRICS_PATH`               | The HTTP path to expose Prometheus metrics.                                        | `/metrics`               |
| `LOG_LEVEL`                  | Logging level (`debug`, `info`, `warn`, `error`).                                  | `info`                   |

### Credential Precedence

When authenticating with the GitHub API, the exporter uses credentials in the following order of precedence:

1. **GitHub App credentials** (`GITHUB_APP=true` with `GITHUB_APP_ID`, `GITHUB_APP_INSTALLATION_ID`, and `GITHUB_APP_KEY_PATH`): If enabled, the exporter authenticates as a GitHub App and ignores any personal access token or token file.
2. **Token file** (`GITHUB_TOKEN_FILE`): If a token file is provided (and GitHub App is not enabled), the exporter reads the token from the specified file.
3. **Direct token** (`GITHUB_TOKEN`): If neither GitHub App nor token file is provided, the exporter uses the token supplied directly via the environment variable.

If none of these credentials are provided, the exporter will make unauthenticated requests, which are subject to very strict rate limits.

## Install and deploy

Run manually from Docker Hub:
```
docker run -d --restart=always -p 9171:9171 -e REPOS="infinityworks/ranch-eye, infinityworks/prom-conf" githubexporter/github-exporter
```

Run manually from Docker Hub (With GitHub App):
```
docker run -d --restart=always -p 9171:9171 --read-only -v ./key.pem:/key.pem -e GITHUB_APP=true -e GITHUB_APP_ID= -e GITHUB_APP_INSTALLATION_ID= -e GITHUB_APP_KEY_PATH=/key.pem <IMAGE_NAME>
```

Build a docker image:
```
docker build -t <image-name> .
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
    image: githubexporter/github-exporter:latest
    environment:
      - REPOS=<REPOS you want to monitor>
      - GITHUB_TOKEN=<your github api token>
```

## Docker compose (GitHub App)

```
github-exporter-github-app:
  tty: true
  stdin_open: true
  expose:
    - 9171
  ports:
    - 9171:9171
  build: .
  environment:
    - LOG_LEVEL=debug
    - LISTEN_PORT=9171
    - GITHUB_APP=true
    - GITHUB_APP_ID=
    - GITHUB_APP_INSTALLATION_ID=
    - GITHUB_APP_KEY_PATH=/key.pem
  restart: unless-stopped
  volumes:
    - "./key.pem:/key.pem:ro"

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

## Metadata
[![](https://images.microbadger.com/badges/image/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own version badge on microbadger.com")
