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

package mailverify

import (
	etcd3 "go.etcd.io/etcd/client/v3"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
)

// MVSvc maile verification server
var MVSvc *MailVerifySvc

// InitEmailService initial the email service
func InitEmailService(c *capability.Capability) {
	config := cc.CloudServer().Service
	etcdCfg, err := config.Etcd.ToConfig()
	if err != nil {
		return
	}
	etcdCli, err := etcd3.New(etcdCfg)
	if err != nil {
		return
	}

	svc := &MailVerifySvc{
		CmsiClient: c.CmsiCli,
		EtcdClient: etcdCli,
	}
	MVSvc = svc

	h := rest.NewHandler()
	// 由于邮箱验证码需要申请权限，暂时关闭邮箱验证码功能
	//h.Add("SendVerifyCode", http.MethodPost, "/mail/send_code", svc.SendVerifyCode)
	//h.Add("VerificationCode", http.MethodPost, "/mail/verify_code", svc.Verification)
	h.Load(c.WebService)
}

// MailVerifySvc mail verification service
type MailVerifySvc struct {
	EtcdClient *etcd3.Client
	CmsiClient cmsi.Client
}
