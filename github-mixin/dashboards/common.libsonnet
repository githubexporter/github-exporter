local grafana = import 'github.com/grafana/grafonnet-lib/grafonnet/grafana.libsonnet';

{
  latestSingleStatPanel(title, format='none')::
    grafana.statPanel.new(title, reducerFunction='last', graphMode='none') +
    {
      fieldConfig: {
        defaults: {
          thresholds: {
            mode: 'absolute',
            steps: [],
          },
          unit: format,
        },
      },
    },
}
