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

// Package cvm ...
package cvm

import (
	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/disk"
	"hcm/cmd/cloud-server/logics/eip"
	"hcm/pkg/api/core"
	rr "hcm/pkg/api/core/recycle-record"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
)

// Interface define cvm interface.
type Interface interface {
	BatchStopCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateAllResult, error)
	BatchDeleteCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateResult, error)
	DestroyRecycledCvm(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo,
		records []rr.CvmRecycleRecord) (*core.BatchOperateResult, error)
	GetNotCmdbRecyclableHosts(kt *kit.Kit, bizHostsIds map[int64][]string) ([]string, error)
	RecyclePreCheck(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo) error
	BatchFinalizeRelRecord(kt *kit.Kit, resType enumor.CloudResourceType,
		status enumor.RecycleRecordStatus, resIds []string) error
}

type cvm struct {
	client     *client.ClientSet
	audit      audit.Interface
	eip        eip.Interface
	disk       disk.Interface
	cmdbClient cmdb.Client
}

// NewCvm new cvm.
func NewCvm(client *client.ClientSet, audit audit.Interface, eip eip.Interface, disk disk.Interface,
	cmdbClient cmdb.Client) Interface {
	return &cvm{
		client:     client,
		audit:      audit,
		eip:        eip,
		disk:       disk,
		cmdbClient: cmdbClient,
	}
}

// AssignedCvmInfo assigned cvm info
type AssignedCvmInfo struct {
	CvmID     string `json:"cvm_id"`
	BkBizID   int64  `json:"bk_biz_id"`
	BkCloudID int64  `json:"bk_cloud_id"`
}

// PreviewAssignedCvmInfo preview assigned cvm info
type PreviewAssignedCvmInfo struct {
	CvmID         string
	AccountBizIDs []int64
	Vendor        enumor.Vendor
	CloudID       string
	InnerIPv4     string
	MacAddr       string
}

// PreviewCvmMatchResult preview cvm match result
type PreviewCvmMatchResult struct {
	BkBizID   int64
	BkCloudID int64
}
