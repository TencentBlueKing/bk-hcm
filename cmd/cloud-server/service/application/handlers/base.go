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

package handlers

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	corecloudzone "hcm/pkg/api/core/cloud/zone"
	dataproto "hcm/pkg/api/data-service/cloud"
	dataprotoimage "hcm/pkg/api/data-service/cloud/image"
	dataprotoregion "hcm/pkg/api/data-service/cloud/region"
	dataprotozone "hcm/pkg/api/data-service/cloud/zone"
	hcprotoinstancetype "hcm/pkg/api/hc-service/instance-type"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/cryptography"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// ApplicationHandler 定义了申请单的表单校验，与itsm对接、审批通过后的资源交付函数
// 创建申请单：CheckReq -> PrepareReq -> CreateITSMTicket -> GenerateApplicationContent -> "SaveToDB"
// 审批通过交付："LoadApplicationFromDB" -> CheckReq -> PrepareReqForContent -> Deliver -> "UpdateStatusToDB"
// Note: 这里创建申请单的请求数据和交付资源的请求数据结构是一样的，这是一种"偷懒"行为，
//  更好的方式是Handler拆分成两种抽象：申请单创建者Creator、申请单交付者Deliverer，然后定义各自的数据结构
type ApplicationHandler interface {
	GetType() enumor.ApplicationType

	// CheckReq 申请单的表单校验
	CheckReq() error
	// PrepareReq 预处理申请单数据
	PrepareReq() error
	// CreateITSMTicket 与ITSM对接，创建ITSM单据
	CreateITSMTicket(serviceID int64, callbackUrl string) (string, error)
	// GenerateApplicationContent 生成存储到DB的申请单内容，Interface格式，便于统一处理
	GenerateApplicationContent() interface{}

	// PrepareReqFromContent 申请单内容从DB里获取后可以进行预处理，便于资源交付时资源请求
	PrepareReqFromContent() error
	// Deliver  审批通过后资源的交付
	Deliver() (status enumor.ApplicationStatus, deliverDetail map[string]interface{}, err error)
}

// BaseApplicationHandler 基础的Handler 一些公共函数和属性处理，可以给到其他具体Handler组合
type BaseApplicationHandler struct {
	ApplicationType enumor.ApplicationType

	Cts       *rest.Contexts
	Client    *client.ClientSet
	EsbClient esb.Client
	Cipher    cryptography.Crypto
}

// GetType 申请单类型
func (a *BaseApplicationHandler) GetType() enumor.ApplicationType {
	return a.ApplicationType
}

// ListBizNames 查询业务名称列表
func (a *BaseApplicationHandler) ListBizNames(bkBizIDs []int64) ([]string, error) {
	// 查询CC业务
	searchResp, err := a.EsbClient.Cmdb().SearchBusiness(a.Cts.Kit.Ctx, &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
	})
	if err != nil {
		return []string{}, fmt.Errorf("call cmdb search business api failed, err: %v", err)
	}
	// 业务ID和Name映射关系
	bizNameMap := map[int64]string{}
	for _, biz := range searchResp.SearchBizResult.Info {
		bizNameMap[biz.BizID] = biz.BizName
	}
	// 匹配出业务名称列表
	bizNames := make([]string, 0, len(bkBizIDs))
	for _, bizID := range bkBizIDs {
		bizNames = append(bizNames, bizNameMap[bizID])
	}

	return bizNames, nil
}

// GetBizName 查询业务名称
func (a *BaseApplicationHandler) GetBizName(bkBizID int64) (string, error) {
	bizNames, err := a.ListBizNames([]int64{bkBizID})
	if err != nil || len(bizNames) != 1 {
		return "", err
	}

	return bizNames[0], nil
}

// GetCloudAreaName 查询云区域名称
func (a *BaseApplicationHandler) GetCloudAreaName(bkCloudAreaID int64) (string, error) {
	res, err := a.EsbClient.Cmdb().SearchCloudArea(
		a.Cts.Kit.Ctx,
		&cmdb.SearchCloudAreaParams{
			Fields: []string{"bk_cloud_id", "bk_cloud_name"},
			Page: cmdb.BasePage{
				Limit: 1,
				Start: 0,
				Sort:  "bk_cloud_id",
			},
			Condition: map[string]interface{}{"bk_cloud_id": bkCloudAreaID},
		},
	)
	if err != nil {
		return "", fmt.Errorf("call cmdb search cloud area api failed, err: %v", err)
	}

	for _, cloudArea := range res.Info {
		if cloudArea.CloudID == bkCloudAreaID {
			return cloudArea.CloudName, nil
		}
	}

	return "", fmt.Errorf("not found bk cloud area by bk_cloud_area_id(%d)", bkCloudAreaID)
}

func (a *BaseApplicationHandler) getPageOfOneLimit() *core.BasePage {
	return &core.BasePage{Count: false, Start: 0, Limit: 1}
}

// GetAccount 查询账号信息
func (a *BaseApplicationHandler) GetAccount(accountID string) (*dataproto.BaseAccountListResp, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "id", Op: filter.Equal.Factory(), Value: accountID},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Account.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataproto.AccountListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found account by id(%s)", accountID)
	}

	return resp.Details[0], nil
}

// GetTCloudRegion 查询云地域信息
func (a *BaseApplicationHandler) GetTCloudRegion(region string) (*corecloud.TCloudRegion, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "region_id", Op: filter.Equal.Factory(), Value: region},
		},
	}
	// 查询
	resp, err := a.Client.DataService().TCloud.Region.ListRegion(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotoregion.TCloudRegionListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found tcloud region by region_id(%s)", region)
	}

	return &resp.Details[0], nil
}

// GetZone 查询可用区
func (a *BaseApplicationHandler) GetZone(vendor enumor.Vendor, region, zone string) (*corecloudzone.BaseZone, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: region},
			filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: zone},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Zone.ListZone(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotozone.ZoneListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s zone by region(%s) and zone cloud_id(%s)", vendor, region, zone)
	}

	return &resp.Details[0], nil
}

// GetTCloudInstanceType 查询机型
func (a *BaseApplicationHandler) GetTCloudInstanceType(
	accountID, region, zone, instanceType string,
) (*hcprotoinstancetype.TCloudInstanceTypeResp, error) {
	resp, err := a.Client.HCService().TCloud.InstanceType.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&hcprotoinstancetype.TCloudInstanceTypeListReq{AccountID: accountID, Region: region, Zone: zone},
	)
	if err != nil {
		return nil, err
	}

	// 遍历查找
	for _, i := range resp {
		if i.InstanceType == instanceType {
			return i, nil
		}
	}

	return nil, fmt.Errorf(
		"not found tcloud instanceType by accountID(%s), region(%s), zone (%s)",
		accountID, region, zone,
	)
}

// ConvertMemoryMBToGB 将内存的MB转换为可用于展示的GB, 特殊展示，不适合其他通用的转换
func (a *BaseApplicationHandler) ConvertMemoryMBToGB(m int64) string {
	if m%1024 == 0 {
		return fmt.Sprintf("%d", m/1024)
	}

	return fmt.Sprintf("%.1f", float64(m/1024))
}

// GetImage 查询镜像
func (a *BaseApplicationHandler) GetImage(
	vendor enumor.Vendor, cloudImageID string,
) (*dataprotoimage.ImageResult, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: cloudImageID},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.ListImage(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataprotoimage.ImageListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s image by cloud_id(%s)", vendor, cloudImageID)
	}

	return resp.Details[0], nil
}

// GetVpc 查询VPC
func (a *BaseApplicationHandler) GetVpc(
	vendor enumor.Vendor, accountID, cloudVpcID string,
) (*corecloud.BaseVpc, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Vpc.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&core.ListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s vpc by cloud_id(%s)", vendor, cloudVpcID)
	}

	return &resp.Details[0], nil
}

// GetSubnet 查询子网
func (a *BaseApplicationHandler) GetSubnet(
	vendor enumor.Vendor, accountID, cloudVpcID, cloudSubnetID string,
) (*corecloud.BaseSubnet, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			filter.AtomRule{Field: "cloud_vpc_id", Op: filter.Equal.Factory(), Value: cloudVpcID},
			filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: cloudSubnetID},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.Subnet.List(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&core.ListReq{
			Filter: reqFilter,
			Page:   a.getPageOfOneLimit(),
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found %s subnet by cloud_id(%s)", vendor, cloudSubnetID)
	}

	return &resp.Details[0], nil
}

// ListSecurityGroup 查询安全组列表
func (a *BaseApplicationHandler) ListSecurityGroup(
	vendor enumor.Vendor, accountID string, cloudSecurityGroupIDs []string,
) ([]corecloud.BaseSecurityGroup, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: cloudSecurityGroupIDs},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.SecurityGroup.ListSecurityGroup(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&dataproto.SecurityGroupListReq{
			Filter: reqFilter,
			Page:   &core.BasePage{Count: false, Start: 0, Limit: uint(len(cloudSecurityGroupIDs))},
		},
	)
	if err != nil {
		return nil, err
	}

	return resp.Details, nil
}
