/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package logicsadmin

import (
	"fmt"

	apisysteminit "hcm/pkg/api/cloud-server/system-init"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/data-service/tenant"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/bkuser"
	cvt "hcm/pkg/tools/converter"
)

// Interface admin logic interface
type Interface interface {
	InitVendorOtherAccount(kt *kit.Kit) (*apisysteminit.OtherAccountInitResult, error)
	GetTenantFromBkUser(kt *kit.Kit) (*bkuser.Tenant, error)
	UpsertLocalTenant(kt *kit.Kit, targetTenant *bkuser.Tenant) (message string, err error)
}

type admin struct {
	c      *client.ClientSet
	bkUser bkuser.Client
}

// NewAdminLogic new admin logic
func NewAdminLogic(c *client.ClientSet, userClient bkuser.Client) Interface {
	return &admin{c: c, bkUser: userClient}
}

// InternalOtherVendorAccountName 内置账号名称
const InternalOtherVendorAccountName = "内置账号"

// InitVendorOtherAccount 查找是否存在vendor为other的账号，若有则返回，没有则创建
func (a *admin) InitVendorOtherAccount(kt *kit.Kit) (*apisysteminit.OtherAccountInitResult, error) {

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("vendor", enumor.Other),
		Page:   core.NewDefaultBasePage(),
	}
	accResp, err := a.c.DataService().Global.Account.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("fail to list other vendor account, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(accResp.Details) > 0 {
		return &apisysteminit.OtherAccountInitResult{ExistsAccountID: accResp.Details[0].ID}, nil
	}

	// 创建other vendor用户
	createReq := &protocloud.AccountCreateReq[protocloud.OtherAccountExtensionCreateReq]{
		Name:     InternalOtherVendorAccountName,
		Managers: []string{"admin"},
		Type:     enumor.ResourceAccount,
		Site:     enumor.InternationalSite,
		Memo:     cvt.ValToPtr(InternalOtherVendorAccountName),
		Extension: &protocloud.OtherAccountExtensionCreateReq{
			CloudID:     string(enumor.Other),
			CloudSecKey: "",
		},
		BkBizIDs: []int64{constant.AttachedAllBiz},
	}
	createResp, err := a.c.DataService().Other.Account.Create(kt, createReq)
	if err != nil {
		logs.Errorf("fail to create other vendor account, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &apisysteminit.OtherAccountInitResult{CreatedAccountID: createResp.ID}, nil
}

// UpsertLocalTenant 插入或更新租户信息
func (a *admin) UpsertLocalTenant(kt *kit.Kit, targetTenant *bkuser.Tenant) (message string, err error) {
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("tenant_id", kt.TenantID),
		Page:   core.NewDefaultBasePage(),
	}
	localTenantResp, err := a.c.DataService().Global.Tenant.List(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list local tenant, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(localTenantResp.Details) > 0 {
		// 2.1 存在则更新
		localTenant := localTenantResp.Details[0]
		status := targetTenant.GetStatus()
		// 	更新租户
		updateReq := &tenant.UpdateTenantReq{Items: []tenant.UpdateTenantField{{
			ID:     localTenant.ID,
			Status: status,
		}}}
		err := a.c.DataService().Global.Tenant.Update(kt, updateReq)
		if err != nil {
			return "", err
		}
		logs.Infof("tenant updated: %s, local id: %s, rid: %s", targetTenant.String(), localTenant.ID, kt.Rid)
		return fmt.Sprintf("tenant update success, %s", localTenant.ID), nil
	}

	// 2.2 不存在则创建
	createReq := &tenant.CreateTenantReq{
		Items: []tenant.CreateTenantField{{
			TenantID: kt.TenantID,
			Status:   targetTenant.GetStatus(),
		}},
	}
	created, err := a.c.DataService().Global.Tenant.Create(kt, createReq)
	if err != nil {
		return "", err
	}
	if len(created.IDs) < 1 {
		return "", fmt.Errorf("tenant created but no any id has been returned")
	}
	createdID := created.IDs[0]
	logs.Infof("tenant created: %s, local id: %s, rid: %s", targetTenant.String(), createdID, kt.Rid)
	return fmt.Sprintf("tenant create success, %s", createdID), nil
}

// GetTenantFromBkUser 尝试从bk user获取租户信息
func (a *admin) GetTenantFromBkUser(kt *kit.Kit) (*bkuser.Tenant, error) {
	if !cc.TenantEnable() {
		return nil, fmt.Errorf("tenant is not enabled")
	}

	// 1. 查找是否是合法租户
	tenantResult, err := a.bkUser.ListTenant(kt)
	if err != nil {
		logs.Errorf("fail to list tenant by bk user, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tenantList := tenantResult.Data
	var targetTenant *bkuser.Tenant
	for _, t := range tenantList {
		if t.Id == kt.TenantID {
			targetTenant = cvt.ValToPtr(t)
			break
		}
	}
	if targetTenant == nil {
		logs.Infof("tenant not found by tenant id: %s, tenant list: %s, rid: %s",
			kt.TenantID, tenantList, kt.Rid)
		return nil, fmt.Errorf("invalid tenant: %s", kt.TenantID)
	}
	return targetTenant, nil
}
