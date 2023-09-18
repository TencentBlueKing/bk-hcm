mockgen:
	mockgen -destination tcloud/tcloud_mock.go  -package=mocktcloud -typed -source=../tcloud/interface.go

init-tools:
	# 安装gomock
	go install go.uber.org/mock/mockgen@latest

default:mockgen