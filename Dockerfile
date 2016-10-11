FROM python:3.5-alpine

RUN pip install prometheus_client requests

ENV BIND_PORT 9171

ADD . /usr/src/app
WORKDIR /usr/src/app

CMD ["python", "github_exporter.py"]
