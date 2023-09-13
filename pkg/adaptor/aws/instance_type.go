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

package aws

import (
	"strings"

	typesinstancetype "hcm/pkg/adaptor/types/instance-type"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// ListInstanceType ...
// reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstanceTypes.html
func (a *Aws) ListInstanceType(kt *kit.Kit, opt *typesinstancetype.AwsInstanceTypeListOption) (
	*typesinstancetype.AwsInstanceTypeListResult, error,
) {
	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := &ec2.DescribeInstanceTypesInput{
		MaxResults: opt.Page.MaxResults,
		NextToken:  opt.Page.NextToken,
	}

	resp, err := client.DescribeInstanceTypesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("describe aws instance type failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	its := make([]*typesinstancetype.AwsInstanceType, 0, len(resp.InstanceTypes))
	for _, it := range resp.InstanceTypes {
		its = append(its, toAwsInstanceType(it))
	}

	return &typesinstancetype.AwsInstanceTypeListResult{Details: its, NextToken: resp.NextToken}, nil
}

func toAwsInstanceType(info *ec2.InstanceTypeInfo) *typesinstancetype.AwsInstanceType {
	it := new(typesinstancetype.AwsInstanceType)
	it.InstanceType = converter.PtrToVal(info.InstanceType)

	if info.MemoryInfo != nil {
		it.Memory = converter.PtrToVal(info.MemoryInfo.SizeInMiB)
	}

	if info.VCpuInfo != nil {
		it.CPU = converter.PtrToVal(info.VCpuInfo.DefaultCores)
	}

	if info.GpuInfo != nil {
		for _, gpu := range info.GpuInfo.Gpus {
			if gpu != nil {
				it.GPU += converter.PtrToVal(gpu.Count)
			}
		}
	}

	if info.FpgaInfo != nil {
		for _, fpga := range info.FpgaInfo.Fpgas {
			if fpga != nil {
				it.FPGA += converter.PtrToVal(fpga.Count)
			}
		}
	}

	if info.NetworkInfo != nil {
		it.NetworkPerformance = converter.PtrToVal(info.NetworkInfo.NetworkPerformance)
	}

	if info.InstanceStorageInfo != nil {
		it.DiskSizeInGB = converter.PtrToVal(info.InstanceStorageInfo.TotalSizeInGB)
		diskType := make([]string, 0)
		for _, one := range info.InstanceStorageInfo.Disks {
			if one != nil {
				diskType = append(diskType, converter.PtrToVal(one.Type))
			}
		}
		it.DiskType = strings.Join(diskType, ",")
	}

	if info.ProcessorInfo != nil {
		architecture := make([]string, 0)
		for _, one := range info.ProcessorInfo.SupportedArchitectures {
			architecture = append(architecture, converter.PtrToVal(one))
		}
		it.Architecture = strings.Join(architecture, ",")
	}

	return it
}
