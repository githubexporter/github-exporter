FROM golang:1.11 as build
LABEL maintainer "Infinity Works"

COPY ./ /infinityworks/github-exporter
WORKDIR /infinityworks/github-exporter

RUN go test ./... \
     && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/main .

FROM alpine:3.6

RUN apk --no-cache add ca-certificates \
     && addgroup exporter \
     && adduser -S -G exporter exporter
USER exporter
COPY --from=build /bin/main /bin/main
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT [ "/bin/main" ]
