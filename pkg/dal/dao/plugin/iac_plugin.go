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

// Package daoplugin defines DAO for iac_plugin.
package daoplugin

import (
	"encoding/json"
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// IaCPlugin only used for iac_plugin.
type IaCPlugin interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableplugin.IaCPluginTable) ([]string, error)
	// UpdateByID 按 ID 更新，并在同一事务内写 update 审计。expectState 不为空时作为更新条件，状态不符返回
	// errf.Aborted。记录没有任何字段实际发生变化时（例如重复提交同样的值）不写审计。
	UpdateByID(kt *kit.Kit, id string, expectState enumor.IaCPluginState, model *tableplugin.IaCPluginTable) error
	// BatchUpdateState 按插件名把处于 fromStates 的记录批量改为 toState，每条被改的记录各写一条 update 审计。
	BatchUpdateState(kt *kit.Kit, pluginName string, fromStates []enumor.IaCPluginState,
		toState enumor.IaCPluginState) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableplugin.IaCPluginTable], error)
}

var _ IaCPlugin = new(Dao)

// Dao iac plugin dao.
type Dao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx batch create iac_plugin with transaction.
func (d *Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableplugin.IaCPluginTable) ([]string, error) {
	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	for index := range models {
		if err := models[index].InsertValidate(); err != nil {
			return nil, err
		}
	}

	ids, err := d.IDGen.Batch(kt, table.IaCPluginTable, len(models))
	if err != nil {
		return nil, err
	}
	for index := range models {
		models[index].ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`, table.IaCPluginTable,
		tableplugin.IaCPluginColumns.ColumnExpr(), tableplugin.IaCPluginColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.IaCPluginTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.IaCPluginTable, err)
	}

	audits := slice.Map(models, func(one tableplugin.IaCPluginTable) *tableaudit.AuditTable {
		return d.buildAudit(kt, enumor.Create, one, nil)
	})
	if err = d.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// buildAudit 构造 iac_plugin 的审计记录。IaC 插件是平台级资源，不属于任何业务，bk_biz_id 固定为 -1。
// create 时 detail.data 为新记录；update 时 detail.data 为变更前的记录，detail.changed 为本次变更的字段。
func (d *Dao) buildAudit(kt *kit.Kit, action enumor.AuditAction, data tableplugin.IaCPluginTable,
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

// UpdateByID update iac_plugin by id in one transaction: lock and read the current record, update it, and write an
// update audit only when something actually changed. If expectState is not empty, the current state must equal it.
func (d *Dao) UpdateByID(kt *kit.Kit, id string, expectState enumor.IaCPluginState,
	model *tableplugin.IaCPluginTable) error {

	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	// message 允许被清空，memo 同理。
	opts := utils.NewFieldOptions().AddBlankedFields("message", "memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	_, err = d.Orm.AutoTxn(kt, func(txn *sqlx.Tx, _ *orm.TxnOption) (interface{}, error) {
		// FOR UPDATE 锁住该行，保证"读旧值 -> 比较状态 -> 更新 -> 写审计"在并发下一致。
		olds, err := d.listByIDsForUpdate(kt, txn, []string{id})
		if err != nil {
			return nil, err
		}
		if len(olds) == 0 {
			return nil, errf.Newf(errf.RecordNotFound, "iac plugin %s not found", id)
		}
		old := olds[0]

		if len(expectState) != 0 && old.State != expectState {
			return nil, errf.Newf(errf.Aborted, "iac plugin %s is in state %s, expect %s, it may have been modified",
				id, old.State, expectState)
		}

		changed := diffIaCPlugin(old, model)
		if len(changed) == 0 {
			// 没有字段实际变化：不更新，不写审计。
			return nil, nil
		}

		sql := fmt.Sprintf(`UPDATE %s %s WHERE id = :id`, table.IaCPluginTable, setExpr)
		toUpdate["id"] = id
		if _, err = d.Orm.Txn(txn).Update(kt.Ctx, sql, toUpdate); err != nil {
			logs.Errorf("update %s failed, err: %v, id: %s, rid: %s", table.IaCPluginTable, err, id, kt.Rid)
			return nil, err
		}

		audits := []*tableaudit.AuditTable{d.buildAudit(kt, enumor.Update, old, changed)}
		if err = d.Audit.BatchCreateWithTx(kt, txn, audits); err != nil {
			logs.Errorf("batch create audit failed, err: %v, id: %s, rid: %s", err, id, kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	return err
}

// diffIaCPlugin 返回 model 中相对 old 实际发生变化的字段，key 为列名。model 里为空值的字段视为"不更新"，
// message、memo 例外：指针非 nil 即表示要更新（包括清空）。
func diffIaCPlugin(old tableplugin.IaCPluginTable, model *tableplugin.IaCPluginTable) map[string]interface{} {
	changed := make(map[string]interface{})

	if len(model.State) != 0 && model.State != old.State {
		changed["state"] = model.State
	}
	if model.Message != nil && (old.Message == nil || cvt.PtrToVal(old.Message) != cvt.PtrToVal(model.Message)) {
		changed["message"] = cvt.PtrToVal(model.Message)
	}
	if len(model.Extension) != 0 && !jsonEqual(string(model.Extension), string(old.Extension)) {
		changed["extension"] = json.RawMessage(model.Extension)
	}
	if model.Memo != nil && (old.Memo == nil || cvt.PtrToVal(old.Memo) != cvt.PtrToVal(model.Memo)) {
		changed["memo"] = cvt.PtrToVal(model.Memo)
	}

	return changed
}

// jsonEqual 比较两段 JSON 的语义是否相同（忽略空白与 key 顺序，MySQL 的 JSON 列会重排 key）。
func jsonEqual(a, b string) bool {
	var va, vb interface{}
	if json.Unmarshal([]byte(a), &va) != nil || json.Unmarshal([]byte(b), &vb) != nil {
		return a == b
	}

	return reflect.DeepEqual(va, vb)
}

// listByIDsForUpdate 在事务内按 id 加行锁读取记录。
func (d *Dao) listByIDsForUpdate(kt *kit.Kit, tx *sqlx.Tx, ids []string) ([]tableplugin.IaCPluginTable, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE id IN (:ids) FOR UPDATE`,
		tableplugin.IaCPluginColumns.ColumnExpr(), table.IaCPluginTable)
	rows := make([]tableplugin.IaCPluginTable, 0, len(ids))
	if err := d.Orm.Txn(tx).Select(kt.Ctx, &rows, sql, map[string]interface{}{"ids": ids}); err != nil {
		logs.Errorf("select %s for update failed, err: %v, ids: %v, rid: %s", table.IaCPluginTable, err, ids, kt.Rid)
		return nil, err
	}

	return rows, nil
}

// BatchUpdateState update the state of iac_plugin by plugin name in one transaction, only records in fromStates are
// updated, and each updated record gets its own update audit.
func (d *Dao) BatchUpdateState(kt *kit.Kit, pluginName string, fromStates []enumor.IaCPluginState,
	toState enumor.IaCPluginState) error {

	if len(pluginName) == 0 || len(fromStates) == 0 {
		return errf.New(errf.InvalidParameter, "plugin_name and from_states are required")
	}

	_, err := d.Orm.AutoTxn(kt, func(txn *sqlx.Tx, _ *orm.TxnOption) (interface{}, error) {
		// 先加行锁读出将被修改的记录，同一事务内再更新，保证审计里的"变更前"与实际被改的行一致。
		sql := fmt.Sprintf(`SELECT %s FROM %s WHERE plugin_name = :plugin_name `+
			`AND state IN (:from_states) AND state != :to_state FOR UPDATE`,
			tableplugin.IaCPluginColumns.ColumnExpr(), table.IaCPluginTable)
		changes := make([]tableplugin.IaCPluginTable, 0)
		args := map[string]interface{}{"plugin_name": pluginName, "from_states": fromStates, "to_state": toState}
		if err := d.Orm.Txn(txn).Select(kt.Ctx, &changes, sql, args); err != nil {
			logs.Errorf("select %s for update failed, err: %v, plugin_name: %s, rid: %s",
				table.IaCPluginTable, err, pluginName, kt.Rid)
			return nil, err
		}
		if len(changes) == 0 {
			return nil, nil
		}

		ids := slice.Map(changes, func(one tableplugin.IaCPluginTable) string { return one.ID })
		updateSQL := fmt.Sprintf(`UPDATE %s SET state = :to_state, reviser = :reviser, updated_at = now() WHERE id IN (:ids)`,
			table.IaCPluginTable)
		updateArgs := map[string]interface{}{"to_state": toState, "reviser": kt.User, "ids": ids}
		if _, err := d.Orm.Txn(txn).Update(kt.Ctx, updateSQL, updateArgs); err != nil {
			logs.Errorf("batch update %s state failed, err: %v, plugin_name: %s, rid: %s",
				table.IaCPluginTable, err, pluginName, kt.Rid)
			return nil, err
		}

		audits := slice.Map(changes, func(one tableplugin.IaCPluginTable) *tableaudit.AuditTable {
			return d.buildAudit(kt, enumor.Update, one, map[string]interface{}{"state": toState})
		})
		if err := d.Audit.BatchCreateWithTx(kt, txn, audits); err != nil {
			logs.Errorf("batch create audit failed, err: %v, plugin_name: %s, rid: %s", err, pluginName, kt.Rid)
			return nil, err
		}

		return nil, nil
	})
	return err
}

// List list iac_plugin.
func (d *Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableplugin.IaCPluginTable], error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableplugin.IaCPluginColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.IaCPluginTable, whereExpr)
		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count %s failed, err: %v, filter: %s, rid: %s", table.IaCPluginTable, err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &types.ListResult[tableplugin.IaCPluginTable]{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableplugin.IaCPluginColumns.FieldsNamedExpr(opt.Fields),
		table.IaCPluginTable, whereExpr, pageExpr)

	details := make([]tableplugin.IaCPluginTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select %s failed, err: %v, sql: %s, filter: %v, rid: %s",
			table.IaCPluginTable, err, sql, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListResult[tableplugin.IaCPluginTable]{Count: 0, Details: details}, nil
}
