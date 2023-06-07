### 描述

- 该接口提供版本：v1.0.0+
- 该接口所需权限：资源查看
- 该接口功能描述：查询单个镜像信息

### URL

GET /api/v1/cloud/vendors/{vendor}/images/{id}

#### 路径参数说明
| 参数名称    | 参数类型    | 必选 | 描述    |
|---------|---------|----|-------|
| vendor  | string  | 是  | 云厂商   |
| id      | string  | 是  | 镜像 ID |

### 调用示例
如查询云厂商是 tcloud 的，镜像 ID 是 00000002 的镜像信息
#### 返回参数示例
```json
{
    "code": 0,
    "message": "",
    "data": {
        "id": "00000002",
        "vendor": "tcloud",
        "name": "CentOS 7.6 64位",
        "cloud_id": "img-9qabwvbn",
        "platform": "CentOS",
        "architecture": "x86_64",
        "type": "public",
        "created_at": "2023-01-16T03:30:41Z",
        "updated_at": "2023-01-16T08:39:28Z",
        "extension": {
            "region": "ap-beijing",
            "image_source": "OFFICIAL",
            "image_size": 20
        }
    }
}
```
### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int  | 状态码  |
| message | string | 请求信息 |
| data    | Data | 响应数据 |
#### Data
| 参数名称   | 参数类型   | 描述                                       |
|--------|--------|------------------------------------------|
| id | string | 镜像 ID |
| vendor | string | 云厂商 |
| name | string | 镜像名 |
| cloud_id | string | 镜像 ID |
| platform | string | 镜像平台 |
| architecture | string | 镜像架构 |
| type | string | 镜像类型 | 
| creator | string | 创建者 |
| reviser | string | 更新者 |
| created_at | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at | string | 更新时间，标准格式：2006-01-02T15:04:05Z | 
| extension | PublicImageExtension[vendor] | 各云厂商的差异化字段| 

#### PublicImageExtension[tcloud]

| 参数名称                           | 参数类型 |描述                                                         |
|--------------------------------| -------- |  ------------------------------------------------------------ |
| region | string | 区域 |
| image_source | string | 镜像来源 |
| image_size | uint | 镜像大小 |


#### PublicImageExtension[azure]
空字典

#### PublicImageExtension[huawei]
空字典

#### PublicImageExtension[gcp]
空字典

#### PublicImageExtension [aws]
空字典
