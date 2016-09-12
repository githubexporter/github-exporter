from prometheus_client import start_http_server
from prometheus_client.core import CounterMetricFamily, GaugeMetricFamily, REGISTRY

import json
import requests
import sys
import time
import os


class GitHubCollector(object):

  def collect(self):
      repos = os.getenv('REPOS', default="infinityworksltd/docker-hub-exporter, infinityworksltd/prometheus-rancher-exporter").replace(' ','').split(",")
      print("Starting exporter")
      self._fork_metrics = GaugeMetricFamily('github_forks', 'Gauge of forks from the public API', labels=["repo", "user"])
      self._star_metrics = GaugeMetricFamily('github_stars', 'Gauge of stars from the public API', labels=["repo", "user"])
      self._open_issues_metrics = GaugeMetricFamily('github_open_issues', 'Gauge of issues from the public API', labels=["repo", "user"])
      self._watchers_metrics = GaugeMetricFamily('github_watchers', 'Gauge of watchers from the public API', labels=["repo", "user"])


      for repo in repos:
          print("Getting JSON for " + repo)
          self._get_json(repo)
          print("Getting Metrics for " + repo)
          self._get_metrics()
          print ("Metrics Updated for " + repo)

      yield self._fork_metrics
      yield self._star_metrics
      yield self._open_issues_metrics
      yield self._watchers_metrics

  def _get_json(self, repo):
      print("Getting JSON Payload for " + repo)

      repo_url = 'https://api.github.com/repos/{0}'.format(repo)
      print(repo_url)
      response = requests.get(repo_url)
      self._response_json = json.loads(response.content.decode('UTF-8'))


  def _get_metrics(self):
      repo_name = self._response_json['name']
      user_name = self._response_json['owner']['login']
      self._fork_metrics.add_metric([repo_name, user_name], value=self._response_json['forks'])
      self._star_metrics.add_metric([repo_name, user_name], value=self._response_json['stargazers_count'])
      self._open_issues_metrics.add_metric([repo_name, user_name], value=self._response_json['open_issues'])
      self._watchers_metrics.add_metric([repo_name, user_name], value=self._response_json['watchers'])


if __name__ == '__main__':
  start_http_server(int(os.getenv('BIND_PORT')))
  REGISTRY.register(GitHubCollector())

  while True: time.sleep(1)
