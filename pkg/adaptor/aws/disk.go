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
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateDisk 创建云硬盘
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateVolume.html
// SDK: https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateVolumeWithContext
func (a *Aws) CreateDisk(kt *kit.Kit, opt *disk.AwsDiskCreateOption) (*ec2.Volume, error) {
	return a.createDisk(kt, opt)
}

func (a *Aws) createDisk(kt *kit.Kit, opt *disk.AwsDiskCreateOption) (*ec2.Volume, error) {
	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req, err := opt.ToCreateVolumeInput()
	if err != nil {
		return nil, err
	}

	return client.CreateVolumeWithContext(kt.Ctx, req)
}

// ListDisk 查看云硬盘
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DescribeVolumes.html
func (a *Aws) ListDisk(kt *kit.Kit, opt *disk.AwsDiskListOption) ([]*ec2.Volume, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "aws disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ec2.DescribeVolumesInput)

	if opt.Page != nil {
		req.MaxResults = opt.Page.MaxResults
		req.NextToken = opt.Page.NextToken
	}

	resp, err := client.DescribeVolumesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Volumes, nil
}
