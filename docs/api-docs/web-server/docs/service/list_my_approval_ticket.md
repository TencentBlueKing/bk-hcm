### 描述

- 该接口提供版本：v1.1.23+。
- 该接口所需权限：。
- 该接口功能描述：查询Itsm待我审批单据列表。

### URL

POST /api/v1/web/tickets/types/my_approval/list

### 输入参数

| 参数名称 | 参数类型   | 必选 | 描述   |
|------|--------|----|------|
| page | object | 是  | 分页设置 |

#### page

| 参数名称   | 参数类型    | 必选 | 描述                 |
|--------|---------|----|--------------------|
| start	 | uint32	 | 否	 | 记录开始位置，start 起始值为0 |
| limit	 | uint32	 | 否	 | 每页限制条数，最大500，不能为0  |

### 调用示例

```json
{
  "page": {
    "start": 0,
    "limit": 500
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "sn": "REQ20230726000025",
      "id": 20007,
      "title": "申请配置平台1个操作权限",
      "service_id": 55,
      "service_type": "request",
      "meta": {
        "priority": {
          "key": "",
          "name": "",
          "order": 0
        }
      },
      "bk_biz_id": -1,
      "current_status": "RUNNING",
      "create_at": "2023-07-26 16:00:07",
      "creator": "lampardtang",
      "is_supervise_needed": true,
      "flow_id": 3743,
      "supervise_type": "EMPTY",
      "supervisor": "",
      "service_name": "自定义权限申请审批流程",
      "current_status_display": "处理中",
      "current_steps": [
        {
          "id": 40847,
          "tag": "DEFAULT",
          "name": "系统管理员审批"
        }
      ],
      "priority_name": "--",
      "current_processors": "",
      "can_comment": false,
      "can_operate": false,
      "waiting_approve": true,
      "followers": [],
      "comment_id": "-1",
      "can_supervise": false,
      "can_withdraw": true,
      "sla": [],
      "sla_color": ""
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int32        | 状态码  |
| message | string       | 请求信息 |
| data    | array object | 响应数据 |

#### data[n]

#### data.detail[n]

| 参数名称                   | 参数类型         | 描述              |
|------------------------|--------------|-----------------|
| sn                     | string       | 单号              |
| id                     | int          | 单据id            |
| title                  | string       | 单据标题            |
| service_id             | int          | 服务id            |
| service_type           | string       | 服务类型            |
| service_name           | string       | 服务名称            |
| bk_biz_id              | int          | 业务id，无业务关联为-1   |
| catalog_id             | int          | 服务目录id          |
| current_status         | string       | 单据当前状态          |
| current_status_display | string       | 单据当前状态          |
| current_steps          | array object | 单据当前步骤          |
| flow_id                | int          | 流程版本id          |
| comment_id             | string       | 单据评价id          |
| is_commented           | bool         | 单据是否已评价         |
| updated_by             | string       | 最近更新者           |
| update_at              | string       | 最近更新时间          |
| end_at                 | string       | 结束时间            |
| creator                | string       | 提单人             |
| create_at              | string       | 创建时间            |
| is_biz_need            | bool         | 是否与业务关联         |
| is_supervise_needed    | bool         | 是否需要督办          |
| supervise_type         | string       | 督办人类型           |
| supervisor             | string       | 督办人             |
| can_supervise          | bool         | 是否督办            |
| priority_name          | string       | 优先级             |
| can_comment            | bool         | 是否可以评论          |
| can_operate            | bool         | 是否可以操作          |
| waiting_approve        | bool         | 是否为待审批状态        |
| current_processors     | string       | 当前处理人           |
| can_withdraw           | bool         | 是否可以撤销          |
| followers              | array string | 关注人             |
| sla                    | array string | 配置的sla名称        |
| sla_color              | string       | 触发sla后，在前端响应的颜色 |

#### data.detail[n].current_steps

| 参数名称     | 参数类型   | 描述     |
|----------|--------|--------|
| id       | int64  | id     |
| tag      | string | 标签     |
| name     | string | 名称     |
| state_id | int64  | 当前流程ID |
