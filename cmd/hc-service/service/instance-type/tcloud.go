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

package instancetype

import (
	typesinstancetype "hcm/pkg/adaptor/types/instance-type"
	proto "hcm/pkg/api/hc-service/instance-type"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListForTCloud ...
func (i *instanceTypeAdaptor) ListForTCloud(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.TCloudInstanceTypeListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := i.adaptor.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typesinstancetype.TCloudInstanceTypeListOption{
		Region: req.Region,
		Zone:   req.Zone,
	}

	its, err := client.ListInstanceType(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor to list tcloud instance type failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	data := make([]*proto.TCloudInstanceTypeResp, 0, len(its))
	for _, one := range its {
		data = append(data, &proto.TCloudInstanceTypeResp{
			Zone:           one.Zone,
			InstanceType:   one.InstanceType,
			InstanceFamily: one.InstanceFamily,
			GPU:            one.GPU,
			CPU:            one.CPU,
			Memory:         one.Memory,
			FPGA:           one.FPGA,
		})
	}

	return data, nil
}
