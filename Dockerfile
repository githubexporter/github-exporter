FROM golang:1.14.8-stretch as build
LABEL maintainer="Infinity Works"

ENV GO111MODULE=on

COPY ./ /go/src/github.com/infinityworks/github-exporter
WORKDIR /go/src/github.com/infinityworks/github-exporter

RUN go mod download \
    && go test ./... \
    && CGO_ENABLED=0 GOOS=linux go build -o /bin/main


RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/main


FROM alpine:3.11.3

RUN apk --no-cache add ca-certificates \
     && addgroup exporter \
     && adduser -S -G exporter exporter
ADD VERSION .
USER exporter
COPY --from=build /bin/main /bin/main
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT [ "/bin/main" ]
