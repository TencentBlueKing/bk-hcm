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

package datasvc

import (
	dataproto "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
)

// DiskCvmRelManager ...
type DiskCvmRelManager struct {
	DiskID  string
	CvmID   string
	DataCli *dataservice.Client
}

// Create ...
func (m *DiskCvmRelManager) Create(kt *kit.Kit) error {
	req := &dataproto.DiskCvmRelBatchCreateReq{
		Rels: []dataproto.DiskCvmRelCreateReq{{DiskID: m.DiskID, CvmID: m.CvmID}},
	}
	return m.DataCli.Global.BatchCreateDiskCvmRel(kt.Ctx, kt.Header(), req)
	// TODO 更新主机和云盘状态
}

// Delete ...
func (m *DiskCvmRelManager) Delete(kt *kit.Kit) error {
	req := &dataproto.DiskCvmRelDeleteReq{Filter: &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "disk_id",
				Op:    filter.Equal.Factory(),
				Value: m.DiskID,
			}, &filter.AtomRule{Field: "cvm_id", Op: filter.Equal.Factory(), Value: m.CvmID},
		},
	}}
	return m.DataCli.Global.DeleteDiskCvmRel(kt.Ctx, kt.Header(), req)
	// TODO 更新主机和云盘状态
}
