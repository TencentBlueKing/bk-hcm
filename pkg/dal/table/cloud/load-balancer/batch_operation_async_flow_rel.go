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

package tablelb

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// BatchOperationAsyncFlowRelColumns defines all the tcloud_lb_url_rule table's columns.
var BatchOperationAsyncFlowRelColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, BatchOperationAsyncFlowRelColumnsDescriptor)

// BatchOperationAsyncFlowRelColumnsDescriptor is batch_task table's column descriptors.
var BatchOperationAsyncFlowRelColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "batch_operation_id", NamedC: "batch_operation_id", Type: enumor.String},
	{Column: "audit_id", NamedC: "audit_id", Type: enumor.Numeric},
	{Column: "flow_id", NamedC: "flow_id", Type: enumor.String},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// BatchOperationAsyncFlowRelTable 批量操作任务表
type BatchOperationAsyncFlowRelTable struct {
	ID string `db:"id" validate:"lte=64" json:"id"`

	BatchOperationID string     `db:"batch_operation_id" validate:"lte=64" json:"batch_operation_id"`
	AuditID          *uint64    `db:"audit_id" json:"audit_id"`
	FlowID           string     `db:"flow_id" validate:"lte=64" json:"flow_id"`
	Creator          string     `db:"creator" validate:"lte=64" json:"creator"`
	CreatedAt        types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
}

// TableName return tcloud_lb_url_rule table name.
func (tlbur BatchOperationAsyncFlowRelTable) TableName() table.Name {
	return table.BatchOperationAsyncFlowRelTable
}

// InsertValidate tcloud_lb_url_rule table when insert.
func (tlbur BatchOperationAsyncFlowRelTable) InsertValidate() error {
	if err := validator.Validate.Struct(tlbur); err != nil {
		return err
	}

	if len(tlbur.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(tlbur.BatchOperationID) == 0 {
		return errors.New("batch_operation_id is required")
	}

	if len(tlbur.FlowID) == 0 {
		return errors.New("flow_id is required")
	}

	if tlbur.AuditID == nil {
		return errors.New("audit_id is required")
	}

	return nil
}

// UpdateValidate tcloud_lb_url_rule table when update.
func (tlbur BatchOperationAsyncFlowRelTable) UpdateValidate() error {
	if err := validator.Validate.Struct(tlbur); err != nil {
		return err
	}

	if len(tlbur.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
