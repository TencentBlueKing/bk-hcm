### 描述

- 该接口提供版本：v1.0.0+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：关联 eip。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/eips/associate

### 输入参数

#### TCloud
| 参数名称                 | 参数类型     | 必选  | 描述     |
|----------------------|----------|-----|--------|
| bk_biz_id            | int64    | 是   | 业务ID   |
| eip_id               | string   | 是   | Eip ID |
| cvm_id               | string   | 是   | 虚拟机ID  |
| network_interface_id | string   | 是   | 网络接口ID |

#### Aws
| 参数名称            | 参数类型      | 必选   | 描述     |
|-----------------|-----------|------|--------|
| bk_biz_id       | int64     | 是    | 业务ID   |
| eip_id          | string    | 是    | Eip ID |
| cvm_id          | string    | 是    | 虚拟机ID  |

#### HuaWei
| 参数名称                 | 参数类型       | 必选    | 描述     |
|----------------------|------------|-------|--------|
| bk_biz_id            | int64      | 是     | 业务ID   |
| eip_id               | string     | 是     | Eip ID |
| network_interface_id | string     | 是     | 网络接口ID |

#### Gcp
| 参数名称                 | 参数类型     | 必选    | 描述      |
|----------------------|----------|-------|---------|
| bk_biz_id            | int64    | 是     | 业务ID    |
| eip_id               | string   | 是     | Eip ID  |
| network_interface_id | string   | 是     | 网络接口ID  |

#### Azure
| 参数名称                 | 参数类型      | 必选    | 描述     |
|----------------------|-----------|-------|--------|
| bk_biz_id            | int64     | 是     | 业务ID   |
| eip_id               | string    | 是     | Eip ID |
| network_interface_id | string    | 是     | 网络接口ID |

### 调用示例

```json
{
  "eip_id": "00001111",
  "cvm_id": "00001112",
  "network_interface_id": "00001112"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
