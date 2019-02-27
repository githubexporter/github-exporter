FROM golang:1.12-alpine as build
LABEL maintainer="Infinity Works"

RUN apk --no-cache add ca-certificates build-base git

COPY ./ /github-exporter
WORKDIR /github-exporter

RUN go get \
 && go test ./... \
 && go build -o /bin/main

FROM alpine:3.9

RUN apk --no-cache add ca-certificates \
     && addgroup exporter \
     && adduser -S -G exporter exporter
USER exporter
COPY --from=build /bin/main /bin/main
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT [ "/bin/main" ]
