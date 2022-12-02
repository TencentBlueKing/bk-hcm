# version
PRO_DIR   = $(shell pwd)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
VERSION   = $(shell echo ${ENV_BK_hcm_VERSION})
DEBUG     = $(shell echo ${ENV_BK_hcm_ENABLE_DEBUG})

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

ui: pre
	@echo -e "\e[34;1mBuilding Front...\033[0m"
	@cd ${PRO_DIR}/front && npm i && npm run build && cp -rf paas-server ${PRO_DIR}/front/dist/
	@mv ${PRO_DIR}/front/dist ${OUTPUT_DIR}/front
	@echo -e "\e[34;1mBuild Front Success!\n\033[0m"

clean:
	@cd ${PRO_DIR}/cmd && make clean
	@rm -rf ${PRO_DIR}/build

init-tools:
	# for gofumpt
	go install mvdan.cc/gofumpt@latest
	# for golines
	go install github.com/segmentio/golines@latest
	# for golangci-lint
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

fmt:
	golines ./ -m 120 -w --base-formatter gofmt --no-reformat-tags
	gofumpt -l -w .

lint:
	golangci-lint run
