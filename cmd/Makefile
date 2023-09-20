all:
	@cd data-service && make
	@cd hc-service && make
	@cd api-server && make
	@cd auth-server && make
	@cd cloud-server && make
	@cd web-server && make
	@cd task-server && make

package:
	@cd data-service && make package
	@cd hc-service && make package
	@cd api-server && make package
	@cd auth-server && make package
	@cd cloud-server && make package
	@cd web-server && make package
	@cd task-server && make package

docker:
	@cd data-service && make docker
	@cd hc-service && make docker
	@cd api-server && make docker
	@cd auth-server && make docker
	@cd cloud-server && make docker
	@cd web-server && make docker
	@cd task-server && make docker

clean:
	@cd data-service && make clean
	@cd hc-service && make clean
	@cd api-server && make clean
	@cd auth-server && make clean
	@cd cloud-server && make clean
	@cd web-server && make clean
	@cd task-server && make clean
