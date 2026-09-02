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
 * to the current version of the project delivered to anyone in the future.
 */

package lblogic

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	"hcm/pkg/api/core"
	coreasync "hcm/pkg/api/core/async"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	coretask "hcm/pkg/api/core/task"
	"hcm/pkg/api/data-service/task"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/require"
)

var cloudServerSetting = new(cc.CloudServerSetting)

func init() {
	cc.InitRuntime(cloudServerSetting)
}

func setClbCondSyncConfig(deleteBatchSize, upsertBatchSize int) {
	// 大批量阈值取整型上限，保证用例走常规批次大小
	setClbCondSyncLargeConfig(deleteBatchSize, upsertBatchSize, math.MaxInt, upsertBatchSize)
}

func setClbCondSyncLargeConfig(deleteBatchSize, upsertBatchSize, largeThreshold, largeBatchSize int) {
	cloudServerSetting.ConcurrentConfig.ClbCondSync = cc.ClbCondSync{
		DeleteBatchSize:      cvt.ValToPtr(deleteBatchSize),
		UpsertBatchSize:      cvt.ValToPtr(upsertBatchSize),
		LargeUpsertThreshold: cvt.ValToPtr(largeThreshold),
		LargeUpsertBatchSize: cvt.ValToPtr(largeBatchSize),
	}
}

func TestBuildBatches(t *testing.T) {
	setClbCondSyncConfig(2, 3)

	opt := &CondSyncLoadBalancerOption{
		Vendor:    enumor.TCloud,
		AccountID: "acc-1",
	}
	executor := &CondSyncLoadBalancerExecutor{opt: opt}

	tests := []struct {
		name        string
		regionDiffs []LoadBalancerRegionDiff
		wantBatches []struct {
			region   string
			isDelete bool
			count    int
		}
	}{
		{
			name:        "empty diffs",
			regionDiffs: nil,
			wantBatches: nil,
		},
		{
			name: "single region delete only",
			regionDiffs: []LoadBalancerRegionDiff{
				{Region: "ap-guangzhou", Delete: makeBriefs("lb-1", "lb-2", "lb-3")},
			},
			wantBatches: []struct {
				region   string
				isDelete bool
				count    int
			}{
				{"ap-guangzhou", true, 2},
				{"ap-guangzhou", true, 1},
			},
		},
		{
			name: "single region upsert only",
			regionDiffs: []LoadBalancerRegionDiff{
				{Region: "ap-guangzhou", Create: makeBriefs("lb-1", "lb-2"), Update: makeBriefs("lb-3")},
			},
			wantBatches: []struct {
				region   string
				isDelete bool
				count    int
			}{
				{"ap-guangzhou", false, 3},
			},
		},
		{
			name: "multi region delete before upsert",
			regionDiffs: []LoadBalancerRegionDiff{
				{Region: "ap-guangzhou", Delete: makeBriefs("lb-d1"), Create: makeBriefs("lb-c1")},
				{Region: "ap-shanghai", Delete: makeBriefs("lb-d2"), Create: makeBriefs("lb-c2")},
			},
			wantBatches: []struct {
				region   string
				isDelete bool
				count    int
			}{
				{"ap-guangzhou", true, 1},
				{"ap-shanghai", true, 1},
				{"ap-guangzhou", false, 1},
				{"ap-shanghai", false, 1},
			},
		},
		{
			name: "delete batch size split",
			regionDiffs: []LoadBalancerRegionDiff{
				{Region: "ap-guangzhou", Delete: makeBriefs("lb-1", "lb-2", "lb-3", "lb-4", "lb-5")},
			},
			wantBatches: []struct {
				region   string
				isDelete bool
				count    int
			}{
				{"ap-guangzhou", true, 2},
				{"ap-guangzhou", true, 2},
				{"ap-guangzhou", true, 1},
			},
		},
		{
			name: "upsert batch size split",
			regionDiffs: []LoadBalancerRegionDiff{
				{Region: "ap-guangzhou", Create: makeBriefs("lb-1", "lb-2", "lb-3", "lb-4", "lb-5", "lb-6", "lb-7")},
			},
			wantBatches: []struct {
				region   string
				isDelete bool
				count    int
			}{
				{"ap-guangzhou", false, 3},
				{"ap-guangzhou", false, 3},
				{"ap-guangzhou", false, 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := executor.buildBatches(tt.regionDiffs)

			if tt.wantBatches == nil {
				require.Empty(t, batches)
				return
			}

			require.Len(t, batches, len(tt.wantBatches))
			for i, want := range tt.wantBatches {
				require.Equal(t, want.region, batches[i].region)
				require.Equal(t, want.isDelete, batches[i].isDelete)
				require.Len(t, batches[i].details, want.count)
			}
		})
	}
}

func TestBuildBatchesLargeUpsert(t *testing.T) {
	// 阈值 4：清理批 2 条一批，同步数量未超阈值时 3 条一批，超过阈值后 5 条一批
	setClbCondSyncLargeConfig(2, 3, 4, 5)

	executor := &CondSyncLoadBalancerExecutor{opt: &CondSyncLoadBalancerOption{
		Vendor:    enumor.TCloud,
		AccountID: "acc-1",
	}}

	// 同步数量等于阈值，仍按常规批次大小拆分
	batches := executor.buildBatches([]LoadBalancerRegionDiff{
		{Region: "ap-guangzhou", Create: makeBriefs("lb-1", "lb-2"), Update: makeBriefs("lb-3", "lb-4")},
	})
	require.Len(t, batches, 2)
	require.Len(t, batches[0].details, 3)
	require.Len(t, batches[1].details, 1)

	// 多地域同步数量之和超过阈值，各地域均改用大批次大小
	batches = executor.buildBatches([]LoadBalancerRegionDiff{
		{Region: "ap-guangzhou", Create: makeBriefs("lb-1", "lb-2", "lb-3")},
		{Region: "ap-shanghai", Update: makeBriefs("lb-4", "lb-5", "lb-6")},
	})
	require.Len(t, batches, 2)
	require.Equal(t, "ap-guangzhou", batches[0].region)
	require.Len(t, batches[0].details, 3)
	require.Equal(t, "ap-shanghai", batches[1].region)
	require.Len(t, batches[1].details, 3)

	// 清理批次不受同步数量阈值影响
	batches = executor.buildBatches([]LoadBalancerRegionDiff{
		{Region: "ap-guangzhou", Delete: makeBriefs("lb-d1", "lb-d2", "lb-d3"),
			Create: makeBriefs("lb-1", "lb-2", "lb-3", "lb-4", "lb-5")},
	})
	require.Len(t, batches, 3)
	require.True(t, batches[0].isDelete)
	require.Len(t, batches[0].details, 2)
	require.True(t, batches[1].isDelete)
	require.Len(t, batches[1].details, 1)
	require.False(t, batches[2].isDelete)
	require.Len(t, batches[2].details, 5)
}

func TestCondSyncLoadBalancerOptionBkBizID(t *testing.T) {
	tests := []struct {
		name string
		opt  *CondSyncLoadBalancerOption
		want int64
	}{
		{name: "nil option", opt: nil, want: constant.UnassignedBiz},
		{name: "nil bk_biz_id", opt: &CondSyncLoadBalancerOption{}, want: constant.UnassignedBiz},
		{name: "biz entry", opt: &CondSyncLoadBalancerOption{BkBizID: 213}, want: 213},
		{name: "unassigned", opt: &CondSyncLoadBalancerOption{BkBizID: constant.UnassignedBiz},
			want: constant.UnassignedBiz},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.opt.bkBizID())
		})
	}
}

func TestNewTaskDetails(t *testing.T) {
	opt := &CondSyncLoadBalancerOption{
		AccountID: "acc-1",
		BkBizID:   213,
	}
	executor := &CondSyncLoadBalancerExecutor{opt: opt, bkBizID: opt.bkBizID()}
	briefs := []corelb.LoadBalancerBrief{
		{CloudID: "lb-1", Region: "ap-nanjing", Address: "1.1.1.1", Domain: "lb.example.com"},
	}

	details := executor.newTaskDetails(enumor.CondSyncLbOpUpdate, briefs)
	require.Len(t, details, 1)
	require.Equal(t, enumor.CondSyncLbOpUpdate, details[0].param.Op)
	require.Equal(t, "acc-1", details[0].param.AccountID)
	require.Equal(t, "lb-1", details[0].param.CloudLbID)
	require.Equal(t, "ap-nanjing", details[0].param.Region)
	require.Equal(t, "1.1.1.1", details[0].param.ClbVipDomain)
	require.Equal(t, "lb.example.com", details[0].param.Domain)
	require.Equal(t, int64(213), executor.bkBizID)

	raw, err := json.Marshal(details[0].param)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"op"`)
	require.Contains(t, string(raw), `"account_id"`)
	require.Contains(t, string(raw), `"cloud_lb_id"`)
	require.Contains(t, string(raw), `"clb_vip_domain"`)
	require.Contains(t, string(raw), `"domain"`)
	require.NotContains(t, string(raw), `"cloud_id"`)
	require.NotContains(t, string(raw), `"bk_biz_id"`)
	require.NotContains(t, string(raw), " / ")
}

func TestNewTaskDetailsVipFallback(t *testing.T) {
	executor := &CondSyncLoadBalancerExecutor{opt: &CondSyncLoadBalancerOption{AccountID: "acc-1"}}

	ipv6Details := executor.newTaskDetails(enumor.CondSyncLbOpCreate, []corelb.LoadBalancerBrief{
		{CloudID: "lb-v6", AddressIPv6: "::1"},
	})
	require.Equal(t, "::1", ipv6Details[0].param.ClbVipDomain)
	require.Empty(t, ipv6Details[0].param.Domain)
}

func TestBuildFlowTasks(t *testing.T) {
	setClbCondSyncConfig(2, 3)

	opt := &CondSyncLoadBalancerOption{
		Vendor:    enumor.TCloud,
		AccountID: "acc-1",
	}
	executor := &CondSyncLoadBalancerExecutor{opt: opt}

	regionDiffs := []LoadBalancerRegionDiff{
		{Region: "ap-guangzhou", Delete: makeBriefs("lb-d1", "lb-d2"), Create: makeBriefs("lb-c1")},
	}
	executor.batches = executor.buildBatches(regionDiffs)

	// Simulate taskDetailID assignment after createTaskDetails
	for _, batch := range executor.batches {
		for i, detail := range batch.details {
			detail.taskDetailID = "detail-" + detail.param.CloudLbID
			_ = i
		}
	}

	flowTasks := executor.buildFlowTasks()

	require.Len(t, flowTasks, 2)

	// First task: delete, no DependOn
	require.Equal(t, enumor.ActionBatchTaskSyncLbDelete, flowTasks[0].ActionName)
	require.Empty(t, flowTasks[0].DependOn)

	// Second task: upsert, depends on first
	require.Equal(t, enumor.ActionBatchTaskSyncLbUpsert, flowTasks[1].ActionName)
	require.Len(t, flowTasks[1].DependOn, 1)
	require.Equal(t, flowTasks[0].ActionID, flowTasks[1].DependOn[0])

	// Verify params
	deleteParams, ok := flowTasks[0].Params.(*actionlb.BatchTaskSyncLbDeleteOption)
	require.True(t, ok)
	require.Equal(t, "ap-guangzhou", deleteParams.Region)
	require.Equal(t, []string{"lb-d1", "lb-d2"}, deleteParams.CloudIDs)
	require.Len(t, deleteParams.ManagementDetailIDs, 2)

	upsertParams, ok := flowTasks[1].Params.(*actionlb.BatchTaskSyncLbUpsertOption)
	require.True(t, ok)
	require.Equal(t, "ap-guangzhou", upsertParams.Region)
	require.Equal(t, []string{"lb-c1"}, upsertParams.CloudIDs)
	require.Len(t, upsertParams.ManagementDetailIDs, 1)
}

func TestBuildFlowTasksEmpty(t *testing.T) {
	setClbCondSyncConfig(2, 3)

	opt := &CondSyncLoadBalancerOption{
		Vendor:    enumor.TCloud,
		AccountID: "acc-1",
	}
	executor := &CondSyncLoadBalancerExecutor{opt: opt}
	executor.batches = executor.buildBatches(nil)

	flowTasks := executor.buildFlowTasks()
	require.Empty(t, flowTasks)
}

func makeBriefs(cloudIDs ...string) []corelb.LoadBalancerBrief {
	briefs := make([]corelb.LoadBalancerBrief, 0, len(cloudIDs))
	for _, id := range cloudIDs {
		briefs = append(briefs, corelb.LoadBalancerBrief{CloudID: id})
	}
	return briefs
}

// stubTaskManagementLister stubs taskManagementLister for testing.
type stubTaskManagementLister struct {
	resp *task.ListManagementResult
	err  error
}

func (s *stubTaskManagementLister) List(kt *kit.Kit, req *core.ListReq) (*task.ListManagementResult, error) {
	return s.resp, s.err
}

// stubFlowLister stubs flowLister for testing.
type stubFlowLister struct {
	resp *ts.ListFlowResult
	err  error
}

func (s *stubFlowLister) ListFlow(kt *kit.Kit, req *core.ListReq) (*ts.ListFlowResult, error) {
	return s.resp, s.err
}

func TestCheckRunningCondSyncTask(t *testing.T) {
	kt := kit.New()

	tests := []struct {
		name        string
		opt         *CondSyncLoadBalancerOption
		mgmtResp    *task.ListManagementResult
		mgmtErr     error
		flowResp    *ts.ListFlowResult
		flowErr     error
		wantErr     bool
		wantErrCode int32
	}{
		{
			name: "no running management, allow",
			opt: &CondSyncLoadBalancerOption{
				Vendor:    enumor.TCloud,
				AccountID: "acc-1",
			},
			mgmtResp: &task.ListManagementResult{Details: []coretask.Management{}},
			wantErr:  false,
		},
		{
			name: "running management with empty flow_ids, allow",
			opt: &CondSyncLoadBalancerOption{
				Vendor:    enumor.TCloud,
				AccountID: "acc-1",
			},
			mgmtResp: &task.ListManagementResult{Details: []coretask.Management{
				{ID: "task-1", FlowIDs: []string{}},
			}},
			wantErr: false,
		},
		{
			name: "running management with all terminal flows, allow",
			opt: &CondSyncLoadBalancerOption{
				Vendor:    enumor.TCloud,
				AccountID: "acc-1",
			},
			mgmtResp: &task.ListManagementResult{Details: []coretask.Management{
				{ID: "task-1", FlowIDs: []string{"flow-1", "flow-2"}},
			}},
			flowResp: &ts.ListFlowResult{Details: []coreasync.AsyncFlow{
				{ID: "flow-1", State: enumor.FlowSuccess},
				{ID: "flow-2", State: enumor.FlowFailed},
			}},
			wantErr: false,
		},
		{
			name: "running management with non-terminal flow, reject",
			opt: &CondSyncLoadBalancerOption{
				Vendor:    enumor.TCloud,
				AccountID: "acc-1",
			},
			mgmtResp: &task.ListManagementResult{Details: []coretask.Management{
				{ID: "task-1", FlowIDs: []string{"flow-1"}},
			}},
			flowResp: &ts.ListFlowResult{Details: []coreasync.AsyncFlow{
				{ID: "flow-1", State: enumor.FlowRunning},
			}},
			wantErr:     true,
			wantErrCode: errf.TooManyRequest,
		},
		{
			name: "management list error",
			opt: &CondSyncLoadBalancerOption{
				Vendor:    enumor.TCloud,
				AccountID: "acc-1",
			},
			mgmtErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "flow list error",
			opt: &CondSyncLoadBalancerOption{
				Vendor:    enumor.TCloud,
				AccountID: "acc-1",
			},
			mgmtResp: &task.ListManagementResult{Details: []coretask.Management{
				{ID: "task-1", FlowIDs: []string{"flow-1"}},
			}},
			flowErr: errors.New("task server error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgmtCli := &stubTaskManagementLister{resp: tt.mgmtResp, err: tt.mgmtErr}
			flowCli := &stubFlowLister{resp: tt.flowResp, err: tt.flowErr}

			err := CheckRunningCondSyncTask(kt, mgmtCli, flowCli, tt.opt)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrCode != 0 {
					ef := errf.Error(err)
					require.NotNil(t, ef)
					require.Equal(t, tt.wantErrCode, ef.Code)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
