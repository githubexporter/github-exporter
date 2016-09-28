from prometheus_client import start_http_server
from prometheus_client.core import CounterMetricFamily, GaugeMetricFamily, REGISTRY

import json, requests, sys, time, os, ast, signal

class GitHubCollector(object):

  def collect(self):
    repos = os.getenv('REPOS', default="infinityworksltd/docker-hub-exporter, infinityworksltd/prometheus-rancher-exporter").replace(' ','').split(",")
    metrics = {'forks': 'forks',
               'stars': 'stargazers_count',
               'open_issues': 'open_issues',
               'watchers': 'watchers_count',
               'has_issues': 'has_issues',
               'subscribers': 'subscribers_count'}

    METRIC_PREFIX = 'github_repo'
    LABELS = ['repo', 'user']
    gauges = {}

    for metric in metrics:
      gauges[metric] = GaugeMetricFamily('%s_%s' % (METRIC_PREFIX, metric), '%s' % metric, value=None, labels=LABELS)
        
    for repo in repos:
      self._get_json(repo)
      self._check_api_limit()
      self._get_metrics(gauges, metrics)

    for metric in metrics:
      yield gauges[metric]


  def _get_json(self, repo):
    print("Getting JSON Payload for " + repo)
    repo_url = 'https://api.github.com/repos/{0}'.format(repo)
    print(repo_url)
    response = requests.get(repo_url)
    self._response_json = json.loads(response.content.decode('UTF-8'))

  def _check_api_limit(self):
    rate_limit_url = "https://api.github.com/rate_limit"
    if os.getenv('GITHUB_TOKEN'):
      print("Authentication token supplied")
      payload = {"access_token":os.environ["GITHUB_TOKEN"],}
      R = requests.get(rate_limit_url,params=payload)
    else:
      R = requests.get(rate_limit_url)
    limit_js = ast.literal_eval(R.text)
    remaining = limit_js["rate"]["remaining"]
    print("Requests remaing this hour", remaining)

    if not remaining:
      print("Rate limit exceeded, sleeping for 60 seconds")
      time.sleep(60)

  def _get_metrics(self, gauges, metrics):
    repo_name = self._response_json['name']
    user_name = self._response_json['owner']['login']

    for metric, field in metrics.items():
      print("Metric %s being set from %s from the API" % (metric, field))
      gauges[metric].add_metric([repo_name, user_name], value=self._response_json[field])


def sigterm_handler(_signo, _stack_frame):
  sys.exit(0)

if __name__ == '__main__':
  start_http_server(int(os.getenv('BIND_PORT')))
  REGISTRY.register(GitHubCollector())
  
  signal.signal(signal.SIGTERM, sigterm_handler)
  while True: time.sleep(int(2)
