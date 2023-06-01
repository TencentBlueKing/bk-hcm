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
	"strings"

	"hcm/pkg/adaptor/types/cvm"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateSecurityGroup create security group.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateSecurityGroup.html
func (a *Aws) CreateSecurityGroup(kt *kit.Kit, opt *securitygroup.AwsCreateOption) (string, error) {

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
		TagSpecifications: []*ec2.TagSpecification{
			{
				// reference: https://docs.aws.amazon.com/zh_cn/AWSEC2/latest/APIReference/API_TagSpecification.html
				ResourceType: aws.String("security-group"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(tagKeyForResourceName),
						Value: aws.String(opt.Name),
					},
				},
			},
		},
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
func (a *Aws) ListSecurityGroup(kt *kit.Kit, opt *securitygroup.AwsListOption) ([]securitygroup.AwsSG,
	*ec2.DescribeSecurityGroupsOutput, error) {

	if opt == nil {
		return nil, nil, errf.New(errf.InvalidParameter, "security group list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, nil, err
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
		if !strings.Contains(err.Error(), ErrSGNotFound) {
			logs.Errorf("list aws security group failed, err: %v, rid: %s", err, kt.Rid)
		}

		return nil, nil, err
	}

	sgs := make([]securitygroup.AwsSG, 0, len(resp.SecurityGroups))
	for _, one := range resp.SecurityGroups {
		sgs = append(sgs, securitygroup.AwsSG{one})
	}
	return sgs, resp, nil
}

// DeleteSecurityGroup delete security group.
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DeleteSecurityGroup.html
func (a *Aws) DeleteSecurityGroup(kt *kit.Kit, opt *securitygroup.AwsDeleteOption) error {
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

// SecurityGroupCvmAssociate reference:
// https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_ModifyInstanceAttribute.html
func (a *Aws) SecurityGroupCvmAssociate(kt *kit.Kit, opt *securitygroup.AwsAssociateCvmOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "associate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	listCvmOpt := &cvm.AwsListOption{
		Region:   opt.Region,
		CloudIDs: []string{opt.CloudCvmID},
	}
	_, resp, err := a.ListCvm(kt, listCvmOpt)
	if err != nil {
		return fmt.Errorf("associate security group to query cvm detail failed, err: %v", err)
	}

	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return fmt.Errorf("cvm(cloud_id=%s) not found", opt.CloudCvmID)
	}

	sgIDs := make([]*string, 0)
	for _, sg := range resp.Reservations[0].Instances[0].SecurityGroups {
		if converter.PtrToVal(sg.GroupId) == opt.CloudSecurityGroupID {
			return fmt.Errorf("cvm: %s already associated security group: %s", opt.CloudCvmID, opt.CloudSecurityGroupID)
		}
		sgIDs = append(sgIDs, sg.GroupId)
	}
	sgIDs = append(sgIDs, aws.String(opt.CloudSecurityGroupID))

	req := &ec2.ModifyInstanceAttributeInput{
		Groups:     sgIDs,
		InstanceId: aws.String(opt.CloudCvmID),
	}
	_, err = client.ModifyInstanceAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("associate aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupCvmDisassociate reference:
// https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_ModifyInstanceAttribute.html
func (a *Aws) SecurityGroupCvmDisassociate(kt *kit.Kit, opt *securitygroup.AwsAssociateCvmOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "disassociate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	listCvmOpt := &cvm.AwsListOption{
		Region:   opt.Region,
		CloudIDs: []string{opt.CloudCvmID},
	}
	_, resp, err := a.ListCvm(kt, listCvmOpt)
	if err != nil {
		return fmt.Errorf("disassociate security group to query cvm detail failed, err: %v", err)
	}

	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return fmt.Errorf("cvm(cloud_id=%s) not found", opt.CloudCvmID)
	}

	sgIDs := make([]*string, 0)
	hit := false
	for _, sg := range resp.Reservations[0].Instances[0].SecurityGroups {
		if sg.GroupId != nil && converter.PtrToVal(sg.GroupId) == opt.CloudSecurityGroupID {
			hit = true
			continue
		}
		sgIDs = append(sgIDs, sg.GroupId)
	}

	if !hit {
		return fmt.Errorf("cvm: %s not assoociate security group: %s", opt.CloudCvmID, opt.CloudSecurityGroupID)
	}

	if len(sgIDs) == 0 {
		return errors.New("the last security group of the cvm is not allowed to disassociate")
	}

	req := &ec2.ModifyInstanceAttributeInput{
		Groups:     sgIDs,
		InstanceId: aws.String(opt.CloudCvmID),
	}
	_, err = client.ModifyInstanceAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("disassociate aws security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
