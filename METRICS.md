# Metrics

Below are an example of the metrics as exposed by this exporter. 

```
# HELP github_repo_open_issues open_issues
# TYPE github_repo_open_issues gauge
github_repo_open_issues{repo="docker-hub-exporter",user="infinityworksltd"} 1.0
github_repo_open_issues{repo="prometheus-rancher-exporter",user="infinityworksltd"} 2.0
# HELP github_repo_watchers watchers
# TYPE github_repo_watchers gauge
github_repo_watchers{repo="docker-hub-exporter",user="infinityworksltd"} 1.0
github_repo_watchers{repo="prometheus-rancher-exporter",user="infinityworksltd"} 6.0
# HELP github_repo_stars stars
# TYPE github_repo_stars gauge
github_repo_stars{repo="docker-hub-exporter",user="infinityworksltd"} 1.0
github_repo_stars{repo="prometheus-rancher-exporter",user="infinityworksltd"} 6.0
# HELP github_repo_forks forks
# TYPE github_repo_forks gauge
github_repo_forks{repo="docker-hub-exporter",user="infinityworksltd"} 0.0
github_repo_forks{repo="prometheus-rancher-exporter",user="infinityworksltd"} 9.0
# HELP github_repo_size_kb Size in KB for given repository
# TYPE github_repo_size_kb gauge
github_repo_size_kb{repo="docker-hub-exporter",user="infinityworksltd"} 44
github_repo_size_kb{repo="prometheus-rancher-exporter",user="infinityworksltd"} 7242
# HELP github_rate_limit Number of API queries allowed in a 60 minute window
# TYPE github_rate_limit gauge
github_rate_limit 60
# HELP github_rate_remaining Number of API queries remaining in the current window
# TYPE github_rate_remaining gauge
github_rate_remaining 38
# HELP github_rate_reset The time at which the current rate limit window resets in UTC epoch seconds
# TYPE github_rate_reset gauge
github_rate_reset 1.493139756e+09
# HELP github_size_kb Size in KB for given repository
# TYPE github_size_kb gauge
github_size_kb{repo="CRUST",user="infinityworksltd"} 44
```