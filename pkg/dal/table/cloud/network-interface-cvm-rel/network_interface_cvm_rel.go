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

package networkinterfacecvmrel

import (
	"errors"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/utils"
)

// NetworkInterfaceCvmRelColumns defines all the network interface and cvm rel table's columns.
var NetworkInterfaceCvmRelColumns = utils.MergeColumns(nil, NetworkInterfaceCvmRelTableColumnDescriptor)

// NetworkInterfaceCvmRelTableColumnDescriptor is network interface's column descriptors.
var NetworkInterfaceCvmRelTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "cvm_id", NamedC: "cvm_id", Type: enumor.String},
	{Column: "network_interface_id", NamedC: "network_interface_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// NetworkInterfaceCvmRelTable 网络接口跟主机的关联表
type NetworkInterfaceCvmRelTable struct {
	// ID 主键 ID
	ID uint64 `db:"id" validate:"required" json:"id"`
	// CvmID 主机ID
	CvmID string `db:"cvm_id" validate:"required,lte=64" json:"cvm_id"`
	// NetworkInterfaceID 网络接口ID
	NetworkInterfaceID string `db:"network_interface_id" validate:"required,lte=64" json:"network_interface_id"`
	// Creator 创建者
	Creator string `db:"creator" validate:"required,lte=64" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt *time.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
}

// TableName return azure network interface and cvm rel table name.
func (n NetworkInterfaceCvmRelTable) TableName() table.Name {
	return table.NetworkInterfaceTable
}

// InsertValidate network interface and cvm rel table when insert.
func (n NetworkInterfaceCvmRelTable) InsertValidate() error {
	if err := validator.Validate.Struct(n); err != nil {
		return err
	}

	if len(n.CvmID) == 0 {
		return errors.New("cvm_id is required")
	}

	if len(n.NetworkInterfaceID) == 0 {
		return errors.New("network_interface_id is required")
	}

	if len(n.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate network interface and cvm rel table when update.
func (n NetworkInterfaceCvmRelTable) UpdateValidate() error {
	if err := validator.Validate.Struct(n); err != nil {
		return err
	}

	if len(n.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
