from prometheus_client import start_http_server
from prometheus_client.core import CounterMetricFamily, GaugeMetricFamily, REGISTRY

import json, requests, sys, time, os, ast, signal

class GitHubCollector(object):

  def collect(self):

    metrics = {'forks': 'forks',
               'stars': 'stargazers_count',
               'open_issues': 'open_issues',
               'watchers': 'watchers_count',
               'has_issues': 'has_issues',}

    METRIC_PREFIX = 'github_repo'
    LABELS = ['repo', 'user', 'private']
    gauges = {}

    # Setup metric counters from prometheus_client.core
    for metric in metrics:
      gauges[metric] = GaugeMetricFamily('%s_%s' % (METRIC_PREFIX, metric), '%s' % metric, value=None, labels=LABELS)

    # Check the API rate limit
    self._check_api_limit()

    # loop through specified repositories and organizations and collect metrics
    if os.getenv('REPOS'):
      repos = os.getenv('REPOS').replace(' ','').split(",")
      self._repo_urls = []
      for repo in repos:
        print(repo + " added to collection array")
        self._repo_urls.extend('https://api.github.com/repos/{0}'.format(repo).split(","))
      self._collect_repo_metrics(gauges, metrics)
      print("Metrics collected for individually specified repositories")

    if os.getenv('ORGS'):
      orgs = os.getenv('ORGS').replace(' ','').split(",")
      self._org_urls = []
      for org in orgs:
        print(org + " added to collection array")
        self._org_urls.extend('https://api.github.com/orgs/{0}/repos'.format(org).split(","))
      self._collect_org_metrics(gauges, metrics)
      print("Metrics collected for repositories listed under specified organization")

    # Yield all metrics returned
    for metric in metrics:
      yield gauges[metric]

  def _collect_repo_metrics(self, gauges, metrics):
    for url in self._repo_urls:
      response_json = self._get_json(url)
      self._add_metrics(gauges, metrics, response_json)

  def _collect_org_metrics(self, gauges, metrics):
    for url in self._org_urls:
      response_json = self._get_json(url)
      for repo in response_json:
        self._add_metrics(gauges, metrics, repo)

  def _get_github_token(self):
    if os.getenv('GITHUB_TOKEN'):
      return os.getenv('GITHUB_TOKEN')
    elif os.getenv('GITHUB_TOKEN_FILE'):
      return open(os.getenv('GITHUB_TOKEN_FILE'), 'r').read().rstrip()
    else:
      return None

  def _get_json(self, url: str):
    """
    using github core api
    rate limit 5000 per hours
    """
    print("Getting JSON Payload for " + url)
    gh_token = self._get_github_token()
    if gh_token:
      payload = {"access_token": gh_token}
      r = requests.get(url,params=payload)
    else:
      r = requests.get(url)
    result = json.loads(r.content.decode('UTF-8'))
    if result is None:
      raise ValueError("Github API is broken, try again")
    return self._pagination(r, result)

  def _pagination(self, response: requests.Response, result):
    if "Link" not in response.headers:
      return result
    links = dict()
    for i in response.headers["Link"].split(","):
      url, rel = i.split(";")
      rel = rel[6:-1]
      url = url[1:-1]
      links[rel] = url
      if "next" in links:
        assert type(result) is list
        return result + self._get_json(links["next"])
      else:
        return result

  def _check_api_limit(self):
    rate_limit_url = "https://api.github.com/rate_limit"
    gh_token = self._get_github_token()
    if gh_token:
      print("Authentication token detected: " + gh_token)
      payload = {"access_token": gh_token}
      R = requests.get(rate_limit_url,params=payload)
    else:
      R = requests.get(rate_limit_url)

    limit_js = ast.literal_eval(R.text)
    remaining = limit_js["rate"]["remaining"]
    print("Requests remaing this hour", remaining)
    if not remaining:
      print("Rate limit exceeded, try enabling authentication")

  def _add_metrics(self, gauges, metrics, response_json):
    for metric, field in metrics.items():
      gauges[metric].add_metric([response_json['name'], response_json['owner']['login'], str(response_json['private']).lower()], value=response_json[field])

def sigterm_handler(_signo, _stack_frame):
  sys.exit(0)

if __name__ == '__main__':
  # Ensure we have something to export
  print("Starting Exporter")
  if not (os.getenv('REPOS') or os.getenv('ORGS')):
    print("No repositories or organizations specified, exiting")
    exit(1)
  start_http_server(int(os.getenv('BIND_PORT')))
  REGISTRY.register(GitHubCollector())

  signal.signal(signal.SIGTERM, sigterm_handler)
  while True: time.sleep(1)
