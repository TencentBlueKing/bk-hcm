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
	apisysteminit "hcm/pkg/api/cloud-server/system-init"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// Interface admin logic interface
type Interface interface {
	InitVendorOtherAccount(kt *kit.Kit) (*apisysteminit.OtherAccountInitResult, error)
}

type admin struct {
	c *client.ClientSet
}

// NewAdminLogic new admin logic
func NewAdminLogic(c *client.ClientSet) Interface {
	return &admin{c: c}
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
		UsageBizIDs: []int64{constant.AttachedAllBiz},
	}
	createResp, err := a.c.DataService().Other.Account.Create(kt, createReq)
	if err != nil {
		logs.Errorf("fail to create other vendor account, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &apisysteminit.OtherAccountInitResult{CreatedAccountID: createResp.ID}, nil
}
