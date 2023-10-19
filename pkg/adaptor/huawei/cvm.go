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

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
)

// ListCvm list cvm.
// reference: https://support.huaweicloud.com/api-ecs/zh-cn_topic_0094148850.html
func (h *HuaWei) ListCvm(kt *kit.Kit, opt *typecvm.HuaWeiListOption) ([]typecvm.HuaWeiCvm, error) {

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
		return nil, err
	}

	cvms := make([]typecvm.HuaWeiCvm, 0, len(converter.PtrToVal(resp.Servers)))
	for _, one := range converter.PtrToVal(resp.Servers) {
		cvms = append(cvms, typecvm.HuaWeiCvm{one})
	}

	return cvms, err
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

	resp, err := client.BatchStartServers(req)
	if err != nil {
		logs.Errorf("batch start huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断批量操作任务是否失败
	handler := &jobPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.SubJob, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, []*string{resp.JobId}, types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	// 等待主机状态改变
	startHandler := &startCvmPollingHandler{
		opt.Region,
	}
	startPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: startHandler}
	_, err = startPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs),
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
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

	resp, err := client.BatchStopServers(req)
	if err != nil {
		logs.Errorf("batch stop huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断批量操作任务是否失败
	handler := &jobPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.SubJob, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, []*string{resp.JobId}, types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	// 等待主机状态改变
	stopHandler := &stopCvmPollingHandler{
		opt.Region,
	}
	stopPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: stopHandler}
	_, err = stopPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs), types.NewBatchOperateCvmPollerOpt())
	if err != nil {
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

	resp, err := client.BatchRebootServers(req)
	if err != nil {
		logs.Errorf("batch reboot huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 判断批量操作任务是否失败
	handler := &jobPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.SubJob, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, []*string{resp.JobId}, types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	// 等待主机状态改变
	rebootHandler := &rebootCvmPollingHandler{
		opt.Region,
	}
	rebootPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: rebootHandler}
	_, err = rebootPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs), types.NewBatchOperateCvmPollerOpt())
	if err != nil {
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

	handler := &resetpwdCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(h, kt, converter.SliceToPtr(opt.CloudIDs),
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	return err
}

// CreateCvm reference: https://support.huaweicloud.com/api-ecs/ecs_02_0101.html
func (h *HuaWei) CreateCvm(kt *kit.Kit, opt *typecvm.HuaWeiCreateOption) (*poller.BaseDoneResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "reset pwd option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new ecs client failed, err: %v", err)
	}

	volumeType, err := opt.RootVolume.VolumeType.RootVolumeType()
	if err != nil {
		return nil, err
	}

	chargingMode, err := opt.InstanceCharge.ChargingMode.ChargingMode()
	if err != nil {
		return nil, err
	}
	req := &model.CreateServersRequest{
		XClientToken: opt.ClientToken,
		Body: &model.CreateServersRequestBody{
			DryRun: converter.ValToPtr(opt.DryRun),
			Server: &model.PrePaidServer{
				ImageRef:  opt.CloudImageID,
				FlavorRef: opt.InstanceType,
				Name:      opt.Name,
				AdminPass: converter.ValToPtr(opt.Password),
				Vpcid:     opt.CloudVpcID,
				Nics: []model.PrePaidServerNic{
					{
						SubnetId: opt.CloudSubnetID,
					},
				},
				RootVolume: &model.PrePaidServerRootVolume{
					Volumetype: volumeType,
					Size:       converter.ValToPtr(opt.RootVolume.SizeGB),
				},
				Count:            converter.ValToPtr(opt.RequiredCount),
				AvailabilityZone: converter.ValToPtr(opt.Zone),
				Description:      opt.Description,
				Extendparam: &model.PrePaidServerExtendParam{
					ChargingMode: converter.ValToPtr(chargingMode),
					IsAutoRenew:  converter.ValToPtr(model.GetPrePaidServerExtendParamIsAutoRenewEnum().TRUE),
					IsAutoPay:    converter.ValToPtr(model.GetPrePaidServerExtendParamIsAutoPayEnum().TRUE),
				},
			},
		},
	}

	if opt.InstanceCharge.PeriodType != nil {
		periodType, err := opt.InstanceCharge.PeriodType.PeriodType()
		if err != nil {
			return nil, err
		}
		req.Body.Server.Extendparam.PeriodType = converter.ValToPtr(periodType)
		req.Body.Server.Extendparam.PeriodNum = opt.InstanceCharge.PeriodNum
	}

	if opt.InstanceCharge.IsAutoRenew != nil {
		if *opt.InstanceCharge.IsAutoRenew {
			req.Body.Server.Extendparam.IsAutoRenew = converter.ValToPtr(
				model.GetPrePaidServerExtendParamIsAutoRenewEnum().TRUE)
		} else {
			req.Body.Server.Extendparam.IsAutoRenew = converter.ValToPtr(
				model.GetPrePaidServerExtendParamIsAutoRenewEnum().FALSE)
		}
	}

	if len(opt.CloudSecurityGroupIDs) != 0 {
		req.Body.Server.SecurityGroups = new([]model.PrePaidServerSecurityGroup)
		for _, sgID := range opt.CloudSecurityGroupIDs {
			*req.Body.Server.SecurityGroups = append(*req.Body.Server.SecurityGroups, model.PrePaidServerSecurityGroup{
				Id: converter.ValToPtr(sgID),
			})
		}
	}

	if len(opt.DataVolume) != 0 {
		req.Body.Server.DataVolumes = new([]model.PrePaidServerDataVolume)
		for _, one := range opt.DataVolume {
			volType, err := one.VolumeType.DataVolumeType()
			if err != nil {
				return nil, err
			}
			*req.Body.Server.DataVolumes = append(*req.Body.Server.DataVolumes, model.PrePaidServerDataVolume{
				Volumetype: volType,
				Size:       one.SizeGB,
			})
		}
	}

	resp, err := client.CreateServers(req)
	if err != nil {
		logs.Errorf("create huawei cvm failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if opt.DryRun {
		return new(poller.BaseDoneResult), nil
	}

	handler := &createCvmPollingHandler{
		opt.Region,
	}
	respPoller := poller.Poller[*HuaWei, []model.ServerDetail, poller.BaseDoneResult]{Handler: handler}
	result, err := respPoller.PollUntilDone(h, kt, converter.SliceToPtr(converter.PtrToVal(resp.ServerIds)),
		types.NewBatchCreateCvmPollerOption())
	if err != nil {
		return nil, err
	}

	return result, err
}

type jobPollingHandler struct {
	region string
}

// Done ...
func (h *jobPollingHandler) Done(jobs []model.SubJob) (bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
		FailedMessage:   "",
	}
	for _, job := range jobs {
		if converter.PtrToVal(job.Status) == model.GetSubJobStatusEnum().RUNNING {
			return false, result
		}

		if converter.PtrToVal(job.Status) == model.GetSubJobStatusEnum().FAIL {
			result.FailedCloudIDs = append(result.FailedCloudIDs, converter.PtrToVal(job.Entities.ServerId))
			result.FailedMessage = converter.PtrToVal(job.FailReason)
		}

		if converter.PtrToVal(job.Status) == model.GetSubJobStatusEnum().SUCCESS {
			result.SuccessCloudIDs = append(result.SuccessCloudIDs, converter.PtrToVal(job.Entities.ServerId))
		}
	}

	return true, result
}

// Poll ...
func (h *jobPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.SubJob, error) {
	if len(cloudIDs) == 0 {
		return nil, errors.New("job id is required")
	}

	ecsCli, err := client.clientSet.ecsClient(h.region)
	if err != nil {
		logs.Errorf("new ecs client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("new ecs client failed, err: %v", err)
	}

	req := &model.ShowJobRequest{
		JobId: *cloudIDs[0],
	}
	resp, err := ecsCli.ShowJob(req)
	if err != nil {
		logs.Errorf("show job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return converter.PtrToVal(resp.Entities.SubJobs), nil
}

type startCvmPollingHandler struct {
	region string
}

// Done ...
func (h *startCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "ACTIVE")
}

// Poll ...
func (h *startCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
}

type stopCvmPollingHandler struct {
	region string
}

// Done ...
func (h *stopCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "SHUTOFF")
}

// Poll ...
func (h *stopCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
}

type resetpwdCvmPollingHandler struct {
	region string
}

// Done ...
func (h *resetpwdCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "ACTIVE")
}

// Poll ...
func (h *resetpwdCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
}

type rebootCvmPollingHandler struct {
	region string
}

// Done ...
func (h *rebootCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "ACTIVE")
}

// Poll ...
func (h *rebootCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
}

func done(cvms []model.ServerDetail, succeed string) (bool, *poller.BaseDoneResult) {
	result := new(poller.BaseDoneResult)

	flag := true
	for _, instance := range cvms {
		// not done
		if instance.Status != succeed {
			flag = false
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, instance.Id)
	}

	return flag, result
}

func poll(client *HuaWei, kt *kit.Kit, region string, cloudIDs []*string) ([]model.ServerDetail, error) {
	cloudIDSplit := slice.Split(cloudIDs, core.HuaWeiQueryLimit)

	cvms := make([]model.ServerDetail, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(model.ListServersDetailsRequest)
		req.ServerId = converter.ValToPtr(strings.Join(converter.PtrToSlice(partIDs), ","))
		req.Limit = converter.ValToPtr(int32(core.HuaWeiQueryLimit))

		cvmCli, err := client.clientSet.ecsClient(region)
		if err != nil {
			return nil, err
		}

		resp, err := cvmCli.ListServersDetails(req)
		if err != nil {
			return nil, err
		}

		cvms = append(cvms, *resp.Servers...)
	}

	return cvms, nil
}

type createCvmPollingHandler struct {
	region string
}

// Done ...
func (h *createCvmPollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
	}
	flag := true
	for _, instance := range cvms {
		// 创建中
		if instance.Status == "BUILD" {
			flag = false
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, instance.Id)
			continue
		}

		// 生产失败
		if instance.Status == "ERROR" || instance.Status == "UNKNOWN" {
			result.FailedCloudIDs = append(result.FailedCloudIDs, instance.Id)
			result.FailedMessage = instance.Fault.String()
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, instance.Id)
	}

	return flag, result
}

// Poll ...
func (h *createCvmPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {

	cloudIDSplit := slice.Split(cloudIDs, core.HuaWeiQueryLimit)

	cvms := make([]model.ServerDetail, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		req := new(model.ListServersDetailsRequest)
		req.ServerId = converter.ValToPtr(strings.Join(converter.PtrToSlice(partIDs), ","))
		req.Limit = converter.ValToPtr(int32(core.HuaWeiQueryLimit))

		cvmCli, err := client.clientSet.ecsClient(h.region)
		if err != nil {
			return nil, err
		}

		resp, err := cvmCli.ListServersDetails(req)
		if err != nil {
			return nil, err
		}

		cvms = append(cvms, *resp.Servers...)
	}

	if len(cvms) != len(cloudIDs) {
		return nil, fmt.Errorf("query cvm count: %d not equal return count: %d", len(cloudIDs), len(cvms))
	}

	return cvms, nil
}

var _ poller.PollingHandler[*HuaWei, []model.ServerDetail, poller.BaseDoneResult] = new(createCvmPollingHandler)
