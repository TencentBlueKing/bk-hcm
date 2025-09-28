# version
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
VERSION   = $(ENV_BK_HCM_VERSION)
DEBUG     = $(ENV_BK_HCM_ENABLE_DEBUG)
MOCK 	= $(ENV_BK_HCM_ENABLE_MOCK)

# 通过tags控制是否编译出mock的服务，目前只有hc-service用到了GOTAGS参数
ifneq ($(MOCK),)
	export GOTAGS=-tags mock
endif


# output directory for release package and version for command line
ifeq ("$(VERSION)", "")
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-hcm
	export LDVersionFLAG = "-X hcm/pkg/version.BUILDTIME=${BUILDTIME} \
		-X hcm/pkg/version.DEBUG=${DEBUG}"
else
	GITHASH   = $(shell git rev-parse HEAD)
	export OUTPUT_DIR = ${PRO_DIR}/build/bk-hcm-${VERSION}
	export LDVersionFLAG = "-X hcm/pkg/version.VERSION=${VERSION} \
    	-X hcm/pkg/version.BUILDTIME=${BUILDTIME} \
    	-X hcm/pkg/version.GITHASH=${GITHASH} \
    	-X hcm/pkg/version.DEBUG=${DEBUG}"
endif

export GO111MODULE=on

include ./scripts/makefile/uname.mk

default: all

# 创建编译文件存储目录
pre:
	@echo -e "\033[34;1mBuilding...\n\033[0m"
	mkdir -p ${OUTPUT_DIR}

# 本地测试前后端编译
all: pre ui server suite
	@echo -e "\033[32;1mBuild All Success!\n\033[0m"

# 后端本地测试编译
server: pre changelog template
	@cd ${PRO_DIR}/cmd && make
	@echo -e "\033[32;1mBuild Server Success!\n\033[0m"

# 二进制出包编译
package: pre ui api ver changelog template
	@echo -e "\033[34;1mPackaging...\n\033[0m"
	@mkdir -p ${OUTPUT_DIR}/bin
	@mkdir -p ${OUTPUT_DIR}/etc
	@mkdir -p ${OUTPUT_DIR}/install
	@mkdir -p ${OUTPUT_DIR}/install/sql
	@cp -f ${PRO_DIR}/scripts/install/migrate.sh ${OUTPUT_DIR}/install/
	@cp -rf ${PRO_DIR}/scripts/sql/* ${OUTPUT_DIR}/install/sql/
	@cd ${PRO_DIR}/cmd && make package
	@echo -e "\033[32;1mPackage All Success!\n\033[0m"

# 容器化编译
docker: pre ui ver changelog template
	@echo -e "\033[34;1mMake Dockering...\n\033[0m"
	@cp -rf ${PRO_DIR}/docs/support-file/docker/* ${OUTPUT_DIR}/
	@mv ${OUTPUT_DIR}/front ${OUTPUT_DIR}/bk-hcm-webserver/
	@mv ${OUTPUT_DIR}/changelog ${OUTPUT_DIR}/bk-hcm-webserver/
	@mv ${OUTPUT_DIR}/template ${OUTPUT_DIR}/bk-hcm-webserver/
	@cp -rf ${PRO_DIR}/scripts/sql ${OUTPUT_DIR}/bk-hcm-dataservice/
	@cd ${PRO_DIR}/cmd && make docker
	@echo -e "\033[32;1mMake Docker All Success!\n\033[0m"

# 编译前端
ui: pre
	@echo -e "\033[34;1mBuilding Front...\033[0m"
	@cd ${PRO_DIR}/front && npm i && npm run build
	@mv ${PRO_DIR}/front/dist ${OUTPUT_DIR}/front
	@echo -e "\033[32;1mBuild Front Success!\n\033[0m"

# 添加版本日志到编译文件中
changelog: pre
	@cp -rf ${PRO_DIR}/docs/support-file/changelog ${OUTPUT_DIR}/
	@echo -e "\033[32;1mPackaging ChangeLog Success!\n\033[0m"

# 添加模板文件到编译文件中
template: pre
	@cp -rf ${PRO_DIR}/docs/support-file/template ${OUTPUT_DIR}/
	@echo -e "\033[32;1mPackaging Template Success!\n\033[0m"

# 添加Api文档到编译文件中
api: pre
	@echo -e "\033[34;1mPackaging API Docs...\033[0m"
	@mkdir -p ${OUTPUT_DIR}/api/
	@mkdir -p ${OUTPUT_DIR}/api/api-server
	@cp -f docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml ${OUTPUT_DIR}/api/api-server
	@tar -czf ${OUTPUT_DIR}/api/api-server/zh.tgz -C docs/api-docs/api-server/docs zh
	@echo -e "\033[32;1mPackaging API Docs Done\n\033[0m"

# 添加版本信息到编译文件中
ver: pre
	@echo ${VERSION} > ${OUTPUT_DIR}/VERSION
	@cp -rf ${PRO_DIR}/CHANGELOG.md ${OUTPUT_DIR}


suite: pre 
	@make -C ${PRO_DIR}/test/suite
	@cp -rf  ${PRO_DIR}/test/suite/suite-test ${OUTPUT_DIR}/
	@rm -rf  ${PRO_DIR}/test/suite/suite-test

mockgen:
	make -C ${PRO_DIR}/pkg/adaptor/mock mockgen

# 初始化下载项目开发依赖工具
init-tools:
	# 前端代码检查依赖工具下载
	curl -o- -L https://yarnpkg.com/install.sh | bash
	# 下载gomock
	make -C  ${PRO_DIR}/pkg/adaptor/mock init-tools



# 清理编译文件
clean:
	@make -C ${PRO_DIR}/cmd clean
	@rm -rf ${PRO_DIR}/build
