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
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ListCvm list cvm.
// reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html
func (a *Aws) ListCvm(kt *kit.Kit, opt *typecvm.AwsListOption) ([]typecvm.AwsCvm, *ec2.DescribeInstancesOutput, error) {
	if opt == nil {
		return nil, nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, nil, err
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
		if !strings.Contains(err.Error(), ErrCvmNotFound) {
			logs.Errorf("list aws cvm failed, err: %v, rid: %s", err, kt.Rid)
		}

		return nil, nil, err
	}

	cvms := make([]typecvm.AwsCvm, 0)
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			cvms = append(cvms, typecvm.AwsCvm{instance})
		}
	}

	return cvms, resp, nil
}

// CountCvm 返回单个地域下的ec2 instance 数量，基于 DescribeInstancesWithContext接口
// reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html
func (a *Aws) CountCvm(kt *kit.Kit, region string) (int32, error) {

	client, err := a.clientSet.ec2Client(region)
	if err != nil {
		return 0, err
	}

	req := new(ec2.DescribeInstancesInput)
	total := 0
	req.MaxResults = converter.ValToPtr(int64(core.AwsQueryLimit))

	for {
		resp, err := client.DescribeInstancesWithContext(kt.Ctx, req)
		if err != nil {
			logs.Errorf("count aws cvm failed, err: %v, region:%s, rid: %s", err, region, kt.Rid)
			return 0, err
		}
		for _, rsv := range resp.Reservations {
			total += len(rsv.Instances)
		}
		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}
	return int32(total), nil
}

// GetCvmNameFromTags ...
func GetCvmNameFromTags(tags []*ec2.Tag) *string {
	if len(tags) == 0 {
		return nil
	}

	for _, tagPtr := range tags {
		if tagPtr != nil && tagPtr.Key != nil && *tagPtr.Key == "ImportMode" {
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

	handler := &startAwsCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*Aws, []*ec2.Instance, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(a, kt, converter.SliceToPtr(opt.CloudIDs),
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
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

	handler := &stopAwsCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*Aws, []*ec2.Instance, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(a, kt, converter.SliceToPtr(opt.CloudIDs),
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
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

	handler := &rebootAwsCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*Aws, []*ec2.Instance, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(a, kt, converter.SliceToPtr(opt.CloudIDs),
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	return nil
}

// CreateCvm reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RunInstances.html
func (a *Aws) CreateCvm(kt *kit.Kit, opt *typecvm.AwsCreateOption) (*poller.BaseDoneResult, error) {
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

	userData, err := genCvmBase64UserData(kt, client, opt.CloudImageID, opt.Password)
	if err != nil {
		return nil, fmt.Errorf("gen cvm base64 user data failed, err: %v", err)
	}

	req := buildCreateCvmReq(opt, userData)

	resp, err := client.RunInstancesWithContext(kt.Ctx, req)
	if err != nil {
		// 参数预校验报错，正常现象
		if strings.Contains(err.Error(), ErrDryRunSuccess) {
			return new(poller.BaseDoneResult), nil
		}

		logs.Errorf("run instances failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	cloudIDs := make([]*string, 0, len(resp.Instances))
	for _, one := range resp.Instances {
		cloudIDs = append(cloudIDs, one.InstanceId)
	}

	// 等待生产成功
	handler := &createCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*Aws, []*ec2.Instance, poller.BaseDoneResult]{Handler: handler}
	result, err := respPoller.PollUntilDone(a, kt, cloudIDs, types.NewBatchCreateCvmPollerOption())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func buildCreateCvmReq(opt *typecvm.AwsCreateOption, userData string) *ec2.RunInstancesInput {
	req := &ec2.RunInstancesInput{
		DryRun:       aws.Bool(opt.DryRun),
		ClientToken:  opt.ClientToken,
		ImageId:      aws.String(opt.CloudImageID),
		InstanceType: aws.String(opt.InstanceType),
		MaxCount:     aws.Int64(opt.RequiredCount),
		MinCount:     aws.Int64(opt.RequiredCount),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(tagKeyForResourceName),
						Value: aws.String(opt.Name),
					},
				},
			},
		},
		UserData: aws.String(userData),
		Placement: &ec2.Placement{
			AvailabilityZone: aws.String(opt.Zone),
		},
	}

	// 如果弹性IP指定了子网，则外部不能设置子网
	if opt.PublicIPAssigned {
		req.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              converter.ValToPtr(int64(0)),
				SubnetId:                 aws.String(opt.CloudSubnetID),
				AssociatePublicIpAddress: aws.Bool(opt.PublicIPAssigned),
				Groups:                   aws.StringSlice(opt.CloudSecurityGroupIDs),
			},
		}
	} else {
		req.SubnetId = aws.String(opt.CloudSubnetID)
		req.SecurityGroupIds = aws.StringSlice(opt.CloudSecurityGroupIDs)
	}

	if len(opt.BlockDeviceMapping) != 0 {
		req.BlockDeviceMappings = make([]*ec2.BlockDeviceMapping, len(opt.BlockDeviceMapping))
		for index, volume := range opt.BlockDeviceMapping {
			req.BlockDeviceMappings[index] = &ec2.BlockDeviceMapping{
				DeviceName: volume.DeviceName,
			}

			if volume.Ebs != nil {
				req.BlockDeviceMappings[index].Ebs = &ec2.EbsBlockDevice{
					Iops:       volume.Ebs.Iops,
					VolumeSize: aws.Int64(volume.Ebs.VolumeSizeGB),
					VolumeType: aws.String(string(volume.Ebs.VolumeType)),
				}
			}
		}
	}
	return req
}

// BatchAssociateSecurityGroup batch associate security group.
// reference: https://docs.amazonaws.cn/en_us/AWSEC2/latest/APIReference/API_ModifyInstanceAttribute.html
func (a *Aws) BatchAssociateSecurityGroup(kt *kit.Kit, opt *typecvm.AwsAssociateSecurityGroupsOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "option is required")
	}
	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	req := &ec2.ModifyInstanceAttributeInput{
		InstanceId: aws.String(opt.CloudCvmID),
		Groups:     aws.StringSlice(opt.CloudSecurityGroupIDs),
	}

	_, err = client.ModifyInstanceAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("batch associate security group failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	return nil
}

type startAwsCvmPollingHandler struct {
	region string
}

// Done ...
func (h *startAwsCvmPollingHandler) Done(cvms []*ec2.Instance) (bool, *poller.BaseDoneResult) {
	return done(cvms, 16)
}

// Poll ...
func (h *startAwsCvmPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]*ec2.Instance, error) {
	return poll(client, kt, h.region, cloudIDs)
}

type stopAwsCvmPollingHandler struct {
	region string
}

// Done ...
func (h *stopAwsCvmPollingHandler) Done(cvms []*ec2.Instance) (bool, *poller.BaseDoneResult) {
	return done(cvms, 80)
}

// Poll ...
func (h *stopAwsCvmPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]*ec2.Instance, error) {
	return poll(client, kt, h.region, cloudIDs)
}

type rebootAwsCvmPollingHandler struct {
	region string
}

// Done ...
func (h *rebootAwsCvmPollingHandler) Done(cvms []*ec2.Instance) (bool, *poller.BaseDoneResult) {
	return done(cvms, 16)
}

// Poll ...
func (h *rebootAwsCvmPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]*ec2.Instance, error) {
	return poll(client, kt, h.region, cloudIDs)
}

func done(cvms []*ec2.Instance, succeed int64) (bool, *poller.BaseDoneResult) {
	result := new(poller.BaseDoneResult)

	flag := true
	for _, instance := range cvms {
		// not done
		if converter.PtrToVal(instance.State.Code) != succeed {
			flag = false
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, converter.PtrToVal(instance.InstanceId))
	}

	return flag, result
}

func poll(client *Aws, kt *kit.Kit, region string, cloudIDs []*string) ([]*ec2.Instance, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.AwsQueryLimit)

	cvms := make([]*ec2.Instance, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(ec2.DescribeInstancesInput)
		req.InstanceIds = partIDs

		cvmCli, err := client.clientSet.ec2Client(region)
		if err != nil {
			return nil, err
		}

		resp, err := cvmCli.DescribeInstancesWithContext(kt.Ctx, req)
		if err != nil {
			return nil, err
		}

		for _, reservation := range resp.Reservations {
			cvms = append(cvms, reservation.Instances...)
		}
	}

	return cvms, nil
}

type createCvmPollingHandler struct {
	region string
}

// Done ...
func (h *createCvmPollingHandler) Done(cvms []*ec2.Instance) (bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
	}
	flag := true
	for _, instance := range cvms {
		// 创建中
		if converter.PtrToVal(instance.State.Code) == 0 {
			flag = false
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, *instance.InstanceId)
			continue
		}

		// 生产失败
		if converter.PtrToVal(instance.State.Code) == 48 {
			result.FailedCloudIDs = append(result.FailedCloudIDs, *instance.InstanceId)
			result.FailedMessage = converter.PtrToVal(instance.StateReason.Message)
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, *instance.InstanceId)
	}

	return flag, result
}

// Poll ...
func (h *createCvmPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]*ec2.Instance, error) {

	cloudIDSplit := slice.Split(cloudIDs, core.AwsQueryLimit)

	cvms := make([]*ec2.Instance, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(ec2.DescribeInstancesInput)
		req.InstanceIds = partIDs

		cvmCli, err := client.clientSet.ec2Client(h.region)
		if err != nil {
			return nil, err
		}

		resp, err := cvmCli.DescribeInstancesWithContext(kt.Ctx, req)
		if err != nil {
			return nil, err
		}

		for _, reservation := range resp.Reservations {
			cvms = append(cvms, reservation.Instances...)
		}
	}

	if len(cvms) != len(cloudIDs) {
		return nil, fmt.Errorf("query cvm count: %d not equal return count: %d", len(cloudIDs), len(cvms))
	}

	return cvms, nil
}

var _ poller.PollingHandler[*Aws, []*ec2.Instance, poller.BaseDoneResult] = new(createCvmPollingHandler)

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
