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

package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// BinaryPlugin outside command execution engine.
type BinaryPlugin[IN any, OUT any] struct {
	// PluginPath refer to binary file path
	PluginPath string
	Args       []string
}

// Init checks params
func (e *BinaryPlugin[IN, OUT]) Init(pluginPath string, args ...string) error {
	f, err := os.Stat(pluginPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("plugin %s does not existed", pluginPath)
	}
	if !f.Mode().IsRegular() {
		return fmt.Errorf("%s is not regular file", pluginPath)
	}
	e.PluginPath = pluginPath
	e.Args = args
	return nil
}

// NewPlugin create new BinaryPlugin instance.
func NewPlugin[IN any, OUT any](pluginPath string, args ...string) (*BinaryPlugin[IN, OUT], error) {
	p := new(BinaryPlugin[IN, OUT])
	err := p.Init(pluginPath, args...)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Execute execute outside engine.
func (e *BinaryPlugin[IN, OUT]) Execute(input *IN) (output *OUT, err error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("encode input %+v failed, err %+v", input, err)
	}

	c := exec.Command(e.PluginPath, e.Args...)
	c.Stdin = bytes.NewBuffer(inputBytes)

	outputBytes, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("execute command failed, err %+v", err)
	}

	out := new(OUT)
	if err := json.Unmarshal(outputBytes, out); err != nil {
		return nil, fmt.Errorf("decode output %s failed, err %+v", string(outputBytes), err)
	}
	return out, nil
}
