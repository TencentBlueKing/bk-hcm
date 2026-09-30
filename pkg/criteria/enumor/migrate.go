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

package enumor

import "fmt"

// MigrationStatus is the status of a migration record or a migration audit row.
type MigrationStatus string

const (
	// MigrationStatusRunning is the in-progress status.
	MigrationStatusRunning MigrationStatus = "running"
	// MigrationStatusSuccess is the successful terminal status.
	MigrationStatusSuccess MigrationStatus = "success"
	// MigrationStatusFailed is the failed terminal status.
	MigrationStatusFailed MigrationStatus = "failed"
)

// Validate checks whether the migration status is one of the declared values.
func (s MigrationStatus) Validate() error {
	switch s {
	case MigrationStatusRunning, MigrationStatusSuccess, MigrationStatusFailed:
		return nil
	default:
		return fmt.Errorf("unsupported migration status: %s", s)
	}
}

// MigrationSkipReason is the reason a migration is skipped in one run.
type MigrationSkipReason string

const (
	// MigrationSkipApplied means a success record of the same ID exists.
	MigrationSkipApplied MigrationSkipReason = "applied"
	// MigrationSkipAboveMaxVersion means the migration is above the version limit of the run.
	MigrationSkipAboveMaxVersion MigrationSkipReason = "above_max_version"
	// MigrationSkipBaseline means init adopt records the migration as success without running it.
	MigrationSkipBaseline MigrationSkipReason = "baseline"
)
