FROM quay.io/prometheus/busybox:latest

ADD simple-api /bin/simple-api

ENTRYPOINT ["/bin/simple-api"]
