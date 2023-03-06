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
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/model"
)

// ListEip ...
// reference: https://support.huaweicloud.com/api-eip/eip_api_0003.html
func (h *HuaWei) ListEip(kt *kit.Kit, opt *eip.HuaWeiEipListOption) (*eip.HuaWeiEipListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := h.clientSet.eipClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(model.ListPublicipsRequest)

	if len(opt.CloudIDs) > 0 {
		req.Id = &opt.CloudIDs
	}

	if len(opt.Ips) > 0 {
		req.PublicIpAddress = &opt.Ips
	}

	if opt.Limit != nil {
		req.Limit = opt.Limit
	}

	if opt.Marker != nil {
		req.Marker = opt.Marker
	}

	resp, err := client.ListPublicips(req)
	if err != nil {
		logs.Errorf("list huawei eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	eips := make([]*eip.HuaWeiEip, len(*resp.Publicips))
	for idx, publicIp := range *resp.Publicips {
		status := publicIp.Status.Value()
		eips[idx] = &eip.HuaWeiEip{
			CloudID:       *publicIp.Id,
			Region:        opt.Region,
			Status:        &status,
			PublicIp:      publicIp.PublicIpAddress,
			PrivateIp:     publicIp.PrivateIpAddress,
			PortID:        publicIp.PortId,
			BandwidthId:   publicIp.BandwidthId,
			BandwidthName: publicIp.BandwidthName,
			BandwidthSize: publicIp.BandwidthSize,
		}
	}

	return &eip.HuaWeiEipListResult{Details: eips}, nil
}
