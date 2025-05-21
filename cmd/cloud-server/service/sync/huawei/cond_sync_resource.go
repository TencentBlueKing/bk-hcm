/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// CondSyncParams 条件同步选项
type CondSyncParams struct {
	AccountID string   `json:"account_id" validate:"required"`
	Regions   []string `json:"regions,omitempty" validate:"max=20"`
}

// CondSyncFunc sync resource by given condition
type CondSyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error

var condSyncFuncMap = map[enumor.CloudResourceType]CondSyncFunc{
	enumor.SecurityGroupCloudResType: CondSyncSecurityGroup,
}

// GetCondSyncFunc ...
func GetCondSyncFunc(res enumor.CloudResourceType) (syncFunc CondSyncFunc, ok bool) {
	syncFunc, ok = condSyncFuncMap[res]
	return syncFunc, ok
}

// CondSyncSecurityGroup ...
func CondSyncSecurityGroup(kt *kit.Kit, cliSet *client.ClientSet, params *CondSyncParams) error {
	syncReq := sync.HuaWeiSyncReq{
		AccountID: params.AccountID,
	}
	for i := range params.Regions {
		syncReq.Region = params.Regions[i]
		err := cliSet.HCService().HuaWei.SecurityGroup.SyncSecurityGroup(kt.Ctx, kt.Header(), &syncReq)
		if err != nil {
			logs.Errorf("[%s] conditional sync security group failed, err: %v, req: %+v, rid: %s",
				enumor.HuaWei, err, syncReq, kt.Rid)
			return err
		}
		logs.Infof("[%s] conditional sync security group end, req: %+v, rid: %s", enumor.HuaWei, syncReq, kt.Rid)
	}
	return nil
}
