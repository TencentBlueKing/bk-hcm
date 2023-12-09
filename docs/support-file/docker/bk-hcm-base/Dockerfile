FROM debian:bullseye-slim

#time zone setting
ENV TimeZone=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TimeZone /etc/localtime \
    && echo $TimeZone >/etc/timezone\
    && apt update \
    && apt install -y -q --no-install-recommends \
    ca-certificates iputils-ping telnet procps curl less \
    && apt-get clean \
    && rm -r /var/lib/apt/lists/* \
    && mkdir -p /data/hcm/etc/certs \
    && mkdir -p /data/hcm/logs

WORKDIR /data/hcm