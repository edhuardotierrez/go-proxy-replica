FROM alpine:latest

USER root

ENV APP_HOME=/app \
    PATH=$APP_HOME:$PATH \
    LOG_LEVEL=ERROR

COPY go-proxy-replica /app/go-proxy-replica

RUN chmod +x /app/go-proxy-replica  \
    && mkdir /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE [80, 443]
CMD ["/app/go-proxy-replica"]