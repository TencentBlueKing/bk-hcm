/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package thirdparty ...
package thirdparty

import (
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/bkbotapproval"
	"hcm/pkg/thirdparty/api-gateway/bkchatapi"
	"hcm/pkg/thirdparty/api-gateway/bkdbm"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"
	"hcm/pkg/thirdparty/api-gateway/usermgr"
	"hcm/pkg/thirdparty/caiche"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/dvmapi"
	"hcm/pkg/thirdparty/erpapi"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/l5api"
	"hcm/pkg/thirdparty/ngateapi"
	"hcm/pkg/thirdparty/safetyapi"
	"hcm/pkg/thirdparty/tcaplusapi"
	"hcm/pkg/thirdparty/tgwapi"
	"hcm/pkg/thirdparty/tjjapi"
	"hcm/pkg/thirdparty/tmpapi"
	"hcm/pkg/thirdparty/xrayapi"
	"hcm/pkg/thirdparty/xshipapi"

	"github.com/prometheus/client_golang/prometheus"
)

// Client third party client
type Client struct {
	CVM             cvmapi.CVMClientInterface
	OldCVM          cvmapi.CVMClientInterface
	DVM             dvmapi.DVMClientInterface
	Tjj             tjjapi.TjjClientInterface
	Xship           xshipapi.XshipClientInterface
	Erp             erpapi.ErpClientInterface
	Tmp             tmpapi.TMPClientInterface
	Xray            xrayapi.XrayClientInterface
	Tcaplus         tcaplusapi.TcaplusClientInterface
	TGW             tgwapi.TgwClientInterface
	L5              l5api.L5ClientInterface
	Safety          safetyapi.SafetyClientInterface
	TencentCloudOpt cc.TCloudCli
	BkChat          bkchatapi.BkChatClientInterface
	Sops            sopsapi.SopsClientInterface
	ITSM            itsm.Client
	Ngate           ngateapi.NgateClientInterface
	CaiChe          caiche.CaiCheClientInterface
	UserMgr         usermgr.Client
	BkDbm           bkdbm.Client
	BkBotApproval   bkbotapproval.Client
	Esb             esb.Client
}

// NewClient new third party client
func NewClient(opts cc.ClientConfig, reg prometheus.Registerer) (*Client, error) {
	// 实例化非蓝鲸第三方服务client
	client, err := newNoBKThirdClient(opts, reg)
	if err != nil {
		return nil, err
	}

	// 实例化API网关client
	client, err = newApiGWClient(opts, reg, client)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// newNoBKThirdClient 实例化第三方服务client
func newNoBKThirdClient(opts cc.ClientConfig, reg prometheus.Registerer) (*Client, error) {
	cvmConf := cvmapi.CVMCli{CvmApiAddr: opts.CvmOpt.CvmApiAddr, CvmLaunchPassword: opts.CvmOpt.CvmLaunchPassword}
	cvm, err := cvmapi.NewCVMClientInterface(cvmConf, reg)
	if err != nil {
		logs.Errorf("failed to new cvm api client, conf: %v, err: %v", cvmConf, err)
		return nil, err
	}

	oldCvmConf := cvmapi.CVMCli{CvmApiAddr: opts.CvmOpt.CvmOldApiAddr,
		CvmLaunchPassword: opts.CvmOpt.CvmLaunchPassword}
	oldCvm, err := cvmapi.NewCVMClientInterface(oldCvmConf, reg)
	if err != nil {
		logs.Errorf("failed to new cvm api client, conf: %v, err: %v", oldCvmConf, err)
		return nil, err
	}

	dvm, err := dvmapi.NewDVMClientInterface(opts.DvmOpt, reg)
	if err != nil {
		logs.Errorf("failed to new dvm api client, err: %v", err)
		return nil, err
	}

	tjj, err := tjjapi.NewTjjClientInterface(opts.TjjOpt, reg)
	if err != nil {
		logs.Errorf("failed to new tjj api client, err: %v", err)
		return nil, err
	}

	xship, err := xshipapi.NewXshipClientInterface(opts.XshipOpt, reg)
	if err != nil {
		logs.Errorf("failed to new xship api client, err: %v", err)
		return nil, err
	}

	erp, err := erpapi.NewErpClientInterface(opts.ErpOpt, reg)
	if err != nil {
		logs.Errorf("failed to new erp api client, err: %v", err)
		return nil, err
	}

	tmp, err := tmpapi.NewTMPClientInterface(opts.TmpOpt, reg)
	if err != nil {
		logs.Errorf("failed to new tmp api client, err: %v", err)
		return nil, err
	}

	xray, err := xrayapi.NewXrayClientInterface(opts.Xray, reg)
	if err != nil {
		logs.Errorf("failed to new xray api client, err: %v", err)
		return nil, err
	}

	tcaplus, err := tcaplusapi.NewTcaplusClientInterface(opts.Tcaplus, reg)
	if err != nil {
		logs.Errorf("failed to new tcaplus api client, err: %v", err)
		return nil, err
	}

	tgw, err := tgwapi.NewTgwClientInterface(opts.TGW, reg)
	if err != nil {
		logs.Errorf("failed to new tgw api client, err: %v", err)
		return nil, err
	}

	l5, err := l5api.NewL5ClientInterface(opts.L5, reg)
	if err != nil {
		logs.Errorf("failed to new l5 api client, err: %v", err)
		return nil, err
	}

	safety, err := safetyapi.NewSafetyClientInterface(opts.Safety, reg)
	if err != nil {
		logs.Errorf("failed to new safety api client, err: %v", err)
		return nil, err
	}

	ngateCli, err := ngateapi.NewNgateClientInterface(opts.Ngate, reg)
	if err != nil {
		logs.Errorf("failed to new ngate api client, err: %v", err)
		return nil, err
	}

	caiCheCli, err := caiche.NewCaiCheClientInterface(opts.CaiChe, reg)
	if err != nil {
		logs.Errorf("failed to new caiche api client, err: %v", err)
		return nil, err
	}

	client := &Client{
		CVM:             cvm,
		OldCVM:          oldCvm,
		DVM:             dvm,
		Tjj:             tjj,
		Xship:           xship,
		TencentCloudOpt: opts.TCloudOpt,
		Erp:             erp,
		Tmp:             tmp,
		Xray:            xray,
		Tcaplus:         tcaplus,
		TGW:             tgw,
		L5:              l5,
		Safety:          safety,
		Ngate:           ngateCli,
		CaiChe:          caiCheCli,
	}
	return client, nil
}

// newApiGWClient 实例化API网关client
func newApiGWClient(opts cc.ClientConfig, reg prometheus.Registerer, client *Client) (*Client, error) {
	sops, err := sopsapi.NewSopsClientInterface(opts.Sops, reg)
	if err != nil {
		logs.Errorf("failed to new sops api client, err: %v", err)
		return nil, err
	}
	client.Sops = sops

	bkchat, err := bkchatapi.NewBkChatClientInterface(opts.BkChat, reg)
	if err != nil {
		logs.Errorf("failed to new bkchat api client, err: %v", err)
		return nil, err
	}
	client.BkChat = bkchat

	itsmCli, err := itsm.NewClient(&opts.ITSM, reg)
	if err != nil {
		logs.Errorf("failed to new itsm api client, err: %v", err)
		return nil, err
	}
	client.ITSM = itsmCli

	dbmCli, err := bkdbm.NewClient(&opts.BkDbm, reg)
	if err != nil {
		logs.Errorf("create bkdbm client failed, err: %v", err)
		return nil, err
	}
	client.BkDbm = dbmCli

	bkBotApprovalCli, err := bkbotapproval.NewClient(&opts.BkBotApproval, reg)
	if err != nil {
		logs.Errorf("create bkbotapproval client failed, err: %v", err)
		return nil, err
	}
	client.BkBotApproval = bkBotApprovalCli

	client.Esb = esb.EsbClient()

	return client, nil
}
