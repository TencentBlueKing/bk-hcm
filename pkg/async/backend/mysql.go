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

package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/task"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// backend mysql backend
type backend struct {
	kt  *kit.Kit
	dao dao.Set
}

// NewBackend create backend instance
func NewBackend(dao dao.Set) Backend {
	return &backend{
		dao: dao,
	}
}

// SetBackendKit set backend kit
func (b *backend) SetBackendKit(kt *kit.Kit) {
	b.kt = kt
}

// ConsumeOnePendingFlow consume one pending flow
func (b *backend) ConsumeOnePendingFlow() (*task.AsyncFlow, error) {
	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "state",
					Op:    filter.Equal.Factory(),
					Value: enumor.FlowPending,
				},
			},
		},
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: 1,
			Sort:  "created_at",
			Order: core.Descending,
		},
	}
	daoResp, err := b.dao.AsyncFlow().List(b.kt, opt)
	if err != nil {
		logs.Errorf("[async] [module-backends] list async flow err: %v, rid: %s", err, b.kt.Rid)
		return nil, fmt.Errorf("list async flow failed, err: %v", err)
	}

	if len(daoResp.Details) != 1 {
		if len(daoResp.Details) != 0 {
			return nil, errors.New("get flow num wrong")
		}
		return nil, errors.New("flow num is 0")
	}

	if err := b.dao.AsyncFlow().UpdateByIDCAS(b.kt, daoResp.Details[0].ID, enumor.FlowRunning); err != nil {
		logs.Errorf("[async] [module-backends] update flow state to running err: %v, rid: %s", err, b.kt.Rid)
		return nil, err
	}

	return &task.AsyncFlow{
		ID:        daoResp.Details[0].ID,
		Name:      daoResp.Details[0].Name,
		State:     daoResp.Details[0].State,
		Memo:      daoResp.Details[0].Memo,
		Reason:    daoResp.Details[0].Reason,
		ShareData: daoResp.Details[0].ShareData,
		Revision: core.Revision{
			Creator:   daoResp.Details[0].Creator,
			Reviser:   daoResp.Details[0].Reviser,
			CreatedAt: daoResp.Details[0].CreatedAt.String(),
			UpdatedAt: daoResp.Details[0].CreatedAt.String(),
		},
	}, nil
}

// GetFlowsByCount get flows by count from backend
func (b *backend) GetFlowsByCount(flowCount int) ([]task.AsyncFlow, error) {
	opt := &types.ListOption{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "state",
					Op:    filter.Equal.Factory(),
					Value: enumor.FlowPending,
				},
			},
		},
		Page: &core.BasePage{
			Count: false,
			Start: 0,
			Limit: uint(flowCount),
			Sort:  "created_at",
			Order: core.Descending,
		},
	}
	daoResp, err := b.dao.AsyncFlow().List(b.kt, opt)
	if err != nil {
		logs.Errorf("[async] [module-backends] list async flow err: %v, rid: %s", err, b.kt.Rid)
		return nil, fmt.Errorf("list async flow failed, err: %v", err)
	}

	if len(daoResp.Details) != flowCount {
		return nil, errors.New("input num not equal output num")
	}

	ret := make([]task.AsyncFlow, 0, len(daoResp.Details))
	for _, one := range daoResp.Details {
		tmp := task.AsyncFlow{
			ID:        one.ID,
			Name:      one.Name,
			State:     one.State,
			Memo:      one.Memo,
			Reason:    one.Reason,
			ShareData: one.ShareData,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.CreatedAt.String(),
			},
		}
		ret = append(ret, tmp)
	}

	return ret, nil
}

// AddFlow add flow into backend
func (b *backend) AddFlow(req *taskserver.AddFlowReq) (string, error) {
	flowIds, err := b.dao.Txn().AutoTxn(b.kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tableasync.AsyncFlowTable, 0)
		models = append(models, tableasync.AsyncFlowTable{
			Name:      string(req.FlowName),
			State:     enumor.FlowPending,
			Reason:    constant.DefaultJsonValue,
			ShareData: constant.DefaultJsonValue,
			Creator:   b.kt.User,
			Reviser:   b.kt.User,
		})

		ids, err := b.dao.AsyncFlow().BatchCreateWithTx(b.kt, txn, models)
		if err != nil {
			return nil, fmt.Errorf("create async flow failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("[async] [module-backends] create async flow err: %v, rid: %s", err, b.kt.Rid)
		return "", err
	}

	ids, ok := flowIds.([]string)
	if !ok {
		return "", fmt.Errorf("create async flow but return id type not string, id type: %v",
			reflect.TypeOf(flowIds).String())
	}

	if len(ids) <= 0 {
		return "", errors.New("no flow id")
	}

	return ids[0], nil
}

// SetFlowChange set flow's change
func (b *backend) SetFlowChange(flowID string, flowChange *FlowChange) error {
	if flowChange == nil {
		return errors.New("there is change to this flow")
	}

	if err := flowChange.State.Validate(); err != nil {
		return err
	}

	_, err := b.dao.Txn().AutoTxn(b.kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		reasonJson, err := json.Marshal(&tableasync.AsyncFlowReason{
			Message: flowChange.Reason,
		})
		if err != nil {
			logs.Errorf("[async] [module-backends] marshal flow reason err: %v, rid: %s", err, b.kt.Rid)
			return nil, err
		}

		if flowChange.ShareData == "" {
			flowChange.ShareData = constant.DefaultJsonValue
		}

		model := &tableasync.AsyncFlowTable{
			State:     flowChange.State,
			ShareData: tabletypes.JsonField(flowChange.ShareData),
			Reason:    tabletypes.JsonField(reasonJson),
			Reviser:   b.kt.User,
		}

		if err := b.dao.AsyncFlow().UpdateByIDWithTx(b.kt, txn, flowID, model); err != nil {
			logs.Errorf("[async] [module-backends] update flow err: %v, rid: %s", err, b.kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		logs.Errorf("[async] [module-backends] update flow err: %v, rid: %s", err, b.kt.Rid)
		return err
	}

	return nil
}

// GetFlowByID get flow by id
func (b *backend) GetFlowByID(flowID string) (*task.AsyncFlow, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", flowID),
		Page:   core.NewDefaultBasePage(),
	}
	daoResp, err := b.dao.AsyncFlow().List(b.kt, opt)
	if err != nil {
		logs.Errorf("[async] [module-backends] list async flow err: %v, rid: %s", err, b.kt.Rid)
		return nil, fmt.Errorf("list async flow failed, err: %v", err)
	}

	if len(daoResp.Details) != 1 {
		return nil, errors.New("get flow not one")
	}

	one := daoResp.Details[0]
	return &task.AsyncFlow{
		ID:        one.ID,
		Name:      one.Name,
		State:     one.State,
		Memo:      one.Memo,
		Reason:    one.Reason,
		ShareData: one.ShareData,
		Revision: core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.CreatedAt.String(),
		},
	}, nil
}

// GetFlows get flows from backend
func (b *backend) GetFlows(req *taskserver.FlowListReq) ([]*task.AsyncFlow, error) {
	if req == nil {
		return nil, errors.New("req is null")
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
	}
	daoResp, err := b.dao.AsyncFlow().List(b.kt, opt)
	if err != nil {
		logs.Errorf("[async] [module-backends] list async flow err: %v, rid: %s", err, b.kt.Rid)
		return nil, fmt.Errorf("list async flow failed, err: %v", err)
	}

	ret := make([]*task.AsyncFlow, 0, len(daoResp.Details))
	for _, one := range daoResp.Details {
		tmp := &task.AsyncFlow{
			ID:        one.ID,
			Name:      one.Name,
			State:     one.State,
			Memo:      one.Memo,
			Reason:    one.Reason,
			ShareData: one.ShareData,
			Revision: core.Revision{
				Creator:   one.Creator,
				Reviser:   one.Reviser,
				CreatedAt: one.CreatedAt.String(),
				UpdatedAt: one.CreatedAt.String(),
			},
		}
		ret = append(ret, tmp)
	}

	return ret, nil
}

// AddTasks add tasks into backend
func (b *backend) AddTasks(tasks []task.AsyncFlowTask) error {
	if len(tasks) <= 0 {
		return fmt.Errorf("no task into addTasks")
	}

	_, err := b.dao.Txn().AutoTxn(b.kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tableasync.AsyncFlowTaskTable, 0, len(tasks))
		for _, task := range tasks {
			if task.Params == "" {
				task.Params = constant.DefaultJsonValue
			}
			if task.Reason == "" {
				task.Reason = constant.DefaultJsonValue
			}
			if task.ShareData == "" {
				task.ShareData = constant.DefaultJsonValue
			}
			models = append(models, tableasync.AsyncFlowTaskTable{
				ID:          task.ID,
				FlowID:      task.FlowID,
				FlowName:    task.FlowName,
				ActionName:  task.ActionName,
				Params:      task.Params,
				RetryCount:  task.RetryCount,
				TimeoutSecs: task.TimeoutSecs,
				DependOn:    strings.Join(task.DependOn, ","),
				State:       task.State,
				Memo:        task.Memo,
				Reason:      task.Reason,
				ShareData:   task.ShareData,
				Creator:     b.kt.User,
				Reviser:     b.kt.User,
			})
		}
		err := b.dao.AsyncFlowTask().BatchCreateWithTx(b.kt, txn, models)
		if err != nil {
			return nil, fmt.Errorf("create async task failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("[async] [module-backends] create async task err: %v, rid: %s", err, b.kt.Rid)
		return err
	}

	return nil
}

// GetTasks get tasks from backend
func (b *backend) GetTasks(taskIDs []string) ([]task.AsyncFlowTask, error) {
	if len(taskIDs) <= 0 {
		return nil, fmt.Errorf("no task id into getTasks")
	}

	allDaoRespDetails := make([]tableasync.AsyncFlowTaskTable, 0)
	split := slice.Split(taskIDs, constant.BatchOperationMaxLimit)
	for _, partIDs := range split {
		opt := &types.ListOption{
			Filter: tools.ContainersExpression("id", partIDs),
			Page:   core.NewDefaultBasePage(),
		}
		daoResp, err := b.dao.AsyncFlowTask().List(b.kt, opt)
		if err != nil {
			logs.Errorf("[async] [module-backends] list async flow task err: %v, rid: %s", err, b.kt.Rid)
			return nil, fmt.Errorf("list async flow task failed, err: %v", err)
		}

		allDaoRespDetails = append(allDaoRespDetails, daoResp.Details...)
	}

	ret := make([]task.AsyncFlowTask, 0, len(allDaoRespDetails))
	for _, one := range allDaoRespDetails {
		dependOn := strings.Split(one.DependOn, ",")
		tmp := task.AsyncFlowTask{
			ID:          one.ID,
			FlowID:      one.FlowID,
			FlowName:    one.FlowName,
			ActionName:  one.ActionName,
			Params:      one.Params,
			RetryCount:  one.RetryCount,
			TimeoutSecs: one.TimeoutSecs,
			DependOn:    dependOn,
			State:       one.State,
			Memo:        one.Memo,
			Reason:      one.Reason,
			ShareData:   one.ShareData,
		}
		ret = append(ret, tmp)
	}

	return ret, nil
}

// GetTasksByFlowID get tasks by flow id from backend
func (b *backend) GetTasksByFlowID(flowID string) ([]task.AsyncFlowTask, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("flow_id", flowID),
		Page:   core.NewDefaultBasePage(),
	}
	daoResp, err := b.dao.AsyncFlowTask().List(b.kt, opt)
	if err != nil {
		logs.Errorf("[async] [module-backends] list async flow task err: %v, rid: %s", err, b.kt.Rid)
		return nil, fmt.Errorf("list async flow task failed, err: %v", err)
	}

	if len(daoResp.Details) <= 0 {
		return nil, errors.New("can not find tasks by flow id")
	}

	ret := make([]task.AsyncFlowTask, 0, len(daoResp.Details))
	for _, one := range daoResp.Details {
		dependOn := strings.Split(one.DependOn, ",")
		if one.DependOn == "" {
			dependOn = []string{}
		}
		tmp := task.AsyncFlowTask{
			ID:          one.ID,
			FlowID:      one.FlowID,
			FlowName:    one.FlowName,
			ActionName:  one.ActionName,
			Params:      one.Params,
			RetryCount:  one.RetryCount,
			TimeoutSecs: one.TimeoutSecs,
			DependOn:    dependOn,
			State:       one.State,
			Memo:        one.Memo,
			Reason:      one.Reason,
			ShareData:   one.ShareData,
		}
		ret = append(ret, tmp)
	}

	return ret, nil
}

// SetTaskChange set task's change
func (b *backend) SetTaskChange(taskID string, taskChange *TaskChange) error {
	if err := taskChange.State.Validate(); err != nil {
		return err
	}

	_, err := b.dao.Txn().AutoTxn(b.kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		reasonJson, err := json.Marshal(&tableasync.AsyncFlowTaskReason{
			Message: taskChange.Reason,
		})
		if err != nil {
			logs.Errorf("[async] [module-backends] marshal task reason err: %v, rid: %s", err, b.kt.Rid)
			return nil, err
		}

		if taskChange.ShareData == "" {
			taskChange.ShareData = constant.DefaultJsonValue
		}

		model := &tableasync.AsyncFlowTaskTable{
			State:     taskChange.State,
			Reason:    tabletypes.JsonField(reasonJson),
			ShareData: tabletypes.JsonField(taskChange.ShareData),
			Reviser:   b.kt.User,
		}

		if err := b.dao.AsyncFlowTask().UpdateByIDWithTx(b.kt, txn, taskID, model); err != nil {
			logs.Errorf("[async] [module-backends] update task err: %v, rid: %s", err, b.kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		logs.Errorf("[async] [module-backends] update task err: %v, rid: %s", err, b.kt.Rid)
		return err
	}

	return nil
}

// MakeTaskIDs make task ids
func (b *backend) MakeTaskIDs(num int) ([]string, error) {
	return b.dao.AsyncFlowTask().GenIDs(b.kt, num)
}
