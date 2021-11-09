FROM alpine:latest

USER root

ENV APP_HOME=/app \
    PATH=$APP_HOME:$PATH \
    LOG_LEVEL=ERROR

COPY http-proxy-replica /app/http-proxy-replica

RUN chmod +x /app/http-proxy-replica  \
    && mkdir /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE [80, 443]
CMD ["/app/http-proxy-replica"]