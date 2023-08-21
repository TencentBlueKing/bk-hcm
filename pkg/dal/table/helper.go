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
	"database/sql/driver"
	"fmt"

	"hcm/pkg/tools/json"
)

// SqlValue define sql value func.
func SqlValue(data interface{}) (driver.Value, error) {
	if data == nil {
		return nil, nil
	}

	return json.Marshal(data)
}

// SqlScan define sql scan func.
func SqlScan(data interface{}, raw interface{}) error {
	if data == nil || raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &data); err != nil {
			return fmt.Errorf("decode into failed, err: %v", err)
		}
		return nil

	case string:
		if err := json.Unmarshal([]byte(v), &data); err != nil {
			return fmt.Errorf("decode into sets failed, err: %v", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported raw type: %T", v)
	}
}
