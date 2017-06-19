FROM golang:1.8.0-alpine as build
LABEL maintainer "Infinity Works"

RUN apk --update add ca-certificates \
     && apk --update add --virtual build-deps git
COPY ./ /go/src/github.com/infinityworksltd/github-exporter
WORKDIR /go/src/github.com/infinityworksltd/github-exporter
RUN go get \
 && go test ./... \
 && go build -o /bin/main

FROM alpine:3.6

RUN apk --update add ca-certificates \
     && addgroup exporter \
     && adduser -S -G exporter exporter
USER exporter
COPY --from=build /bin/main /bin/main
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT [ "/bin/main" ]
