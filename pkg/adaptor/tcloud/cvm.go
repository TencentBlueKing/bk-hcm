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
	"fmt"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// ListCvm list cvm.
// reference: https://cloud.tencent.com/document/api/213/15728
func (t *TCloud) ListCvm(kt *kit.Kit, opt *typecvm.TCloudListOption) ([]*cvm.Instance, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := cvm.NewDescribeInstancesRequest()
	if len(opt.CloudIDs) != 0 {
		req.InstanceIds = common.StringPtrs(opt.CloudIDs)
	}

	if opt.Page != nil {
		req.Offset = common.Int64Ptr(int64(opt.Page.Offset))
		req.Limit = common.Int64Ptr(int64(opt.Page.Limit))
	}

	resp, err := client.DescribeInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud instance failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Response.InstanceSet, nil
}

// DeleteCvm reference: https://cloud.tencent.com/document/api/213/15723
func (t *TCloud) DeleteCvm(kt *kit.Kit, opt *typecvm.TCloudDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "start cvm option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}

	req := cvm.NewTerminateInstancesRequest()
	req.InstanceIds = common.StringPtrs(opt.CloudIDs)

	_, err = client.TerminateInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("terminate cvm instance failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// StartCvm reference: https://cloud.tencent.com/document/api/213/15735
func (t *TCloud) StartCvm(kt *kit.Kit, opt *typecvm.TCloudStartOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "start cvm option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}
	req := cvm.NewStartInstancesRequest()
	req.InstanceIds = common.StringPtrs(opt.CloudIDs)

	_, err = client.StartInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("start cvm failed, err: %v, ids: %v, rid: %s", err, opt.CloudIDs, kt.Rid)
		return err
	}

	return nil
}

// StopCvm reference: https://cloud.tencent.com/document/api/213/15743
func (t *TCloud) StopCvm(kt *kit.Kit, opt *typecvm.TCloudStopOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "stop cvm option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}

	req := cvm.NewStopInstancesRequest()
	req.InstanceIds = common.StringPtrs(opt.CloudIDs)
	req.StopType = common.StringPtr(string(opt.StopType))
	req.StoppedMode = common.StringPtr(string(opt.StoppedMode))

	_, err = client.StopInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("stop cvm failed, err: %v, ids: %v, rid: %s", err, opt.CloudIDs, kt.Rid)
		return err
	}

	return nil
}

// RebootCvm reference: https://cloud.tencent.com/document/api/213/15742
func (t *TCloud) RebootCvm(kt *kit.Kit, opt *typecvm.TCloudRebootOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "reboot cvm option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}

	req := cvm.NewRebootInstancesRequest()
	req.InstanceIds = common.StringPtrs(opt.CloudIDs)
	req.StopType = common.StringPtr(string(opt.StopType))

	_, err = client.RebootInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("reboot cvm failed, err: %v, ids: %v, rid: %s", err, opt.CloudIDs, kt.Rid)
		return err
	}

	return nil
}

// ResetCvmPwd reference: https://cloud.tencent.com/document/api/213/15736
func (t *TCloud) ResetCvmPwd(kt *kit.Kit, opt *typecvm.TCloudResetPwdOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "reset pwd option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}

	req := cvm.NewResetInstancesPasswordRequest()
	req.InstanceIds = common.StringPtrs(opt.CloudIDs)
	req.Password = common.StringPtr(opt.Password)
	req.UserName = common.StringPtr(opt.UserName)
	req.ForceStop = common.BoolPtr(opt.ForceStop)

	_, err = client.ResetInstancesPasswordWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("reset cvm instance's password failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// CreateCvm reference: https://cloud.tencent.com/document/api/213/15730
// NOTE：返回实例`ID`列表并不代表实例创建成功，可根据 [DescribeInstances](https://cloud.tencent.com/document/api/213/15728)
// 接口查询返回的InstancesSet中对应实例的`ID`的状态来判断创建是否完成；如果实例状态由“PENDING(创建中)”变为“RUNNING(运行中)”，则为创建成功。
func (t *TCloud) CreateCvm(kt *kit.Kit, opt *typecvm.TCloudCreateOption) (ids []string, err error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}

	req := cvm.NewRunInstancesRequest()
	req.Placement = &cvm.Placement{
		Zone: common.StringPtr(opt.Zone),
	}
	req.InstanceType = common.StringPtr(opt.InstanceType)
	req.ImageId = common.StringPtr(opt.ImageID)
	req.InstanceCount = common.Int64Ptr(opt.RequiredCount)
	req.InstanceName = opt.Name
	req.SecurityGroupIds = common.StringPtrs(opt.SecurityGroupIDs)
	req.ClientToken = opt.ClientToken
	req.InstanceChargeType = common.StringPtr(string(opt.InstanceChargeType))
	req.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
		VpcId:    common.StringPtr(opt.VpcID),
		SubnetId: common.StringPtr(opt.SubnetID),
	}
	req.LoginSettings = &cvm.LoginSettings{
		Password: common.StringPtr(opt.Password),
	}

	req.SystemDisk = &cvm.SystemDisk{
		DiskId:   opt.SystemDisk.DiskID,
		DiskSize: opt.SystemDisk.DiskSizeGB,
	}
	if len(opt.SystemDisk.DiskType) != 0 {
		req.SystemDisk.DiskType = common.StringPtr(string(opt.SystemDisk.DiskType))
	}

	if len(opt.DataDisk) != 0 {
		req.DataDisks = make([]*cvm.DataDisk, 0, len(opt.DataDisk))
		for _, one := range opt.DataDisk {
			disk := &cvm.DataDisk{
				DiskSize: one.DiskSizeGB,
				DiskId:   one.DiskID,
			}

			if len(one.DiskType) != 0 {
				disk.DiskType = common.StringPtr(string(one.DiskType))
			}
			req.DataDisks = append(req.DataDisks, disk)
		}
	}

	if opt.InstanceChargePrepaid != nil {
		req.InstanceChargePrepaid = &cvm.InstanceChargePrepaid{
			Period: opt.InstanceChargePrepaid.Period,
		}

		if len(opt.InstanceChargePrepaid.RenewFlag) != 0 {
			req.InstanceChargePrepaid.RenewFlag = common.StringPtr(string(opt.InstanceChargePrepaid.RenewFlag))
		}
	}

	resp, err := client.RunInstancesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud cvm instance failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resourceIDs := make([]string, len(resp.Response.InstanceIdSet))
	for index, resourceID := range resp.Response.InstanceIdSet {
		resourceIDs[index] = *resourceID
	}

	return resourceIDs, nil
}
