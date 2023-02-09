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

package tcloud

import (
	"fmt"

	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
)

// CreateDisk 创建云硬盘
// reference: https://cloud.tencent.com/document/api/362/16312
func (t *TCloud) CreateDisk(kt *kit.Kit, opt *disk.TCloudDiskCreateOption) (*cbs.CreateDisksResponse, error) {
	return t.createDisk(kt, opt)
}

func (t *TCloud) createDisk(kt *kit.Kit, opt *disk.TCloudDiskCreateOption) (*cbs.CreateDisksResponse, error) {
	client, err := t.clientSet.cbsClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req, err := opt.ToCreateDisksRequest()
	if err != nil {
		return nil, err
	}

	return client.CreateDisksWithContext(kt.Ctx, req)
}

// ListDisk 查询云硬盘列表
// reference: https://cloud.tencent.com/document/api/362/16315
func (t *TCloud) ListDisk(kt *kit.Kit, opt *disk.TCloudDiskListOption) ([]*cbs.Disk, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cbsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud cbs client failed, err: %v", err)
	}

	req := cbs.NewDescribeDisksRequest()
	if opt.Page != nil {
		req.Offset = &opt.Page.Offset
		req.Limit = &opt.Page.Limit
	}

	resp, err := client.DescribeDisksWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Response.DiskSet, nil

}
