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
	typesinstancetype "hcm/pkg/adaptor/types/instance-type"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// ListInstanceType ...
// reference: https://cloud.tencent.com/document/api/213/17378
func (t *TCloudImpl) ListInstanceType(kt *kit.Kit, opt *typesinstancetype.TCloudInstanceTypeListOption) (
	[]typesinstancetype.TCloudInstanceType, error,
) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := cvm.NewDescribeZoneInstanceConfigInfosRequest()
	req.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("zone"),
			Values: []*string{&opt.Zone},
		},
		{
			Name:   common.StringPtr("instance-charge-type"),
			Values: []*string{&opt.InstanceChargeType},
		},
	}

	resp, err := client.DescribeZoneInstanceConfigInfosWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud instance type failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	its := make([]typesinstancetype.TCloudInstanceType, 0, len(resp.Response.InstanceTypeQuotaSet))
	for _, it := range resp.Response.InstanceTypeQuotaSet {
		its = append(its, typesinstancetype.TCloudInstanceType{
			InstanceType:      converter.PtrToVal(it.InstanceType),
			InstanceFamily:    converter.PtrToVal(it.InstanceFamily),
			GPU:               converter.PtrToVal(it.Gpu),
			CPU:               converter.PtrToVal(it.Cpu),
			Memory:            converter.PtrToVal(it.Memory) * 1024, // Note: 为保持与其他云一致，内存单位调整为MB
			FPGA:              converter.PtrToVal(it.Fpga),
			Status:            converter.PtrToVal(it.Status),
			CpuType:           converter.PtrToVal(it.CpuType),
			InstanceBandwidth: converter.PtrToVal(it.InstanceBandwidth),
			InstancePps:       converter.PtrToVal(it.InstancePps),
			Price:             converter.PtrToVal(it.Price),
			TypeName:          converter.PtrToVal(it.TypeName),
		})
	}

	return its, nil
}
