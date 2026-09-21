/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package tcloud

import (
	"fmt"
	"strings"

	loadbalancer "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

type formItem struct {
	Label string
	Value string
}

// RenderItsmTitle 渲染ITSM单据标题
func (a *ApplicationOfCreateTCloudLB) RenderItsmTitle() (string, error) {
	return fmt.Sprintf("申请新增[%s]负载均衡(%s)", a.Vendor().GetNameZh(), cvt.PtrToVal(a.req.Name)), nil
}

// RenderItsmForm 渲染ITSM表单
func (a *ApplicationOfCreateTCloudLB) RenderItsmForm() (string, error) {
	req := a.req

	formItems := make([]formItem, 0)

	// 基本通用信息
	baseInfoFormItems, err := a.renderBaseInfo()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, baseInfoFormItems...)

	exclusiveFormItems, err := a.renderExclusiveClusterForm()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, exclusiveFormItems...)

	// 网络
	networkFormItems, err := a.renderNetwork()
	if err != nil {
		return "", err
	}
	formItems = append(formItems, networkFormItems...)

	// 计费
	formItems = append(formItems, a.renderInstanceChargeForm()...)

	// 购买数量
	count := cvt.PtrToVal(req.RequireCount)
	if count == 0 {
		count = 1
	}
	formItems = append(formItems, formItem{Label: "购买数量", Value: fmt.Sprintf("%d", count)})

	// 备注
	if req.Memo != "" {
		formItems = append(formItems, formItem{Label: "备注", Value: req.Memo})
	}

	// 转换为ITSM表单内容数据
	content := make([]string, 0, len(formItems))
	for _, i := range formItems {
		content = append(content, fmt.Sprintf("%s: %s", i.Label, i.Value))
	}
	return strings.Join(content, "\n"), nil
}

func (a *ApplicationOfCreateTCloudLB) renderBaseInfo() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 业务
	bizName, err := a.GetBizName(req.BkBizID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "业务", Value: bizName})

	// 云账号
	accountInfo, err := a.GetAccount(req.AccountID)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云账号", Value: accountInfo.Name})

	// 云厂商
	formItems = append(formItems, formItem{Label: "云厂商", Value: a.Vendor().GetNameZh()})

	// 云地域
	regionInfo, err := a.GetTCloudRegion(req.Region)
	if err != nil {
		return formItems, err
	}
	formItems = append(formItems, formItem{Label: "云地域", Value: regionInfo.RegionName})

	// 可用区
	zones := append(req.Zones, req.BackupZones...)
	if len(zones) > 0 {
		zoneInfos, err := a.GetZones(a.Vendor(), req.Region, zones)
		if err != nil {
			return formItems, err
		}
		zoneStr := strings.Builder{}
		for i, info := range zoneInfos {
			if i > 0 {
				zoneStr.WriteRune(',')
			}
			zoneStr.WriteString(info.Name)
			if len(info.NameCn) != 0 {
				zoneStr.WriteRune('(')
				zoneStr.WriteString(info.NameCn)
				zoneStr.WriteRune(')')
			}
		}
		formItems = append(formItems, formItem{Label: "可用区", Value: zoneStr.String()})
	}

	// 名称
	formItems = append(formItems, formItem{Label: "名称", Value: cvt.PtrToVal(req.Name)})

	// 规格：独占型由 exclusive 合成，空 sla_type 展示共享型，其余透传性能容量档位
	formItems = append(formItems, formItem{Label: "规格", Value: renderSlaType(req)})

	// 运营商
	isp := "BGP"
	if req.VipIsp != nil {
		isp = *req.VipIsp
	}
	if req.LoadBalancerType == loadbalancer.InternalLoadBalancerType {
		isp = "内网流量"
	}
	formItems = append(formItems, formItem{Label: "运营商", Value: isp})

	if req.InternetMaxBandwidthOut != nil {
		formItems = append(formItems,
			formItem{Label: "带宽", Value: fmt.Sprintf("%dMbps", cvt.PtrToVal(req.InternetMaxBandwidthOut))})
	}

	return formItems, nil
}

func (a *ApplicationOfCreateTCloudLB) renderNetwork() ([]formItem, error) {
	req := a.req
	formItems := make([]formItem, 0)

	// 内网公网类型
	formItems = append(formItems, formItem{Label: "类型", Value: LoadBalancerTypeMap[req.LoadBalancerType]})

	formItems = append(formItems, formItem{Label: "IP版本", Value: IPVersionNameMap[req.AddressIPVersion]})

	// VPC
	vpcName := "未指定-默认VPC"
	if req.CloudVpcID != nil {
		vpcInfo, err := a.GetVpc(a.Vendor(), req.AccountID, cvt.PtrToVal(req.CloudVpcID))
		if err != nil {
			return formItems, err
		}
		vpcName = fmt.Sprintf("%s(%s)", vpcInfo.CloudID, vpcInfo.Name)
	}
	formItems = append(formItems, formItem{Label: "VPC", Value: vpcName})

	// subnet
	subnetName := "未指定"
	if req.CloudVpcID != nil && req.CloudSubnetID != nil {
		// 子网
		subnetInfo, err := a.GetSubnet(a.Vendor(), req.AccountID, cvt.PtrToVal(req.CloudVpcID),
			cvt.PtrToVal(req.CloudSubnetID))
		if err != nil {
			return formItems, err
		}
		subnetName = fmt.Sprintf("%s(%s)", subnetInfo.CloudID, subnetInfo.Name)
	}
	formItems = append(formItems, formItem{Label: "子网", Value: subnetName})

	// EIP信息
	eipID := cvt.PtrToVal(req.CloudEipID)
	if len(eipID) > 0 {
		eipInfo, err := a.GetEip(a.Vendor(), req.AccountID, eipID)
		if err != nil {
			return formItems, err
		}
		formItems = append(formItems, formItem{
			Label: "EIP",
			Value: fmt.Sprintf("%s(%s)", eipInfo.PublicIp, eipInfo.CloudID),
		})
	}

	return formItems, nil
}

func (a *ApplicationOfCreateTCloudLB) renderInstanceChargeForm() []formItem {
	req := a.req
	formItems := make([]formItem, 0)

	payMode := "按量计费"
	if req.InternetChargeType != nil {
		payMode = LoadBalancerNetworkChargeTypeNameMap[*req.InternetChargeType]
	}
	if payMode == "" {
		payMode = string(*req.InternetChargeType)
	}
	// 计费模式
	formItems = append(formItems, formItem{Label: "网络计费模式", Value: payMode})

	// 是否自动续费
	if req.AutoRenew != nil && *req.AutoRenew {
		formItems = append(formItems, formItem{Label: "是否自动续费", Value: "是"})
	} else {
		formItems = append(formItems, formItem{Label: "是否自动续费", Value: "否"})
	}

	return formItems
}

// renderSlaType 合成 ITSM「规格」展示：独占型优先于 sla_type；空 sla_type 视为共享型。
func renderSlaType(req *hclb.TCloudLoadBalancerCreateReq) string {
	if req.IsExclusive() {
		return exclusiveSlaTypeName
	}
	if slaType := cvt.PtrToVal(req.SlaType); slaType != "" {
		return slaType
	}
	return sharedSlaTypeName
}

// renderExclusiveClusterForm 独占型申请单追加七层标签、指定 VIP、四层集群信息；非独占型不追加。
func (a *ApplicationOfCreateTCloudLB) renderExclusiveClusterForm() ([]formItem, error) {
	if !a.req.IsExclusive() {
		return nil, nil
	}

	return renderExclusiveClusterItems(cvt.PtrToVal(a.req.ClusterTag), cvt.PtrToVal(a.req.Vip),
		a.listTgwClusterItsmLines()), nil
}

// renderExclusiveClusterItems 组装独占集群相关的 ITSM 表单项，便于单测覆盖展示口径。
func renderExclusiveClusterItems(clusterTag, vip string, tgwLines []string) []formItem {
	tgwValue := emptyItsmValue
	if len(tgwLines) != 0 {
		tgwValue = strings.Join(tgwLines, "；")
	}

	return []formItem{
		{Label: "七层独占集群标签", Value: emptyOrItsmValue(clusterTag)},
		{Label: "指定VIP", Value: emptyOrItsmValue(vip)},
		{Label: "四层独占集群", Value: tgwValue},
	}
}

// listTgwClusterItsmLines 按 cloud_cluster_ids 查本地四层独占集群，拼出 ITSM 展示行；查询失败或本地未命中时名称
// 使用「本地未同步」，仍带上云上 ID，不阻断提单。
func (a *ApplicationOfCreateTCloudLB) listTgwClusterItsmLines() []string {
	cloudClusterIDs := a.req.CloudClusterIDs
	if len(cloudClusterIDs) == 0 {
		return nil
	}

	localByCloudID := a.listLocalTgwClusters(cloudClusterIDs)
	lines := make([]string, 0, len(cloudClusterIDs))
	for _, cloudID := range cloudClusterIDs {
		local := localByCloudID[cloudID]
		lines = append(lines, formatTgwClusterItsmLine(local.Name, cloudID, local.ClusterTag))
	}
	return lines
}

// listLocalTgwClusters 按云上 ID 查询当前业务下的四层独占集群，查询失败返回空映射。
func (a *ApplicationOfCreateTCloudLB) listLocalTgwClusters(cloudClusterIDs []string) map[string]localTgwItsmCluster {
	req := &core.ListReq{
		Fields: []string{"cloud_id", "name", "cluster_tag"},
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("bk_biz_id", a.req.BkBizID),
			tools.RuleEqual("cluster_type", enumor.TGWClusterType),
			tools.RuleIn("cloud_id", cloudClusterIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := a.Client.DataService().Global.ListExclusiveCluster(a.Cts.Kit, req)
	if err != nil {
		logs.Errorf("list exclusive cluster for itsm form failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return nil
	}

	localByCloudID := make(map[string]localTgwItsmCluster, len(result.Details))
	for _, one := range result.Details {
		localByCloudID[one.CloudID] = localTgwItsmCluster{Name: one.Name, ClusterTag: one.ClusterTag}
	}
	return localByCloudID
}

// localTgwItsmCluster 四层独占集群用于 ITSM 展示的本地字段。
type localTgwItsmCluster struct {
	Name       string
	ClusterTag string
}

// formatTgwClusterItsmLine 格式化单条四层独占集群 ITSM 展示，口径与申请单详情页一致。
func formatTgwClusterItsmLine(name, cloudID, clusterTag string) string {
	if name == "" {
		name = localUnsyncedName
	}
	if clusterTag == "" {
		return fmt.Sprintf("%s（云上ID：%s）", name, cloudID)
	}
	return fmt.Sprintf("%s（云上ID：%s，标签：%s）", name, cloudID, clusterTag)
}

// emptyOrItsmValue 空字符串时返回 ITSM 占位符。
func emptyOrItsmValue(value string) string {
	if value == "" {
		return emptyItsmValue
	}
	return value
}
