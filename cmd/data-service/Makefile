SERVER = bk-hcm-dataservice
ARGS   = -c ./etc/data_service.yaml

include ../../scripts/makefile/common.mk

include ../../scripts/makefile/uname.mk

default:
	@echo "\033[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) data_service.go
	@cp -r etc $(BINDIR)
	@sh $(GENERATOR) '$(SERVER)' '$(ARGS)' '$(BINDIR)'
	@echo "\033[32;1mBuild $(SERVER) success!\n\033[0m"

package:
	@echo "\033[34;1mPackaging $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(PKGBIN) data_service.go
	@cp -f etc/data_service.yaml $(PKGETC)/
	@echo "\033[32;1mPackage $(SERVER) success!\n\033[0m"

docker:
	@echo "\033[34;1mBuilding $(SERVER)...\033[0m"
	go build -ldflags ${LDVersionFLAG} -o $(BIN) data_service.go
	@echo "\033[32;1mBuild $(SERVER) success!\n\033[0m"

clean:
	@rm -rf $(BINDIR) $(LOCALBUILD)
