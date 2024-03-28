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

package constant

import "regexp"

const (
	// TimeStdFormat is the system's standard time format to store or to query.
	TimeStdFormat = "2006-01-02T15:04:05Z07:00"
	// DateLayout is the date layout with '%Y-%m-%d'
	DateLayout = "2006-01-02"
	// DateTimeLayout is the date layout with '%Y-%m-%d %H:%M:%S'
	DateTimeLayout = "2006-01-02 15:04:05"
)

// TimeStdRegexp is a regular expression to match the TimeStdFormat
var TimeStdRegexp = regexp.
	MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[Z+]([0-9]{2}:[0-9]{2})*$`)
