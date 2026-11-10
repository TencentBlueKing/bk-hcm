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

package plugin

import (
	"fmt"
	"net/http"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

func (svc *service) initIaCPluginRoutes(h *rest.Handler) {
	h.Add("CreateIaCPlugin", http.MethodPost, "/iac_plugins/create", svc.CreateIaCPlugin)
	h.Add("UpdateIaCPlugin", http.MethodPatch, "/iac_plugins/update", svc.UpdateIaCPlugin)
	h.Add("BatchDeleteIaCPlugin", http.MethodDelete, "/iac_plugins/batch", svc.BatchDeleteIaCPlugin)
	h.Add("ListIaCPlugin", http.MethodPost, "/iac_plugins/list", svc.ListIaCPlugin)
}

// CreateIaCPlugin create iac plugin record.
func (svc *service) CreateIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IaCPluginCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create iac plugin decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("create iac plugin validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := []tableplugin.IaCPluginTable{{
			PluginName:    req.PluginName,
			PluginVersion: req.PluginVersion,
			Protocol:      req.Protocol,
			State:         req.State,
			Message:       cvt.ValToPtr(req.Message),
			Extension:     req.Extension,
			Memo:          req.Memo,
			Creator:       cts.Kit.User,
			Reviser:       cts.Kit.User,
		}}

		ids, err := svc.dao.IaCPlugin().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("create iac plugin failed, err: %v, plugin: %s@%s, rid: %s",
				err, req.PluginName, req.PluginVersion, cts.Kit.Rid)
			return nil, err
		}

		return &core.BatchCreateResult{IDs: ids}, nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// BatchDeleteIaCPlugin 按过滤条件查出目标记录，再按 ID 精准删除并写入审计。
func (svc *service) BatchDeleteIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IaCPluginBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete iac plugin decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete iac plugin validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	records := make([]tableplugin.IaCPluginTable, 0)
	listOpt := &types.ListOption{
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	for {
		listResp, err := svc.dao.IaCPlugin().List(cts.Kit, listOpt)
		if err != nil {
			logs.Errorf("list iac plugin before delete failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("list iac plugin before delete failed, err: %v", err)
		}
		records = append(records, listResp.Details...)

		if uint(len(listResp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		listOpt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	if len(records) == 0 {
		return nil, nil
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, batch := range slice.Split(records, constant.BatchOperationMaxLimit) {
			delIDs := slice.Map(batch, func(one tableplugin.IaCPluginTable) string { return one.ID })
			delFilter := tools.ContainersExpression("id", delIDs)
			if err := svc.dao.IaCPlugin().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
				logs.Errorf("delete iac plugin failed, err: %v, delIDs: %v, rid: %s", err, delIDs, cts.Kit.Rid)
				return nil, err
			}

			audits := slice.Map(batch, func(one tableplugin.IaCPluginTable) *tableaudit.AuditTable {
				return buildIaCPluginAudit(cts.Kit, enumor.Delete, one, nil)
			})
			if err := svc.dao.Audit().BatchCreateWithTx(cts.Kit, txn, audits); err != nil {
				logs.Errorf("batch create delete iac plugin audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// UpdateIaCPlugin update iac plugin record. current_state makes it a conditional update.
func (svc *service) UpdateIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IaCPluginUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("update iac plugin decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("update iac plugin validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listResp, err := svc.dao.IaCPlugin().List(cts.Kit, &types.ListOption{
		Filter: tools.EqualExpression("id", req.ID),
		Page:   &core.BasePage{Limit: 1},
	})
	if err != nil {
		logs.Errorf("list iac plugin before update failed, err: %v, id: %s, rid: %s", err, req.ID, cts.Kit.Rid)
		return nil, err
	}
	if len(listResp.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "iac plugin %s not found", req.ID)
	}
	// 审计需要记录变更前的快照与插件名。
	old := listResp.Details[0]

	model := &tableplugin.IaCPluginTable{
		State:     req.State,
		Message:   req.Message,
		Extension: req.Extension,
		Memo:      req.Memo,
		Reviser:   cts.Kit.User,
	}

	flt := tools.EqualExpression("id", req.ID)
	if len(req.CurrentState) != 0 {
		flt = tools.ExpressionAnd(
			tools.RuleEqual("id", req.ID),
			tools.RuleEqual("state", req.CurrentState),
		)
	}

	changed, err := cvt.StructToMap(req)
	if err != nil {
		logs.Errorf("convert update iac plugin req to map failed, err: %v, id: %s, rid: %s", err, req.ID, cts.Kit.Rid)
		return nil, err
	}
	delete(changed, "id")
	delete(changed, "current_state")

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := svc.dao.IaCPlugin().UpdateWithTx(cts.Kit, txn, flt, model)
		if err != nil {
			logs.Errorf("update iac plugin failed, err: %v, id: %s, rid: %s", err, req.ID, cts.Kit.Rid)
			return nil, err
		}
		if effected == 0 {
			return nil, errf.Newf(errf.RecordNotUpdate, "iac plugin %s not updated", req.ID)
		}

		audits := []*tableaudit.AuditTable{buildIaCPluginAudit(cts.Kit, enumor.Update, old, changed)}
		if err := svc.dao.Audit().BatchCreateWithTx(cts.Kit, txn, audits); err != nil {
			logs.Errorf("batch create update iac plugin audit failed, err: %v, id: %s, rid: %s",
				err, req.ID, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListIaCPlugin list iac plugin records.
func (svc *service) ListIaCPlugin(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list iac plugin decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("list iac plugin validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.IaCPlugin().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list iac plugin failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &dataplugin.IaCPluginListResult{Count: result.Count}, nil
	}

	return &dataplugin.IaCPluginListResult{Details: slice.Map(result.Details, convTableToIaCPlugin)}, nil
}

func buildIaCPluginAudit(kt *kit.Kit, action enumor.AuditAction, data tableplugin.IaCPluginTable,
	changed map[string]interface{}) *tableaudit.AuditTable {

	return &tableaudit.AuditTable{
		ResID:    data.ID,
		ResName:  data.PluginName,
		ResType:  enumor.IaCPluginAuditResType,
		Action:   action,
		BkBizID:  constant.UnassignedBiz,
		Operator: kt.User,
		Source:   kt.GetRequestSource(),
		Rid:      kt.Rid,
		AppCode:  kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data:    data,
			Changed: changed,
		},
	}
}

func convTableToIaCPlugin(one tableplugin.IaCPluginTable) coreplugin.IaCPlugin {
	return coreplugin.IaCPlugin{
		ID:            one.ID,
		PluginName:    one.PluginName,
		PluginVersion: one.PluginVersion,
		Protocol:      one.Protocol,
		State:         one.State,
		Message:       cvt.PtrToVal(one.Message),
		Extension:     one.Extension,
		Memo:          one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}
