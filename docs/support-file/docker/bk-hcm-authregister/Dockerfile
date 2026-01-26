FROM mirrors.tencent.com/build/blueking/bk-hcm-base:v1.2
RUN yum update -y \
    && yum install -y --setopt=install_weak_deps=False jq \
    && yum clean all