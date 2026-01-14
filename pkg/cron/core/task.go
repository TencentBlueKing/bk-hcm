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

// Package core ...
package core

import (
	"time"

	"hcm/pkg/kit"
)

// Task defines the interface for the Task.
type Task interface {
	// Name return the name of the task.
	Name() string
	// Next return the next time to run the task.
	Next() (time.Time, error)
	// Do execute the task.
	Do(kt *kit.Kit) error
	// GetURL get the url of the task, require every task to have external api in service.
	GetURL() string
}
