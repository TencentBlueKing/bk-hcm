# HCM 编译指南

---

## 编译环境

- golang >= 1.18
- mysql >= 8.0.17

#### 将go mod设置为auto
```
go env -w GO111MODULE="auto"
```

## 源码下载

``` shell
git clone https://github.com/TencentBlueKing/bk-hcm.git hcm
```

## 下载项目所需依赖
``` shell
cd hcm

go mod tidy
```

 go mod是Golang的包管理工具，若没有开启，可以进行下面操作:
 ``` shell
 go env -w GO111MODULE="auto"

或

 go env -w GO111MODULE="on"
 ```

## 编译

#### 编译共有三种模式

##### 模式一：同时编译前端UI和后端服务

``` shell
make 
```

此模式编译后会同时生成前端UI文件和后端服务文件。

##### 模式二：仅编译后端服务

``` shell
make server
```

此模式下仅会编译生成后端服务文件。

##### 模式三：仅编译前端UI

``` shell
make ui
```

此模式下仅会编译生成前端UI文件。

### 打包

``` shell
make package
```