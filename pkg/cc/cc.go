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

// Package cc defines hcm service config center.
package cc

import (
	"sync"

	"hcm/pkg/logs"
)

var runtimeOnce sync.Once

// rt is the runtime Setting which is loaded from
// config file.
// It can be called only after LoadSettings is executed successfully.
var rt *runtime

func initRuntime(s Setting) {
	runtimeOnce.Do(func() {
		rt = &runtime{
			settings: s,
		}
	})
}

type runtime struct {
	lock     sync.Mutex
	settings Setting
}

// Ready is used to test if the runtime configuration is
// initialized with load from file success and already
// ready to use.
func (r *runtime) Ready() bool {
	if r == nil {
		return false
	}

	if r.settings == nil {
		return false
	}

	return true
}

// ApiServer return api server Setting.
func ApiServer() ApiServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty api server setting")
		return ApiServerSetting{}
	}

	s, ok := rt.settings.(*ApiServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get api server setting", ServiceName())
		return ApiServerSetting{}
	}

	return *s
}

// CloudServer return cloud server Setting.
func CloudServer() CloudServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty cloud server setting")
		return CloudServerSetting{}
	}

	s, ok := rt.settings.(*CloudServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get cloud server setting", ServiceName())
		return CloudServerSetting{}
	}

	return *s
}

// DataService return data service Setting.
func DataService() DataServiceSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty data service setting")
		return DataServiceSetting{}
	}

	s, ok := rt.settings.(*DataServiceSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get data service setting", ServiceName())
		return DataServiceSetting{}
	}

	return *s
}

// HCService return hc service Setting.
func HCService() HCServiceSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty hc service setting")
		return HCServiceSetting{}
	}

	s, ok := rt.settings.(*HCServiceSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get hc service setting", ServiceName())
		return HCServiceSetting{}
	}

	return *s
}

// AuthServer return auth server Setting.
func AuthServer() AuthServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty auth server setting")
		return AuthServerSetting{}
	}

	s, ok := rt.settings.(*AuthServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get auth server setting", ServiceName())
		return AuthServerSetting{}
	}

	return *s
}

// WebServer return web server Setting.
func WebServer() WebServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty web server setting")
		return WebServerSetting{}
	}

	s, ok := rt.settings.(*WebServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get web server setting", ServiceName())
		return WebServerSetting{}
	}

	return *s
}

// TaskServer return task server Setting.
func TaskServer() TaskServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty task server setting")
		return TaskServerSetting{}
	}

	s, ok := rt.settings.(*TaskServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get task server setting", ServiceName())
		return TaskServerSetting{}
	}

	return *s
}

// AccountServer return account server Setting.
func AccountServer() AccountServerSetting {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	if !rt.Ready() {
		logs.ErrorDepthf(1, "runtime not ready, return empty task server setting")
		return AccountServerSetting{}
	}

	s, ok := rt.settings.(*AccountServerSetting)
	if !ok {
		logs.ErrorDepthf(1, "current %s service can not get account server setting", ServiceName())
		return AccountServerSetting{}
	}

	return *s
}
