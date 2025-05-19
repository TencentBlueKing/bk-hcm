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
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	coretask "hcm/pkg/api/core/task"
	datatask "hcm/pkg/api/data-service/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
)

const (
	// BatchTaskDefaultRetryTimes 批量任务默认重试次数
	BatchTaskDefaultRetryTimes = 3
	// BatchTaskDefaultRetryDelayMinMS 批量任务默认重试最小延迟时间
	BatchTaskDefaultRetryDelayMinMS = 600
	// BatchTaskDefaultRetryDelayMaxMS 批量任务默认重试最大延迟时间
	BatchTaskDefaultRetryDelayMaxMS = 1000
)

func listTaskDetail(kt *kit.Kit, ids []string) ([]coretask.Detail, error) {
	result := make([]coretask.Detail, 0, len(ids))
	for _, idBatch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		// 查询任务状态
		detailListReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", idBatch),
			Page:   core.NewDefaultBasePage(),
		}
		detailResp, err := actcli.GetDataService().Global.TaskDetail.List(kt, detailListReq)
		if err != nil {
			logs.Errorf("fail to query task detail, err: %v, ids: %s, rid: %s", err, ids, kt.Rid)
			return nil, err
		}
		if len(detailResp.Details) != len(idBatch) {
			return nil, fmt.Errorf("some of task management detail ids not found, want: %d, got: %d",
				len(ids), len(detailResp.Details))
		}
		result = append(result, detailResp.Details...)
	}

	return result, nil
}

func batchUpdateTaskDetailState(kt *kit.Kit, ids []string, state enumor.TaskDetailState) error {

	detailUpdates := make([]datatask.UpdateTaskDetailField, min(len(ids), constant.BatchOperationMaxLimit))
	for _, idBatch := range slice.Split(ids, constant.BatchOperationMaxLimit) {
		for i := range idBatch {
			detailUpdates[i] = datatask.UpdateTaskDetailField{ID: ids[i], State: state}
		}
		updateTaskReq := &datatask.UpdateDetailReq{Items: detailUpdates[:len(idBatch)]}
		rangeMS := [2]uint{BatchTaskDefaultRetryDelayMinMS, BatchTaskDefaultRetryDelayMaxMS}
		policy := retry.NewRetryPolicy(0, rangeMS)
		err := policy.BaseExec(kt, func() error {
			err := actcli.GetDataService().Global.TaskDetail.Update(kt, updateTaskReq)
			if err != nil {
				logs.Errorf("fail to update task detail state to %s, err: %v, ids: %s, rid: %s",
					state, err, idBatch, kt.Rid)
				return err
			}
			return nil
		})
		if err != nil {
			logs.Errorf("fail to update task detail state to %s after retry, err: %v, ids: %s, rid: %s",
				state, err, idBatch, kt.Rid)
			return err
		}
	}

	return nil
}

func batchUpdateTaskDetailResultState(kt *kit.Kit, ids []string, state enumor.TaskDetailState,
	result any, reason error) error {

	detailUpdates := make([]datatask.UpdateTaskDetailField, min(len(ids), constant.BatchOperationMaxLimit))
	for _, idBatch := range slice.Split(ids, constant.BatchOperationMaxLimit) {
		for i := range idBatch {
			field := datatask.UpdateTaskDetailField{ID: ids[i], State: state, Result: result}
			if reason != nil {
				// 需要截取否则超出DB字段长度限制，会更新状态失败
				runesReason := []rune(reason.Error())
				field.Reason = string(runesReason[:min(1000, len(runesReason))])
			}
			detailUpdates[i] = field
		}
		updateTaskReq := &datatask.UpdateDetailReq{Items: detailUpdates[:len(idBatch)]}
		rangeMS := [2]uint{BatchTaskDefaultRetryDelayMinMS, BatchTaskDefaultRetryDelayMaxMS}
		policy := retry.NewRetryPolicy(0, rangeMS)
		err := policy.BaseExec(kt, func() error {
			err := actcli.GetDataService().Global.TaskDetail.Update(kt, updateTaskReq)
			if err != nil {
				logs.Errorf("fail to update task detail result state to %s, err: %v, ids: %s, rid: %s",
					state, err, idBatch, kt.Rid)
				return err
			}
			return nil
		})
		if err != nil {
			logs.Errorf("fail to update task detail result state to %s after retry, err: %v, ids: %s, rid: %s",
				state, err, idBatch, kt.Rid)
			return err
		}
	}
	return nil
}

func getListenerWithLb(kt *kit.Kit, lblID string) (*corelb.BaseLoadBalancer,
	*corelb.BaseListener, error) {

	// 查询监听器数据
	listenerReq := &core.ListReq{
		Filter: tools.EqualExpression("id", lblID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	}
	lblResp, err := actcli.GetDataService().Global.LoadBalancer.ListListener(kt, listenerReq)
	if err != nil {
		logs.Errorf("fail to list tcloud listener, err: %v, id: %s, rid: %s", err, lblID, kt.Rid)
		return nil, nil, err
	}
	if len(lblResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.InvalidParameter, "lbl not found")
	}
	listener := lblResp.Details[0]

	// 查询负载均衡
	lbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", listener.LbID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	}
	lbResp, err := actcli.GetDataService().Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
	if err != nil {
		logs.Errorf("fail to list tcloud load balancer, err: %v, id: %s, rid: %s", err, listener.LbID, kt.Rid)
		return nil, nil, err
	}
	if len(lbResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "lb not found")
	}
	lb := lbResp.Details[0]
	return &lb, &listener, nil
}

// 七层规则不支持设置检查端口
func isHealthCheckChange(req *corelb.TCloudHealthCheckInfo, db *corelb.TCloudHealthCheckInfo, isL7 bool) bool {
	if req == nil {
		// 请求为空，默认参数
		return false
	}
	if db == nil {
		// 数据库为空，认为是默认参数
		return false
	}
	if !assert.IsPtrInt64Equal(req.HealthSwitch, db.HealthSwitch) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.TimeOut, db.TimeOut) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.IntervalTime, db.IntervalTime) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.HealthNum, db.HealthNum) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.UnHealthNum, db.UnHealthNum) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.HttpCode, db.HttpCode) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpCheckPath, db.HttpCheckPath) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpCheckDomain, db.HttpCheckDomain) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpCheckMethod, db.HttpCheckMethod) {
		return true
	}
	// 七层规则不支持设置检查端口, 这里不比较该数据
	if isL7 && !assert.IsPtrInt64Equal(req.CheckPort, db.CheckPort) {
		return true
	}
	if !assert.IsPtrStringEqual(req.ContextType, db.ContextType) {
		return true
	}
	if !assert.IsPtrStringEqual(req.SendContext, db.SendContext) {
		return true
	}
	if !assert.IsPtrStringEqual(req.RecvContext, db.RecvContext) {
		return true
	}
	if !assert.IsPtrStringEqual(req.CheckType, db.CheckType) {
		return true
	}
	if !assert.IsPtrStringEqual(req.HttpVersion, db.HttpVersion) {
		return true
	}
	if !assert.IsPtrInt64Equal(req.SourceIpType, db.SourceIpType) {
		return true
	}
	if !assert.IsPtrStringEqual(req.ExtendedCode, db.ExtendedCode) {
		return true
	}
	return false
}

func isListenerCertChange(want *corelb.TCloudCertificateInfo, db *corelb.TCloudCertificateInfo) bool {
	if want == nil {
		// 请求为空，默认参数
		return false
	}
	if db == nil {
		// 数据库为空，认为是默认参数
		return false
	}

	if !assert.IsPtrStringEqual(want.SSLMode, db.SSLMode) {
		return true
	}

	if !assert.IsPtrStringEqual(want.CaCloudID, db.CaCloudID) {
		return true
	}

	// 都有，但是数量不相等
	if len(db.CertCloudIDs) != len(want.CertCloudIDs) {
		// 数量不相等
		return true
	}
	// 要求证书按顺序相等。
	for i := range want.CertCloudIDs {
		if db.CertCloudIDs[i] != want.CertCloudIDs[i] {
			return true
		}
	}
	return false
}

// batchListListenerByIDs 根据监听器ID数组，批量获取监听器列表
func batchListListenerByIDs(kt *kit.Kit, lblIDs []string) ([]corelb.BaseListener, error) {
	if len(lblIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "listener ids is required")
	}

	// 查询监听器列表
	req := &core.ListReq{
		Filter: tools.ContainersExpression("id", lblIDs),
		Page:   core.NewDefaultBasePage(),
	}
	lblList := make([]corelb.BaseListener, 0)
	for {
		lblResp, err := actcli.GetDataService().Global.LoadBalancer.ListListener(kt, req)
		if err != nil {
			logs.Errorf("failed to list tcloud listener, err: %v, lblIDs: %v, rid: %s", err, lblIDs, kt.Rid)
			return nil, err
		}

		lblList = append(lblList, lblResp.Details...)
		if uint(len(lblResp.Details)) < core.DefaultMaxPageLimit {
			break
		}

		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return lblList, nil
}
