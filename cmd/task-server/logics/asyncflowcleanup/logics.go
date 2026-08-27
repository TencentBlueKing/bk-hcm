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

// Package asyncflowcleanup 提供 async_flow / async_flow_task 两张表的历史数据清理能力。
package asyncflowcleanup

import (
	"errors"
	"sync"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// 清理入口的两种「预期内跳过」，调用方用 errors.Is 判定。
// 定义为哨兵错误而非只靠 errf.Aborted 错误码：Aborted 是全项目通用的业务异常码，
// 下层任何一处返回 Aborted 的真实故障都会被误判成跳过，从而被静默吞掉。
var (
	// ErrCleanupDisabled 清理开关已关闭，本轮不执行。
	ErrCleanupDisabled = errf.New(errf.Aborted, "async flow and task cleanup is disabled")
	// ErrCleanupRunning 上一轮清理尚未结束，本轮不执行。
	ErrCleanupRunning = errf.New(errf.Aborted, "async flow and task cleanup is already running")
)

// flowBatchSize 单批处理的 flow 条数, 最大不能超过500。
const flowBatchSize uint = 100

// scanWindow 定位起点游标时单次扫描读取的行数。
const scanWindow = core.DefaultMaxPageLimit

// cleanupFlowNames 参与清理的 flow name 白名单，需要清理新的 flow 类型时在此追加即可。
// 元素个数受 filter 的 IN 上限（filter.DefaultMaxInLimit）约束。
var cleanupFlowNames = []enumor.FlowName{
	enumor.FlowBillMainAccountSummary,
}

// IsSkipped reports whether the error means the cleanup round is skipped as expected,
// 而不是执行过程中真的出错了。
func IsSkipped(err error) bool {
	return errors.Is(err, ErrCleanupDisabled) || errors.Is(err, ErrCleanupRunning)
}

// Result describes the outcome of one cleanup round.
type Result struct {
	// DeletedFlowCount 本轮累计删除的 async_flow 条数
	DeletedFlowCount int `json:"deleted_flow_count"`
	// DeletedTaskCount 本轮累计删除的 async_flow_task 条数
	DeletedTaskCount int `json:"deleted_task_count"`
	// DurationMs 本轮总耗时，单位：毫秒。不用 time.Duration 是因为它序列化后是纳秒整数，运维不易读。
	DurationMs int64 `json:"duration_ms"`
	// Interrupted 标记本轮是否因服务优雅退出被中断，为 true 时剩余超期数据留待下一周期处理
	Interrupted bool `json:"interrupted"`
}

// Logics holds the async flow and task cleanup logics.
type Logics struct {
	store   flowStore
	tenants tenantLister
	// mu 进程内防重入锁，保证同一时刻只有一轮清理在执行
	mu sync.Mutex
	// startCursors 按租户记录已确认可以跳过的主键位置，轮次之间保留，供下一轮续扫。
	startCursors map[string]string
}

// NewLogics creates a new async flow and task cleanup logics.
func NewLogics(daoSet dao.Set, ds *dataservice.Client) *Logics {
	return &Logics{
		store:        &daoFlowStore{dao: daoSet},
		tenants:      &dsTenantLister{ds: ds},
		startCursors: make(map[string]string),
	}
}

// Cleanup executes one cleanup round, it returns ErrCleanupDisabled if the cleanup is
// disabled by configuration, or ErrCleanupRunning if another round is still running.
func (l *Logics) Cleanup(kt *kit.Kit) (*Result, error) {
	if !l.mu.TryLock() {
		return nil, ErrCleanupRunning
	}
	defer l.mu.Unlock()

	cfg := cc.TaskServer().AsyncFlowAndTaskCleanup
	if !cvt.PtrToVal(cfg.Enabled) {
		return nil, ErrCleanupDisabled
	}

	start := time.Now()
	cutoff := start.AddDate(0, 0, -cvt.PtrToVal(cfg.RetentionDays)).Format(constant.TimeStdFormat)
	interval := time.Duration(cvt.PtrToVal(cfg.BatchIntervalMs)) * time.Millisecond

	tenantIDs, err := l.tenants.ListAllTenantIDs(kt)
	if err != nil {
		logs.Errorf("list tenant for async flow and task cleanup failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	logs.Infof("start async flow and task cleanup, cutoff: %s, batchSize: %d, batchInterval: %s, "+
		"tenantCount: %d, operator: %s, rid: %s", cutoff, flowBatchSize, interval, len(tenantIDs), kt.User, kt.Rid)

	result, err := l.cleanupByTenant(kt, tenantIDs, cutoff, interval)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	result.DurationMs = duration.Milliseconds()

	if result.Interrupted {
		logs.Warnf("async flow and task cleanup interrupted, this round is partially done, remaining expired "+
			"data will be handled in the next round, deleted flow: %d, task: %d, duration: %s, rid: %s",
			result.DeletedFlowCount, result.DeletedTaskCount, duration, kt.Rid)
		return result, nil
	}

	logs.Infof("async flow and task cleanup success, deleted flow: %d, task: %d, duration: %s, rid: %s",
		result.DeletedFlowCount, result.DeletedTaskCount, duration, kt.Rid)

	return result, nil
}

// cleanupByTenant 逐个租户执行清理，每个租户用带该租户 id 的子 kit，让 DAO 按租户注入过滤条件。
func (l *Logics) cleanupByTenant(kt *kit.Kit, tenantIDs []string, cutoff string, interval time.Duration) (
	*Result, error) {

	total := new(Result)
	for _, tenantID := range tenantIDs {
		tenantKt := kt.NewSubKitWithTenant(tenantID)

		one, err := l.cleanupLoop(tenantKt, cutoff, interval)
		if err != nil {
			return nil, err
		}

		total.DeletedFlowCount += one.DeletedFlowCount
		total.DeletedTaskCount += one.DeletedTaskCount
		logs.Infof("async flow and task cleanup finished for tenant: %s, deleted flow: %d, task: %d, rid: %s",
			tenantID, one.DeletedFlowCount, one.DeletedTaskCount, tenantKt.Rid)

		// 服务优雅退出，剩余租户留待下一周期处理
		if one.Interrupted {
			total.Interrupted = true
			return total, nil
		}
	}

	return total, nil
}

// cleanupLoop 以主键为游标分批清理单个租户，直到查不到满足条件的 flow 为止。
func (l *Logics) cleanupLoop(kt *kit.Kit, cutoff string, interval time.Duration) (*Result, error) {
	result := new(Result)
	lastFlowID, err := l.locateStartCursor(kt)
	if err != nil {
		return nil, err
	}

	for {
		flowIDs, err := l.store.ListExpiredFlowIDs(kt, cutoff, lastFlowID, flowBatchSize)
		if err != nil {
			logs.Errorf("list expired async flow failed, err: %v, tenant: %s, cutoff: %s, lastFlowID: %s, "+
				"deletedFlow: %d, deletedTask: %d, rid: %s",
				err, kt.TenantID, cutoff, lastFlowID, result.DeletedFlowCount, result.DeletedTaskCount, kt.Rid)
			return nil, err
		}

		if len(flowIDs) == 0 {
			return result, nil
		}

		taskCount, err := l.store.DeleteFlowsWithTasks(kt, flowIDs)
		if err != nil {
			logs.Errorf("delete expired async flow and task failed, err: %v, tenant: %s, flowCount: %d, "+
				"deletedFlow: %d, deletedTask: %d, rid: %s", err, kt.TenantID, len(flowIDs), result.DeletedFlowCount,
				result.DeletedTaskCount, kt.Rid)
			return nil, err
		}

		result.DeletedFlowCount += len(flowIDs)
		result.DeletedTaskCount += taskCount
		lastFlowID = flowIDs[len(flowIDs)-1]
		logs.Infof("async flow and task cleanup batch success, flow: %d, task: %d, cursor: %s, rid: %s",
			len(flowIDs), taskCount, lastFlowID, kt.Rid)

		// 返回条数不足一批，说明已经没有更多超期记录，本租户清理结束
		if len(flowIDs) < int(flowBatchSize) {
			return result, nil
		}

		if sleepBetweenBatch(kt, interval) {
			result.Interrupted = true
			return result, nil
		}
	}
}

// locateStartCursor 用只带主键条件的有界扫描前扫，定位到第一条 name 命中的记录之前，
func (l *Logics) locateStartCursor(kt *kit.Kit) (string, error) {
	lastFlowID := l.startCursors[kt.TenantID]
	scanCount := 0

	for {
		rows, err := l.store.ScanFlowsAfter(kt, lastFlowID, scanWindow)
		if err != nil {
			logs.Errorf("scan async flow for start cursor failed, err: %v, tenant: %s, lastFlowID: %s, rid: %s",
				err, kt.TenantID, lastFlowID, kt.Rid)
			return "", err
		}
		scanCount++

		// 扫到表尾都没有 name 命中的记录，游标停在表尾，本轮不会删到任何东西
		if len(rows) == 0 {
			break
		}

		cursor, found := pickStartCursor(rows, lastFlowID)
		lastFlowID = cursor
		if found {
			break
		}
	}

	l.startCursors[kt.TenantID] = lastFlowID
	logs.Infof("locate async flow cleanup start cursor done, tenant: %s, cursor: %s, scan: %d, rid: %s",
		kt.TenantID, lastFlowID, scanCount, kt.Rid)

	return lastFlowID, nil
}

// pickStartCursor 从一个扫描窗口里找出第一条 name 命中的 flow，返回可用作起点的游标。
func pickStartCursor(rows []flowBrief, lastFlowID string) (cursor string, found bool) {
	if len(rows) == 0 {
		return lastFlowID, false
	}

	for i, one := range rows {
		if !slice.IsItemInSlice(cleanupFlowNames, one.Name) {
			continue
		}

		if i == 0 {
			return lastFlowID, true
		}
		return rows[i-1].ID, true
	}

	return rows[len(rows)-1].ID, false
}

// sleepBetweenBatch 批间限速等待，服务优雅退出时立即中断并返回 true。
func sleepBetweenBatch(kt *kit.Kit, interval time.Duration) (interrupted bool) {
	if interval <= 0 {
		return false
	}

	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-timer.C:
		return false
	case <-kt.Ctx.Done():
		return true
	}
}
