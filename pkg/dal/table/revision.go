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
	"time"

	"hcm/pkg/criteria/enumor"
)

// RevisionColumns defines all the Revision table's columns.
var RevisionColumns = mergeColumns(nil, RevisionColumnDescriptor)

// RevisionColumnDescriptor is Revision's column descriptors.
var RevisionColumnDescriptor = ColumnDescriptors{
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// Revision is a resource's status information
type Revision struct {
	Creator   string     `db:"creator" json:"creator"`
	Reviser   string     `db:"reviser" json:"reviser"`
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

// IsEmpty test whether a revision is empty or not.
func (r Revision) IsEmpty() bool {
	if len(r.Creator) != 0 {
		return false
	}

	if len(r.Reviser) != 0 {
		return false
	}

	if !r.CreatedAt.IsZero() {
		return false
	}

	if !r.UpdatedAt.IsZero() {
		return false
	}

	return true
}

const lagSeconds = 5 * 60

// ValidateCreate validate revision when created
func (r Revision) ValidateCreate() error {
	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if r.CreatedAt != nil && !r.CreatedAt.IsZero() {
		return errors.New("created_at can not be set, it is generated through db")
	}

	if r.UpdatedAt != nil && !r.UpdatedAt.IsZero() {
		return errors.New("updated_at can not be set, it is generated through db")
	}

	return nil
}

// ValidateUpdate validate revision when updated
func (r Revision) ValidateUpdate() error {
	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	if r.CreatedAt != nil && !r.CreatedAt.IsZero() {
		return errors.New("created_at can not be updated")
	}

	if r.UpdatedAt != nil && !r.UpdatedAt.IsZero() {
		return errors.New("updated_at can not be set, it is generated through db")
	}

	return nil
}

// CreatedRevisionColumns defines all the Revision table's columns.
var CreatedRevisionColumns = mergeColumns(nil, CreatedRevisionColumnDescriptor)

// CreatedRevisionColumnDescriptor is CreatedRevision's column descriptors.
var CreatedRevisionColumnDescriptor = ColumnDescriptors{
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// CreatedRevision is a resource's reversion information being created.
type CreatedRevision struct {
	Creator   string     `db:"creator" json:"creator"`
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
}

// Validate revision when created
func (r CreatedRevision) Validate() error {
	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if r.CreatedAt != nil && !r.CreatedAt.IsZero() {
		return errors.New("create_at can not be set, it is generated through db")
	}

	return nil
}
