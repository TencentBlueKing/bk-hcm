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

// Package service ...
package service

import (
	"errors"
	"fmt"
	"net/http"

	"hcm/cmd/auth-server/options"
	"hcm/cmd/auth-server/service/auth"
	"hcm/cmd/auth-server/service/iam"
	"hcm/cmd/auth-server/service/initial"
	"hcm/pkg/cc"
	apicli "hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/iam/client"
	pkgauth "hcm/pkg/iam/sdk/auth"
	"hcm/pkg/iam/sys"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/tools/ssl"
)

// Service do all the data service's work
type Service struct {
	client *ClientSet
	serve  *http.Server
	state  serviced.State

	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
	// disableWriteOpt defines which biz's write operation needs to be disabled
	disableWriteOpt *options.DisableWriteOption

	// iam logic module.
	iam *iam.IAM
	// initial logic module.
	initial *initial.Initial
	// auth logic module.
	auth *auth.Auth
}

// NewService create a service instance.
func NewService(sd serviced.Discover, iamSettings cc.IAM, esbSettings cc.Esb, disableAuth bool,
	disableWriteOpt *options.DisableWriteOption) (*Service, error) {

	cli, err := newClientSet(sd, iamSettings, esbSettings, disableAuth)
	if err != nil {
		return nil, fmt.Errorf("new client set failed, err: %v", err)
	}

	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}

	s := &Service{
		client:          cli,
		state:           state,
		disableAuth:     disableAuth,
		disableWriteOpt: disableWriteOpt,
	}

	if err = s.initLogicModule(); err != nil {
		return nil, err
	}

	return s, nil
}

func newClientSet(sd serviced.Discover, iamSettings cc.IAM, esbSettings cc.Esb, disableAuth bool) (*ClientSet, error) {

	logs.Infof("start initialize the client set.")

	tlsConfig := new(ssl.TLSConfig)
	if iamSettings.TLS.Enable() {
		tlsConfig = &ssl.TLSConfig{
			InsecureSkipVerify: iamSettings.TLS.InsecureSkipVerify,
			CertFile:           iamSettings.TLS.CertFile,
			KeyFile:            iamSettings.TLS.KeyFile,
			CAFile:             iamSettings.TLS.CAFile,
			Password:           iamSettings.TLS.Password,
		}
	}

	// initiate system api client set.
	restCli, err := restcli.NewClient(tlsConfig)
	if err != nil {
		return nil, err
	}
	apiClientSet := apicli.NewClientSet(restCli, sd)

	logs.Infof("initialize system api client set success.")

	cfg := &client.Config{
		Address:   iamSettings.Endpoints,
		AppCode:   iamSettings.AppCode,
		AppSecret: iamSettings.AppSecret,
		SystemID:  sys.SystemIDHCM,
		TLS:       tlsConfig,
	}
	iamCli, err := client.NewClient(cfg, metrics.Register())
	if err != nil {
		return nil, err
	}

	iamSys, err := sys.NewSys(iamCli)
	if err != nil {
		return nil, fmt.Errorf("new iam sys failed, err: %v", err)
	}
	logs.Infof("initialize iam sys success.")

	// initialize iam auth sdk
	iamLgc, err := iam.NewIAM(apiClientSet.DataService(), iamSys, disableAuth)
	if err != nil {
		return nil, fmt.Errorf("new iam logics failed, err: %v", err)
	}

	esbClient, err := esb.NewClient(&esbSettings, metrics.Register())

	cmdbCfg := cc.AuthServer().Cmdb
	cmdbCli, err := cmdb.NewClient(&cmdbCfg, metrics.Register())
	if err != nil {
		return nil, err
	}

	authSdk, err := pkgauth.NewAuth(iamCli, iamLgc, esbClient)
	if err != nil {
		return nil, fmt.Errorf("new iam auth sdk failed, err: %v", err)
	}
	logs.Infof("initialize iam auth sdk success.")

	cs := &ClientSet{
		ds:      apiClientSet.DataService(),
		sys:     iamSys,
		auth:    authSdk,
		cmdbCli: cmdbCli,
	}
	logs.Infof("initialize the client set success.")
	return cs, nil
}

// ClientSet defines configure server's all the depends api client.
type ClientSet struct {
	// data service's sys api
	ds *dataservice.Client
	// iam sys related operate.
	sys *sys.Sys
	// auth related operate.
	auth pkgauth.Authorizer
	// cmdb client.
	cmdbCli cmdb.Client
}

// initLogicModule init logic module.
func (s *Service) initLogicModule() error {
	var err error

	s.initial, err = initial.NewInitial(s.client.sys, s.disableAuth)
	if err != nil {
		return err
	}

	s.iam, err = iam.NewIAM(s.client.ds, s.client.sys, s.disableAuth)
	if err != nil {
		return err
	}

	s.auth, err = auth.NewAuth(s.client.auth, s.client.ds, s.disableAuth, s.client.cmdbCli, s.disableWriteOpt)
	if err != nil {
		return err
	}

	return nil
}
