FROM golang:1.17-alpine AS build

WORKDIR /src/
COPY *.go go.* VERSION /src/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=v$(cat VERSION)'" -o /app/server

FROM alpine:latest

USER root
COPY --from=build /app/* /app/
COPY ./docker.entrypoint.sh /entrypoint.sh

ENV LOG_LEVEL="ERROR" \
    AUTOTLS_DOMAINS="" \
    AUTOTLS_EMAIL=""

RUN apk --no-cache add ca-certificates curl \
    && chmod +x /app/server && chmod +x /entrypoint.sh \
    && mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

CMD ["/entrypoint.sh", "start-server"]