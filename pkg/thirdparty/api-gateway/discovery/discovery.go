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

package discovery

import (
	"fmt"
	"sync"
)

// Discovery used to third-party service discovery.
type Discovery struct {
	Name    string
	Servers []string
	index   int
	sync.Mutex
}

// GetServers get third-party service server host.
func (d *Discovery) GetServers() ([]string, error) {
	d.Lock()
	defer d.Unlock()
	num := len(d.Servers)
	if num == 0 {
		return []string{}, fmt.Errorf("there is no %s server can be used", d.Name)
	}
	if d.index < num-1 {
		d.index = d.index + 1
		return append(d.Servers[d.index-1:], d.Servers[:d.index-1]...), nil
	}
	d.index = 0
	return append(d.Servers[num-1:], d.Servers[:num-1]...), nil
}
