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

package tableselection

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// SchemeTableColumns defines all the scheme table's columns.
var SchemeTableColumns = utils.MergeColumns(nil, SchemeTableColumnDescriptor)

// SchemeTableColumnDescriptor is scheme table column descriptors.
var SchemeTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "biz_type", NamedC: "biz_type", Type: enumor.String},
	{Column: "vendors", NamedC: "vendors", Type: enumor.Json},
	{Column: "deployment_architecture", NamedC: "deployment_architecture", Type: enumor.Json},
	{Column: "cover_ping", NamedC: "cover_ping", Type: enumor.Numeric},
	{Column: "composite_score", NamedC: "composite_score", Type: enumor.Numeric},
	{Column: "net_score", NamedC: "net_score", Type: enumor.Numeric},
	{Column: "cost_score", NamedC: "cost_score", Type: enumor.Numeric},
	{Column: "cover_rate", NamedC: "cover_rate", Type: enumor.Numeric},
	{Column: "user_distribution", NamedC: "user_distribution", Type: enumor.Json},
	{Column: "result_idc_ids", NamedC: "result_idc_ids", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.String},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.String},
}

// SchemeTable scheme表
type SchemeTable struct {
	// ID scheme ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" validate:"min=-1" json:"bk_biz_id"`
	// Name scheme名称
	Name string `db:"name" validate:"max=255" json:"name"`
	// BizType 业务类型
	BizType string `db:"biz_type" json:"biz_type" validate:"max=64"`
	// Vendors 机房供应商
	Vendors types.StringArray `db:"vendors" json:"vendors"`
	// DeploymentArchitecture 部署方式
	DeploymentArchitecture types.StringArray `db:"deployment_architecture" json:"deployment_architecture"`
	// CoverPing 网络延迟
	CoverPing float64 `db:"cover_ping" json:"cover_ping"`
	// CompositeScore 综合评分
	CompositeScore float64 `db:"composite_score" json:"composite_score"`
	// NetScore 网络评分
	NetScore float64 `db:"net_score" json:"net_score"`
	// CostScore 成本评分
	CostScore float64 `db:"cost_score" json:"cost_score"`
	// CoverRate 覆盖率
	CoverRate float64 `db:"cover_rate" json:"cover_rate"`
	// UserDistribution 用户分布占比信息
	UserDistribution types.AreaInfos `db:"user_distribution" json:"user_distribution"`
	// ResultIdcIDs 推荐机房ID列表
	ResultIdcIDs types.StringArray `db:"result_idc_ids" json:"result_idc_ids"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return scheme table name.
func (v SchemeTable) TableName() table.Name {
	return table.CloudSelectionSchemeTable
}

// InsertValidate validate scheme table on insert.
func (v SchemeTable) InsertValidate() error {
	if len(v.Name) == 0 {
		return errors.New("name can not be nil")
	}

	if v.BkBizID == 0 {
		return errors.New("biz id can not be empty")
	}

	if len(v.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(v)
}

// UpdateValidate validate scheme table on update.
func (v SchemeTable) UpdateValidate() error {
	if err := validator.Validate.Struct(v); err != nil {
		return err
	}

	if len(v.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(v.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
