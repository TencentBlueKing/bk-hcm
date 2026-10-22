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

package tfruntime

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hcm/pkg/kit"
)

func TestProcessManagerStub(t *testing.T) {
	m, err := NewProcessManager(ProcessConfig{PluginName: "tcloud", ExecPath: "/bin/true"})
	require.NoError(t, err)

	exec, err := m.Start(kit.New())
	assert.Nil(t, exec)
	assert.True(t, errors.Is(err, ErrNotImplemented))

	assert.False(t, m.Exited())
	assert.True(t, errors.Is(m.CheckHealth(), ErrNotImplemented))

	// Close 幂等。
	assert.NoError(t, m.Close(context.Background()))
	assert.NoError(t, m.Close(context.Background()))
}

func TestNewProcessManagerValidate(t *testing.T) {
	_, err := NewProcessManager(ProcessConfig{ExecPath: "/bin/true"})
	assert.Error(t, err)

	_, err = NewProcessManager(ProcessConfig{PluginName: "tcloud"})
	assert.Error(t, err)
}

func TestDefaultStarter(t *testing.T) {
	s := NewStarter(StarterOptions{LogDir: "/tmp/log"})

	m, err := s.NewProcessManager(kit.New(), "tcloud", "1.0.0", "/bin/true")
	require.NoError(t, err)
	assert.NotNil(t, m)

	_, err = s.NewProcessManager(kit.New(), "", "1.0.0", "/bin/true")
	assert.Error(t, err)
}

func TestDiagnosticsError(t *testing.T) {
	warnOnly := []Diagnostic{{Severity: SeverityWarning, Summary: "deprecated"}}
	assert.False(t, HasError(warnOnly))
	assert.NoError(t, DiagnosticsToError(warnOnly))
	assert.NoError(t, DiagnosticsToError(nil))

	diags := []Diagnostic{
		{Severity: SeverityWarning, Summary: "deprecated"},
		{
			Severity:  SeverityError,
			Summary:   "missing secret",
			Detail:    "secret_key=abc",
			Attribute: AttributePath{"auth", "secret"},
		},
	}
	assert.True(t, HasError(diags))

	// 业务错误经包装后仍可被 errors.As 识别，传输错误不能。
	err := fmt.Errorf("configure failed: %w", DiagnosticsToError(diags))
	var de *DiagnosticsError
	require.True(t, errors.As(err, &de))
	assert.Len(t, de.Diagnostics, 2)

	assert.False(t, errors.As(errors.New("context deadline exceeded"), &de))

	// Error() 只含 Error 级 Summary 与属性路径，不含 Warning 与 Detail。
	assert.Contains(t, de.Error(), "missing secret (auth.secret)")
	assert.NotContains(t, de.Error(), "deprecated")
	assert.NotContains(t, de.Error(), "secret_key=abc")
}
