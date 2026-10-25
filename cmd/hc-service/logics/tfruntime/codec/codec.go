/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package codec

import (
	"errors"
	"math/big"

	"hcm/pkg/plugin/tfplugin5"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Unknown is the only representation of a value the plugin has not computed.
// A missing attribute is null, not unknown.
var Unknown any = unknownSentinel{}

type unknownSentinel struct{}

// Codec implements executor.ValueCodec with msgpack dynamic values.
type Codec struct{}

// Encode writes a caller object to a msgpack DynamicValue.
// A type mismatch returns an error and no DynamicValue.
func (Codec) Encode(schema *tfplugin5.Schema, value any) (*tfplugin5.DynamicValue, error) {
	typ, err := schemaObject(schema)
	if err != nil {
		return nil, err
	}
	converted, err := convert(typ, value, "")
	if err != nil {
		return nil, err
	}
	raw, err := converted.MarshalMsgPack(typ)
	if err != nil {
		return nil, err
	}
	return &tfplugin5.DynamicValue{Msgpack: raw}, nil
}

// Decode reads msgpack with the schema that produced it.
// JSON-only values are rejected. A null object comes back as nil.
func (Codec) Decode(schema *tfplugin5.Schema, value *tfplugin5.DynamicValue) (any, error) {
	typ, err := schemaObject(schema)
	if err != nil {
		return nil, err
	}
	if value == nil || len(value.GetMsgpack()) == 0 {
		return nil, errors.New("msgpack is required")
	}
	decoded, err := tftypes.ValueFromMsgPack(value.GetMsgpack(), typ)
	if err != nil {
		return nil, err
	}
	return fromValue(decoded)
}

func fromValue(value tftypes.Value) (any, error) {
	if !value.IsKnown() {
		return Unknown, nil
	}
	if value.IsNull() {
		return nil, nil
	}
	switch value.Type().(type) {
	case tftypes.List, tftypes.Set, tftypes.Tuple:
		var items []tftypes.Value
		if err := value.As(&items); err != nil {
			return nil, err
		}
		out := make([]any, len(items))
		for i, item := range items {
			got, err := fromValue(item)
			if err != nil {
				return nil, err
			}
			out[i] = got
		}
		return out, nil
	case tftypes.Object, tftypes.Map:
		var attrs map[string]tftypes.Value
		if err := value.As(&attrs); err != nil {
			return nil, err
		}
		out := make(map[string]any, len(attrs))
		for name, item := range attrs {
			got, err := fromValue(item)
			if err != nil {
				return nil, err
			}
			out[name] = got
		}
		return out, nil
	}
	if value.Type().Is(tftypes.String) {
		var got string
		if err := value.As(&got); err != nil {
			return nil, err
		}
		return got, nil
	}
	if value.Type().Is(tftypes.Bool) {
		var got bool
		if err := value.As(&got); err != nil {
			return nil, err
		}
		return got, nil
	}
	if value.Type().Is(tftypes.Number) {
		got := new(big.Float)
		if err := value.As(got); err != nil {
			return nil, err
		}
		return got, nil
	}
	return nil, pathError("", "type mismatch")
}
