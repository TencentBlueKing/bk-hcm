### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：获取指定模板文件。

### URL

GET /api/v1/web/templates/{filename}


### 输入参数

| 参数名称     | 参数类型   | 必选 | 描述   |
|----------|--------|----|------|
| filename | string | 是  | 文件名称 |

#### filename 可选项

| filename                                    | 说明            |
|---------------------------------------------|---------------|
| 1_hcm_clb_tcp_udp_listener_template.xlsx    | 创建四层监听器模板文件   |
| 2_hcm_clb_http_https_listener_template.xlsx | 创建七层监听器模板文件   |
| 3_hcm_clb_bind_rs_url_ruler_template.xlsx   | 创建URL规则模板文件   |
| 4_hcm_clb_bind_rs_tcp_udp_template.xlsx     | 四层监听器绑定RS模板文件 |
| 5_hcm_clb_url_rule_http_https_template.xlsx | 七层监听器绑定RS模板文件 |

### 响应示例

#### 导出成功结果示例

Content-Type: application/octet-stream
Content-Disposition: attachment; filename="Xxxxx.xlsx"
[二进制文件流]
