ARG GOLANG_VERSION=1.25.2
FROM golang:${GOLANG_VERSION} AS build
LABEL maintainer="githubexporter"

COPY ./ /go/src/github.com/githubexporter/github-exporter
WORKDIR /go/src/github.com/githubexporter/github-exporter

RUN go mod download \
    && go test ./... \
    && CGO_ENABLED=0 GOOS=linux go build -o /bin/main

FROM gcr.io/distroless/static AS runtime

ADD VERSION .
COPY --from=build /bin/main /bin/main
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT [ "/bin/main" ]
