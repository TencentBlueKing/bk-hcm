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

package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hcm/cmd/iac-server/options"
	"hcm/pkg/cc"

	"github.com/stretchr/testify/assert"
)

// TestRun_FastFailWhenEtcdUnreachable tests fast failure when Etcd endpoint is unreachable (AC-004).
func TestRun_FastFailWhenEtcdUnreachable(t *testing.T) {
	cc.InitService(cc.IacServerName)

	yamlContent := `
network:
  bindIP: 127.0.0.1
  port: 19605
service:
  etcd:
    endpoints:
      - 127.0.0.1:22379
    dialTimeoutMS: 200
log:
  logDir: ` + t.TempDir() + `
  toStdErr: true
`
	cfgFile := filepath.Join(t.TempDir(), "iac_server_unreachable.yaml")
	err := os.WriteFile(cfgFile, []byte(yamlContent), 0600)
	assert.NoError(t, err)

	opt := &options.Option{
		Sys: &cc.SysOption{
			ConfigFile: cfgFile,
		},
	}

	err = Run(opt)
	if !assert.Error(t, err) {
		return
	}
	assert.True(t, strings.Contains(err.Error(), "new service discovery failed"),
		"error %q should contain 'new service discovery failed'", err.Error())
}
