FROM alpine:latest
WORKDIR /home
COPY enphaseLocalToInflux.linux.amd64 enphaseLocalToInflux.linux.amd64
COPY config.yaml config.yaml
CMD while true; do ./enphaseLocalToInflux.linux.amd64; sleep 60; done