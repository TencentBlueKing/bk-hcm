# version
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
VERSION   = $(shell echo ${ENV_BK_HCM_VERSION})
DEBUG     = $(shell echo ${ENV_BK_HCM_ENABLE_DEBUG})

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

pre:
	@echo -e "\e[34;1mBuilding...\n\033[0m"
	mkdir -p ${OUTPUT_DIR}

all: pre ui server
	@cd ${PRO_DIR}/cmd && make
	@echo -e "\e[34;1mBuild All Success!\n\033[0m"

server: pre
	@cd ${PRO_DIR}/cmd && make
	@echo -e "\e[34;1mBuild Server Success!\n\033[0m"

package: pre ui api
	@echo -e "\e[34;1mPackaging...\n\033[0m"
	@mkdir -p ${OUTPUT_DIR}/bin
	@mkdir -p ${OUTPUT_DIR}/etc
	@mkdir -p ${OUTPUT_DIR}/install
	@mkdir -p ${OUTPUT_DIR}/install/sql
	@cp -rf ${PRO_DIR}/scripts/sql/* ${OUTPUT_DIR}/install/sql/
	@cd ${PRO_DIR}/cmd && make package
	@echo -e "\e[34;1mPackage All Success!\n\033[0m"

ui: pre
	@echo -e "\e[34;1mBuilding Front...\033[0m"
	@cd ${PRO_DIR}/front && npm i && npm run build
	@mv ${PRO_DIR}/front/dist ${OUTPUT_DIR}/front
	@echo -e "\e[34;1mBuild Front Success!\n\033[0m"

api: pre
	@echo -e "\e[34;1mPackaging API Docs...\033[0m"
	@mkdir -p ${OUTPUT_DIR}/api/
	@mkdir -p ${OUTPUT_DIR}/api/api-server
	@cp -f docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml ${OUTPUT_DIR}/api/api-server
	@tar -czf ${OUTPUT_DIR}/api/api-server/zh.tgz -C docs/api-docs/api-server/docs zh
	@echo -e "\e[34;1mPackaging API Docs Done\n\033[0m"

clean:
	@cd ${PRO_DIR}/cmd && make clean
	@rm -rf ${PRO_DIR}/build

init-tools:
	# for golangci-lint
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	golangci-lint run
