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

package core

import (
	"encoding/json"
	"errors"

	"hcm/pkg/logs"
)

// ExtMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type ExtMessage []byte

// MarshalStruct marshal struct.
func MarshalStruct(data interface{}) (ExtMessage, error) {
	marshal, err := json.Marshal(data)
	if err != nil {
		logs.ErrorDepthf(1, "marshal struct to ExtMessage failed, err: %v, data: %v", err, data)
		return nil, err
	}

	return marshal, nil
}

// MarshalJSON returns m as the JSON encoding of m.
func (m ExtMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *ExtMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.ExtMessage: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

var _ json.Marshaler = (*ExtMessage)(nil)
var _ json.Unmarshaler = (*ExtMessage)(nil)
