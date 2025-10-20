ARG GOLANG_VERSION=1.25.3
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS build
LABEL maintainer="githubexporter"

ARG TARGETOS
ARG TARGETARCH

COPY ./ /go/src/github.com/githubexporter/github-exporter
WORKDIR /go/src/github.com/githubexporter/github-exporter

RUN go mod download \
    && go test ./... \
    && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/main

FROM gcr.io/distroless/static AS runtime

ADD VERSION .
COPY --from=build /bin/main /bin/main
ENV LISTEN_PORT=9171
EXPOSE $LISTEN_PORT
ENTRYPOINT [ "/bin/main" ]
