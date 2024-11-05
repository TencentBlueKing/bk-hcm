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
	"hcm/pkg/runtime/filter"
)

// ListNIIDByCvm 查询主机对应的网络接口
func (a *BaseApplicationHandler) ListNIIDByCvm(cvmIDs []string) ([]string, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "cvm_id", Op: filter.In.Factory(), Value: cvmIDs},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.NetworkInterfaceCvmRel.ListNetworkCvmRels(
		a.Cts.Kit,
		&core.ListReq{
			Filter: reqFilter,
			Page:   core.NewDefaultBasePage(),
		},
	)
	if err != nil {
		return nil, err
	}

	niIDs := make([]string, 0, len(resp.Details))
	for _, rel := range resp.Details {
		niIDs = append(niIDs, rel.NetworkInterfaceID)
	}

	return niIDs, nil
}
