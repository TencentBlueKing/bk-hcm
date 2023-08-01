FROM debian:bullseye-slim

RUN apt-get update \
    && apt-get install -y -q --no-install-recommends \
    ca-certificates \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/*
RUN mkdir -p /data/hcm/etc/certs
RUN mkdir -p /data/hcm/logs

COPY bk-hcm-cloudserver /data/hcm/

#time zone setting
ENV TimeZone=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TimeZone /etc/localtime && echo $TimeZone > /etc/timezone

CMD ["/data/hcm/bk-hcm-cloudserver", "--config-file", "/data/hcm/etc/config.yaml"]
