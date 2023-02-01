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

package huawei

import (
	"hcm/pkg/adaptor/types/disk"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/region"
)

// CreateDisk 创建云硬盘
// reference: https://support.huaweicloud.com/api-evs/evs_04_2003.html
func (h *Huawei) CreateDisk(opt *disk.HuaWeiDiskCreateOption) (*model.CreateVolumeResponse, error) {
	return h.createDisk(opt)
}

func (h *Huawei) createDisk(opt *disk.HuaWeiDiskCreateOption) (*model.CreateVolumeResponse, error) {
	client, err := h.clientSet.evsClient(region.ValueOf(opt.Region))
	if err != nil {
		return nil, err
	}

	req, err := opt.ToCreateVolumeRequest()
	if err != nil {
		return nil, err
	}

	return client.CreateVolume(req)
}
