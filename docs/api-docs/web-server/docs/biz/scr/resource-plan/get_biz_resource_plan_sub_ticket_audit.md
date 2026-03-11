### 描述

- 该接口提供版本：v1.8.5.6+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询资源预测子单的审批流，包括审批状态、当前审批阶段等。

### URL

GET /api/v1/woa/bizs/{bk_biz_id}/plans/resources/sub_tickets/{sub_ticket_id}/audit

### 输入参数

无

### 调用示例

无

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "00000001",
    "admin_audit": {
      "status": "auditing",
      "current_steps": [
        {
          "name": "管理员审批",
          "processors": [
            "zhangsan",
            "lisi"
          ],
          "processors_auth": {
            "zhangsan": true,
            "lisi": true
          }
        }
      ],
      "logs": [
        {
          "name": "管理员审批",
          "operator": "lisi",
          "operate_info": "审批意见",
          "operate_at": "2024-11-06 12:03:12"
        }
      ]
    },
    "crp_audit": {
      "crp_sn": "XQ000001",
      "crp_url": "http://crp/ticket/XQ000001",
      "status": "init",
      "status_name": "待审批",
      "message": "如果失败，这里会写原因",
      "current_steps": [
        {
          "state_id": "",
          "name": "规划经理审批",
          "processors": [
            "lisi",
            "wangwu"
          ],
          "processors_auth": {
            "lisi": true,
            "wangwu": false
          }
        }
      ],
      "logs": [
        {
          "operator": "xxxx",
          "operate_at": "2024-11-06 12:03:12",
          "message": "同意",
          "name": "部门管理员审批"
        }
      ]
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	   | object | 响应数据                      |

#### data

| 参数名称        | 参数类型   | 描述                      |
|-------------|--------|-------------------------|
| id          | string | 资源预测需求子单ID              |
| admin_audit | object | 部门审批信息                  |
| crp_audit   | object | crp审批信息，没有CRP审批阶段时为null |

#### data.admin_audit

| 参数名称            | 参数类型         | 描述                                                        |
|-----------------|--------------|-----------------------------------------------------------|
| status          | string       | 审批状态（枚举值：init, auditing, rejected, done, failed, revoked） |
| current_steps   | object array | 当前审批阶段                                                    |
| logs            | object array | 审批历史列表                                                    |

#### data.crp_audit

| 参数名称          | 参数类型         | 描述                                                        |
|---------------|--------------|-----------------------------------------------------------|
| crp_sn        | string       | CRP流程单号                                                   |
| crp_url       | string       | CRP流程单链接                                                  |
| status        | string       | 审批状态（枚举值：init, auditing, rejected, done, failed, revoked） |
| status_name   | string       | 审批状态名称                                                    |
| message       | string       | 审批失败原因                                                    |
| current_steps | object array | 当前审批阶段                                                    |
| logs          | object array | 审批历史列表                                                    |

#### data.admin_audit.current_steps[n]

| 参数名称            | 参数类型         | 描述       |
|-----------------|--------------|----------|
| name            | string       | 步骤名称     |
| processors      | string array | 审批人列表    |
| processors_auth | object       | 审批人是否有权限 |

#### data.admin_audit.logs[n]

| 参数名称       | 参数类型   | 描述     |
|------------|--------|--------|
| operator   | string | 审批人    |
| operate_at | string | 处理时间   |
| name       | string | 步骤名称   |
| operate_info | string | 审批意见 |

#### data.crp_audit.current_steps[n]

| 参数名称            | 参数类型         | 描述       |
|-----------------|--------------|----------|
| state_id        | string       | 步骤ID     |
| name            | string       | 步骤名称     |
| processors      | string array | 审批人列表    |
| processors_auth | object       | 审批人是否有权限 |

#### data.crp_audit.logs[n]

| 参数名称       | 参数类型   | 描述     |
|------------|--------|--------|
| operator   | string | 审批人    |
| operate_at | string | 处理时间   |
| message    | string | 审批结果信息 |
| name       | string | 步骤名称   |
