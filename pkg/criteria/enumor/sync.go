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

// SyncStatus is sync status.
type SyncStatus string

// Validate SyncStatus.
func (v SyncStatus) Validate() error {
	switch v {
	case SyncSuccess:
	case SyncFailed:
	case NotSync:
	case Syncing:
	default:
		return fmt.Errorf("unsupported sync status: %s", v)
	}

	return nil
}

const (
	// SyncSuccess status
	SyncSuccess SyncStatus = "sync_success"
	// SyncFailed status
	SyncFailed SyncStatus = "sync_failed"
	// NotSync status
	NotSync SyncStatus = "not_sync"
	// Syncing status
	Syncing SyncStatus = "syncing"
)
