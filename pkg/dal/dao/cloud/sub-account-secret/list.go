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

// Package daosubaccountsecret is the dao layer for sub account secret.
package daosubaccountsecret

import (
	"encoding/json"
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ListJoinAccountAndSubAccount lists sub_account_secret rows joined with sub_account and account for biz scope.
func (dao *SubAccountSecretDao) ListJoinAccountAndSubAccount(kt *kit.Kit, opt *types.ListSecretJoinAccountOption) (
	*types.ListSubAccountSecretBizJoinDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list sub account secret biz options is nil")
	}
	if opt.BkBizID <= 0 {
		return nil, errf.New(errf.InvalidParameter, "bk_biz_id is invalid")
	}
	if opt.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is required")
	}
	if err := opt.Page.Validate(core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	sortCol := joinListSortColumn(opt.Page.Sort)

	whereSQL, whereArgs, err := buildJoinWhere(opt)
	if err != nil {
		return nil, err
	}

	joinSQL := fmt.Sprintf(
		`FROM %s AS secret
		INNER JOIN %s AS sub_account ON secret.sub_account_id = sub_account.id
		INNER JOIN %s AS account ON secret.account_id = account.id`,
		table.SubAccountSecretTable, table.SubAccountTable, table.AccountTable,
	)

	ormOpts := orm.NewInjectTenantIDOpt(kt.TenantID)

	if opt.Page.Count {
		sqlStr := fmt.Sprintf(`SELECT COUNT(*) %s %s`, joinSQL, whereSQL)
		count, err := dao.Orm.ModifySQLOpts(ormOpts).Do().Count(kt.Ctx, sqlStr, whereArgs)
		if err != nil {
			logs.Errorf("count sub account secret biz join failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		return &types.ListSubAccountSecretBizJoinDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, &types.PageSQLOption{
		Sort: types.SortOption{Sort: sortCol, ForceOverlap: true},
	})
	if err != nil {
		return nil, err
	}

	selectSQL := fmt.Sprintf(
		`SELECT secret.*, account.managers AS account_managers, account.name AS account_name,
		sub_account.managers AS sub_account_managers,
		sub_account.name AS sub_account_name,
		sub_account.extension AS sub_account_extension %s %s %s`,
		joinSQL, whereSQL, pageExpr,
	)

	details := make([]types.SubAccountSecretBizJoinRow, 0)
	err = dao.Orm.ModifySQLOpts(ormOpts).Do().Select(kt.Ctx, &details, selectSQL, whereArgs)
	if err != nil {
		logs.Errorf("select sub account secret biz join failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.ListSubAccountSecretBizJoinDetails{Count: 0, Details: details}, nil
}

// buildJoinWhere builds WHERE clause and named args for biz join list.
func buildJoinWhere(opt *types.ListSecretJoinAccountOption) (string, map[string]interface{}, error) {
	if opt == nil {
		return "", nil, fmt.Errorf("list secret join account option is nil")
	}

	whereExprs := make([]string, 0)
	args := make(map[string]interface{})

	// 云厂商条件过滤和二级账号资源账号条件过滤
	whereExprs = append(whereExprs, "secret.vendor = :vendor AND account.type = :account_type")

	args["vendor"] = string(opt.Vendor)
	args["account_type"] = string(enumor.ResourceAccount)

	// （最大查询范围）查询符合以下条件的三级账号的密钥：三级账号的业务是当前业务，或三级账号所属二级账号的管理业务是当前业务
	bizScope := "(JSON_CONTAINS(sub_account.bk_biz_ids, CAST(:bk_biz_id AS JSON)) OR " +
		"account.bk_biz_id = :bk_biz_id)"
	whereExprs = append(whereExprs, bizScope)
	args["bk_biz_id"] = opt.BkBizID

	if len(opt.IDs) > 0 {
		whereExprs = append(whereExprs, "secret.id IN (:ids)")
		args["ids"] = opt.IDs
	}
	if len(opt.Status) > 0 {
		whereExprs = append(whereExprs, "secret.status IN (:status)")
		args["status"] = opt.Status
	}
	if len(opt.AccountIDs) > 0 {
		whereExprs = append(whereExprs, "secret.account_id IN (:account_ids)")
		args["account_ids"] = opt.AccountIDs
	}
	if len(opt.SubAccountIDs) > 0 {
		whereExprs = append(whereExprs, "secret.sub_account_id IN (:sub_account_ids)")
		args["sub_account_ids"] = opt.SubAccountIDs
	}
	if len(opt.AccountManagers) > 0 {
		amJSON, err := json.Marshal(opt.AccountManagers)
		if err != nil {
			return "", nil, fmt.Errorf("marshal account_managers for JSON_OVERLAPS: %w", err)
		}
		whereExprs = append(whereExprs,
			"JSON_OVERLAPS(account.managers, CAST(:account_managers AS JSON))")
		args["account_managers"] = string(amJSON)
	}
	if len(opt.SubAccountManagers) > 0 {
		smJSON, err := json.Marshal(opt.SubAccountManagers)
		if err != nil {
			return "", nil, fmt.Errorf("marshal sub_account_managers for JSON_OVERLAPS: %w", err)
		}
		whereExprs = append(whereExprs,
			"JSON_OVERLAPS(sub_account.managers, CAST(:sub_account_managers AS JSON))")
		args["sub_account_managers"] = string(smJSON)
	}

	// 根据vendor构建扩展查询条件
	if !opt.Extension.IsEmpty() {
		var err error
		switch opt.Vendor {
		case enumor.TCloud:
			whereExprs, err = buildExtWhereForTCloud(whereExprs, args, opt.Extension)
		default:
			return "", nil, fmt.Errorf("unsupported vendor: %s", opt.Vendor)
		}

		if err != nil {
			return "", nil, err
		}
	}

	return "WHERE " + strings.Join(whereExprs, " AND "), args, nil
}

func buildExtWhereForTCloud(whereExprs []string, args map[string]interface{}, extension tabletypes.JsonField,
) ([]string, error) {

	tc := new(types.TCloudSubAccountSecretBizJoinExt)
	if err := json.Unmarshal([]byte(extension), tc); err != nil {
		return nil, fmt.Errorf("invalid tcloud extension json: %w", err)
	}
	if err := validator.Validate.Struct(tc); err != nil {
		return nil, err
	}
	if len(tc.CloudSecretIDs) > 0 {
		whereExprs = append(whereExprs,
			`JSON_UNQUOTE(JSON_EXTRACT(secret.extension, '$."cloud_secret_id"')) IN (:cloud_secret_ids)`)
		args["cloud_secret_ids"] = tc.CloudSecretIDs
	}
	if len(tc.CloudMainAccountIDs) > 0 {
		whereExprs = append(whereExprs,
			`JSON_UNQUOTE(JSON_EXTRACT(account.extension, '$."cloud_main_account_id"')) IN (:cloud_main_account_ids)`)
		args["cloud_main_account_ids"] = tc.CloudMainAccountIDs
	}
	if len(tc.CloudSubAccountIDs) > 0 {
		whereExprs = append(whereExprs, `(sub_account.cloud_id IN (:cloud_sub_account_ids_cloud_id) OR `+
			`CAST(JSON_EXTRACT(sub_account.extension, '$.uin') AS CHAR) IN (:cloud_sub_account_ids_uin))`)
		args["cloud_sub_account_ids_cloud_id"] = tc.CloudSubAccountIDs
		args["cloud_sub_account_ids_uin"] = tc.CloudSubAccountIDs
	}
	if tc.ConsoleLogin != nil {
		whereExprs = append(whereExprs,
			`CAST(JSON_EXTRACT(sub_account.extension, '$.console_login') AS SIGNED) = :console_login`)
		args["console_login"] = int64(*tc.ConsoleLogin)
	}

	return whereExprs, nil
}

// joinListSortColumn maps API BasePage.Sort field names to ORDER BY expressions using the secret alias.
// Only listed keys are accepted; anything else falls back to secret.id.
func joinListSortColumn(apiSort string) string {
	switch apiSort {
	case "", "id":
		return "secret.id"
	case "created_at":
		return "secret.created_at"
	case "updated_at":
		return "secret.updated_at"
	case "cloud_created_at":
		return "secret.cloud_created_at"
	default:
		return "secret.id"
	}
}
