FROM golang:1.19-buster AS build

WORKDIR /go/src/github.com/galbirk/github-exporter

COPY . .

RUN go mod download

RUN go build -o /main

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /main /main 

USER nonroot:nonroot

ENTRYPOINT ["/main"]