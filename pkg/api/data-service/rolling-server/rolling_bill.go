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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	rs "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/runtime/filter"
)

// BatchCreateRollingBillReq batch create request
type BatchCreateRollingBillReq struct {
	Bills []RollingBillCreateReq `json:"bills" validate:"required,max=100"`
}

// Validate ...
func (c *BatchCreateRollingBillReq) Validate() error {
	if len(c.Bills) == 0 || len(c.Bills) > 100 {
		return errf.Newf(errf.InvalidParameter, "bills count should between 1 and 100")
	}
	for _, item := range c.Bills {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return validator.Validate.Struct(c)
}

// RollingBillCreateReq create request
type RollingBillCreateReq struct {
	// BkBizID 业务ID
	BkBizID int64 `json:"bk_biz_id" validate:"required"`
	// DeliveredCore 已交付核心数
	DeliveredCore uint64 `json:"delivered_core"`
	// ReturnedCore 已退还核心数
	ReturnedCore uint64 `json:"returned_core"`
	// ExemptedReturnedCore 减免退还核心数
	ExemptedReturnedCore uint64 `json:"exempted_returned_core"`
	// NotReturnedCore 未退还核心数
	NotReturnedCore uint64 `json:"not_returned_core"`
	// Year 记录账单的年份
	Year int `json:"year" validate:"required"`
	// Month 记录账单的月份
	Month int `json:"month" validate:"required"`
	// Day 记录账单的天
	Day int `json:"day" validate:"required"`
	// Creator 创建者
	Creator string `json:"creator"`

	// DataDate 日期
	DataDate string `json:"data_date" validate:"required"`
	// ProductID 运营产品id
	ProductID int64 `json:"product_id" validate:"required"`
	// BusinessSetID 一级业务id
	BusinessSetID int64 `json:"business_set_id" validate:"required"`
	// BusinessSetName 一级业务名称
	BusinessSetName string `json:"business_set_name" validate:"required"`
	// BusinessID 二级业务id
	BusinessID int64 `json:"business_id" validate:"required"`
	// BusinessName 二级业务名称
	BusinessName string `json:"business_name" validate:"required"`
	// BusinessModID 三级业务id
	BusinessModID int64 `json:"business_mod_id"`
	// BusinessModName 三级业务名称
	BusinessModName string `json:"business_mod_name"`
	// Uin uin
	Uin string `json:"uin"`
	// AppID app id
	AppID string `json:"app_id"`
	// User 使用人
	User string `json:"user"`
	// CityID 城市ID
	CityID int64 `json:"city_id"`
	// CampusID 园区ID
	CampusID int64 `json:"campus_id"`
	// IdcUnitID 管理单元ID
	IdcUnitID int64 `json:"idc_unit_id"`
	// IdcUnitName 管理单元名称
	IdcUnitName string `json:"idc_unit_name"`
	// ModuleID module id
	ModuleID int64 `json:"module_id"`
	// ModuleName module名称
	ModuleName string `json:"module_name"`
	// ZoneID 可用区ID
	ZoneID int64 `json:"zone_id"`
	// ZoneName 可用区名称
	ZoneName string `json:"zone_name"`
	// PlatformID 平台ID
	PlatformID int64 `json:"platform_id" validate:"required"`
	// ResClassID 资源规格ID
	ResClassID int64 `json:"res_class_id" validate:"required"`
	// ClusterID 集群ID
	ClusterID string `json:"cluster_id"`
	// PlatformResID 最小粒度资源ID
	PlatformResID string `json:"platform_res_id"`
	// BandwidthTypeID 带宽类型ID
	BandwidthTypeID int64 `json:"bandwidth_type_id"`
	// OperatorNameID 运营商ID
	OperatorNameID int64 `json:"operator_name_id"`
	// Amount 核算用量
	Amount float64 `json:"amount"`
	// AmountInCurrentDate 参考日用量
	AmountInCurrentDate float64 `json:"amount_in_current_date"`
	// Cost 成本
	Cost float64 `json:"cost"`
	// ExtendDetail 扩展详情
	ExtendDetail string `json:"extend_detail"`
}

// Validate ...
func (c *RollingBillCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// RollingBillListReq list request
type RollingBillListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	Fields []string           `json:"fields" validate:"omitempty"`
}

// Validate ...
func (req *RollingBillListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RollingBillListResult list result
type RollingBillListResult = core.ListResultT[*rs.OBSBillItemRolling]
