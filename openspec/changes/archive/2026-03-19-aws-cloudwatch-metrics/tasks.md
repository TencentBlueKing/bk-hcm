## 1. CloudWatch client 基础设施

- [x] 1.1 `pkg/adaptor/aws/client.go`：`clientSet` 新增 `cloudWatchClient(region string) (*cloudwatch.CloudWatch, error)` 方法，引入 `github.com/aws/aws-sdk-go/service/cloudwatch`
- [x] 1.2 `pkg/adaptor/aws/cloudwatch.go`：`Aws` 结构体新增 `GetMetricData` 和 `ListMetrics` adaptor 方法，封装 SDK 调用、分页处理和类型转换

## 2. CloudWatch 数据类型定义

- [x] 2.1 `pkg/api/hc-service/instance-type/aws_cloudwatch.go`：更新 `AwsAssumeRoleGetMetricDataReq`（root_account_id + main_account_id + role_chain + region + metric_data_queries + start_time + end_time）
- [x] 2.2 定义 `MetricDataQueryParam` 结构体（Id、Namespace、MetricName、Dimensions、Stat、Period）
- [x] 2.3 定义 `AwsAssumeRoleGetMetricDataResp`（MetricDataResults 数组，每个包含 Id、Timestamps、Values）
- [x] 2.4 更新 `AwsAssumeRoleListMetricsReq`（root_account_id + main_account_id + role_chain + region + namespace + metric_name + dimensions）
- [x] 2.5 定义 `AwsAssumeRoleListMetricsResp`（Metrics 数组，每个包含 Namespace、MetricName、Dimensions）

## 3. hc-service CloudWatch handler

- [x] 3.1 `cmd/hc-service/service/`：更新 GetMetricData handler（用 root_account_id 调 AwsRoot + main_account_id 查 CloudID → AwsWithAssumeRole → adaptor.GetMetricData）
- [x] 3.2 更新 ListMetrics handler（同上模式 → adaptor.ListMetrics）
- [x] 3.3 注册 hc-service 路由：`POST /vendors/aws/assume_role/cloudwatch/metric_data/get` 和 `POST /vendors/aws/assume_role/cloudwatch/metrics/list`

## 4. hc-service client 方法

- [x] 4.1 `pkg/client/hc-service/aws/`：新增 `GetAssumeRoleMetricData` client 方法
- [x] 4.2 新增 `ListAssumeRoleMetrics` client 方法

## 5. cloud-server 入口与开放接口

- [x] 5.1 `cmd/cloud-server/service/`：更新 GetMetricData 资源视角 handler（移除 CloudID→AccountID 反查）
- [x] 5.2 更新 ListMetrics 资源视角 handler
- [x] 5.3 注册 cloud-server 路由：`POST /vendors/aws/assume_role/cloudwatch/metric_data/get` 和 `POST /vendors/aws/assume_role/cloudwatch/metrics/list`
- [x] 5.4 `docs/api-docs/api-server/api/bk_apigw_resources_bk-hcm.yaml`：注册两个 CloudWatch 开放接口

## 6. 接口文档

- [x] 6.1 `docs/api-docs/web-server/docs/resource/get_aws_assume_role_metric_data.md`：更新 GetMetricData 接口文档
- [x] 6.2 `docs/api-docs/web-server/docs/resource/list_aws_assume_role_metrics.md`：更新 ListMetrics 接口文档
