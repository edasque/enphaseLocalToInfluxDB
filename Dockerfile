FROM alpine:latest
WORKDIR /home
COPY enphaseLocalToInflux.linux.amd64 enphaselocal2influx.linux.amd64
COPY config.yaml config.yaml
CMD while true; do ./enphaselocal2influx.linux.amd64; sleep 300; done