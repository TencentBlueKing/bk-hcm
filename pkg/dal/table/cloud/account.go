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

// Package cloud 描述云资源表
package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

/*
account.go:
- 账号表结构定义：AccountTable
- 账号字段描述：AccountColumnDescriptor
- 账号表插入校验
- 账号表更新校验
- `hcm/pkg/api/core/account` 转账号表结构函数
- 账号表结构转 `hcm/pkg/api/core/account` 函数
*/

// AccountColumns defines all the account table's columns.
var AccountColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, AccountColumnDescriptor)

// AccountColumnDescriptor is Account's column descriptors.
var AccountColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "managers", NamedC: "managers", Type: enumor.Json},
	{Column: "department_id", NamedC: "department_id", Type: enumor.Numeric},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "site", NamedC: "site", Type: enumor.String},
	{Column: "sync_status", NamedC: "sync_status", Type: enumor.String},
	{Column: "price", NamedC: "price", Type: enumor.String},
	{Column: "price_unit", NamedC: "price_unit", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
}

// AccountTable 云账号表
type AccountTable struct {
	// ID 账号自增 ID
	ID uint64 `db:"id"`
	// Name 账号名称
	Name string `db:"name"`
	// Vendor 云厂商
	Vendor string `db:"vendor"`
	// Managers 责任人
	Managers types.StringArray `db:"managers"`
	// DepartmentID 部门 ID
	DepartmentID int64 `db:"department_id"`
	// Type 账号类型(资源账号|登记账号)
	Type string `db:"type"`
	// Site 站点(中国站｜国际站)
	Site string `db:"site"`
	// SyncStatus 账号资源同步状态
	SyncStatus string `db:"sync_status"`
	// Price 账号余额数值
	Price string `db:"price"`
	// PriceUnit 账号余额单位
	PriceUnit string `db:"price_unit"`
	// Extension 云厂商账号差异扩展字段
	Extension types.JsonField `db:"extension"`
	// Creator 创建者
	Creator string `db:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser"`
	// CreatedAt 创建时间
	CreatedAt *time.Time `db:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt *time.Time `db:"updated_at"`
	// Memo 账号信息备注
	Memo *string `db:"memo"`
}

// TableName return account table name.
func (a AccountTable) TableName() table.Name {
	return table.AccountTable
}

// InsertValidate account table when insert.
func (a AccountTable) InsertValidate() error {
	if a.ID != 0 {
		return errors.New("id can not set")
	}

	if a.CreatedAt != nil {
		return errors.New("created_at can not set")
	}

	if a.UpdatedAt != nil {
		return errors.New("update_at can not set")
	}

	// TODO: 添加账号其他信息正则和长度校验。

	return nil
}

// UpdateValidate account table when update.
func (a AccountTable) UpdateValidate() error {
	if a.UpdatedAt != nil {
		return errors.New("update_at can not update")
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	// TODO: 添加账号无法更新字段的校验和可以更新字段的正则及长度校验。

	return nil
}

// ConvAccountTable conv account table from account model.
func ConvAccountTable(model *cloud.Account) (*AccountTable, error) {
	if model == nil {
		return nil, nil
	}

	account := &AccountTable{
		ID:     model.ID,
		Vendor: string(model.Vendor),
	}

	if model.Spec != nil {
		account.Name = model.Spec.Name
		account.DepartmentID = model.Spec.DepartmentID
		account.Type = string(model.Spec.Type)
		account.SyncStatus = string(model.Spec.SyncStatus)
		account.Price = model.Spec.Price
		account.PriceUnit = model.Spec.PriceUnit
		account.Memo = model.Spec.Memo
		account.Site = string(model.Spec.Site)
		account.Managers = model.Spec.Managers
	}

	if model.Extension != nil {
		var extension interface{}
		switch model.Vendor {
		case enumor.TCloud:
			extension = model.Extension.TCloud
		case enumor.AWS:
			extension = model.Extension.Aws
		case enumor.HuaWei:
			extension = model.Extension.HuaWei
		case enumor.GCP:
			extension = model.Extension.Gcp
		case enumor.Azure:
			extension = model.Extension.Azure
		default:
			return nil, fmt.Errorf("unknown vendor type: %s", account.Vendor)
		}

		ext, err := json.Marshal(extension)
		if err != nil {
			return nil, err
		}

		account.Extension = types.JsonField(ext)
	}

	if model.Revision != nil {
		account.Creator = model.Revision.Creator
		account.Reviser = model.Revision.Reviser
		account.CreatedAt = model.Revision.CreatedAt
		account.UpdatedAt = model.Revision.UpdatedAt
	}

	return account, nil
}

// ConvAccount conv account model from account core.
func ConvAccount(table *AccountTable) (*cloud.Account, error) {
	if table == nil {
		return nil, nil
	}

	account := &cloud.Account{
		ID:     table.ID,
		Vendor: enumor.Vendor(table.Vendor),
		Spec: &cloud.AccountSpec{
			Name:         table.Name,
			Managers:     table.Managers,
			DepartmentID: table.DepartmentID,
			Type:         enumor.AccountType(table.Type),
			Site:         enumor.AccountSiteType(table.Site),
			SyncStatus:   enumor.AccountSyncStatus(table.SyncStatus),
			Price:        table.Price,
			PriceUnit:    table.PriceUnit,
			Memo:         table.Memo,
		},
		Extension: &cloud.AccountExtension{},
		Revision: &core.Revision{
			Creator:   table.Creator,
			Reviser:   table.Reviser,
			CreatedAt: table.CreatedAt,
			UpdatedAt: table.UpdatedAt,
		},
	}

	// TODO: 考虑使用反射统一处理
	var err error
	switch account.Vendor {
	case enumor.TCloud:
		account.Extension.TCloud = new(cloud.TCloudAccountExtension)
		err = json.Unmarshal([]byte(table.Extension), account.Extension.TCloud)

	case enumor.AWS:
		account.Extension.Aws = new(cloud.AwsAccountExtension)
		err = json.Unmarshal([]byte(table.Extension), account.Extension.Aws)

	case enumor.HuaWei:
		account.Extension.HuaWei = new(cloud.HuaWeiAccountExtension)
		err = json.Unmarshal([]byte(table.Extension), account.Extension.HuaWei)

	case enumor.GCP:
		account.Extension.Gcp = new(cloud.GcpAccountExtension)
		err = json.Unmarshal([]byte(table.Extension), account.Extension.Gcp)

	case enumor.Azure:
		account.Extension.Azure = new(cloud.AzureAccountExtension)
		err = json.Unmarshal([]byte(table.Extension), account.Extension.Azure)

	default:
		return nil, fmt.Errorf("unknown vendor type: %s", account.Vendor)
	}
	if err != nil {
		return nil, fmt.Errorf("unmarshal extension failed, err: %v, vendor: %s", err, account.Vendor)
	}

	return account, nil
}

// ConvAccountList conv accounts model from accounts core.
func ConvAccountList(list []*AccountTable) ([]*cloud.Account, error) {
	accounts := make([]*cloud.Account, len(list))

	for index, one := range list {
		account, err := ConvAccount(one)
		if err != nil {
			return nil, err
		}

		accounts[index] = account
	}

	return accounts, nil
}
