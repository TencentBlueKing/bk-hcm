### 描述

- 该接口提供版本：v1.8.7+。
- 该接口所需权限：业务-负载均衡操作。
- 该接口功能描述：导出RS。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/{vendor}/targets/export

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述               |
|------------|--------------|----|------------------|
| bk_biz_id  | int64        | 是  | 业务ID             |
| target_ids | string array | 是  | RS的id列表，长度限制5000 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "target_ids": ["0000001", "0000002", "0000003"]
}
```

### 响应示例

#### 导出成功结果示例

Content-Type: application/octet-stream
Content-Disposition: attachment; filename="hcm-clb-202506120902.zip"
[二进制文件流]
