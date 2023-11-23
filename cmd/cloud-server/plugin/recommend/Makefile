SERVER = bk-hcm-cloudserver

include ../../../../scripts/makefile/common.mk

default: install
	@pyinstaller --hidden-import autograd --hidden-import pymoo.cython.non_dominated_sorting -F main.py --clean

package:
	@echo "\033[34;1mMaking $(SERVER) Algorithm Plugin...\033[0m"
	$(MAKE) default
	@mkdir -p $(OUTPUT_DIR)/plugin
	@cp -f ./dist/main $(OUTPUT_DIR)/plugin/algorithm
	@$(MAKE) clean
	@echo "\033[32;1mMake $(SERVER) Algorithm Plugin success!\n\033[0m"

normal:
	@echo "\033[34;1mMaking $(SERVER) Algorithm Plugin...\033[0m"
	$(MAKE) default
	@mkdir -p $(BINDIR)/plugin
	@cp -f ./dist/main $(BINDIR)/plugin/algorithm
	@$(MAKE) clean
	@echo "\033[32;1mMake $(SERVER) Algorithm Plugin success!\n\033[0m"

install:
	@pip install -r requirements.txt
	@pip install pyinstaller

clean:
	@rm -rf main.spec ./dist ./build
