FROM alpine:latest
WORKDIR /home
COPY enphaseLocalToInflux.linux.amd64 enphaselocal2influx.linux.amd64
COPY config.yaml config.yaml
CMD ./enphaselocal2influx.linux.amd64