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
	"errors"
	"fmt"
	"strconv"

	"hcm/pkg/adaptor/types/core"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	common "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// CreateSecurityGroup create security group.
// reference: https://cloud.tencent.com/document/api/215/15806
func (t *TCloudImpl) CreateSecurityGroup(kt *kit.Kit, opt *securitygroup.TCloudCreateOption) (*vpc.SecurityGroup,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	vpcCli, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewCreateSecurityGroupRequest()
	req.GroupName = common.StringPtr(opt.Name)
	req.GroupDescription = opt.Description

	resp, err := vpcCli.CreateSecurityGroupWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp == nil || resp.Response == nil || resp.Response.SecurityGroup == nil {
		return nil, errors.New("create tcloud security group return security group is nil")
	}

	return resp.Response.SecurityGroup, nil
}

// DeleteSecurityGroup delete security group.
// reference: https://cloud.tencent.com/document/api/215/15803
func (t *TCloudImpl) DeleteSecurityGroup(kt *kit.Kit, opt *securitygroup.TCloudDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteSecurityGroupRequest()
	req.SecurityGroupId = common.StringPtr(opt.CloudID)

	_, err = client.DeleteSecurityGroupWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UpdateSecurityGroup update security group.
// reference: https://cloud.tencent.com/document/api/215/15805
func (t *TCloudImpl) UpdateSecurityGroup(kt *kit.Kit, opt *securitygroup.TCloudUpdateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "tcloud security group update option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewModifySecurityGroupAttributeRequest()
	req.SecurityGroupId = common.StringPtr(opt.CloudID)
	req.GroupDescription = opt.Description
	if len(opt.Name) != 0 {
		req.GroupName = common.StringPtr(opt.Name)
	}

	_, err = client.ModifySecurityGroupAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("update tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSecurityGroupNew list security group.
// reference: https://cloud.tencent.com/document/api/215/15808
func (t *TCloudImpl) ListSecurityGroupNew(kt *kit.Kit, opt *securitygroup.TCloudListOption) ([]securitygroup.TCloudSG,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "tcloud security group list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeSecurityGroupsRequest()
	if len(opt.CloudIDs) != 0 {
		req.SecurityGroupIds = common.StringPtrs(opt.CloudIDs)
		req.Limit = converter.ValToPtr(strconv.FormatUint(core.TCloudQueryLimit, 10))
	}

	if opt.Page != nil {
		req.Offset = common.StringPtr(strconv.FormatInt(int64(opt.Page.Offset), 10))
		req.Limit = common.StringPtr(strconv.FormatInt(int64(opt.Page.Limit), 10))
	}

	resp, err := client.DescribeSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	sgs := make([]securitygroup.TCloudSG, 0, len(resp.Response.SecurityGroupSet))
	for _, one := range resp.Response.SecurityGroupSet {
		sgs = append(sgs, securitygroup.TCloudSG{one})
	}
	return sgs, nil
}

// CountSecurityGroup 基于 DescribeSecurityGroupsWithContext
// reference: https://cloud.tencent.com/document/api/215/15808
func (t *TCloudImpl) CountSecurityGroup(kt *kit.Kit, region string) (int32, error) {

	client, err := t.clientSet.VpcClient(region)
	if err != nil {
		return 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeSecurityGroupsRequest()
	req.Limit = converter.ValToPtr("1")
	resp, err := client.DescribeSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count tcloud security group failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return 0, err
	}
	return int32(*resp.Response.TotalCount), nil
}

// SecurityGroupCvmAssociate associate cvm.
// reference: https://cloud.tencent.com/document/api/213/31282
func (t *TCloudImpl) SecurityGroupCvmAssociate(kt *kit.Kit, opt *securitygroup.TCloudAssociateCvmOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "bind option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud cvm client failed, err: %v", err)
	}

	req := cvm.NewAssociateSecurityGroupsRequest()
	req.SecurityGroupIds = []*string{
		common.StringPtr(opt.CloudSecurityGroupID),
	}
	req.InstanceIds = []*string{
		common.StringPtr(opt.CloudCvmID),
	}

	_, err = client.AssociateSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("associate tcloud security group and cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupCvmDisassociate disassociate cvm.
// reference: https://cloud.tencent.com/document/api/213/31281
func (t *TCloudImpl) SecurityGroupCvmDisassociate(kt *kit.Kit, opt *securitygroup.TCloudAssociateCvmOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "bind option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud cvm client failed, err: %v", err)
	}

	req := cvm.NewDisassociateSecurityGroupsRequest()
	req.SecurityGroupIds = []*string{
		common.StringPtr(opt.CloudSecurityGroupID),
	}
	req.InstanceIds = []*string{
		common.StringPtr(opt.CloudCvmID),
	}

	_, err = client.DisassociateSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("disassociate tcloud security group and cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupCvmBatchAssociate  batch associate cvm.
// reference: https://cloud.tencent.com/document/api/213/31282
func (t *TCloudImpl) SecurityGroupCvmBatchAssociate(kt *kit.Kit,
	opt *securitygroup.TCloudBatchAssociateCvmOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "bind option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud cvm client failed, err: %v", err)
	}

	req := cvm.NewAssociateSecurityGroupsRequest()
	// 安全组id 接口只支持传一个
	req.SecurityGroupIds = []*string{
		common.StringPtr(opt.CloudSecurityGroupID),
	}
	req.InstanceIds = common.StringPtrs(opt.CloudCvmIDs)

	_, err = client.AssociateSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("associate tcloud security group and cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupCvmBatchDisassociate batch disassociate cvm.
// reference: https://cloud.tencent.com/document/api/213/31281
func (t *TCloudImpl) SecurityGroupCvmBatchDisassociate(kt *kit.Kit,
	opt *securitygroup.TCloudBatchAssociateCvmOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "bind option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new tcloud cvm client failed, err: %v", err)
	}

	req := cvm.NewDisassociateSecurityGroupsRequest()
	req.SecurityGroupIds = []*string{
		common.StringPtr(opt.CloudSecurityGroupID),
	}
	req.InstanceIds = common.StringPtrs(opt.CloudCvmIDs)

	_, err = client.DisassociateSecurityGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("disassociate tcloud security group and cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
