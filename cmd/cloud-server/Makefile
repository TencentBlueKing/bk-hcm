SERVER = bk-hcm-cloudserver
ARGS   = -c ./etc/cloud_server.yaml

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo -e "\033[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) cloud_server.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo -e "\033[32;1mBuild $(SERVER) success!\n\033[0m"

package:
	@echo -e "\033[34;1mPackaging $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(PKGBIN) cloud_server.go
	@cp -f etc/cloud_server.yaml $(PKGETC)/
	@echo -e "\033[32;1mPackage $(SERVER) success!\n\033[0m"

docker:
	@echo -e "\033[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) cloud_server.go
	@echo -e "\033[32;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
