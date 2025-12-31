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

// Package ressync ...
package ressync

import (
	"context"
	"time"

	"hcm/cmd/woa-server/logics/config"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/utils/wait"
)

// Logics provides management interface for resource sync.
type Logics interface {
	// SyncCapacity sync device capacity info collection
	SyncCapacity() error
	// SyncLeftIP sync left ip config list
	SyncLeftIP() error
}

// logics resource sync logics.
type logics struct {
	sd           serviced.State
	client       *client.ClientSet
	configLogics config.Logics
}

// New creates resource sync logics instance.
func New(sd serviced.State, client *client.ClientSet, configLogics config.Logics) (Logics, error) {
	logic := &logics{
		sd:           sd,
		client:       client,
		configLogics: configLogics,
	}
	go logic.Run()
	return logic, nil
}

// Run starts dispatcher
func (l *logics) Run() {
	ctx := context.Background()
	capacityMinute := cc.WoaServer().ResourceSync.SyncCapacity
	if capacityMinute > 0 {
		// sync capacity every 6 hours
		go wait.JitterUntil(l.SyncCapacity, time.Duration(capacityMinute)*time.Minute, 0.5, true, ctx)
	}

	leftIPMinute := cc.WoaServer().ResourceSync.SyncLeftIP
	if leftIPMinute > 0 {
		// sync left ip every 6 hours
		go wait.JitterUntil(l.SyncLeftIP, time.Duration(leftIPMinute)*time.Minute, 0.5, true, ctx)
	}
}
