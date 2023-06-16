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
	"fmt"
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// CreateDisk 创建云硬盘
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateVolume.html
// SDK: https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateVolumeWithContext
func (a *Aws) CreateDisk(kt *kit.Kit, opt *disk.AwsDiskCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "aws disk create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	diskCloudIDs := make([]*string, 0)

	for i := uint64(1); i <= *opt.DiskCount; i++ {
		resp, err := a.createDisk(kt, opt)
		if err != nil {
			return nil, err
		}
		diskCloudIDs = append(diskCloudIDs, resp.VolumeId)
	}

	respPoller := poller.Poller[*Aws, []disk.AwsDisk, poller.BaseDoneResult]{
		Handler: &createDiskPollingHandler{region: opt.Region},
	}
	return respPoller.PollUntilDone(a, kt, diskCloudIDs, nil)
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
func (a *Aws) ListDisk(kt *kit.Kit, opt *disk.AwsDiskListOption) ([]disk.AwsDisk, *string, error) {
	if opt == nil {
		return nil, nil, errf.New(errf.InvalidParameter, "aws disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return nil, nil, err
	}

	req := new(ec2.DescribeVolumesInput)

	if len(opt.CloudIDs) > 0 {
		req.VolumeIds = converter.SliceToPtr(opt.CloudIDs)
	} else if opt.Page != nil {
		req.MaxResults = opt.Page.MaxResults
		req.NextToken = opt.Page.NextToken
	}

	resp, err := client.DescribeVolumesWithContext(kt.Ctx, req)
	if err != nil {
		if !strings.Contains(err.Error(), ErrDiskNotFound) {
			logs.Errorf("list aws disk failed, err: %v, rid: %s", err, kt.Rid)
		}

		return nil, nil, err
	}

	disks := make([]disk.AwsDisk, 0, len(resp.Volumes))
	for _, one := range resp.Volumes {
		disks = append(disks, disk.AwsDisk{one})
	}

	return disks, resp.NextToken, err
}

// DeleteDisk 删除云盘
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DeleteVolume.html
func (a *Aws) DeleteDisk(kt *kit.Kit, opt *disk.AwsDiskDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "aws disk delete option is required")
	}

	input, err := opt.ToDeleteVolumeInput()
	if err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.DeleteVolumeWithContext(kt.Ctx, input)
	return err
}

// AttachDisk 挂载云盘
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_AttachVolume.html
func (a *Aws) AttachDisk(kt *kit.Kit, opt *disk.AwsDiskAttachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "aws disk attach option is required")
	}

	input, err := opt.ToAttachVolumeInput()
	if err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.AttachVolumeWithContext(kt.Ctx, input)
	if err != nil {
		return err
	}

	respPoller := poller.Poller[*Aws, []disk.AwsDisk, poller.BaseDoneResult]{
		Handler: &attachDiskPollingHandler{region: opt.Region},
	}
	_, err = respPoller.PollUntilDone(a, kt, []*string{&opt.CloudDiskID}, nil)
	return err
}

// DetachDisk 卸载云盘
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_DetachVolume.html
func (a *Aws) DetachDisk(kt *kit.Kit, opt *disk.AwsDiskDetachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "aws disk detach option is required")
	}

	input, err := opt.ToDetachVolumeInput()
	if err != nil {
		return err
	}

	client, err := a.clientSet.ec2Client(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.DetachVolumeWithContext(kt.Ctx, input)
	if err != nil {
		return err
	}

	respPoller := poller.Poller[*Aws, []disk.AwsDisk, poller.BaseDoneResult]{
		Handler: &detachDiskPollingHandler{region: opt.Region},
	}
	_, err = respPoller.PollUntilDone(a, kt, []*string{&opt.CloudDiskID}, nil)
	return err
}

type createDiskPollingHandler struct {
	region string
}

// Done ...
func (h *createDiskPollingHandler) Done(pollResult []disk.AwsDisk) (bool, *poller.BaseDoneResult) {
	successCloudIDs := make([]string, 0)
	unknownCloudIDs := make([]string, 0)

	for _, r := range pollResult {
		if converter.PtrToVal(r.State) == "creating" {
			unknownCloudIDs = append(unknownCloudIDs, *r.VolumeId)
		} else {
			successCloudIDs = append(successCloudIDs, *r.VolumeId)
		}
	}

	isDone := false
	if len(successCloudIDs) != 0 && len(successCloudIDs) == len(pollResult) {
		isDone = true
	}

	return isDone, &poller.BaseDoneResult{
		SuccessCloudIDs: successCloudIDs,
		UnknownCloudIDs: unknownCloudIDs,
	}
}

// Poll ...
func (h *createDiskPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]disk.AwsDisk, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, _, err := client.ListDisk(
		kt,
		&disk.AwsDiskListOption{
			Region:   h.region,
			CloudIDs: cIDs,
			Page:     &core.AwsPage{MaxResults: converter.ValToPtr(int64(filter.DefaultMaxInLimit))},
		},
	)
	return result, err
}

var _ poller.PollingHandler[*Aws, []disk.AwsDisk, poller.BaseDoneResult] = new(createDiskPollingHandler)

type attachDiskPollingHandler struct {
	region string
}

// Done ...
func (h *attachDiskPollingHandler) Done(pollResult []disk.AwsDisk) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if converter.PtrToVal(r.State) != "in-use" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *attachDiskPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]disk.AwsDisk, error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	result, _, err := client.ListDisk(
		kt,
		&disk.AwsDiskListOption{
			Region:   h.region,
			CloudIDs: cIDs,
		},
	)
	return result, err
}

type detachDiskPollingHandler struct {
	region string
}

// Done ...
func (h *detachDiskPollingHandler) Done(pollResult []disk.AwsDisk) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]
	if converter.PtrToVal(r.State) != "available" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *detachDiskPollingHandler) Poll(client *Aws, kt *kit.Kit, cloudIDs []*string) ([]disk.AwsDisk, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, _, err := client.ListDisk(
		kt,
		&disk.AwsDiskListOption{
			Region:   h.region,
			CloudIDs: cIDs,
		},
	)
	return result, err
}
