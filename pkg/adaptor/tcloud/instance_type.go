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
)

// ListInstanceType ...
// reference: https://cloud.tencent.com/document/api/213/15749
func (t *TCloud) ListInstanceType(kt *kit.Kit, opt *typesinstancetype.TCloudListOption) (
	[]typesinstancetype.TCloudInstanceType, error,
) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := opt.ToListRequest()
	resp, err := client.DescribeInstanceTypeConfigsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud instance type failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	its := make([]typesinstancetype.TCloudInstanceType, 0, len(resp.Response.InstanceTypeConfigSet))
	for _, it := range resp.Response.InstanceTypeConfigSet {
		its = append(its, typesinstancetype.TCloudInstanceType{
			Zone:           converter.PtrToVal(it.Zone),
			InstanceType:   converter.PtrToVal(it.InstanceType),
			InstanceFamily: converter.PtrToVal(it.InstanceFamily),
			GPU:            converter.PtrToVal(it.GPU),
			CPU:            converter.PtrToVal(it.CPU),
			Memory:         converter.PtrToVal(it.Memory),
			FPGA:           converter.PtrToVal(it.FPGA),
		})
	}

	return its, nil
}
