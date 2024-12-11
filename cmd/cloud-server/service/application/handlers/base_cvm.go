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
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/tools/slice"
)

// ListCvm 查询主机列表
func (a *BaseApplicationHandler) ListCvm(
	vendor enumor.Vendor, accountID, region string, cloudCvmIDs []string,
) ([]corecvm.BaseCvm, error) {

	result := make([]corecvm.BaseCvm, 0)
	for _, ids := range slice.Split(cloudCvmIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleEqual("account_id", accountID),
				tools.RuleIn("cloud_id", ids),
				tools.RuleEqual("region", region),
			),
			Page: core.NewDefaultBasePage(),
		}
		resp, err := a.Client.DataService().Global.Cvm.ListCvm(a.Cts.Kit, listReq)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.Details...)
	}
	return result, nil
}
