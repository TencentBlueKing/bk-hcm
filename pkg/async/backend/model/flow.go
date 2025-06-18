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

package model

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
)

// Flow 任务流
type Flow struct {
	Name      enumor.FlowName       `json:"name"`
	ShareData *tableasync.ShareData `json:"share_data"`
	Memo      string                `json:"memo"`

	ID        string             `json:"id"`
	State     enumor.FlowState   `json:"state"`
	Reason    *tableasync.Reason `json:"reason"`
	Worker    *string            `json:"worker"`
	Creator   string             `json:"creator"`
	Reviser   string             `json:"reviser"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`

	Tasks    []Task `json:"tasks"`
	TenantID string `json:"tenant_id"`
}

// CreateValidate Flow.
func (f Flow) CreateValidate() error {

	if len(f.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(f.Name) == 0 {
		return errors.New("name is required")
	}

	if len(f.State) == 0 {
		return errors.New("state is required")
	}

	if f.Reason != nil {
		return errors.New("reason can not set")
	}

	if len(f.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(f.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(f.CreatedAt) == 0 {
		return errors.New("updated_at can not set")
	}

	if len(f.UpdatedAt) == 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate Flow.
func (f Flow) UpdateValidate() error {

	if len(f.ID) == 0 {
		return errors.New("id is required")
	}

	if len(f.Name) != 0 {
		return errors.New("name can not set")
	}

	if len(f.Creator) != 0 {
		return errors.New("creator can not set")
	}

	if len(f.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(f.CreatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	if len(f.UpdatedAt) == 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}
