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

// Package rollingserver ...
package rollingserver

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// OBSBillItemRollingColumns defines account_bill_summary's columns.
var OBSBillItemRollingColumns = utils.MergeColumns(nil, OBSBillItemRollingColumnDescriptor)

// OBSBillItemRollingColumnDescriptor is AwsBill's column descriptors.
var OBSBillItemRollingColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "delivered_core", NamedC: "delivered_core", Type: enumor.Numeric},
	{Column: "returned_core", NamedC: "returned_core", Type: enumor.Numeric},
	{Column: "exempted_returned_core", NamedC: "exempted_returned_core", Type: enumor.Numeric},
	{Column: "not_returned_core", NamedC: "not_returned_core", Type: enumor.Numeric},
	{Column: "year", NamedC: "year", Type: enumor.Numeric},
	{Column: "month", NamedC: "month", Type: enumor.Numeric},
	{Column: "day", NamedC: "day", Type: enumor.Numeric},
	{Column: "roll_date", NamedC: "roll_date", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},

	{Column: "data_date", NamedC: "data_date", Type: enumor.String},
	{Column: "product_id", NamedC: "product_id", Type: enumor.Numeric},
	{Column: "business_set_id", NamedC: "business_set_id", Type: enumor.Numeric},
	{Column: "business_set_name", NamedC: "business_set_name", Type: enumor.String},
	{Column: "business_id", NamedC: "business_id", Type: enumor.String},
	{Column: "business_name", NamedC: "business_name", Type: enumor.String},
	{Column: "business_mod_id", NamedC: "business_mod_id", Type: enumor.Numeric},
	{Column: "business_mod_name", NamedC: "business_mod_name", Type: enumor.String},
	{Column: "uin", NamedC: "uin", Type: enumor.String},
	{Column: "app_id", NamedC: "app_id", Type: enumor.String},
	{Column: "user", NamedC: "user", Type: enumor.String},
	{Column: "city_id", NamedC: "city_id", Type: enumor.Numeric},
	{Column: "campus_id", NamedC: "campus_id", Type: enumor.Numeric},
	{Column: "idc_unit_id", NamedC: "idc_unit_id", Type: enumor.Numeric},
	{Column: "idc_unit_name", NamedC: "idc_unit_name", Type: enumor.String},
	{Column: "module_id", NamedC: "module_id", Type: enumor.Numeric},
	{Column: "module_name", NamedC: "module_name", Type: enumor.String},
	{Column: "zone_id", NamedC: "zone_id", Type: enumor.Numeric},
	{Column: "zone_name", NamedC: "zone_name", Type: enumor.String},
	{Column: "platform_id", NamedC: "platform_id", Type: enumor.Numeric},
	{Column: "res_class_id", NamedC: "res_class_id", Type: enumor.Numeric},
	{Column: "cluster_id", NamedC: "cluster_id", Type: enumor.String},
	{Column: "platform_res_id", NamedC: "platform_res_id", Type: enumor.String},
	{Column: "bandwidth_type_id", NamedC: "bandwidth_type_id", Type: enumor.Numeric},
	{Column: "operator_name_id", NamedC: "operator_name_id", Type: enumor.Numeric},
	{Column: "amount", NamedC: "amount", Type: enumor.Numeric},
	{Column: "amount_in_current_date", NamedC: "amount_in_current_date", Type: enumor.Numeric},
	{Column: "cost", NamedC: "cost", Type: enumor.Numeric},
	{Column: "extend_detail", NamedC: "extend_detail", Type: enumor.String},
}

// OBSBillItemRolling Rolling bill item
type OBSBillItemRolling struct {
	// ID 唯一ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// DeliveredCore 已交付核心数
	DeliveredCore uint64 `db:"delivered_core" json:"delivered_core"`
	// ReturnedCore 已退还核心数
	ReturnedCore uint64 `db:"returned_core" json:"returned_core"`
	// ExemptedReturnedCore 减免退还核心数
	ExemptedReturnedCore uint64 `db:"exempted_returned_core" json:"exempted_returned_core"`
	// NotReturnedCore 未退还核心数
	NotReturnedCore uint64 `db:"not_returned_core" json:"not_returned_core"`
	// Year 记录账单的年份
	Year int `db:"year" json:"year"`
	// Month 记录账单的月份
	Month int `db:"month" json:"month"`
	// Day 记录账单的天
	Day int `db:"day" json:"day"`
	// RollDate 记录账单的年月日
	RollDate int `db:"roll_date" json:"roll_date"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`

	// DataDate 日期
	DataDate string `db:"data_date" json:"data_date"`
	// ProductID 运营产品id
	ProductID int64 `db:"product_id" json:"product_id"`
	// BusinessSetID 一级业务id
	BusinessSetID int64 `db:"business_set_id" json:"business_set_id"`
	// BusinessSetName 一级业务名称
	BusinessSetName string `db:"business_set_name" json:"business_set_name"`
	// BusinessID 二级业务id
	BusinessID int64 `db:"business_id" json:"business_id"`
	// BusinessName 二级业务名称
	BusinessName string `db:"business_name" json:"business_name"`
	// BusinessModID 三级业务id
	BusinessModID int64 `db:"business_mod_id" json:"business_mod_id"`
	// BusinessModName 三级业务名称
	BusinessModName string `db:"business_mod_name" json:"business_mod_name"`
	// Uin uin
	Uin string `db:"uin" json:"uin"`
	// AppID app id
	AppID string `db:"app_id" json:"app_id"`
	// User 使用人
	User string `db:"user" json:"user"`
	// CityID 城市ID
	CityID int64 `db:"city_id" json:"city_id"`
	// CampusID 园区ID
	CampusID int64 `db:"campus_id" json:"campus_id"`
	// IdcUnitID 管理单元ID
	IdcUnitID int64 `db:"idc_unit_id" json:"idc_unit_id"`
	// IdcUnitName 管理单元名称
	IdcUnitName string `db:"idc_unit_name" json:"idc_unit_name"`
	// ModuleID module id
	ModuleID int64 `db:"module_id" json:"module_id"`
	// ModuleName module名称
	ModuleName string `db:"module_name" json:"module_name"`
	// ZoneID 可用区ID
	ZoneID int64 `db:"zone_id" json:"zone_id"`
	// ZoneName 可用区名称
	ZoneName string `db:"zone_name" json:"zone_name"`
	// PlatformID 平台ID
	PlatformID int64 `db:"platform_id" json:"platform_id"`
	// ResClassID 资源规格ID
	ResClassID int64 `db:"res_class_id" json:"res_class_id"`
	// ClusterID 集群ID
	ClusterID string `db:"cluster_id" json:"cluster_id"`
	// PlatformResID 最小粒度资源ID
	PlatformResID string `db:"platform_res_id" json:"platform_res_id"`
	// BandwidthTypeID 带宽类型ID
	BandwidthTypeID int64 `db:"bandwidth_type_id" json:"bandwidth_type_id"`
	// OperatorNameID 运营商ID
	OperatorNameID int64 `db:"operator_name_id" json:"operator_name_id"`
	// Amount 核算用量
	Amount float64 `db:"amount" json:"amount"`
	// AmountInCurrentDate 参考日用量
	AmountInCurrentDate float64 `db:"amount_in_current_date" json:"amount_in_current_date"`
	// Cost 成本
	Cost float64 `db:"cost" json:"cost"`
	// ExtendDetail 扩展详情
	ExtendDetail string `db:"extend_detail" json:"extend_detail"`
}

// TableName 返回滚服账单表名
func (b *OBSBillItemRolling) TableName() table.Name {
	return table.OBSBillRollingItemTable
}

// InsertValidate validate rolling bill on insert
func (b *OBSBillItemRolling) InsertValidate() error {
	if len(b.ID) == 0 {
		return errors.New("id is required")
	}

	if b.BkBizID <= 0 {
		return errors.New("bk_biz_id should be > 0")
	}

	if len(b.DataDate) == 0 {
		return errors.New("data_date is required")
	}

	if b.ProductID == 0 {
		return errors.New("product_id is required")
	}

	if b.PlatformID == 0 {
		return errors.New("platform_id is required")
	}

	if b.ResClassID == 0 {
		return errors.New("res_class_id is required")
	}

	if len(b.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if err := validator.Validate.Struct(b); err != nil {
		return err
	}

	return nil
}
