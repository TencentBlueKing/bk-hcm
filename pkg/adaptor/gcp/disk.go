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

package gcp

import (
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"google.golang.org/api/compute/v1"
)

// CreateDisk 创建云硬盘
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/disks/insert
func (g *Gcp) CreateDisk(kt *kit.Kit, opt *disk.GcpDiskCreateOption) (*compute.Operation, error) {
	return g.createDisk(kt, opt)
}

func (g *Gcp) createDisk(kt *kit.Kit, opt *disk.GcpDiskCreateOption) (*compute.Operation, error) {
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	cloudProjectID := g.clientSet.credential.CloudProjectID
	req, err := opt.ToCreateDiskRequest(cloudProjectID)
	if err != nil {
		return nil, err
	}

	var call *compute.DisksInsertCall
	call = client.Disks.Insert(cloudProjectID, opt.Zone, req).Context(kt.Ctx)
	return call.Do()
}

// ListDisk 查看云硬盘
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/disks/list
func (g *Gcp) ListDisk(kit *kit.Kit, opt *disk.GcpDiskListOption) ([]*compute.Disk, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "gcp disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kit)
	if err != nil {
		return nil, err
	}
	cloudProjectID := g.clientSet.credential.CloudProjectID

	yDisks := []*compute.Disk{}
	req := client.Disks.List(cloudProjectID, opt.Region)
	if err := req.Pages(kit.Ctx, func(page *compute.DiskList) error {
		for _, disk := range page.Items {
			yDisks = append(yDisks, disk)
		}
		return nil
	}); err != nil {
		logs.Errorf("failed to list disk, err: %v, rid: %s", err, kit.Rid)
	}

	return yDisks, nil
}
