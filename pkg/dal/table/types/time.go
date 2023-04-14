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

package types

import (
	"database/sql/driver"
	"fmt"
	"time"

	"hcm/pkg/tools/json"
)

// Time ISO 8610时间格式的时间
type Time string

// String ...
func (t *Time) String() string {
	return string(*t)
}

// Scan is used to decode raw message which is read from db into Time.
func (t *Time) Scan(raw interface{}) error {
	if raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case time.Time:
		*t = Time(v.In(time.Local).Format("2006-01-02T15:04:05Z07:00"))
		return nil

	default:
		return fmt.Errorf("unsupported Time raw type: %T", v)
	}
}

// Value encode the Time to a json raw, so that it can be stored to db with json raw.
func (t Time) Value() (driver.Value, error) {
	if len(t) == 0 {
		return "", nil
	}

	return json.Marshal(t)
}
