FROM golang:1.22-bookworm as build
LABEL maintainer="githubexporter"

COPY ./ /go/src/github.com/githubexporter/github-exporter
WORKDIR /go/src/github.com/githubexporter/github-exporter

RUN go mod download \
    && go test ./... \
    && cd cmd/github-exporter \
    && CGO_ENABLED=0 GOOS=linux go build -o /bin/github-exporter

FROM alpine:3

RUN apk --no-cache add ca-certificates \
     && addgroup exporter \
     && adduser -S -G exporter exporter
ADD VERSION .
USER exporter
COPY --from=build /bin/github-exporter /bin/github-exporter
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT ["/bin/github-exporter"]