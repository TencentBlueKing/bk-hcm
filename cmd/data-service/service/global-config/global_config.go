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

// Package globalconfig global config service
package globalconfig

import (
	"fmt"

	"hcm/pkg/api/core"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablegconf "hcm/pkg/dal/table/global-config"
	dtypes "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/util"

	"github.com/jmoiron/sqlx"
)

// BatchCreateGlobalConfigs creates the global config.
func (svc *service) BatchCreateGlobalConfigs(cts *rest.Contexts) (interface{}, error) {
	req := new(datagconf.BatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create global config decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("create global config validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create
	ids, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		globalConfigs := make([]tablegconf.GlobalConfigTable, len(req.Configs))
		for index, config := range req.Configs {
			globalConfigs[index] = tablegconf.GlobalConfigTable{
				ConfigKey:   config.ConfigKey,
				ConfigValue: dtypes.JsonField(util.GetStrByInterface(config.ConfigValue)),
				ConfigType:  config.ConfigType,
				Memo:        config.Memo,
				Creator:     cts.Kit.User,
				Reviser:     cts.Kit.User,
			}
		}

		ids, err := svc.dao.GlobalConfig().CreateWithTx(cts.Kit, txn, globalConfigs)
		if err != nil {
			logs.Errorf("create global config failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("create global config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return ids, nil
}

// ListGlobalConfigs ...
func (svc *service) ListGlobalConfigs(cts *rest.Contexts) (interface{}, error) {
	req := new(datagconf.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list global config decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("list global config validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listOpt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.GlobalConfig().List(cts.Kit, listOpt)
	if err != nil {
		logs.Errorf("list global config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	resp := &datagconf.ListResp{
		Count:   result.Count,
		Details: result.Details,
	}

	return resp, nil
}

// BatchUpdateGlobalConfigs ...
func (svc *service) BatchUpdateGlobalConfigs(cts *rest.Contexts) (interface{}, error) {
	req := new(datagconf.BatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch update global config decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch update global config validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, config := range req.Configs {
			record := &tablegconf.GlobalConfigTable{
				Reviser: cts.Kit.User,
			}

			if config.ConfigValue != nil {
				record.ConfigValue = dtypes.JsonField(util.GetStrByInterface(config.ConfigValue))
			}

			if config.Memo != nil {
				record.Memo = config.Memo
			}

			if err := svc.dao.GlobalConfig().UpdateWithTx(cts.Kit, txn,
				tools.EqualExpression("id", config.ID), record); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update global config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// BatchDeleteGlobalConfigs ...
func (svc *service) BatchDeleteGlobalConfigs(cts *rest.Contexts) (interface{}, error) {
	req := new(datagconf.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete global config decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete global config validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.GlobalConfig().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("delete list global config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("delete list global config failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		delFilter := tools.ContainersExpression("id", delIDs)
		if err = svc.dao.GlobalConfig().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete global config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
