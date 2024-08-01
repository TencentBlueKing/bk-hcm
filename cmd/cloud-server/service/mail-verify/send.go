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

package mail_verify

import (
	"context"
	"fmt"

	etcd3 "go.etcd.io/etcd/client/v3"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/tools/json"
)

// SendVerifyCode send verify code email
func (svc *mailSvc) SendVerifyCode(cts *rest.Contexts) (interface{}, error) {
	req := new(SendVerifyCodeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 1.判断是否已存在相应的验证码，限制发送
	key := svc.GenKey(req.Mail, string(req.Scenes))
	exists, err := svc.CheckCodeExists(key)
	if err != nil {
		logs.Errorf("check verification code exists failed, key: %s err: %v, rid: %s",
			key, err, cts.Kit.Rid)
		return nil, err
	}
	if exists {
		logs.Infof("previous verification code is still valid, key: %s, rid: %s", key, cts.Kit.Rid)
		return nil, fmt.Errorf("previous verification code is still valid, please try again later")
	}

	// 2.生成随机验证码，格式：6位随机数字
	verifyCode := svc.GenVerifyCode()

	// 3.不同场景生成不同邮件内容
	mail, err := svc.GenMailByScenes(req, verifyCode)
	if err != nil {
		logs.Errorf("generate mail object failed, req: %v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	// 4.异步发送验证码邮件
	go func() {
		err = svc.CmsiClient.SendMail(cts.Kit, mail)
		if err != nil {
			logs.Errorf("send verify code mail failed, err: %v, rid: %s", err, cts.Kit.Rid)
		}
	}()

	// 5.存储验证码到etcd，设置对应的过期时间
	err = svc.StoryVerifyCode(key, verifyCode, req.Scenes)
	if err != nil {
		logs.Errorf("story verification code failed, key: %s, verifyCode: %s, err: %v, rid: %s",
			key, verifyCode, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// GenMailByScenes generate mail by scenes
func (svc *mailSvc) GenMailByScenes(req *SendVerifyCodeReq, verifyCode string) (*cmsi.CmsiMail, error) {
	info := new(SecondAccountApplicationInfo)
	if err := json.Unmarshal(req.Info, info); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := info.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 填充邮件内容
	var title, content string
	switch req.Scenes {
	case VerifyScenesSecAccountApp:
		title = SecAccountAppMailTitle
		content = fmt.Sprintf(SecAccountAppMailTemplate,
			info.Vendor, info.AccountName, req.Mail, verifyCode, SecAccountAppCodeTTL)
	}

	mail := &cmsi.CmsiMail{
		Receiver: req.Mail,
		Title:    title,
		Content:  content,
	}

	return mail, nil
}

// CheckCodeExists check if there is a verification code
func (svc *mailSvc) CheckCodeExists(key string) (bool, error) {
	resp, err := svc.EtcdClient.Get(context.Background(), key)
	if err != nil {
		logs.Errorf("call etcd get failed, err: %v", err)
		return false, err
	}

	return len(resp.Kvs) != 0, nil
}

// StoryVerifyCode story verification code
func (svc *mailSvc) StoryVerifyCode(key, verifyCode string, scenes Scenes) error {
	var ttl int64
	switch scenes {
	case VerifyScenesSecAccountApp:
		ttl = SecAccountAppCodeTTL*60 + 20
	}

	lease := etcd3.NewLease(svc.EtcdClient)
	leaseResp, err := lease.Grant(context.Background(), ttl)
	if err != nil {
		logs.Errorf("grant lease failed, err: %v", err)
		return err
	}

	_, err = svc.EtcdClient.Put(context.Background(), key, verifyCode, etcd3.WithLease(leaseResp.ID))
	if err != nil {
		logs.Errorf("put kv with lease failed, key: %s, value: %s, err: %v", key, verifyCode, err)
		return err
	}

	return nil
}

// GenKey generate etcd key
func (svc *mailSvc) GenKey(mail, scenes string) string {
	return fmt.Sprintf(VerificationCodeKeyTemplate, mail, scenes)
}

// GenVerifyCode generate a random verification code
func (svc *mailSvc) GenVerifyCode() string {
	code := svc.RandX.Intn(1000000)
	verifyCode := fmt.Sprintf("%06d", code)

	return verifyCode
}
