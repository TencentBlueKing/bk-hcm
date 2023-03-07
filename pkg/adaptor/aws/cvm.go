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
	"encoding/base64"
	"fmt"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ListCvm list cvm.
// reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html
func (a *Aws) ListCvm(kt *kit.Kit, opt *typecvm.AwsListOption) (*ec2.DescribeInstancesOutput, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(ec2.DescribeInstancesInput)

	if len(opt.CloudIDs) > 0 {
		req.InstanceIds = aws.StringSlice(opt.CloudIDs)
	}

	if opt.Page != nil {
		req.MaxResults = opt.Page.MaxResults
		req.NextToken = opt.Page.NextToken
	}

	resp, err := client.DescribeInstancesWithContext(kt.Ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list aws cvm instances failed, err: %v", err)
	}

	return resp, nil
}

// GetCvmNameFromTags ...
func GetCvmNameFromTags(tags []*ec2.Tag) *string {
	if len(tags) == 0 {
		return nil
	}

	for _, tagPtr := range tags {
		if tagPtr != nil && tagPtr.Key != nil && *tagPtr.Key == "Name" {
			return tagPtr.Value
		}
	}

	return nil
}

// DeleteCvm reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_TerminateInstances.html
func (a *Aws) DeleteCvm(kt *kit.Kit, opt *typecvm.AwsDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice(opt.CloudIDs),
	}
	_, err = client.TerminateInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("terminate cvm failed, err: %v, ids: %v, rid: %s", err, opt.CloudIDs, kt.Rid)
		return err
	}

	return nil
}

// StartCvm reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_StartInstances.html
func (a *Aws) StartCvm(kt *kit.Kit, opt *typecvm.AwsStartOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "start option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.StartInstancesInput{
		InstanceIds: aws.StringSlice(opt.CloudIDs),
	}
	_, err = client.StartInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("start cvm failed, err: %v, ids: %v, rid: %s", err, opt.CloudIDs, kt.Rid)
		return err
	}

	return nil
}

// StopCvm reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_StopInstances.html
func (a *Aws) StopCvm(kt *kit.Kit, opt *typecvm.AwsStopOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "stop option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.StopInstancesInput{
		Force:       aws.Bool(opt.Force),
		Hibernate:   aws.Bool(opt.Hibernate),
		InstanceIds: aws.StringSlice(opt.CloudIDs),
	}
	_, err = client.StopInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("stop cvm failed, err: %v, ids: %v, rid: %s", err, opt.CloudIDs, kt.Rid)
		return err
	}

	return nil
}

// RebootCvm reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RebootInstances.html
func (a *Aws) RebootCvm(kt *kit.Kit, opt *typecvm.AwsRebootOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "reboot option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.RebootInstancesInput{
		InstanceIds: aws.StringSlice(opt.CloudIDs),
	}
	_, err = client.RebootInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("reboot cvm failed, err: %v, ids: %v, rid: %s", err, opt.CloudIDs, kt.Rid)
		return err
	}

	return nil
}

// CreateCvm reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RunInstances.html
func (a *Aws) CreateCvm(kt *kit.Kit, opt *typecvm.AwsCreateOption) ([]*ec2.Instance, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	userData, err := genCvmBase64UserData(kt, client, opt.ImageID, opt.Password)
	if err != nil {
		return nil, fmt.Errorf("gen cvm base64 user data failed, err: %v", err)
	}

	req := &ec2.RunInstancesInput{
		ClientToken:      opt.ClientToken,
		ImageId:          aws.String(opt.ImageID),
		InstanceType:     aws.String(opt.InstanceType),
		MaxCount:         aws.Int64(opt.RequiredCount),
		MinCount:         aws.Int64(opt.RequiredCount),
		SecurityGroupIds: aws.StringSlice(opt.SecurityGroupIDs),
		SubnetId:         aws.String(opt.SubnetID),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(tagKeyForResourceName),
						Value: opt.Name,
					},
				},
			},
		},
		UserData: aws.String(userData),
	}

	if opt.Zone != nil {
		req.Placement = &ec2.Placement{
			AvailabilityZone: opt.Zone,
		}
	}

	if len(opt.BlockDeviceMapping) != 0 {
		req.BlockDeviceMappings = make([]*ec2.BlockDeviceMapping, len(opt.BlockDeviceMapping))
		for index, volume := range opt.BlockDeviceMapping {
			req.BlockDeviceMappings[index].DeviceName = volume.DeviceName

			if volume.Ebs != nil {
				req.BlockDeviceMappings[index].Ebs = &ec2.EbsBlockDevice{
					Iops:       volume.Ebs.Iops,
					VolumeSize: aws.Int64(volume.Ebs.VolumeSizeGB),
					VolumeType: aws.String(volume.Ebs.VolumeType),
				}
			}
		}
	}

	resp, err := client.RunInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run instances failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	return resp.Instances, nil
}

func genCvmBase64UserData(kt *kit.Kit, ec2Client *ec2.EC2, imageID string, passwd string) (string, error) {
	req := new(ec2.DescribeImagesInput)
	req.ImageIds = aws.StringSlice([]string{imageID})
	resp, err := ec2Client.DescribeImagesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("query aws image: %s failed, err: %v, rid: %s", imageID, err, kt.Rid)
		return "", err
	}

	if len(resp.Images) == 0 {
		return "", fmt.Errorf("image: %s not found", imageID)
	}

	if len(resp.Images) != 1 {
		return "", fmt.Errorf("query image: %s, return count: %d not right", imageID, len(resp.Images))
	}

	if resp.Images[0].Platform != nil && (*resp.Images[0].Platform == "Windows" ||
		*resp.Images[0].Platform == "windows") {
		script := fmt.Sprintf(`<script>
net user administrator %s
</script>`, passwd)

		return base64.StdEncoding.EncodeToString([]byte(script)), nil
	}

	script := fmt.Sprintf(`#!/bin/bash
echo root:%s|chpasswd
sed -i 's/PasswordAuthentication/\# PasswordAuthentication/g' /etc/ssh/sshd_config
sed -i 's/PermitRootLogin/\# PermitRootLogin/g' /etc/ssh/sshd_config
sed -i '20 a PasswordAuthentication yes' /etc/ssh/sshd_config
sed -i '20 a PermitRootLogin yes' /etc/ssh/sshd_config
systemctl restart sshd`, passwd)

	return base64.StdEncoding.EncodeToString([]byte(script)), nil
}
