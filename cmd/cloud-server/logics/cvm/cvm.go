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
	"hcm/pkg/client"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/thirdparty/esb"
)

// Interface define cvm interface.
type Interface interface {
	BatchStopCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateAllResult, error)
	BatchDeleteCvm(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateResult, error)
	// DestroyRecycledCvm 销毁已经处于回收状态的Cvm
	DestroyRecycledCvm(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateResult, error)
	GetNotCmdbRecyclableHosts(kt *kit.Kit, bizHostsIds map[int64][]string) ([]string, error)
	// RecyclePreCheck 回收预校验、包含主机状态和CC待回收模块检查
	RecyclePreCheck(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo) *core.BatchOperateAllResult
}

type cvm struct {
	client    *client.ClientSet
	audit     audit.Interface
	eip       eip.Interface
	disk      disk.Interface
	esbClient esb.Client
}

// NewCvm new cvm.
func NewCvm(client *client.ClientSet, audit audit.Interface, eip eip.Interface, disk disk.Interface,
	esbClient esb.Client) Interface {
	return &cvm{
		client:    client,
		audit:     audit,
		eip:       eip,
		disk:      disk,
		esbClient: esbClient,
	}
}
