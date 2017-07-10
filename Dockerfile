FROM golang:1.8.0-alpine
LABEL maintainer "Infinity Works"

EXPOSE 9171

ENV GOPATH=/go
ENV LISTEN_PORT=9171

RUN addgroup exporter \
     && adduser -S -G exporter exporter \
     && apk --update add ca-certificates \
     && apk --update add --virtual build-deps git

COPY ./ /go/src/github.com/infinityworks/github-exporter

WORKDIR /go/src/github.com/infinityworks/github-exporter

RUN go get \
 && go test ./... \
 && go build -o /bin/main

USER exporter

CMD [ "/bin/main" ]
