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

package disk

import (
	"hcm/pkg/adaptor/types/disk"
	proto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// AwsCreateDisk ...
func AwsCreateDisk(da *diskAdaptor, cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsDiskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := da.adaptor.Aws(cts.Kit, req.Base.AccountID)
	if err != nil {
		return nil, err
	}

	diskSize := int64(req.Base.DiskSize)
	opt := &disk.AwsDiskCreateOption{
		Region:   req.Base.Region,
		Zone:     &req.Base.Zone,
		DiskType: &req.Base.DiskType,
		DiskSize: &diskSize,
	}
	client.CreateDisk(cts.Kit, opt)

	// TODO save to data-service

	return nil, nil
}
