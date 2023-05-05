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

// Package discovery defines server discovery operations.
package discovery

import (
	"fmt"

	"hcm/pkg/cc"
)

// Interface discovery interface.
type Interface interface {
	// GetServers 获取服务节点信息
	GetServers() ([]string, error)
}

// DeniedServers are virtual servers instance which is used to deny
// access to illegal services.
func DeniedServers(nm cc.Name) Interface {
	return &deniedServers{name: nm}
}

type deniedServers struct {
	name cc.Name
}

// GetServers is used to return denied errors.
func (ud deniedServers) GetServers() ([]string, error) {
	return nil, fmt.Errorf("access to %s server is not allowed", ud.name)
}
