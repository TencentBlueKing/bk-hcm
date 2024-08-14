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

package huawei

import (
	csdisk "hcm/pkg/api/cloud-server/disk"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/thirdparty/api-gateway/itsm"
)

// PrepareReq ...
func (a *ApplicationOfCreateHuaWeiDisk) PrepareReq() error {
	return nil
}

// GenerateApplicationContent 获取预处理过的数据，以interface格式
func (a *ApplicationOfCreateHuaWeiDisk) GenerateApplicationContent() interface{} {
	// 需要将Vendor也存储进去
	return &struct {
		*csdisk.HuaWeiDiskCreateReq `json:",inline"`
		Vendor                      enumor.Vendor `json:"vendor"`
	}{
		HuaWeiDiskCreateReq: a.req,
		Vendor:              a.Vendor(),
	}
}

// PrepareReqFromContent ...
func (a *ApplicationOfCreateHuaWeiDisk) PrepareReqFromContent() error {
	return nil
}

// GetItsmApprover 获取itsm审批人
func (a *ApplicationOfCreateHuaWeiDisk) GetItsmApprover(managers []string) []itsm.VariableApprover {
	return a.GetItsmPlatformAndAccountApprover(managers, a.req.AccountID)
}

// GetBkBizIDs 获取当前的业务IDs
func (a *ApplicationOfCreateHuaWeiDisk) GetBkBizIDs() []int64 {
	return []int64{a.req.BkBizID}
}
