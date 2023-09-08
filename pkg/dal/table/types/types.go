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

// Package types ...
package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JsonField 对应 db 的 json field 格式字段
type JsonField string

// NewJsonField create a json field from data.
func NewJsonField(data any) (JsonField, error) {
	js, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return JsonField(js), nil
}

// StringArray 对应 db 的json字符串数组格式字段
type StringArray []string

// Scan is used to decode raw message which is read from db into StringArray.
func (str *StringArray) Scan(raw interface{}) error {
	if str == nil || raw == nil {
		*str = make([]string, 0)
		return nil
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &str); err != nil {
			return fmt.Errorf("decode into string array failed, err: %v", err)
		}
		return nil

	case string:
		if err := json.Unmarshal([]byte(v), &str); err != nil {
			return fmt.Errorf("decode into string array failed, err: %v", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported string array raw type: %T", v)
	}
}

// Value encode the StringArray to a json raw, so that it can be stored to db with json raw.
func (str StringArray) Value() (driver.Value, error) {
	if str == nil {
		return "[]", nil
	}

	return json.Marshal(str)
}

// Int64Array 对应 db 的json数字数组格式字段
type Int64Array []int64

// Scan is used to decode raw message which is read from db into Int64Array.
func (i *Int64Array) Scan(raw interface{}) error {
	if i == nil || raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &i); err != nil {
			return fmt.Errorf("decode into int64 array failed, err: %v", err)
		}
		return nil

	case string:
		if err := json.Unmarshal([]byte(v), &i); err != nil {
			return fmt.Errorf("decode into int64 array failed, err: %v", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported int64 array raw type: %T", v)
	}
}

// Value encode the Int64Array to a json raw, so that it can be stored to db with json raw.
func (i Int64Array) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}

	return json.Marshal(i)
}
