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
	"hcm/pkg/adaptor/types/core"
	typesinstancetype "hcm/pkg/adaptor/types/instance-type"
	proto "hcm/pkg/api/hc-service/instance-type"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ListForGcp ...
func (i *instanceTypeAdaptor) ListForGcp(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GcpInstanceTypeListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := i.adaptor.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	data := make([]*proto.GcpInstanceTypeResp, 0)
	// 分页遍历获取所有数据
	nextPageToken := ""
	for {
		opt := &typesinstancetype.GcpInstanceTypeListOption{
			Zone: req.Zone,
			Page: &core.GcpPage{PageSize: int64(filter.DefaultMaxInLimit), PageToken: nextPageToken},
		}

		result, err := client.ListInstanceType(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list gcp instance type failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		if len(result.Details) <= 0 {
			logs.Errorf("request adaptor to list gcp instance type num <= 0, rid: %s", cts.Kit.Rid)
			return nil, err
		}

		for _, it := range result.Details {
			data = append(data, toGcpInstanceTypeResp(it))
		}

		// 判断是否还有下一页
		if result.NextPageToken == "" {
			break
		}
		nextPageToken = result.NextPageToken
	}

	return data, nil
}

func toGcpInstanceTypeResp(it *typesinstancetype.GcpInstanceType) *proto.GcpInstanceTypeResp {
	return &proto.GcpInstanceTypeResp{
		InstanceType: it.InstanceType,
		GPU:          it.GPU,
		CPU:          it.CPU,
		Memory:       it.Memory,
		FPGA:         it.FPGA,
	}
}
