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
	"fmt"
	"strings"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
)

// ListCvm list cvm.
// reference: https://support.huaweicloud.com/api-ecs/zh-cn_topic_0094148850.html
func (h *HuaWei) ListCvm(kt *kit.Kit, opt *typecvm.HuaWeiListOption) (*[]model.ServerDetail, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new ecs client failed, err: %v", err)
	}

	req := new(model.ListServersDetailsRequest)

	if len(opt.CloudIDs) != 0 {
		req.ServerId = converter.ValToPtr(strings.Join(opt.CloudIDs, ","))
	}

	if opt.Page != nil {
		req.Limit = converter.ValToPtr(opt.Page.Limit)
		req.Offset = converter.ValToPtr(opt.Page.Offset)
	}

	resp, err := client.ListServersDetails(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return nil, nil
		}
		logs.Errorf("list huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Servers, err
}

// DeleteCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0103.html
func (h *HuaWei) DeleteCvm(kt *kit.Kit, opt *typecvm.HuaWeiDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	req := &model.DeleteServersRequest{
		Body: &model.DeleteServersRequestBody{
			DeletePublicip: converter.ValToPtr(opt.DeletePublicIP),
			DeleteVolume:   converter.ValToPtr(opt.DeleteVolume),
			Servers:        svrIDs,
		},
	}

	_, err = client.DeleteServers(req)
	if err != nil {
		logs.Errorf("delete huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return err
}

// StartCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0301.html
func (h *HuaWei) StartCvm(kt *kit.Kit, opt *typecvm.HuaWeiStartOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "start option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	req := &model.BatchStartServersRequest{
		Body: &model.BatchStartServersRequestBody{
			OsStart: &model.BatchStartServersOption{
				Servers: svrIDs,
			},
		},
	}

	_, err = client.BatchStartServers(req)
	if err != nil {
		logs.Errorf("batch start huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return err
}

// StopCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0303.html
func (h *HuaWei) StopCvm(kt *kit.Kit, opt *typecvm.HuaWeiStopOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "stop option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	var stopType model.BatchStopServersOptionType
	if opt.Force {
		stopType = model.GetBatchStopServersOptionTypeEnum().SOFT
	} else {
		stopType = model.GetBatchStopServersOptionTypeEnum().HARD
	}

	req := &model.BatchStopServersRequest{
		Body: &model.BatchStopServersRequestBody{
			OsStop: &model.BatchStopServersOption{
				Type:    &stopType,
				Servers: svrIDs,
			},
		},
	}

	_, err = client.BatchStopServers(req)
	if err != nil {
		logs.Errorf("batch stop huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return err
}

// RebootCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0302.html
func (h *HuaWei) RebootCvm(kt *kit.Kit, opt *typecvm.HuaWeiRebootOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "reboot option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	var rebootType model.BatchRebootSeversOptionType
	if opt.Force {
		rebootType = model.GetBatchRebootSeversOptionTypeEnum().SOFT
	} else {
		rebootType = model.GetBatchRebootSeversOptionTypeEnum().HARD
	}

	req := &model.BatchRebootServersRequest{
		Body: &model.BatchRebootServersRequestBody{
			Reboot: &model.BatchRebootSeversOption{
				Type:    rebootType,
				Servers: svrIDs,
			},
		},
	}

	_, err = client.BatchRebootServers(req)
	if err != nil {
		logs.Errorf("batch reboot huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return err
}

// ResetCvmPwd reference: https://support.huaweicloud.com/api-ecs/ecs_02_0306.html
func (h *HuaWei) ResetCvmPwd(kt *kit.Kit, opt *typecvm.HuaWeiResetPwdOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "reset pwd option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	svrIDs := make([]model.ServerId, 0, len(opt.CloudIDs))
	for _, one := range opt.CloudIDs {
		svrIDs = append(svrIDs, model.ServerId{
			Id: one,
		})
	}

	req := &model.BatchResetServersPasswordRequest{
		Body: &model.BatchResetServersPasswordRequestBody{
			NewPassword: opt.Password,
			Servers:     svrIDs,
		},
	}

	_, err = client.BatchResetServersPassword(req)
	if err != nil {
		logs.Errorf("batch reset pwd huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return err
}

// CreateCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0101.html
func (h *HuaWei) CreateCvm(kt *kit.Kit, opt *typecvm.HuaWeiCreateOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "reset pwd option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return fmt.Errorf("new ecs client failed, err: %v", err)
	}

	req := &model.CreateServersRequest{
		XClientToken: opt.ClientToken,
		Body: &model.CreateServersRequestBody{
			Server: &model.PrePaidServer{
				ImageRef:         opt.ImageID,
				FlavorRef:        opt.InstanceType,
				Name:             opt.Name,
				AdminPass:        converter.ValToPtr(opt.Password),
				Vpcid:            opt.VpcID,
				Nics:             nil,
				Count:            converter.ValToPtr(opt.RequiredCount),
				RootVolume:       nil,
				DataVolumes:      nil,
				SecurityGroups:   nil,
				AvailabilityZone: converter.ValToPtr(opt.Zone),
				Description:      opt.Description,
			},
		},
	}

	if len(opt.SecurityGroupIDs) != 0 {
		req.Body.Server.SecurityGroups = new([]model.PrePaidServerSecurityGroup)
		for _, sgID := range opt.SecurityGroupIDs {
			*req.Body.Server.SecurityGroups = append(*req.Body.Server.SecurityGroups, model.PrePaidServerSecurityGroup{
				Id: converter.ValToPtr(sgID),
			})
		}
	}

	req.Body.Server.Nics = make([]model.PrePaidServerNic, len(opt.Nics))
	for _, nic := range opt.Nics {
		req.Body.Server.Nics = append(req.Body.Server.Nics, model.PrePaidServerNic{
			SubnetId:   nic.SubnetID,
			IpAddress:  nic.IPAddress,
			Ipv6Enable: nic.IPv6Enable,
		})
	}

	_, err = client.CreateServers(req)
	if err != nil {
		logs.Errorf("create huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return err
}
