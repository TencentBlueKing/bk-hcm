/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package actionlb

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	actionflow "hcm/cmd/task-server/logics/flow"
	"hcm/pkg/api/core"
	hcsync "hcm/pkg/api/hc-service/sync"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// --------------------------[批量操作-批量删除云上已不存在的负载均衡]-----------------------------

var _ action.Action = new(BatchTaskSyncLbDeleteAction)
var _ action.ParameterAction = new(BatchTaskSyncLbDeleteAction)

// BatchTaskSyncLbDeleteAction 批量操作-批量删除云上已不存在的负载均衡
type BatchTaskSyncLbDeleteAction struct{}

// BatchTaskSyncLbDeleteOption 批量操作-批量删除云上已不存在的负载均衡
type BatchTaskSyncLbDeleteOption struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	AccountID string        `json:"account_id" validate:"required"`
	Region    string        `json:"region" validate:"required"`
	// ManagementDetailIDs 对应的详情行id列表，需要和 CloudIDs 长度对应
	ManagementDetailIDs []string `json:"management_detail_ids" validate:"required,min=1"`
	// CloudIDs 本批次待删除的负载均衡云上ID列表
	CloudIDs []string `json:"cloud_ids" validate:"required,min=1"`
}

// Validate validate option.
func (opt BatchTaskSyncLbDeleteOption) Validate() error {
	if err := validateSyncLbVendor(opt.Vendor); err != nil {
		return err
	}

	if len(opt.ManagementDetailIDs) != len(opt.CloudIDs) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and cloud_ids num not match, %d != %d",
			len(opt.ManagementDetailIDs), len(opt.CloudIDs))
	}

	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskSyncLbDeleteAction) ParameterNew() (params any) {
	return new(BatchTaskSyncLbDeleteOption)
}

// Name return action name
func (act BatchTaskSyncLbDeleteAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskSyncLbDelete
}

// Run 批量删除云上已不存在的负载均衡，按云上ID精确删除，支持重入
func (act BatchTaskSyncLbDeleteAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskSyncLbDeleteOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type is not BatchTaskSyncLbDeleteOption")
	}

	reason, err := prepareSyncLbTaskDetails(kt.Kit(), opt.ManagementDetailIDs)
	if err != nil {
		return nil, err
	}
	if len(reason) > 0 {
		return reason, nil
	}

	defer func() {
		finishSyncLbTaskDetails(kt.Kit(), opt.ManagementDetailIDs, taskErr)
	}()

	// 使用异步来源的kit，令hc-service开启云上接口的限频重试
	if taskErr = act.deleteLoadBalancer(kt.AsyncKit(), opt); taskErr != nil {
		logs.Errorf("batch delete load balancer failed, err: %v, account: %s, region: %s, cloudIDs: %v, rid: %s",
			taskErr, opt.AccountID, opt.Region, opt.CloudIDs, kt.Kit().Rid)
		return nil, taskErr
	}

	logs.Infof("batch delete load balancer success, account: %s, region: %s, count: %d, rid: %s",
		opt.AccountID, opt.Region, len(opt.CloudIDs), kt.Kit().Rid)

	return nil, nil
}

// deleteLoadBalancer 调用hc-service删除本批负载均衡，云上ID由上游条件同步比对得出
func (act BatchTaskSyncLbDeleteAction) deleteLoadBalancer(kt *kit.Kit, opt *BatchTaskSyncLbDeleteOption) error {
	req := &hcsync.TCloudDelLoadBalancerByCondReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  opt.CloudIDs,
	}

	switch opt.Vendor {
	case enumor.TCloud:
		return actcli.GetHCService().TCloud.Clb.DeleteLoadBalancerByCond(kt, req)
	default:
		return fmt.Errorf("unsupport vendor for batch sync load balancer delete: %s", opt.Vendor)
	}
}

// Rollback 批量删除负载均衡支持重入，无需回滚
func (act BatchTaskSyncLbDeleteAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskSyncLbDeleteAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}

// --------------------------[批量操作-批量同步负载均衡]-----------------------------

var _ action.Action = new(BatchTaskSyncLbUpsertAction)
var _ action.ParameterAction = new(BatchTaskSyncLbUpsertAction)

// BatchTaskSyncLbUpsertAction 批量操作-批量同步负载均衡
type BatchTaskSyncLbUpsertAction struct{}

// BatchTaskSyncLbUpsertOption 批量操作-批量同步负载均衡
type BatchTaskSyncLbUpsertOption struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	AccountID string        `json:"account_id" validate:"required"`
	Region    string        `json:"region" validate:"required"`
	// ManagementDetailIDs 对应的详情行id列表，需要和 CloudIDs 长度对应
	ManagementDetailIDs []string `json:"management_detail_ids" validate:"required,min=1"`
	// CloudIDs 本批次待同步的负载均衡云上ID列表
	CloudIDs []string `json:"cloud_ids" validate:"required,min=1"`
	// TagFilters 标签过滤条件，与创建任务时的条件保持一致
	TagFilters core.MultiValueTagMap `json:"tag_filters,omitempty"`
}

// Validate validate option.
func (opt BatchTaskSyncLbUpsertOption) Validate() error {
	if err := validateSyncLbVendor(opt.Vendor); err != nil {
		return err
	}

	if len(opt.ManagementDetailIDs) != len(opt.CloudIDs) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and cloud_ids num not match, %d != %d",
			len(opt.ManagementDetailIDs), len(opt.CloudIDs))
	}

	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskSyncLbUpsertAction) ParameterNew() (params any) {
	return new(BatchTaskSyncLbUpsertOption)
}

// Name return action name
func (act BatchTaskSyncLbUpsertAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskSyncLbUpsert
}

// Run 批量同步负载均衡，同步为覆盖写，支持重入
func (act BatchTaskSyncLbUpsertAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskSyncLbUpsertOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type is not BatchTaskSyncLbUpsertOption")
	}

	reason, err := prepareSyncLbTaskDetails(kt.Kit(), opt.ManagementDetailIDs)
	if err != nil {
		return nil, err
	}
	if len(reason) > 0 {
		return reason, nil
	}

	defer func() {
		finishSyncLbTaskDetails(kt.Kit(), opt.ManagementDetailIDs, taskErr)
	}()

	// 使用异步来源的kit，令hc-service开启云上接口的限频重试
	if taskErr = act.syncLoadBalancer(kt.AsyncKit(), opt); taskErr != nil {
		logs.Errorf("batch sync load balancer failed, err: %v, account: %s, region: %s, cloudIDs: %v, rid: %s",
			taskErr, opt.AccountID, opt.Region, opt.CloudIDs, kt.Kit().Rid)
		return nil, taskErr
	}

	logs.Infof("batch sync load balancer success, account: %s, region: %s, count: %d, rid: %s",
		opt.AccountID, opt.Region, len(opt.CloudIDs), kt.Kit().Rid)

	return nil, nil
}

// syncLoadBalancer 调用hc-service同步本批负载均衡，同步不清理DB，云上已删除的由删除动作处理
func (act BatchTaskSyncLbUpsertAction) syncLoadBalancer(kt *kit.Kit, opt *BatchTaskSyncLbUpsertOption) error {
	req := &hcsync.TCloudSyncLoadBalancerByCondReq{
		AccountID:  opt.AccountID,
		Region:     opt.Region,
		CloudIDs:   opt.CloudIDs,
		TagFilters: opt.TagFilters,
	}

	switch opt.Vendor {
	case enumor.TCloud:
		return actcli.GetHCService().TCloud.Clb.SyncLoadBalancerByCond(kt, req)
	default:
		return fmt.Errorf("unsupport vendor for batch sync load balancer upsert: %s", opt.Vendor)
	}
}

// Rollback 批量同步负载均衡支持重入，无需回滚
func (act BatchTaskSyncLbUpsertAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskSyncLbUpsertAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}

func validateSyncLbVendor(vendor enumor.Vendor) error {
	switch vendor {
	case enumor.TCloud:
		return nil
	default:
		return fmt.Errorf("unsupport vendor for batch sync load balancer: %s", vendor)
	}
}

// prepareSyncLbTaskDetails 校验本批任务详情状态并置为执行中。
// 返回非空 reason 表示本批已被取消，调用方跳过本批且不视为失败。
func prepareSyncLbTaskDetails(kt *kit.Kit, detailIDs []string) (string, error) {
	detailList, err := actionflow.ListTaskDetail(kt, detailIDs)
	if err != nil {
		logs.Errorf("list task detail failed, err: %v, detailIDs: %v, rid: %s", err, detailIDs, kt.Rid)
		return "", err
	}

	for _, detail := range detailList {
		if detail.State == enumor.TaskDetailCancel {
			return fmt.Sprintf("task detail task: %s is canceled", detail.ID), nil
		}
		if detail.State != enumor.TaskDetailInit {
			return "", errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
				detail.ID, detail.State)
		}
	}

	if err = actionflow.BatchUpdateTaskDetailState(kt, detailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("update task detail state to running failed, err: %v, detailIDs: %v, rid: %s",
			err, detailIDs, kt.Rid)
		return "", err
	}

	return "", nil
}

// finishSyncLbTaskDetails 按本批执行结果写回任务详情状态，写回失败只记录日志，不影响动作结果
func finishSyncLbTaskDetails(kt *kit.Kit, detailIDs []string, taskErr error) {
	state := enumor.TaskDetailSuccess
	if taskErr != nil {
		state = enumor.TaskDetailFailed
	}

	if err := actionflow.BatchUpdateTaskDetailResultState(kt, detailIDs, state, nil, taskErr); err != nil {
		logs.Errorf("update task detail state to %s failed, err: %v, detailIDs: %v, rid: %s",
			state, err, detailIDs, kt.Rid)
	}
}
