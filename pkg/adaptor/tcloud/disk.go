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
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/kit"

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
