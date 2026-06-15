## 背景与动机

按业务查询三级账号密钥列表的 Web API（`POST .../sub_account_secrets/join/list`）需要返回分布在 `account`、`sub_account` 上的字段（负责人列表、TCloud 的 `console_login` 及 extension 中的云侧 ID等；**不**在响应中返回 `tenant_id`），而当前 data-service 列表路径只单独查 `sub_account_secret`。若不做联表列表与分层校验，文档中的过滤条件无法正确生效，cloud-server 也无法按已发布的接口契约实现。

## 变更内容

- 增加**业务维度列表**能力：cloud-server 路由、IAM **业务访问**鉴权、与 `list_sub_account_secret.md` 一致的请求/响应类型。
- 扩展 **data-service** 列表能力：在 **DAO 层** 将 `sub_account_secret` 与 `sub_account`、`account` **联表**（SQL 使用表别名），以支持文档中的过滤条件；复用现有模式（如 `ListJoinAccount`、filter 校验、租户注入）。
- 落实**参数校验**（分页规则、数组长度 ≤ 500、厂商扩展字段）、**权限校验**（业务与资源语义与其他 biz 列表 API 一致），以及在契约隐含**作用域**时的**存在性/一致性**约束（例如结果限定在请求的业务与 vendor 下）。
- 填充响应字段：`account_managers`、`sub_account_managers`，以及 TCloud 扩展字段（`cloud_sub_account_id` 来自 `sub_account.extension.uin`、`cloud_main_account_id`、`cloud_secret_id`、`console_login`）；**不包含** `tenant_id`。

## 能力范围（Capabilities）

### 新增能力

- `biz-sub-account-secret-list`：业务作用域下的三级账号密钥列表，联表带出账号/子账号属性，支持过滤、仅统计条数模式及鉴权。

### 修改的既有能力

- 无（`openspec/specs/` 中该区域暂无基线 spec）

## 影响面

- **cloud-server**：`cmd/cloud-server/service/subaccount-secret/`（新增 list 处理函数与路由注册；当前 `list.go` 为空）。
- **data-service**：`cmd/data-service/service/cloud/sub-account-secret/`，以及 `pkg/client/data-service/` 下新增或扩展的列表 API 与客户端。
- **DAO / 类型**：`pkg/dal/dao/cloud/sub-account-secret/`、`pkg/dal/dao/types/`，以及联表结果结构体。
- **API 包**：`pkg/api/cloud-server/`、`pkg/api/data-service/cloud/`、`pkg/api/core/cloud/sub-account-secret/`（列表结果形态，含 managers，不含 `tenant_id`）。
- **文档**：`list_sub_account_secret.md` 已约定响应不返回 `tenant_id`；其余字段命名与 cloud-server 约定一致。
