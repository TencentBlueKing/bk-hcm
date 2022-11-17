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

all: pre
	@cd ${PRO_DIR}/cmd && make
	@echo -e "\e[34;1mBuild All Success!\n\033[0m"

clean:
	@cd ${PRO_DIR}/cmd && make clean
	@rm -rf ${PRO_DIR}/build
