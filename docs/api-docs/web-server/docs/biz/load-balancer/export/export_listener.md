### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-负载均衡操作。
- 该接口功能描述：导出负载均衡excel。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/listeners/export

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述            |
|-----------|--------|----|---------------|
| bk_biz_id | int64  | 是  | 业务ID          |
| listeners | object array | 是  | 监听器信息，长度限制100 |

#### listeners

| 参数名称      | 参数类型   | 必选 | 描述                                       |
|-----------|--------|----|------------------------------------------|
| lb_id     | string | 是  | 负载均衡id，当只传该参数时，代表负载均衡下的全部监听器             |
| lbl_ids   | string array | 否  | 负载均衡监听器id列表，加上不同listeners该参数的总和，长度限制为100 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "listeners": [
    {
      "lb_id": "0000001",
      "lbl_ids": ["0000001"]
    }
  ]
}
```

### 响应示例

#### 导出成功结果示例

Content-Type: application/octet-stream
Content-Disposition: attachment; filename="hcm-clb-202506120902.zip"
[二进制文件流]
