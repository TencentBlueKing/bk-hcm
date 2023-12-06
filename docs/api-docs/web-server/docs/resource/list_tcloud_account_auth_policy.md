### 描述

- 该接口提供版本：v1.2.1+。
- 该接口所需权限：账号录入。
- 该接口功能描述：获取账号已有的权限策略。(代理接口)

### URL

POST /vendors/tcloud/accounts/auth_policies/list

#### 输入参数

| 参数名称             | 参数类型   | 必选 | 描述     |
|------------------|--------|----|--------|
| cloud_secret_id  | string | 是  | 云密钥ID  |
| cloud_secret_key | string | 是  | 云密钥Key |
| uin              | string | 是  | 云账号ID  |
| service_type     | string | 否  | 云服务类型  |

### 调用示例

```json
{
  "cloud_secret_id": "xxxx",
  "cloud_secret_key": "xxxx",
  "uin": 1000
}
```

#### 返回参数示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "Service": {
        "ServiceType": "cdn",
        "ServiceName": "内容分发网络"
      },
      "Action": [
        {
          "Name": "AddCLSTopicDomains",
          "Description": "AddCLSTopicDomains 用于新增域名到某日志主题下"
        }
      ],
      "Policy": [
        {
          "PolicyId": "1",
          "PolicyName": "AdministratorAccess",
          "PolicyType": "Presetting",
          "PolicyDescription": "该策略允许您管理账户内所有用户及其权限、财务相关的信息、云服务资产。"
        }
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| code    | int          | 状态码  |
| message | string       | 请求信息 |
| data    | array object | 响应数据 |

#### data[n]

| 参数名称    | 参数类型         | 描述    |
|---------|--------------|-------|
| Service | object       | 服务    |
| Action  | array object | 接口信息  |
| Policy  | array object | 授权的策略 |

#### Service

| 参数名称        | 参数类型   | 描述  |
|-------------|--------|-----|
| ServiceType | string | 服务  |
| ServiceName | string | 服务名 |

#### Action[n]

| 参数名称        | 参数类型   | 描述  |
|-------------|--------|-----|
| Name        | string | 接口名 |
| Description | string | 描述  |

#### Policy[n]

| 参数名称              | 参数类型   | 描述   |
|-------------------|--------|------|
| PolicyID          | string | 策略ID |
| PolicyName        | string | 策略名  |
| PolicyType        | string | 策略类型 |
| PolicyDescription | string | 策略描述 |
