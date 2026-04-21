FROM mirrors.tencent.com/build/blueking/bk-hcm-base:v1.2
RUN yum update -y \
    && yum install -y --setopt=install_weak_deps=False jq \
    && yum clean all
COPY bin/auth-register.sh /data/bin/auth-register.sh
RUN chmod +x /data/bin/auth-register.sh