# Prometheus GitHub Exporter

Exposes basic metrics for your repositories from the GitHub API, to a Prometheus compatible endpoint.

## Configuration

This exporter is configured via environment variables. All variables are optional unless otherwise stated. Below is a list of supported configuration values:

| Variable                      | Description                                                                        | Default                  |
|-------------------------------|------------------------------------------------------------------------------------|--------------------------|
| `ORGS`                        | Comma-separated list of GitHub organizations to monitor (e.g. `org1,org2`).        |                          |
| `REPOS`                       | Comma-separated list of repositories to monitor (e.g. `user/repo1,user/repo2`).    |                          |
| `USERS`                       | Comma-separated list of GitHub users to monitor (e.g. `user1,user2`).              |                          |
| `GITHUB_TOKEN`                | GitHub personal access token for API authentication.                               |                          |
| `GITHUB_TOKEN_FILE`           | Path to a file containing a GitHub personal access token.                          |                          |
| `GITHUB_APP`                  | Set to `true` to authenticate as a GitHub App.                                     | `false`                  |
| `GITHUB_APP_ID`               | The App ID of the GitHub App. Required if `GITHUB_APP` is `true`.                  |                          |
| `GITHUB_APP_INSTALLATION_ID`  | The Installation ID of the GitHub App. Required if `GITHUB_APP` is `true`.         |                          |
| `GITHUB_APP_KEY_PATH`         | Path to the GitHub App private key file. Required if `GITHUB_APP` is `true`.       |                          |
| `GITHUB_RATE_LIMIT_ENABLED`   | Whether to fetch GitHub API rate limit metrics (`true` or `false`).                | `true`                   |
| `GITHUB_RATE_LIMIT`           | Core API quota threshold for proactive GitHub App token refresh. When the remaining `core` requests drop below this value, a new installation token is requested. `0` disables this behaviour. | `0` |
| `GITHUB_RESULTS_PER_PAGE`     | Number of results to request per page from the GitHub API (max 100).               | `100`                    |
| `FETCH_REPO_RELEASES_ENABLED` | Whether to fetch repository release metrics (`true` or `false`).                   | `true`                   |
| `FETCH_ORGS_CONCURRENCY`      | Number of concurrent requests to make when fetching organization data.             | `1`                      |
| `FETCH_ORG_REPOS_CONCURRENCY` | Number of concurrent requests to make when fetching organization repository data.  | `1`                      |
| `FETCH_USERS_CONCURRENCY`     | Number of concurrent requests to make when fetching user data.                     | `1`                      |
| `FETCH_REPOS_CONCURRENCY`     | Number of concurrent requests to make when fetching repository data.               | `1`                      |
| `API_URL`                     | GitHub API URL. You should not need to change this unless using GitHub Enterprise. | `https://api.github.com` |
| `LISTEN_PORT`                 | The port the exporter will listen on.                                              | `9171`                   |
| `METRICS_PATH`                | The HTTP path to expose Prometheus metrics.                                        | `/metrics`               |
| `LOG_LEVEL`                   | Logging level (`debug`, `info`, `warn`, `error`).                                  | `info`                   |

### Credential Precedence

When authenticating with the GitHub API, the exporter uses credentials in the following order of precedence:

1. **GitHub App credentials** (`GITHUB_APP=true` with `GITHUB_APP_ID`, `GITHUB_APP_INSTALLATION_ID`, and `GITHUB_APP_KEY_PATH`): If enabled, the exporter authenticates as a GitHub App and ignores any personal access token or token file.
2. **Direct token** (`GITHUB_TOKEN`): If a token is supplied directly via the environment variable (and GitHub App is not enabled), the exporter uses it and ignores any token file.
3. **Token file** (`GITHUB_TOKEN_FILE`): If neither GitHub App nor a direct token is provided, the exporter reads the token from the specified file.

If none of these credentials are provided, the exporter will make unauthenticated requests, which are subject to very strict rate limits.

## Permissions

The exporter only makes read requests to the GitHub API. The access it needs depends on whether you are monitoring private repositories and which features you have enabled.

### By scenario

| Scenario                                  | Unauthenticated | Personal access token (classic) | Fine-grained personal access token                                                      | GitHub App                                                       |
|-------------------------------------------|-----------------|---------------------------------|-----------------------------------------------------------------------------------------|------------------------------------------------------------------|
| Public repositories only                  | Yes (see note)  | No scopes                       | No permissions (public repositories are always readable)                                | No extra permissions                                             |
| Private repositories                      | No              | `repo`                          | Metadata, Pull requests and Contents: read, on the selected repositories                | Metadata, Pull requests and Contents: read, installed on them    |
| Private repositories, releases disabled   | No              | `repo`                          | Metadata and Pull requests: read, on the selected repositories                          | Metadata and Pull requests: read, installed on them              |

> [!NOTE]
> Unauthenticated requests are limited to 60 per hour, and each scrape makes at least two requests per repository, so a token is recommended even for public repositories.

Classic tokens have no read-only scope for private repositories, so `repo` also grants write access. Prefer a fine-grained token or a GitHub App where possible.

### By feature

Permissions below are for fine-grained tokens and GitHub Apps, and are only needed for private repositories.

| Feature                                  | Permission            | Endpoint                             | Metrics                                                                                   |
|------------------------------------------|-----------------------|--------------------------------------|-------------------------------------------------------------------------------------------|
| `REPOS`                                  | Metadata: read        | `GET /repos/{owner}/{repo}`          | `github_repo_stars`, `github_repo_forks`, `github_repo_watchers`, `github_repo_size_kb`   |
| `ORGS`                                   | Metadata: read        | `GET /orgs/{org}/repos`, then `GET /repos/{owner}/{repo}` for each repository | As above                                         |
| `USERS`                                  | Metadata: read        | `GET /users/{username}/repos`, then `GET /repos/{owner}/{repo}` for each repository | As above                                   |
| Pull requests (always fetched)           | Pull requests: read   | `GET /repos/{owner}/{repo}/pulls`    | `github_repo_pull_request_count`, `github_repo_open_issues`                               |
| `FETCH_REPO_RELEASES_ENABLED` (default)  | Contents: read        | `GET /repos/{owner}/{repo}/releases` | `github_repo_releases`, `github_repo_release_downloads`                                   |
| `GITHUB_RATE_LIMIT_ENABLED` (default)    | None                  | `GET /rate_limit`                    | `github_rate_limit`, `github_rate_remaining`, `github_rate_reset`                         |
| `GITHUB_RATE_LIMIT`                      | None                  | None beyond the rate limit request   | None. Refreshes the GitHub App token using its private key                                |

The Metadata permission is included automatically in fine-grained tokens and GitHub Apps. No organization or account permissions are needed.

### Limitations

* **A missing permission fails the whole scrape.** If any request fails, the exporter logs the error and returns no repository metrics for that scrape, not just those for the affected repository. For example, a token without Contents access on one private repository loses metrics for all repositories unless `FETCH_REPO_RELEASES_ENABLED=false` is set.
* **`USERS` only finds public repositories**, whatever the credentials. List private repositories owned by a user in `REPOS` instead.
* **Fine-grained tokens and GitHub Apps cover private repositories for one account.** A fine-grained token has a single resource owner, and `GITHUB_APP_INSTALLATION_ID` selects a single installation. To monitor private repositories across several accounts, use a classic token or run one exporter per account.
* **`ORGS` only returns the private repositories the credentials can access.** Others are silently left out rather than causing an error.
* **Draft releases are only visible with push access.** `github_repo_releases` can therefore be higher with a classic `repo` token from a user with push access than with read-only access.

### Organization policies

* Organizations using SAML single sign-on require classic tokens to be [authorized for SSO](https://docs.github.com/en/authentication/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on).
* Organizations can block fine-grained tokens or require approval for them. Until approved, a token can only read public resources.

### Troubleshooting

API responses for fine-grained tokens and GitHub Apps include an `X-Accepted-GitHub-Permissions` header listing the permissions an endpoint requires. See GitHub's [fine-grained token](https://docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens) and [GitHub App](https://docs.github.com/en/rest/authentication/permissions-required-for-github-apps) permission references for more detail.

## Install and deploy

### Helm Chart

> [!NOTE]
> This chart is maintained by a third party and is not officially supported or endorsed by this project. Review it before use, and report chart issues to its maintainers.

* https://github.com/k8sonlab/publiccharts/tree/main/charts/github-exporter

### Docker

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

### Docker Compose

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

### Docker Compose (GitHub App)

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
Once a new pull request has been merged into `master` the following script should be executed locally. The script will trigger a new image build in docker hub with the new image having the tag `v<semantic-version>`. The version is taken from the `VERSION` file and must follow semantic versioning. For more information see [semver.org](https://semver.org/).

Prior to running the following command ensure the number has been increased to desired version in `VERSION`:

```bash
./release-version.sh
```

## Metadata
[![](https://images.microbadger.com/badges/image/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own image badge on microbadger.com") [![](https://images.microbadger.com/badges/version/infinityworks/github-exporter.svg)](http://microbadger.com/images/infinityworks/github-exporter "Get your own version badge on microbadger.com")
