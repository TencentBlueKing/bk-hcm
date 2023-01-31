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

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// CreateSecurityGroup create security group.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0010.html
func (h *Huawei) CreateSecurityGroup(kt *kit.Kit, opt *types.HuaWeiSecurityGroupCreateOption) (
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
func (h *Huawei) DeleteSecurityGroup(kt *kit.Kit, opt *types.SecurityGroupDeleteOption) error {

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
func (h *Huawei) UpdateSecurityGroup(kt *kit.Kit, opt *types.HuaWeiSecurityGroupUpdateOption) error {

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
func (h *Huawei) ListSecurityGroup(kt *kit.Kit, opt *types.HuaWeiSecurityGroupListOption) (*[]model.SecurityGroup,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group update option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
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
		logs.Errorf("update huawei security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.SecurityGroups, nil
}
