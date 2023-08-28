TARGET = suite-test

PWD        = $(shell pwd)
OUTPUT_DIR = $(PWD)/$(TARGET)

export GO111MODULE=on

default:
	@echo -e "\e[34;1mBuilding $(TARGET)...\033[0m"
	@mkdir -p $(OUTPUT_DIR)
	@cd cases/cloud-server && go test -c && mv *.test $(OUTPUT_DIR)/
	@cd cases/hc-service && go test -c && mv *.test $(OUTPUT_DIR)/
	@cd tools && go build -o $(OUTPUT_DIR)/testhelper
	@cd tools && cp start.sh $(OUTPUT_DIR)/ && cp README.md $(OUTPUT_DIR)/
	@echo -e "\e[34;1mBuild $(TARGET) success!\n\033[0m"

clean:
	@rm -rf $(OUTPUT_DIR)
