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

package table

import "fmt"

// Table defines all the database table
// related resources.
type Table interface {
	TableName() Name
}

// Name is database table's name type
type Name string

const (
	// AuditTable is audit table's name
	AuditTable Name = "audit"
	// AccountTable is account table's name.
	AccountTable Name = "account"
	// AccountBizRelTable is account and biz relation table's name.
	AccountBizRelTable Name = "account_biz_rel"
	// IDGenerator is id generator table's name.
	IDGenerator Name = "id_generator"
	// SecurityGroupTable is security group table's name.
	SecurityGroupTable Name = "security_group"
	// VpcTable is vpc table's name.
	VpcTable Name = "vpc"
)

// Validate whether the table name is valid or not.
func (n Name) Validate() error {
	switch n {
	case AuditTable:
	case AccountTable:
	case AccountBizRelTable:
	case IDGenerator:
	default:
		return fmt.Errorf("unknown table name: %s", n)
	}

	return nil
}
