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

package huawei

import (
	"errors"
	"fmt"
	"strings"

	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// CreateSecurityGroup create security group.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0010.html
func (h *HuaWei) CreateSecurityGroup(kt *kit.Kit, opt *securitygroup.HuaWeiCreateOption) (
	*model.SecurityGroupInfo, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &model.CreateSecurityGroupRequest{
		Body: &model.CreateSecurityGroupRequestBody{
			SecurityGroup: &model.CreateSecurityGroupOption{
				Name:        opt.Name,
				Description: opt.Description,
			},
		},
	}
	resp, err := client.CreateSecurityGroup(req)
	if err != nil {
		logs.Errorf("create huawei security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp == nil || resp.SecurityGroup == nil {
		return nil, errors.New("create huawei security group return security group id is nil")
	}

	return resp.SecurityGroup, nil
}

// DeleteSecurityGroup delete security group.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0014.html
func (h *HuaWei) DeleteSecurityGroup(kt *kit.Kit, opt *securitygroup.HuaWeiDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &model.DeleteSecurityGroupRequest{
		SecurityGroupId: opt.CloudID,
	}
	_, err = client.DeleteSecurityGroup(req)
	if err != nil {
		logs.Errorf("delete huawei security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UpdateSecurityGroup update security group.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0013.html
func (h *HuaWei) UpdateSecurityGroup(kt *kit.Kit, opt *securitygroup.HuaWeiUpdateOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group update option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &model.UpdateSecurityGroupRequest{
		SecurityGroupId: opt.CloudID,
		Body: &model.UpdateSecurityGroupRequestBody{
			SecurityGroup: &model.UpdateSecurityGroupOption{
				Description: opt.Description,
			},
		},
	}

	if len(opt.Name) != 0 {
		req.Body.SecurityGroup.Name = &opt.Name
	}

	_, err = client.UpdateSecurityGroup(req)
	if err != nil {
		logs.Errorf("update huawei security group failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSecurityGroup list security group.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0011.html
func (h *HuaWei) ListSecurityGroup(kt *kit.Kit, opt *securitygroup.HuaWeiListOption) ([]securitygroup.HuaWeiSG,
	*model.PageInfo, error) {

	if opt == nil {
		return nil, nil, errf.New(errf.InvalidParameter, "security group update option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(opt.Region)
	if err != nil {
		return nil, nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := new(model.ListSecurityGroupsRequest)
	if len(opt.CloudIDs) != 0 {
		req.Id = sliceToPtr[string](opt.CloudIDs)
	}

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	resp, err := client.ListSecurityGroups(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return nil, nil, nil
		}
		logs.Errorf("update huawei security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	sgs := make([]securitygroup.HuaWeiSG, 0, len(converter.PtrToVal(resp.SecurityGroups)))
	for _, one := range converter.PtrToVal(resp.SecurityGroups) {
		sgs = append(sgs, securitygroup.HuaWeiSG{one})
	}

	return sgs, resp.PageInfo, err
}

// SecurityGroupCvmAssociate associate cvm.
// reference: https://support.huaweicloud.com/api-ecs/ecs_03_0601.html
func (h *HuaWei) SecurityGroupCvmAssociate(kt *kit.Kit, opt *securitygroup.HuaWeiAssociateCvmOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "associate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	req := new(ecsmodel.NovaAssociateSecurityGroupRequest)
	req.ServerId = opt.CloudCvmID
	req.Body = &ecsmodel.NovaAssociateSecurityGroupRequestBody{
		AddSecurityGroup: &ecsmodel.NovaAddSecurityGroupOption{
			Name: opt.CloudSecurityGroupID,
		},
	}

	_, err = client.NovaAssociateSecurityGroup(req)
	if err != nil {
		logs.Errorf("associate tcloud security group and cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SecurityGroupCvmDisassociate disassociate cvm.
// reference: https://support.huaweicloud.com/api-ecs/ecs_03_0601.html
func (h *HuaWei) SecurityGroupCvmDisassociate(kt *kit.Kit, opt *securitygroup.HuaWeiAssociateCvmOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "disassociate option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	req := new(ecsmodel.NovaDisassociateSecurityGroupRequest)
	req.ServerId = opt.CloudCvmID
	req.Body = &ecsmodel.NovaDisassociateSecurityGroupRequestBody{
		RemoveSecurityGroup: &ecsmodel.NovaRemoveSecurityGroupOption{
			Name: opt.CloudSecurityGroupID,
		},
	}

	_, err = client.NovaDisassociateSecurityGroup(req)
	if err != nil {
		logs.Errorf("disassociate tcloud security group and cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
