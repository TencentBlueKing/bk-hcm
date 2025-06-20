### 描述

- 该接口提供版本：v1.8.1+。
- 该接口所需权限：IaaS资源操作。
- 该接口功能描述：批量重装虚拟机。

### URL

POST /api/v1/cloud/cvms/batch/reset_async

### 输入参数

| 参数名称        | 参数类型         | 必选 | 描述                     |
|-------------|--------------|----|------------------------|
| hosts       | object array | 是  | 虚拟机的Host列表, 最多支持500台主机 |
| pwd         | string       | 是  | 重装密码                   |
| pwd_confirm | string       | 是  | 重装确认密码                 |

#### hosts[n]
| 参数名称        | 参数类型    | 必选 | 描述                                          |
|----------------|-----------|------|---------------------------------------------|
| id	         | string	 | 是   | 主机唯一ID                                      |
| device_type    | string	 | 是   | 机型                                          |
| image_name_old | string	 | 是   | 原镜像名称                                       |
| cloud_image_id | string	 | 是   | 新镜像云ID                                      |
| image_name     | string	 | 是   | 新镜像名称                                       |
| image_type     | string	 | 是   | 新镜像类型(PUBLIC_IMAGE:公共镜像 PRIVATE_IMAGE:私有镜像) |

### 调用示例

```json
{
  "hosts": [
    {
      "id": "00000001",
      "device_type": "SA5.4XLARGE32",
      "image_name_old": "Tencent OS 001",
      "cloud_image_id": "img-002",
      "image_name": "Tencent OS 002",
      "image_type": "PUBLIC_IMAGE"
    }
  ],
  "pwd": "xxxxxx",
  "pwd_confirm": "xxxxxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_management_id": "xxxxxx"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data参数说明

| 参数名称               | 参数类型   | 描述     |
|--------------------|--------|--------|
| task_management_id | string | 任务管理id |

