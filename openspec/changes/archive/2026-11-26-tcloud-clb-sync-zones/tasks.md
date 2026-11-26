# Tasks

## 1. 数据服务（`data-service`）批量更新协议与落库支持

- [x] 1.1 在 `pkg/api/data-service/cloud/load_balancer.go` 的 `LoadBalancerExtUpdateReq[T]` 结构体中新增可选字段 `Zones []string`（`json:"zones,omitempty"`），并通过 `go build ./pkg/api/data-service/cloud/...` 验证编译通过
- [x] 1.2 在 `cmd/data-service/service/cloud/load-balancer/update.go` 的 `batchUpdateLoadBalancer` 中将 `lb.Zones` 映射写入 `tablelb.LoadBalancerTable.Zones`，并通过 `go build ./cmd/data-service/...` 验证编译通过且未传 `Zones`（`nil`）时不影响现有更新字段

## 2. 腾讯云 CLB 资源同步（`hc-service/logics/res-sync/tcloud`）可用区提取与回填

- [x] 2.1 在 `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go` 中实现 `getTCloudZones(cloud typeslb.TCloudClb) []string`（公网 `OPEN` 提取非空 `MasterZone.Zone`，内网 `INTERNAL` 提取非空 `cloud.Zones`，无有效值返回 `nil`），并在 `convCloudToDBCreate` 中改用 `getTCloudZones(cloud)` 设置 `lb.Zones`
- [x] 2.2 在 `cmd/hc-service/logics/res-sync/tcloud/load_balancer.go` 的 `convCloudToDBUpdate` 中写入 `Zones: getTCloudZones(cloud)`，并在 `isLBChange` 中增加 `zones := getTCloudZones(cloud); len(zones) > 0 && !slices.Equal(db.Zones, zones)` 的变更判定
- [x] 2.3 在 `cmd/hc-service/logics/res-sync/tcloud/load_balancer_test.go` 中新增单元测试（`TestGetTCloudZones`、扩展转换与变更检测单测），覆盖公网非空/空主可用区、内网可用区过滤空值、新建与更新转换回填、云端返回空时不覆盖本地已有值、本地缺失或不一致时触发变更，运行 `go test -v ./cmd/hc-service/logics/res-sync/tcloud -run "TestGetTCloud|TestConvCloudToDB|TestIsLBChange"` 验证全部通过

## 3. 整体规范与静态检查验证

- [x] 3.1 检查所有变更文件符合 Go 代码规范（单行 <= 120 列、中文注释、无方案代号或单据号腐蚀注释），运行 `gofmt -s -w` 与 `go vet ./pkg/api/data-service/cloud/... ./cmd/data-service/service/cloud/load-balancer/... ./cmd/hc-service/logics/res-sync/tcloud/...` 以及 `go test ./cmd/hc-service/logics/res-sync/tcloud/...` 验证全部通过
