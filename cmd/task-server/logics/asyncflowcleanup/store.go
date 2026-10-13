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

package asyncflowcleanup

import (
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// flowStore 定义清理逻辑依赖的 DB 操作，抽成接口是为了让主流程在单元测试中可以注入桩实现。
type flowStore interface {
	// ListExpiredFlowIDs 以主键为游标查询一批超期 flow 的 id。
	ListExpiredFlowIDs(kt *kit.Kit, cutoff, lastFlowID string, limit uint) ([]string, error)
	// ScanFlowsAfter 以主键为游标向后扫描固定行数的 flow 概要，仅用于定位起点游标。
	ScanFlowsAfter(kt *kit.Kit, lastFlowID string, limit uint) ([]flowBrief, error)
	// DeleteFlowsWithTasks 在同一事务内删除给定 flow 及其名下全部 task，返回删除的 task 条数。
	DeleteFlowsWithTasks(kt *kit.Kit, flowIDs []string) (int, error)
}

// flowBrief 定位起点游标所需的最小字段集。
type flowBrief struct {
	// ID flow 主键，同时用作扫描游标
	ID string
	// Name flow 名称，用于筛出目标 flow
	Name enumor.FlowName
}

// tenantLister 列出清理需要覆盖的租户 id，抽成接口同样是为了单元测试可注入桩实现。
type tenantLister interface {
	// ListAllTenantIDs 列出全部租户 id。
	ListAllTenantIDs(kt *kit.Kit) ([]string, error)
}

var _ tenantLister = new(dsTenantLister)

type dsTenantLister struct {
	ds *dataservice.Client
}

// ListAllTenantIDs 分页取回全部租户 id。
func (l *dsTenantLister) ListAllTenantIDs(kt *kit.Kit) ([]string, error) {
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewDefaultBasePage(),
	}

	tenantIDs := make([]string, 0)
	for {
		result, err := l.ds.Global.Tenant.List(kt, req)
		if err != nil {
			logs.Errorf("list tenant failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			tenantIDs = append(tenantIDs, one.TenantID)
		}

		if len(result.Details) < int(req.Page.Limit) {
			return tenantIDs, nil
		}
		req.Page.Start += uint32(req.Page.Limit)
	}
}

var _ flowStore = new(daoFlowStore)

type daoFlowStore struct {
	dao dao.Set
}

// buildExpiredFlowFilter 构造超期 flow 的过滤条件。
func buildExpiredFlowFilter(cutoff, lastFlowID string) *filter.Expression {
	return tools.ExpressionAnd(
		tools.RuleIn("name", cleanupFlowNames),
		tools.RuleLessThan("updated_at", cutoff),
		tools.RuleIDGreaterThan(lastFlowID),
	)
}

// buildScanFilter 构造定位起点游标用的过滤条件，只放主键游标一条规则。
func buildScanFilter(lastFlowID string) *filter.Expression {
	return tools.ExpressionAnd(tools.RuleIDGreaterThan(lastFlowID))
}

// buildFlowTaskFilter 构造某批 flow 名下 task 的过滤条件，flow_id 走 idx_flow_id 索引。
func buildFlowTaskFilter(flowIDs []string, lastTaskID string) *filter.Expression {
	return tools.ExpressionAnd(
		tools.RuleIn("flow_id", flowIDs),
		tools.RuleIDGreaterThan(lastTaskID),
	)
}

// ListExpiredFlowIDs 以主键为游标查询一批超期 flow 的 id。
func (s *daoFlowStore) ListExpiredFlowIDs(kt *kit.Kit, cutoff, lastFlowID string, limit uint) ([]string, error) {
	opt := &types.ListOption{
		Fields: []string{"id"},
		Filter: buildExpiredFlowFilter(cutoff, lastFlowID),
		Page:   &core.BasePage{Start: 0, Limit: limit, Sort: "id", Order: core.Ascending},
	}

	result, err := s.dao.AsyncFlow().List(kt, opt)
	if err != nil {
		logs.Errorf("list async flow failed, err: %v, cutoff: %s, lastFlowID: %s, rid: %s",
			err, cutoff, lastFlowID, kt.Rid)
		return nil, err
	}

	ids := make([]string, len(result.Details))
	for i, one := range result.Details {
		ids[i] = one.ID
	}

	return ids, nil
}

// ScanFlowsAfter 以主键为游标向后扫描固定行数的 flow 概要。
func (s *daoFlowStore) ScanFlowsAfter(kt *kit.Kit, lastFlowID string, limit uint) ([]flowBrief, error) {
	opt := &types.ListOption{
		Fields: []string{"id", "name"},
		Filter: buildScanFilter(lastFlowID),
		Page:   &core.BasePage{Start: 0, Limit: limit, Sort: "id", Order: core.Ascending},
	}

	result, err := s.dao.AsyncFlow().List(kt, opt)
	if err != nil {
		logs.Errorf("scan async flow failed, err: %v, lastFlowID: %s, limit: %d, rid: %s",
			err, lastFlowID, limit, kt.Rid)
		return nil, err
	}

	briefs := make([]flowBrief, len(result.Details))
	for i, one := range result.Details {
		briefs[i] = flowBrief{ID: one.ID, Name: one.Name}
	}

	return briefs, nil
}

// DeleteFlowsWithTasks 在同一事务内删除给定 flow 及其名下全部 task。
func (s *daoFlowStore) DeleteFlowsWithTasks(kt *kit.Kit, flowIDs []string) (int, error) {
	taskIDs, err := s.listTaskIDsByFlowIDs(kt, flowIDs)
	if err != nil {
		return 0, err
	}

	_, err = s.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, _ *orm.TxnOption) (interface{}, error) {
		return nil, s.deleteBatchWithTx(kt, txn, flowIDs, taskIDs)
	})
	if err != nil {
		logs.Errorf("delete async flow and task failed, err: %v, flowCount: %d, taskCount: %d, rid: %s", err,
			len(flowIDs), len(taskIDs), kt.Rid)
		return 0, err
	}

	return len(taskIDs), nil
}

// deleteBatchWithTx 先删 task 再删 flow，顺序不可颠倒：
// 万一事务外发生意外，残留的「flow 在、task 已删」可被下一轮重新识别，而孤儿 task 则永远无人认领。
func (s *daoFlowStore) deleteBatchWithTx(kt *kit.Kit, txn *sqlx.Tx, flowIDs, taskIDs []string) error {
	for _, chunk := range slice.Split(taskIDs, int(filter.DefaultMaxInLimit)) {
		if err := s.dao.AsyncFlowTask().DeleteWithTx(kt, txn, tools.ContainersExpression("id", chunk)); err != nil {
			logs.Errorf("delete async flow task failed, err: %v, count: %d, rid: %s", err, len(chunk), kt.Rid)
			return err
		}
	}

	for _, chunk := range slice.Split(flowIDs, int(filter.DefaultMaxInLimit)) {
		if err := s.dao.AsyncFlow().DeleteWithTx(kt, txn, tools.ContainersExpression("id", chunk)); err != nil {
			logs.Errorf("delete async flow failed, err: %v, count: %d, rid: %s", err, len(chunk), kt.Rid)
			return err
		}
	}

	return nil
}

// listTaskIDsByFlowIDs 查询给定 flow 名下的全部 task id，flow id 与结果均按 IN 上限分片翻页。
func (s *daoFlowStore) listTaskIDsByFlowIDs(kt *kit.Kit, flowIDs []string) ([]string, error) {
	taskIDs := make([]string, 0)

	for _, batch := range slice.Split(flowIDs, int(filter.DefaultMaxInLimit)) {
		lastTaskID := ""
		for {
			opt := &types.ListOption{
				Fields: []string{"id"},
				Filter: buildFlowTaskFilter(batch, lastTaskID),
				Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "id", Order: core.Ascending},
			}

			result, err := s.dao.AsyncFlowTask().List(kt, opt)
			if err != nil {
				logs.Errorf("list async flow task failed, err: %v, flowCount: %d, rid: %s",
					err, len(batch), kt.Rid)
				return nil, err
			}

			for _, one := range result.Details {
				taskIDs = append(taskIDs, one.ID)
			}

			if uint(len(result.Details)) < core.DefaultMaxPageLimit {
				break
			}
			lastTaskID = result.Details[len(result.Details)-1].ID
		}
	}

	return taskIDs, nil
}
