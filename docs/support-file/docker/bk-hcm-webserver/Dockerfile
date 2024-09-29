FROM hub.bktencent.com/blueking/bk-hcm-base:v1.1

COPY bk-hcm-webserver /data/hcm/
COPY front /data/hcm/front
COPY changelog /data/hcm/changelog

CMD ["/data/hcm/bk-hcm-webserver", "--config-file", "/data/hcm/etc/config.yaml"]
