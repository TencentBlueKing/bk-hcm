FROM mirrors.tencent.com/tencentos/tencentos4-minimal

#time zone setting
ENV TimeZone=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TimeZone /etc/localtime \
    && echo $TimeZone >/etc/timezone\
    && yum update -y \
    && yum install -y --setopt=install_weak_deps=False \
    iputils telnet procps-ng iproute \
    && yum clean all \
    && mkdir -p /data/hcm/etc/certs \
    && mkdir -p /data/hcm/logs

WORKDIR /data/hcm