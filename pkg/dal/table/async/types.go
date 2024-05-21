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

package tableasync

import (
	"database/sql/driver"

	"hcm/pkg/dal/table/types"
)

// Reason define async flow reason
type Reason struct {
	Message  string `json:"message,omitempty"`
	PreState string `json:"pre_state,omitempty"`
	// 改为rollback的次数
	RollbackCount uint `json:"rollback_count,omitempty"`
}

// Scan is used to decode raw message which is read from db into Reason.
func (d *Reason) Scan(raw interface{}) error {
	return types.Scan(raw, d)
}

// Value encode the Reason to a json raw, so that it can be stored to db with json raw.
func (d Reason) Value() (driver.Value, error) {
	return types.Value(d)
}
