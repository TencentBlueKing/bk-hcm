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
	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	tableplugin "hcm/pkg/dal/table/plugin"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

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
			Extension:     tabletypes.JsonField(req.Extension),
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

// UpdateIaCPlugin update iac plugin record, expect_state makes it a conditional update.
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

	model := &tableplugin.IaCPluginTable{
		State:     req.State,
		Message:   req.Message,
		Extension: tabletypes.JsonField(req.Extension),
		Memo:      req.Memo,
		Reviser:   cts.Kit.User,
	}
	if err := svc.dao.IaCPlugin().UpdateByID(cts.Kit, req.ID, req.ExpectState, model); err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchUpdateIaCPluginState batch update the state of the iac plugin by plugin name.
func (svc *service) BatchUpdateIaCPluginState(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IaCPluginBatchUpdateStateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch update iac plugin state decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch update iac plugin state validate failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.dao.IaCPlugin().BatchUpdateState(cts.Kit, req.PluginName, req.FromStates, req.ToState); err != nil {
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

func convTableToIaCPlugin(one tableplugin.IaCPluginTable) coreplugin.IaCPlugin {
	return coreplugin.IaCPlugin{
		ID:            one.ID,
		PluginName:    one.PluginName,
		PluginVersion: one.PluginVersion,
		Protocol:      one.Protocol,
		State:         one.State,
		Message:       cvt.PtrToVal(one.Message),
		Extension:     []byte(one.Extension),
		Memo:          one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}
