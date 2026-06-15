## 1. 枚举常量定义

- [x] 1.1 在 `pkg/criteria/enumor/application.go` 中为 `ApplicationOperation` 定义全量 `Op` 前缀常量（覆盖所有现有 `ApplicationType` 对应的 operation，以及新增细粒度操作如 `OpCreateSubAccount`、`OpUpdateSubAccount`）
- [x] 1.2 为 `ApplicationOperation` 添加 `Validate()` 方法，校验合法枚举值

## 2. Handler 基础层扩展

- [x] 2.1 在 `pkg/criteria/enumor/application.go` 中为 `ApplicationHandler` 接口新增 `GetOperation() enumor.ApplicationOperation` 方法定义（`cmd/cloud-server/service/application/handlers/handler.go`）
- [x] 2.2 在 `BaseApplicationHandler`（`handlers/base.go`）中新增 `operation enumor.ApplicationOperation` 字段，并实现 `GetOperation()` 方法
- [x] 2.3 更新 `NewBaseApplicationHandler` 函数签名，新增 `operation enumor.ApplicationOperation` 参数

## 3. 各业务 Handler 构造函数更新

- [x] 3.1 更新 `handlers/account/init.go`：`NewApplicationOfAddAccount` 传入 `enumor.OpAddAccount`
- [x] 3.2 更新 `handlers/cvm/tcloud/init.go`：传入 `enumor.OpCreateCvm`
- [x] 3.3 更新 `handlers/cvm/aws/init.go`：传入 `enumor.OpCreateCvm`
- [x] 3.4 更新 `handlers/cvm/huawei/init.go`：传入 `enumor.OpCreateCvm`
- [x] 3.5 更新 `handlers/cvm/gcp/init.go`：传入 `enumor.OpCreateCvm`
- [x] 3.6 更新 `handlers/cvm/azure/init.go`：传入 `enumor.OpCreateCvm`
- [x] 3.7 更新 `handlers/disk/tcloud/init.go`：传入 `enumor.OpCreateDisk`
- [x] 3.8 更新 `handlers/disk/aws/init.go`：传入 `enumor.OpCreateDisk`
- [x] 3.9 更新 `handlers/disk/huawei/init.go`：传入 `enumor.OpCreateDisk`
- [x] 3.10 更新 `handlers/disk/gcp/init.go`：传入 `enumor.OpCreateDisk`
- [x] 3.11 更新 `handlers/disk/azure/init.go`：传入 `enumor.OpCreateDisk`
- [x] 3.12 更新 `handlers/vpc/tcloud/init.go`：传入 `enumor.OpCreateVpc`
- [x] 3.13 更新 `handlers/vpc/aws/init.go`：传入 `enumor.OpCreateVpc`
- [x] 3.14 更新 `handlers/vpc/huawei/init.go`：传入 `enumor.OpCreateVpc`
- [x] 3.15 更新 `handlers/vpc/gcp/init.go`：传入 `enumor.OpCreateVpc`
- [x] 3.16 更新 `handlers/vpc/azure/init.go`：传入 `enumor.OpCreateVpc`
- [x] 3.17 更新 `handlers/load_balancer/tcloud/init.go`：传入 `enumor.OpCreateLoadBalancer`
- [x] 3.18 更新 `handlers/main-account/create-main-account/init.go`：传入 `enumor.OpCreateMainAccount`
- [x] 3.19 更新 `handlers/main-account/update-main-account/init.go`：传入 `enumor.OpUpdateMainAccount`

## 4. 创建申请单流程改造

- [x] 4.1 在 `cmd/cloud-server/service/application/create.go` 的 `createApplication` 函数中，将 `ApplicationCreateReq.Operation` 设置为 `handler.GetOperation()`
- [x] 4.2 在 `create.go` 中定义 `needBkBizIDsOps` 白名单 map，将 `bkBizIDs` 的判断逻辑从对比 `applicationType` 改为对比 `handler.GetOperation()` 与白名单

## 5. 查询接口改造

- [x] 5.1 在 `pkg/api/data-service/application.go` 的 `ApplicationResp` 中新增 `Operation string` 字段（json tag: `"operation"`）
- [x] 5.2 确认 data-service 侧 list/get 接口的数据库查询已包含 `operation` 列（`ApplicationColumnDescriptor` 已有，确认 DAO 层 select 不遗漏）

## 6. 数据迁移

- [x] 6.1 编写存量数据迁移 SQL 脚本（文件放在 `scripts/sql/` 或约定的 migration 目录）：
  ```sql
  UPDATE application SET operation = `type` WHERE operation = '' OR operation IS NULL;
  ```
- [x] 6.2 在部署文档或 changelog 中说明需在上线前执行此迁移脚本

## 7. 编译 & 验证

- [x] 7.1 执行 `go build ./...` 确认全量编译无报错（重点验证 `NewBaseApplicationHandler` 签名变更后所有调用处均已更新）
- [ ] 7.2 手动或通过接口测试验证：创建申请单后，`application` 表的 `operation` 字段有正确值
- [ ] 7.3 验证 `ListBizApplications` 响应中包含 `operation` 字段，且支持以 `operation` 作为 filter 条件查询
