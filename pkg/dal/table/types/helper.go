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
	"reflect"

	"hcm/pkg/tools/json"
)

// Scan is used to decode raw message which is read from db into dest.
func Scan(raw interface{}, dest interface{}) error {
	if dest == nil || raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &dest); err != nil {
			return fmt.Errorf("[]byte decode into %s failed, err: %v", reflect.TypeOf(dest).String(), err)
		}
		return nil

	case string:
		if err := json.Unmarshal([]byte(v), &dest); err != nil {
			return fmt.Errorf("string decode into %s failed, err: %v", reflect.TypeOf(dest).String(), err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported raw type: %T", v)
	}
}

// Value encode the source to a json raw, so that it can be stored to db with json raw.
func Value(source interface{}) (driver.Value, error) {
	if source == nil {
		return "", nil
	}

	return json.Marshal(source)
}
