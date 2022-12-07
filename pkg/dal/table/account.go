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

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
)

// AccountColumns defines all the account table's columns.
var AccountColumns = mergeColumns(insertWithoutPrimaryID, AccountColumnDescriptor)

// AccountColumnDescriptor is Account's column descriptors.
var AccountColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
	},
	mergeColumnDescriptors("spec", AccountSpecColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// Account defines an account's detail information
type Account struct {
	// ID is an auto-increased value, which is an account's unique identity.
	ID uint64 `db:"id" json:"id"`
	// Spec is a collection of account's specifics defined with user
	Spec *AccountSpec `db:"spec" json:"spec"`
	// Revision record this account's revision information
	Revision *Revision `db:"revision" json:"revision"`
}

// TableName is the account's database table name.
func (a Account) TableName() types.Name {
	return types.AccountTable
}

// ValidateCreate validate account's info when created.
func (a Account) ValidateCreate() error {
	if a.ID != 0 {
		return errors.New("id can not be empty")
	}

	if a.Spec == nil {
		return errors.New("spec can not be empty")
	}

	if err := a.Spec.ValidateCreate(); err != nil {
		return err
	}

	if a.Revision == nil {
		return errors.New("revision can not be empty")
	}

	if err := a.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate account's info when update.
func (a Account) ValidateUpdate() error {
	if a.ID > 0 {
		return errors.New("id can not be updated")
	}

	if a.Spec == nil {
		return errors.New("spec can not be empty")
	}

	if err := a.Spec.ValidateUpdate(); err != nil {
		return err
	}

	if a.Revision == nil {
		return errors.New("revision can not be empty")
	}

	if err := a.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// AccountSpecColumns defines all the account spec's columns.
var AccountSpecColumns = mergeColumns(nil, AccountSpecColumnDescriptor)

// AccountSpecColumnDescriptor is AccountSpec's column descriptors.
var AccountSpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
}

// AccountSpec is a collection of account's specifics defined with user
type AccountSpec struct {
	Name string  `db:"name" json:"name"`
	Memo *string `db:"memo" json:"memo"`
}

// ValidateCreate validate spec when created.
func (as *AccountSpec) ValidateCreate() error {
	if as == nil {
		return errors.New("spec is nil")
	}

	if err := validator.ValidateName(as.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(as.Memo, false); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate spec when updated.
func (as *AccountSpec) ValidateUpdate() error {
	if as == nil {
		return errors.New("spec can not be empty")
	}

	if err := validator.ValidateName(as.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(as.Memo, false); err != nil {
		return err
	}

	return nil
}
