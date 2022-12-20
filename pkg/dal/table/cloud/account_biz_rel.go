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

package cloud

import (
	"errors"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/utils"
)

// AccountBizRelColumns defines all the account and biz relation table's columns.
var AccountBizRelColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, AccountBizRelColumnDescriptor)

// AccountBizRelColumnDescriptor is account and biz relation table column descriptors.
var AccountBizRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "tenant_id", NamedC: "tenant_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// AccountBizRelTable 云账户与业务关联表
type AccountBizRelTable struct {
	// BkBizID 蓝鲸业务 ID
	// TODO 这种外部的依赖ID是否需要预留为string？
	BkBizID int64 `db:"bk_biz_id"`
	// AccountID 云账号主键 ID
	AccountID string `db:"account_id"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id"`
	// Creator 创建者
	Creator string `db:"creator"`
	// CreatedAt 创建时间
	CreatedAt *time.Time `db:"created_at"`
}

// TableName return account table name.
func (a AccountBizRelTable) TableName() table.Name {
	return table.AccountBizRelTable
}

// InsertValidate account table when insert.
func (a AccountBizRelTable) InsertValidate() error {
	if a.CreatedAt != nil {
		return errors.New("created_at can not set")
	}

	if a.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	if len(a.AccountID) == 0 {
		return errors.New("account id is required")
	}

	if len(a.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}
