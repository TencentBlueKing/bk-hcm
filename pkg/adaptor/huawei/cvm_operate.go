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

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
)

/**
主机相关操作的接口集合
*/

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
	startHandler := &cvmOperatePollingHandler{
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
	stopHandler := &cvmOperatePollingHandler{
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
	rebootHandler := &cvmOperatePollingHandler{
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

	handler := &cvmOperatePollingHandler{
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

type cvmOperatePollingHandler struct {
	region string
}

// Done ...
func (h *cvmOperatePollingHandler) Done(cvms []model.ServerDetail) (bool, *poller.BaseDoneResult) {
	return done(cvms, "ACTIVE")
}

// Poll ...
func (h *cvmOperatePollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]model.ServerDetail, error) {
	return poll(client, kt, h.region, cloudIDs)
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
