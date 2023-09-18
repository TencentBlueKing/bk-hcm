SERVER = bk-hcm-hcservice
ARGS   = -c ./etc/hc_service.yaml

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\033[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} ${GOTAGS} -o $(BIN) hc_service.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\033[32;1mBuild $(SERVER) success!\n\033[0m"

package:
	echo hc_service package ${GOTAGS}
	@echo -e "\033[34;1mPackaging $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG}  ${GOTAGS}  -o $(PKGBIN) hc_service.go
	@cp -f etc/hc_service.yaml $(PKGETC)/
	@echo -e "\033[32;1mPackage $(SERVER) success!\n\033[0m"

docker:
	@echo -e "\033[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) hc_service.go
	@echo -e "\033[32;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
