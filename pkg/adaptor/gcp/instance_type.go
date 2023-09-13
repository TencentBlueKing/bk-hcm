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
	typesinstancetype "hcm/pkg/adaptor/types/instance-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"google.golang.org/api/compute/v1"
)

// ListInstanceType ...
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/machineTypes/list
func (g *Gcp) ListInstanceType(
	kt *kit.Kit, opt *typesinstancetype.GcpInstanceTypeListOption,
) (*typesinstancetype.GcpInstanceTypeListResult, error) {
	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	req := client.MachineTypes.List(g.CloudProjectID(), opt.Zone).Context(kt.Ctx)
	if opt.Page != nil {
		req.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := req.Do()
	if err != nil {
		logs.Errorf("list instance type failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	its := make([]*typesinstancetype.GcpInstanceType, 0, len(resp.Items))
	for _, machineType := range resp.Items {
		if machineType != nil {
			gcpInstanceType := toGcpInstanceType(machineType)
			if gcpInstanceType != nil {
				its = append(its, gcpInstanceType)
			}
		}
	}

	return &typesinstancetype.GcpInstanceTypeListResult{Details: its, NextPageToken: resp.NextPageToken}, nil
}

func toGcpInstanceType(machineType *compute.MachineType) *typesinstancetype.GcpInstanceType {
	if machineType.Deprecated != nil {
		return nil
	}

	return &typesinstancetype.GcpInstanceType{
		InstanceType: machineType.Name,
		Memory:       machineType.MemoryMb,
		CPU:          machineType.GuestCpus,
		Kind:         machineType.Kind,
	}
}
