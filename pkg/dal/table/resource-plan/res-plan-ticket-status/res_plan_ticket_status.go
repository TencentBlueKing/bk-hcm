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

package resplanticketstatus

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanTicketStatusColumns defines all the resource plan ticket status table's columns.
var ResPlanTicketStatusColumns = utils.MergeColumns(nil, ResPlanTicketStatusColumnDescriptor)

// ResPlanTicketStatusColumnDescriptor is ResPlanTicketStatusTable's column descriptors.
var ResPlanTicketStatusColumnDescriptor = utils.ColumnDescriptors{
	{Column: "ticket_id", NamedC: "ticket_id", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "itsm_sn", NamedC: "itsm_sn", Type: enumor.String},
	{Column: "itsm_url", NamedC: "itsm_url", Type: enumor.String},
	{Column: "crp_sn", NamedC: "crp_sn", Type: enumor.String},
	{Column: "crp_url", NamedC: "crp_url", Type: enumor.String},
	{Column: "message", NamedC: "message", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanTicketStatusTable is used to save resource's resource plan ticket status information.
type ResPlanTicketStatusTable struct {
	// TicketID 单据表唯一ID
	TicketID string `db:"ticket_id" json:"ticket_id" validate:"lte=64"`
	// Status 单据状态
	Status enumor.RPTicketStatus `db:"status" json:"status" validate:"lte=64"`
	// ItsmSN 关联的ITSM单据编码
	ItsmSN string `db:"itsm_sn" json:"itsm_sn" validate:"lte=64"`
	// ItsmURL 关联的ITSM单据链接
	ItsmURL string `db:"itsm_url" json:"itsm_url" validate:"lte=64"`
	// CrpSN 关联的CRP单据编码
	CrpSN string `db:"crp_sn" json:"crp_sn" validate:"lte=64"`
	// CrpURL 关联的CRP单据链接
	CrpURL string `db:"crp_url" json:"crp_url" validate:"lte=64"`
	// Message 单据失败信息
	Message string `db:"message" json:"message"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName is the recycleRecord's database table name.
func (r ResPlanTicketStatusTable) TableName() table.Name {
	return table.ResPlanTicketStatusTable
}

// InsertValidate validate resource plan ticket status on insertion.
func (r ResPlanTicketStatusTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.TicketID) == 0 {
		return errors.New("ticket id can not be empty")
	}

	if err := r.Status.Validate(); err != nil {
		return err
	}

	return nil
}

// UpdateValidate validate resource plan ticket status on update.
func (r ResPlanTicketStatusTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Status) > 0 {
		if err := r.Status.Validate(); err != nil {
			return err
		}
	}

	return nil
}
