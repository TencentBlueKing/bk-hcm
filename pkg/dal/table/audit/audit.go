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

// Package audit ...
package audit

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AuditColumns defines all the audit table's columns.
var AuditColumns = utils.MergeColumns(utils.InsertWithoutPrimaryID, AuditColumnDescriptor)

// AuditColumnDescriptor is AuditTable's column descriptors.
var AuditColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "res_id", NamedC: "res_id", Type: enumor.String},
	{Column: "cloud_res_id", NamedC: "cloud_res_id", Type: enumor.String},
	{Column: "res_name", NamedC: "res_name", Type: enumor.String},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "action", NamedC: "action", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "operator", NamedC: "operator", Type: enumor.String},
	{Column: "source", NamedC: "source", Type: enumor.String},
	{Column: "rid", NamedC: "rid", Type: enumor.String},
	{Column: "app_code", NamedC: "app_code", Type: enumor.String},
	{Column: "detail", NamedC: "detail", Type: enumor.Json},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// AuditTable is used to save resource's audit information.
type AuditTable struct {
	ID         uint64                   `db:"id" json:"id"`
	ResID      string                   `db:"res_id" json:"res_id" validate:"lte=64"`
	CloudResID string                   `db:"cloud_res_id" json:"cloud_res_id" validate:"lte=255"`
	ResName    string                   `db:"res_name" json:"res_name" validate:"lte=255"`
	ResType    enumor.AuditResourceType `db:"res_type" json:"res_type" validate:"lte=50"`
	Action     enumor.AuditAction       `db:"action" json:"action" validate:"lte=20"`
	BkBizID    int64                    `db:"bk_biz_id" json:"bk_biz_id"`
	Vendor     enumor.Vendor            `db:"vendor" json:"vendor" validate:"lte=16"`
	AccountID  string                   `db:"account_id" json:"account_id" validate:"lte=64"`
	Operator   string                   `db:"operator" json:"operator" validate:"lte=64"`
	Source     enumor.RequestSourceType `db:"source" json:"source" validate:"lte=20"`
	Rid        string                   `db:"rid" json:"rid" validate:"lte=64"`
	AppCode    string                   `db:"app_code" json:"app_code" validate:"lte=64"`
	Detail     *BasicDetail             `db:"detail" json:"detail" validate:"-"`
	CreatedAt  types.Time               `db:"created_at" json:"created_at"`
}

// CreateValidate audit when created
func (a AuditTable) CreateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if !a.ResType.Exist() {
		return fmt.Errorf("resource type: %s not support", a.ResType)
	}

	if !a.Action.Exist() {
		return fmt.Errorf("action: %s not support", a.Action)
	}

	if !a.Source.Exist() {
		return fmt.Errorf("source: %s not support", a.Source)
	}

	return nil
}

// TableName is the audit's database table name.
func (a AuditTable) TableName() table.Name {
	return table.AuditTable
}

// BasicDetail defines the audit's basic details.
type BasicDetail struct {
	Data    interface{} `json:"data,omitempty"`
	Changed interface{} `json:"changed,omitempty"`
}

// BasicDetailRaw audit's basic details with RawMessage for later decode
type BasicDetailRaw struct {
	Data    json.RawMessage `json:"data,omitempty"`
	Changed json.RawMessage `json:"changed,omitempty"`
}

// Scan is used to decode raw message which is read from db into a structured
// ScopeSelector instance.
func (detail *BasicDetail) Scan(raw interface{}) error {
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
func (detail *BasicDetail) Value() (driver.Value, error) {
	if detail == nil {
		return nil, errors.New("auditBasicDetail is not initialized, can not be encoded")
	}

	return json.Marshal(detail)
}
