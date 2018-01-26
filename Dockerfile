FROM alpine:3.7

ENV LISTEN 0.0.0.0:9382
EXPOSE 9382

RUN apk add -U ca-certificates

COPY ./gcp-exporter /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/gcp-exporter", "service"]
