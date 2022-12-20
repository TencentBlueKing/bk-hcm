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

package audit

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/utils"
)

// AuditColumns defines all the audit table's columns.
var AuditColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, AuditColumnDescriptor)

// AuditColumnDescriptor is Audit's column descriptors.
var AuditColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "res_id", NamedC: "res_id", Type: enumor.Numeric},
	{Column: "action", NamedC: "action", Type: enumor.String},
	{Column: "rid", NamedC: "rid", Type: enumor.String},
	{Column: "app_code", NamedC: "app_code", Type: enumor.String},
	{Column: "detail", NamedC: "detail", Type: enumor.Json},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "account_id", NamedC: "account_id", Type: enumor.Numeric},
	{Column: "tenant_id", NamedC: "tenant_id", Type: enumor.String},
	{Column: "operator", NamedC: "operator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// Audit is used to save resource's audit information.
type Audit struct {
	ID           uint64                   `db:"id" json:"id"`
	ResourceType enumor.AuditResourceType `db:"res_type" json:"resource_type"`
	ResourceID   string                   `db:"res_id" json:"resource_id"`
	Action       enumor.AuditAction       `db:"action" json:"action"`
	Rid          string                   `db:"rid" json:"rid"`
	AppCode      string                   `db:"app_code" json:"app_code"`
	Detail       *AuditBasicDetail        `db:"detail" json:"detail"`
	BizID        uint64                   `db:"bk_biz_id" json:"bk_biz_id"`
	AccountID    string                   `db:"account_id" json:"account_id"`
	TenantID     string                   `db:"tenant_id" json:"tenant_id"`
	Operator     string                   `db:"operator" json:"operator"`
	CreatedAt    *time.Time               `db:"created_at" json:"created_at"`
}

// CreateValidate audit when created
func (a Audit) CreateValidate() error {
	if len(a.ResourceType) == 0 {
		return errors.New("resource type can not be empty")
	}

	if len(a.ResourceID) == 0 {
		return errors.New("resource id can not be empty")
	}

	if len(a.Action) == 0 {
		return errors.New("action can not be empty")
	}

	if len(a.Rid) == 0 {
		return errors.New("request id can not be empty")
	}

	if len(a.Operator) == 0 {
		return errors.New("operator can not be empty")
	}

	if a.CreatedAt != nil && !a.CreatedAt.IsZero() {
		return errors.New("create_at can not be set, it is generated through db")
	}

	return nil
}

// TableName is the audit's database table name.
func (a Audit) TableName() table.Name {
	return table.AuditTable
}

// AuditBasicDetail defines the audit's basic details.
type AuditBasicDetail struct {
	Data    interface{} `json:"data,omitempty"`
	Changed interface{} `json:"changed,omitempty"`
}

// Scan is used to decode raw message which is read from db into a structured
// ScopeSelector instance.
func (detail *AuditBasicDetail) Scan(raw interface{}) error {
	if detail == nil {
		return errors.New("auditBasicDetail is not initialized")
	}

	if raw == nil {
		return errors.New("raw is nil, can not be decoded")
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &detail); err != nil {
			return fmt.Errorf("decode into auditBasicDetail failed, err: %v", err)
		}
		return nil
	case string:
		if err := json.Unmarshal([]byte(v), &detail); err != nil {
			return fmt.Errorf("decode into auditBasicDetail failed, err: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported auditBasicDetail raw type: %T", v)
	}
}

// Value encode the scope selector to a json raw, so that it can be stored to db with json raw.
func (detail *AuditBasicDetail) Value() (driver.Value, error) {
	if detail == nil {
		return nil, errors.New("auditBasicDetail is not initialized, can not be encoded")
	}

	return json.Marshal(detail)
}
