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

// Package orgtopo ...
package orgtopo

import (
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/client"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
)

// OrgTopoTiming timing sync org topo.
func OrgTopoTiming(c *client.ClientSet, state serviced.State, intervalMin time.Duration) {
	r := &orgTopoSvc{
		client: c,
		state:  state,
	}

	go r.orgTopoTiming(intervalMin)
}

func (ots *orgTopoSvc) orgTopoTiming(intervalMin time.Duration) {
	for {
		time.Sleep(intervalMin * time.Minute)

		kt := core.NewBackendKit()
		if !ots.state.IsMaster() {
			logs.Infof("sync org topo timing, but is not master, skip, rid: %s", kt.Rid)
			time.Sleep(10 * time.Minute)
			continue
		}

		err := ots.SyncOrgTopo(kt)
		if err != nil {
			logs.Errorf("sync org topo timing failed, err: %+v, rid: %s", err, kt.Rid)
			time.Sleep(10 * time.Minute)
			continue
		}
	}
}
