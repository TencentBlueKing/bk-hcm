## 1. API 与类型

- [x] 1.1 增加 data-service / core 侧业务列表请求与响应类型（含 TCloud extension 过滤、`console_login` 输出）
- [x] 1.2 增加 cloud-server 侧列表请求类型（与接口文档字段对齐）

## 2. DAO

- [x] 2.1 定义联表扫描结果结构体与 `ListSubAccountSecretBizJoin` 方法（JOIN + WHERE + 分页/统计）
- [x] 2.2 扩展 `SubAccountSecret` 接口与 `dao.Set` 装配（若需）

## 3. Data-Service

- [x] 3.1 实现 `ListSubAccountSecretJoinExt` handler：校验、`bk_biz_id` 业务域、调用 DAO、组装响应
- [x] 3.2 注册路由 `POST /vendors/{vendor}/sub_account_secrets/join/list`

## 4. Client

- [x] 4.1 TCloud `SubAccountSecretClient` 增加 `ListSubAccountSecretJoinExt` 调用

## 5. Cloud-Server

- [x] 5.1 实现 `ListSubAccountSecretJoinExt`：路径参数、`meta.Find` 级业务鉴权、调 DS
- [x] 5.2 注册路由 `POST /bizs/{bk_biz_id}/vendors/{vendor}/sub_account_secrets/join/list`（与文档一致）

## 6. 校验

- [x] 6.1 `go build ./...` 通过相关模块
