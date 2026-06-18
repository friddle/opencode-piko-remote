ARG REGISTRY=docker.io
FROM ${REGISTRY}/openresty/openresty:1.25.3.2-5-alpine-fat

RUN apk add --no-cache curl bash

ENV PIKO_UPSTREAM_PORT=8022
ENV LISTEN_PORT=8080

COPY --from=ghcr.io/andydunstall/piko:latest /bin/piko /usr/local/bin/piko

WORKDIR /app

COPY build/start.sh /app/start.sh
COPY build/nginx.conf /etc/nginx/nginx.conf
COPY build/piko.conf /etc/nginx/conf.d/piko.conf

RUN mkdir -p /var/log/nginx && chmod +x /app/start.sh

EXPOSE 8088 8022 8023

CMD ["/app/start.sh"]
