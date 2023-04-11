all:
	@cd data-service && make
	@cd hc-service && make
	@cd api-server && make
	@cd auth-server && make
	@cd cloud-server && make
	@cd web-server && make

package:
	@cd data-service && make package
	@cd hc-service && make package
	@cd api-server && make package
	@cd auth-server && make package
	@cd cloud-server && make package
	@cd web-server && make package

clean:
	@cd data-service && make clean
	@cd hc-service && make clean
	@cd api-server && make clean
	@cd auth-server && make clean
	@cd cloud-server && make clean
	@cd web-server && make clean
