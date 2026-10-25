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

package tfruntime

import (
	"testing"

	"hcm/cmd/hc-service/logics/tfruntime/executor"
	"hcm/pkg/plugin/tfplugin5"
)

type staticCodec struct {
	value *tfplugin5.DynamicValue
}

func (c staticCodec) Encode(*tfplugin5.Schema, any) (*tfplugin5.DynamicValue, error) {
	return c.value, nil
}

func (c staticCodec) Decode(*tfplugin5.Schema, *tfplugin5.DynamicValue) (any, error) {
	return "decoded", nil
}

func TestProcessManager_SetValueCodec(t *testing.T) {
	manager := NewProcessManager("tencentcloud", "/bin/true", nil, "", "")
	exec := &executor.PluginExecutor{}
	manager.mu.Lock()
	manager.executor = exec
	manager.mu.Unlock()

	manager.SetValueCodec(staticCodec{value: &tfplugin5.DynamicValue{Msgpack: []byte("packed")}})
	if manager.codec == nil {
		t.Fatal("codec was not stored")
	}
}
