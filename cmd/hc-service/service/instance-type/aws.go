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
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ListForAws ...
func (i *instanceTypeAdaptor) ListForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsInstanceTypeListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := i.adaptor.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	data := make([]*proto.AwsInstanceTypeResp, 0)
	// 分页遍历获取所有数据
	nextToken := ""
	for {
		opt := &typesinstancetype.AwsInstanceTypeListOption{
			Region: req.Region,
			Page:   &core.AwsPage{MaxResults: converter.ValToPtr(int64(100))},
		}
		if nextToken != "" {
			opt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		result, err := client.ListInstanceType(cts.Kit, opt)
		if err != nil {
			logs.Errorf("request adaptor to list aws instance type failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		if len(result.Details) <= 0 {
			logs.Errorf("request adaptor to list aws instance type num <= 0, rid: %s", cts.Kit.Rid)
			return nil, err
		}

		for _, it := range result.Details {
			data = append(data, toAwsInstanceTypeResp(it))
		}

		// 判断是否还有下一页
		if result.NextToken == nil || *result.NextToken == "" {
			break
		}
		nextToken = *result.NextToken
	}

	return data, nil
}

func toAwsInstanceTypeResp(it *typesinstancetype.AwsInstanceType) *proto.AwsInstanceTypeResp {
	return &proto.AwsInstanceTypeResp{
		InstanceFamily:     it.InstanceFamily,
		InstanceType:       it.InstanceType,
		GPU:                it.GPU,
		GPUMemory:          it.GPUMemory,
		GPUName:            it.GPUName,
		GPUManufacturer:    it.GPUManufacturer,
		CPU:                it.CPU,
		Memory:             it.Memory,
		FPGA:               it.FPGA,
		NetworkPerformance: it.NetworkPerformance,
		DiskSizeInGB:       it.DiskSizeInGB,
		Architecture:       it.Architecture,
		DiskType:           it.DiskType,
	}
}

// ListAssumeRoleInstanceTypeForAws lists instance types via AssumeRole cross-account access.
func (i *instanceTypeAdaptor) ListAssumeRoleInstanceTypeForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleInstanceTypeListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := i.adaptor.AwsWithAssumeRole(cts.Kit, req.CloudID, req.RoleChain)
	if err != nil {
		return nil, err
	}

	data := make([]*proto.AwsInstanceTypeResp, 0)
	nextToken := ""
	for {
		opt := &typesinstancetype.AwsInstanceTypeListOption{
			Region: req.Region,
			Page:   &core.AwsPage{MaxResults: converter.ValToPtr(int64(100))},
		}
		if nextToken != "" {
			opt.Page.NextToken = converter.ValToPtr(nextToken)
		}

		result, err := client.ListInstanceType(cts.Kit, opt)
		if err != nil {
			logs.Errorf("list aws assume role instance types failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		if len(result.Details) <= 0 {
			break
		}

		for _, it := range result.Details {
			data = append(data, toAwsInstanceTypeResp(it))
		}

		if result.NextToken == nil || *result.NextToken == "" {
			break
		}
		nextToken = *result.NextToken
	}

	return data, nil
}

// ListAssumeRoleInstanceForAws lists EC2 instances via AssumeRole cross-account access.
func (i *instanceTypeAdaptor) ListAssumeRoleInstanceForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleInstanceListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := i.adaptor.AwsWithAssumeRole(cts.Kit, req.CloudID, req.RoleChain)
	if err != nil {
		return nil, err
	}

	ec2Client, err := client.GetEC2Client(req.Region)
	if err != nil {
		return nil, err
	}

	data := make([]*proto.AwsAssumeRoleInstanceResp, 0)
	input := &ec2.DescribeInstancesInput{
		MaxResults: converter.ValToPtr(int64(100)),
	}

	for {
		resp, err := ec2Client.DescribeInstancesWithContext(cts.Kit.Ctx, input)
		if err != nil {
			logs.Errorf("list aws assume role instances failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		for _, reservation := range resp.Reservations {
			for _, inst := range reservation.Instances {
				item := &proto.AwsAssumeRoleInstanceResp{
					InstanceID:   aws.StringValue(inst.InstanceId),
					InstanceType: aws.StringValue(inst.InstanceType),
					PrivateIP:    aws.StringValue(inst.PrivateIpAddress),
					PublicIP:     aws.StringValue(inst.PublicIpAddress),
					Region:       req.Region,
				}
				if inst.State != nil {
					item.State = aws.StringValue(inst.State.Name)
				}
				if inst.Placement != nil {
					item.Zone = aws.StringValue(inst.Placement.AvailabilityZone)
				}
				data = append(data, item)
			}
		}

		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}

	return data, nil
}
