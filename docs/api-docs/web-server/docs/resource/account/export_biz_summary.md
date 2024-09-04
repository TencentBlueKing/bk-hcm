### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：账单查看。
- 该接口功能描述：导出根据业务聚合的账单信息。

### URL

POST /api/v1/account/bills/biz_summarys/export

### 输入参数

| 参数名称         | 参数类型      | 必选 | 描述               |
|--------------|-----------|----|------------------|
| bill_year    | int       | 是  | 账单年份             |
| bill_month   | int       | 是  | 账单月份             |
| export_limit | int       | 是  | 导出限制条数, 0-200000 |
| bk_biz_ids   | int array | 是  | 业务id             |



### 调用示例

#### 获取详细信息请求参数示例

导出2024年6月的, 业务id为2005000002的账单, 限制条数为100条.

```json
{
  "bill_year": 2024,
  "bill_month": 6,
  "bk_biz_ids": [2005000002],
  "export_limit": 100
}
```



### 响应示例

#### 导出成功结果示例

Content-Type: application/octet-stream
Content-Disposition: attachment; filename="bill_summary_biz.csv.zip"
[二进制文件流]
