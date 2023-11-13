FROM hub.bktencent.com/blueking/bk-hcm-base:v1.0

COPY bk-hcm-webserver /data/hcm/
COPY front /data/hcm/front

CMD ["/data/hcm/bk-hcm-webserver", "--config-file", "/data/hcm/etc/config.yaml"]
