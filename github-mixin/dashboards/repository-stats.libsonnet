local common = import 'common.libsonnet';
local grafana = import 'github.com/grafana/grafonnet-lib/grafonnet/grafana.libsonnet';

local dashboardWidth = 24;

local metric(metric_name, title, format='none') = {
  name: metric_name,
  title: title,
  format: format,
};

local latestRepoStatPanel(metric) =
  common.latestSingleStatPanel(metric.title, metric.format)
  .addTarget(grafana.prometheus.target(metric.name + '{user=~"$user",repo=~"$repo"}'));

local graphPanel(metric) =
  grafana.graphPanel.new(
    metric.title,
    min=0,
    legend_show=false,
    format=metric.format
  )
  .addTarget(grafana.prometheus.target(metric.name + '{user=~"$user",repo=~"$repo"}'));

local setGridPos(panels, startY, panelWidth, panelHeight) =
  if panelWidth > dashboardWidth then
    error 'panelWidth cannot be larger than dashboardWidth'
  else
    local calculate(index) = {
      gridPos: {
        x: (index % std.floor(dashboardWidth / panelWidth)) * panelWidth,
        y: startY + std.floor((index * panelWidth) / dashboardWidth),
        w: panelWidth,
        h: panelHeight,
      },
    };

    std.mapWithIndex(function(index, panel) panel + calculate(index), panels);

local maxY(panels) = std.foldl(std.max, [p.gridPos.y + p.gridPos.h for p in panels], 0);

local repoPanels(metrics) =
  local statPanels = std.map(latestRepoStatPanel, metrics);
  local statPanelsWithGridPos = setGridPos(statPanels, 0, 4, 4);

  local statPanelsMaxY = maxY(statPanelsWithGridPos);

  local graphRowPanel = { title: 'Graphs', type: 'row' };
  local graphRowPanelWithGridPos = setGridPos([graphRowPanel], statPanelsMaxY, dashboardWidth, 1);

  local graphPanels = std.map(graphPanel, metrics);
  local graphPanelsWithGridPos = setGridPos(graphPanels, statPanelsMaxY + 1, 8, 8);

  std.flattenArrays([statPanelsWithGridPos, graphRowPanelWithGridPos, graphPanelsWithGridPos]);

grafana.dashboard.new('GitHub Repository Stats', uid='github-repo-stats', editable=true)
.addTemplate(
  grafana.template.datasource(
    'datasource',
    'prometheus',
    'Prometheus'
  )
)
.addTemplate(
  grafana.template.new(
    'user',
    '$datasource',
    'label_values(user)',
    refresh='load'
  )
)
.addTemplate(
  grafana.template.new(
    'repo',
    '$datasource',
    'label_values(github_repo_open_issues{user="$user"}, repo)',
    refresh='load'
  )
)
.addPanels(
  repoPanels(
    [
      metric('github_repo_open_issues', 'Open Issues'),
      metric('github_repo_pull_request_count', 'Open Pull Requests'),
      metric('github_repo_forks', 'Forks'),
      metric('github_repo_stars', 'Stars'),
      metric('github_repo_watchers', 'Watchers'),
      metric('github_repo_size_kb', 'Repository Size', format='deckbytes'),
    ]
  )
)
