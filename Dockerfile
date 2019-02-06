FROM alpine:3.6
LABEL maintainer "Infinity Works"

RUN apk --no-cache add ca-certificates \
     && addgroup exporter \
     && adduser -S -G exporter exporter
USER exporter
COPY ./github-exporter /
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT [ "/github-exporter" ]
