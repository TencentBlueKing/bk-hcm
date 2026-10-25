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

package executor

import "hcm/pkg/plugin/tfplugin5"

// ValueCodec encodes caller values and decodes plugin values for a schema.
// Encode returns the generated DynamicValue so the executor can place it on a request.
// The executor does not implement msgpack. A later change supplies the codec.
type ValueCodec interface {
	Encode(schema *tfplugin5.Schema, value any) (*tfplugin5.DynamicValue, error)
	Decode(schema *tfplugin5.Schema, value *tfplugin5.DynamicValue) (any, error)
}
