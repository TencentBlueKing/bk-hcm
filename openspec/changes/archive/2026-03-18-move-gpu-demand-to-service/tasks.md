## 1. Service层新建GPU需求业务逻辑文件

- [x] 1.1 在 `cmd/woa-server/service/plan/` 下新建 `gpu_demand_order.go` 文件，将 logics 层 `gpu_demand_order.go` 中的所有业务方法和辅助函数迁移过来，将 `(c *Controller)` 接收器改为 `(s *service)`，将 `c.client` 改为 `s.client`
- [x] 1.2 将不依赖结构体字段的纯函数（`validateOrderDetails`、`validateDetailFixedFields`、`validateDetailExtension`、`buildDetails`、`convertCellValue`、`convertEnumValue`）保持为包级函数
- [x] 1.3 优化迁移后方法的命名和注释，使函数名更简洁、注释更清晰地表达功能（如 `ExcelImportGpuDemand` → `parseGpuDemandExcel`，`CreateGpuDemandOrder` → `createGpuDemandOrder`，`OverwriteGpuDemandOrder` → `overwriteGpuDemandOrder`），方法不再导出因为不需要通过接口暴露

## 2. 修改Service层Handler调用方式

- [x] 2.1 修改 `cmd/woa-server/service/plan/gpu_demand_excel.go` 中的 `ExcelImportGpuDemand` handler，将 `s.planController.ExcelImportGpuDemand(...)` 改为 `s.parseGpuDemandExcel(...)`
- [x] 2.2 修改 `cmd/woa-server/service/plan/gpu_demand_excel.go` 中的 `CreateGpuDemandOrder` handler，将 `s.planController.CreateGpuDemandOrder(...)` 改为 `s.createGpuDemandOrder(...)`
- [x] 2.3 修改 `cmd/woa-server/service/plan/gpu_demand_excel.go` 中的 `OverwriteGpuDemandOrder` handler，将 `s.planController.OverwriteGpuDemandOrder(...)` 改为 `s.overwriteGpuDemandOrder(...)`

## 3. 清理Logics层

- [x] 3.1 从 `cmd/woa-server/logics/plan/plan.go` 的 `Logics` 接口定义中移除 `ExcelImportGpuDemand`、`CreateGpuDemandOrder`、`OverwriteGpuDemandOrder` 三个方法声明
- [x] 3.2 删除 `cmd/woa-server/logics/plan/gpu_demand_order.go` 文件（所有内容已迁移到 service 层）
- [x] 3.3 清理 `plan.go` 中因移除方法而不再需要的 import 项（如 `io`、`woaapi` 如果只被 GPU 方法使用的话）

## 4. 验证与编译

- [x] 4.1 运行 `go build ./cmd/woa-server/...` 确认编译通过，无未解析的引用
- [x] 4.2 检查是否有其他代码通过 `plan.Logics` 接口调用了这三个被移除的方法，如有则同步修改
- [x] 4.3 运行相关单元测试确认无回归
