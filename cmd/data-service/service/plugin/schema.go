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
	"reflect"

	"hcm/pkg/api/core"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

func (svc *service) initIacPluginSchemaRoutes(h *rest.Handler) {
	h.Add("BatchCreateIacPluginSchema", http.MethodPost,
		"/iac_plugin_schema/batch/create", svc.BatchCreateIacPluginSchema)
	h.Add("BatchUpdateIacPluginSchema", http.MethodPatch,
		"/iac_plugin_schema/batch/update", svc.BatchUpdateIacPluginSchema)
	h.Add("BatchDeleteIacPluginSchema", http.MethodDelete,
		"/iac_plugin_schema/batch", svc.BatchDeleteIacPluginSchema)
	h.Add("ListIacPluginSchema", http.MethodPost, "/iac_plugin_schema/list", svc.ListIacPluginSchema)
}

// BatchCreateIacPluginSchema 批量创建 iac_plugin_schema 记录。
func (svc *service) BatchCreateIacPluginSchema(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IacPluginSchemaBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch create iac plugin schema decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch create iac plugin schema validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createdIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tableplugin.IacPluginSchemaTable, 0, len(req.Items))
		for _, item := range req.Items {
			models = append(models, tableplugin.IacPluginSchemaTable{
				PluginName:    item.PluginName,
				PluginVersion: item.PluginVersion,
				SchemaHash:    item.SchemaHash,
				SchemaDetail:  item.SchemaDetail,
				Creator:       cts.Kit.User,
				Reviser:       cts.Kit.User,
			})
		}

		ids, err := svc.dao.IacPluginSchema().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("batch create iac plugin schema in db failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create iac plugin schema transaction failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	ids, ok := createdIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create iac plugin schema return ids type is not []string: %v",
			reflect.TypeOf(createdIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateIacPluginSchema 批量更新 iac_plugin_schema 记录。
func (svc *service) BatchUpdateIacPluginSchema(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IacPluginSchemaBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch update iac plugin schema decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch update iac plugin schema validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, item := range req.Items {
			model := &tableplugin.IacPluginSchemaTable{
				SchemaHash:   item.SchemaHash,
				SchemaDetail: item.SchemaDetail,
				Reviser:      cts.Kit.User,
			}

			if err := svc.dao.IacPluginSchema().UpdateWithTx(cts.Kit, txn,
				tools.EqualExpression("id", item.ID), model); err != nil {
				logs.Errorf("update iac plugin schema in db failed, id: %s, err: %v, rid: %s",
					item.ID, err, cts.Kit.Rid)
				return nil, err
			}
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update iac plugin schema transaction failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// BatchDeleteIacPluginSchema 批量删除 iac_plugin_schema 记录。
func (svc *service) BatchDeleteIacPluginSchema(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IacPluginSchemaBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete iac plugin schema decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete iac plugin schema validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.IacPluginSchema().DeleteWithTx(cts.Kit, txn, req.Filter); err != nil {
			logs.Errorf("delete iac plugin schema in db failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch delete iac plugin schema transaction failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// ListIacPluginSchema 列表查询 iac_plugin_schema，支持 fields 投影。
func (svc *service) ListIacPluginSchema(cts *rest.Contexts) (interface{}, error) {
	req := new(dataplugin.IacPluginSchemaListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list iac plugin schema decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("list iac plugin schema validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	res, err := svc.dao.IacPluginSchema().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list iac plugin schema failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if req.Page.Count {
		return &dataplugin.IacPluginSchemaListResult{Count: res.Count}, nil
	}

	return &dataplugin.IacPluginSchemaListResult{Details: res.Details}, nil
}
