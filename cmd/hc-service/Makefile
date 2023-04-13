SERVER = bk-hcm-hcservice
ARGS   = -c ./etc/hc_service.yaml

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) hc_service.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

package:
	@echo -e "\e[34;1mPackaging $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(PKGBIN) hc_service.go
	@cp -f etc/hc_service.yaml $(PKGETC)/
	@echo -e "\e[34;1mPackage $(SERVER) success!\n\033[0m"

docker:
	@echo -e "\e[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) hc_service.go
	@echo -e "\e[34;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
