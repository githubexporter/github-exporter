FROM python:3.5-alpine

RUN pip install prometheus_client requests

ENV BIND_PORT 9171
ENV IMAGES "prom/prometheus, prom/node-exporter"
ENV INTERVAL 5

ADD . /usr/src/app
WORKDIR /usr/src/app

CMD ["python", "github_exporter.py"]
