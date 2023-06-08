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

package networkcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/cloud/cvm"
	networkinterfacedao "hcm/pkg/dal/dao/cloud/network-interface"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	nitable "hcm/pkg/dal/table/cloud/network-interface"
	nicvmreltable "hcm/pkg/dal/table/cloud/network-interface-cvm-rel"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// NiCvmRel only used for network interface and cvm relation.
type NiCvmRel interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []nicvmreltable.NetworkInterfaceCvmRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListNetworkInterfaceCvmRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	ListJoinNetworkInterface(kt *kit.Kit, cvmIDs []string, vendor enumor.Vendor) (
		*types.ListCvmRelsJoinNetworkInterfaceDetails, error)
}

// NiCvmRel ...
var _ NiCvmRel = new(NiCvmRelDao)

// NiCvmRelDao ni_cvm_rel dao.
type NiCvmRelDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// ListJoinNetworkInterface list cvm rel with network interface detail.
func (dao NiCvmRelDao) ListJoinNetworkInterface(kt *kit.Kit, cvmIDs []string, vendor enumor.Vendor) (
	*types.ListCvmRelsJoinNetworkInterfaceDetails, error) {

	if len(cvmIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "cvm ids is required")
	}

	sql := fmt.Sprintf(
		`SELECT %s, %s FROM %s AS rel LEFT JOIN %s AS ni ON rel.network_interface_id = ni.id WHERE 
		rel.cvm_id IN (:cvm_ids) AND ni.vendor = :vendor`,
		nitable.NetworkInterfaceColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		tools.BaseRelJoinSqlBuild(
			"rel",
			"ni",
			"id",
			"cvm_id",
		),
		table.NetworkInterfaceCvmRelTable,
		table.NetworkInterfaceTable,
	)

	details := make([]*types.NetworkInterfaceWithCvmID, 0)
	if err := dao.Orm.Do().Select(kt.Ctx, &details, sql,
		map[string]interface{}{"cvm_ids": cvmIDs, "vendor": vendor}); err != nil {
		logs.ErrorJson("select network interface cvm rels join network failed, err: %v, sql: (%s), rid: %s",
			err, sql, kt.Rid)
		return nil, err
	}

	return &types.ListCvmRelsJoinNetworkInterfaceDetails{Details: details}, nil
}

// BatchCreateWithTx batch create network interface cvm rel with transaction.
func (dao NiCvmRelDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	rels []nicvmreltable.NetworkInterfaceCvmRelTable) error {

	// 校验关联资源是否存在
	networkInterfaceIDs := make([]string, 0)
	cvmIDs := make([]string, 0)
	cvmIDMap := make(map[string]bool, 0)
	for _, rel := range rels {
		if _, ok := cvmIDMap[rel.CvmID]; ok {
			continue
		}

		networkInterfaceIDs = append(networkInterfaceIDs, rel.NetworkInterfaceID)
		cvmIDs = append(cvmIDs, rel.CvmID)
		cvmIDMap[rel.CvmID] = true
	}

	netMap, err := networkinterfacedao.ListNetworkInterface(kt, dao.Orm, networkInterfaceIDs)
	if err != nil {
		logs.Errorf("list create network interface failed, err: %v, ids: %v, rid: %s",
			err, networkInterfaceIDs, kt.Rid)
		return err
	}

	if len(netMap) != len(networkInterfaceIDs) {
		logs.Errorf("get network interface count not right, ids: %v, netCount: %d, err: %v, rid: %s",
			networkInterfaceIDs, len(netMap), err, kt.Rid)
		return fmt.Errorf("get network interface count not right")
	}

	cvmMap, err := cvm.ListCvm(kt, dao.Orm, cvmIDs)
	if err != nil {
		logs.Errorf("list network interface cvm failed, err: %v, ids: %v, rid: %s",
			err, networkInterfaceIDs, kt.Rid)
		return err
	}

	if len(cvmMap) != len(cvmIDs) {
		logs.Errorf("get cvm count not right, ids: %v, count: %d, err: %v, rid: %s",
			cvmIDs, len(cvmMap), err, kt.Rid)
		return fmt.Errorf("get cvm count not right")
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.NetworkInterfaceCvmRelTable,
		nicvmreltable.NetworkInterfaceCvmRelColumns.ColumnExpr(),
		nicvmreltable.NetworkInterfaceCvmRelColumns.ColonNameExpr())

	if err := dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.NetworkInterfaceCvmRelTable, err, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", table.NetworkInterfaceCvmRelTable, err)
	}

	return nil
}

// List network interface cvm rel.
func (dao NiCvmRelDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListNetworkInterfaceCvmRelDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(
		nicvmreltable.NetworkInterfaceCvmRelColumns.ColumnTypes())), core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.NetworkInterfaceCvmRelTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count network interface cvm rels failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListNetworkInterfaceCvmRelDetails{Count: &count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		nicvmreltable.NetworkInterfaceCvmRelColumns.FieldsNamedExpr(opt.Fields),
		table.NetworkInterfaceCvmRelTable, whereExpr, pageExpr)

	details := make([]*nicvmreltable.NetworkInterfaceCvmRelTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select network cvm rels failed, filter: %s, err: %v, rid: %s", opt.Filter, err, kt.Rid)
		return nil, err
	}

	return &types.ListNetworkInterfaceCvmRelDetails{Details: details}, nil
}

// DeleteWithTx delete network interface cvm rel with transaction.
func (dao NiCvmRelDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.NetworkInterfaceCvmRelTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete network cvm rels failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
