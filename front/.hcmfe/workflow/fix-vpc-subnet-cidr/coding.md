# Coding — fix-vpc-subnet-cidr

**TAPD**: 1069995598139198037

## 改什么

`getSaveData` 组初始子网时，腾讯云、华为云、Azure、GCP 的 `ipv4_cidr` 改用表单里的子网网段。

**文件**: `src/views/service/service-apply/hooks/use-vpc-form-data.ts`

## 改动点

四处 `saveData.subnet.ipv4_cidr` 从 `formData.ipv4_cidr` 改为 `subnet.ipv4_cidr`：

- 腾讯云
- 华为云
- Azure
- GCP

保持不动：

- `saveData.ipv4_cidr = formData.ipv4_cidr`（VPC 网段）
- 华为云 `gateway_ip` 已经按 `subnet.ipv4_cidr` 计算
- AWS 不组初始子网
- 不恢复「子网必须落在 VPC 内」的前端校验
- 不改后端，不改历史单据

## 验收

- VPC `x.x.x.x/21`、子网 `x.x.x.x/22` 时，提交体里 VPC 为 `/21`，子网为 `/22`
- 四朵云同一规则
- 用户未再改子网、两条网段相同时，提交体里的子网网段仍与表单一致
