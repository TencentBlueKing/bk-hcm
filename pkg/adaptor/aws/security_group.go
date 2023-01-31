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
	"errors"
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateSecurityGroup create security group.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateSecurityGroup.html
func (a *Aws) CreateSecurityGroup(kt *kit.Kit, opt *types.AwsSecurityGroupCreateOption) (string, error) {

	if opt == nil {
		return "", errf.New(errf.InvalidParameter, "security group create option is required")
	}

	if err := opt.Validate(); err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return "", fmt.Errorf("new ec2 client failed, err: %v", err)
	}

	req := &ec2.CreateSecurityGroupInput{
		Description: opt.Description,
		GroupName:   aws.String(opt.Name),
	}

	if len(opt.CloudVpcID) != 0 {
		req.VpcId = aws.String(opt.CloudVpcID)
	}

	resp, err := client.CreateSecurityGroupWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if resp == nil || resp.GroupId == nil {
		return "", errors.New("create tcloud security group return security group id is nil")
	}

	return *resp.GroupId, nil
}

// ListSecurityGroup list security group.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DescribeSecurityGroups.html
func (a *Aws) ListSecurityGroup(kt *kit.Kit, opt *types.AwsSecurityGroupListOption) (*ec2.DescribeSecurityGroupsOutput,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ec2.DescribeSecurityGroupsInput)

	if len(opt.CloudIDs) > 0 {
		req.GroupIds = aws.StringSlice(opt.CloudIDs)
	}

	if opt.Page != nil {
		req.MaxResults = opt.Page.MaxResults
		req.NextToken = opt.Page.NextToken
	}

	resp, err := client.DescribeSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp, nil
}

// DeleteSecurityGroup delete security group.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DeleteSecurityGroup.html
func (a *Aws) DeleteSecurityGroup(kt *kit.Kit, opt *types.SecurityGroupDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(opt.CloudID),
	}
	if _, err = client.DeleteSecurityGroupWithContext(kt.Ctx, req); err != nil {
		logs.Errorf("delete aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
