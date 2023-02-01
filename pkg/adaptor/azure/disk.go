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

package azure

import (
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/kit"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// CreateDisk 创建云硬盘
// reference: https://learn.microsoft.com/en-us/rest/api/compute/disks/list?source=recommendations&tabs=Go#disklist
func (a *Azure) CreateDisk(kt *kit.Kit, opt *disk.AzureDiskCreateOption) (*armcompute.Disk, error) {
	return a.createDisk(kt, opt)
}

func (a *Azure) createDisk(kt *kit.Kit, opt *disk.AzureDiskCreateOption) (*armcompute.Disk, error) {
	client, err := a.clientSet.diskClient()
	if err != nil {
		return nil, err
	}

	diskReq, err := opt.ToCreateDiskRequest()
	if err != nil {
		return nil, err
	}

	pollerResp, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.Name, *diskReq, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Disk, nil
}
