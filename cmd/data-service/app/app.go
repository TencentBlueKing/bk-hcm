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
	"fmt"

	"hcm/cmd/data-service/options"
	"hcm/cmd/data-service/service"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/shutdown"
)

// Run start the api server
func Run(opt *options.Option) error {
	as := new(hcServer)
	if err := as.prepare(opt); err != nil {
		return err
	}

	if err := as.service.ListenAndServeRest(); err != nil {
		return err
	}

	shutdown.RegisterFirstShutdown(as.finalizer)
	shutdown.WaitShutdown(20)
	return nil
}

type hcServer struct {
	service *service.Service
}

// prepare do prepare jobs before run api discover.
func (as *hcServer) prepare(opt *options.Option) error {
	// load settings from config file.
	if err := cc.LoadSettings(opt.Sys); err != nil {
		return fmt.Errorf("load settings from config files failed, err: %v", err)
	}

	logs.InitLogger(cc.DataService().Log.Logs())

	logs.Infof("load settings from config file success.")

	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("initialize service failed, err: %v", err)
	}
	as.service = svc

	return nil
}

func (as *hcServer) finalizer() {
	return
}
