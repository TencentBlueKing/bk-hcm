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
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/runtime/filter"
)

// ListDiskIDByCvm 查询主机对应的硬盘
func (a *BaseApplicationHandler) ListDiskIDByCvm(cvmIDs []string) ([]string, error) {
	reqFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: "cvm_id", Op: filter.In.Factory(), Value: cvmIDs},
		},
	}
	// 查询
	resp, err := a.Client.DataService().Global.ListDiskCvmRel(
		a.Cts.Kit,
		&dataproto.DiskCvmRelListReq{
			Filter: reqFilter,
			Page:   &core.BasePage{Count: false, Start: 0, Limit: uint(len(cvmIDs) * 20)}, // 每台主机最多挂20块硬盘
		},
	)
	if err != nil {
		return nil, err
	}

	diskIDs := make([]string, 0, len(resp.Details))
	for _, rel := range resp.Details {
		diskIDs = append(diskIDs, rel.DiskID)
	}

	return diskIDs, nil
}
