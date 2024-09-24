### 描述

- 该接口提供版本：v1.6.7+。
- 该接口所需权限：无。
- 该接口功能描述：获取通知中心信息。

### URL

GET /api/v1/web/notice/current_announcements

### 输入参数

| 参数名称      | 参数类型    | 必选 | 描述                  |
|-----------|---------|----|---------------------|
| limit	    | int	    | 否	 | 每页返回的结果数。           |
| offset	   | int	    | 否	 | 返回结果的初始索引。          |

以上参数都是URLParam

### 调用示例


```
limit=10&offset=0
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": [
    {
      "id": 1,
      "title": "这是通知标题",
      "content": "这是查询参数指定语言的通知内容",
      "content_list": [
        {
          "content": "这是通知内容",
          "language": "zh-cn"
        }
      ],
      "announce_type": "event",
      "start_time": "2024-09-10T10:42:04.9967548+08:00",
      "end_time": "2024-09-10T10:42:04.9967548+08:00"
    }
  ]
}
```

### 响应参数说明

| 参数名称    | 参数类型     | 描述   |
|---------|----------|------|
| code    | int32    | 状态码  |
| message | string   | 请求信息 |
| data    | []object | 响应数据 |

#### data参数说明
| 参数名称          | 参数类型     | 描述                                            |
|---------------|----------|-----------------------------------------------|
| title         | string   | 标题                                            |
| content       | string   | 返回指定语言的公告内容(如果指定语言的公告不存在，则优先返回en语言，再没有就返回第一条) |
| content_list  | []object | 语言内容列表                                        |
| announce_type | string   | 公告类型: "event" (活动通知)/ "announce"(平台公告)        |
| start_time    | string   | 开始时间                                          |
| end_time      | string   | 结束时间                                          |

#### data.content_list参数说明
| 参数名称     | 参数类型   | 描述             |
|----------|--------|----------------|
| content  | string | 公告内容           |
| language | string | 公告语言(zh-cn/en) |

