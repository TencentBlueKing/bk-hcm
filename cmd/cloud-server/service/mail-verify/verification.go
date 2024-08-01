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

	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Verification verification code
func (svc *mailSvc) Verification(cts *rest.Contexts) (interface{}, error) {
	req := new(VerificationReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// etcd取出验证码，判断
	key := svc.GenKey(req.Mail, string(req.Scenes))
	VerifyCode, err := svc.GetVerifyCode(key)
	if err != nil {
		logs.Errorf("get verification code failed, key: %s err: %v, rid: %s", key, err, cts.Kit.Rid)
		return nil, err
	}

	if len(VerifyCode) == 0 || req.VerifyCode != VerifyCode {
		logs.Infof("verification failed, mail: %s, scenes: %s", req.Mail, req.Scenes)
		return false, nil
	}

	// 验证通过，异步删除本次的验证码
	go func() {
		_, err = svc.EtcdClient.Delete(context.Background(), key)
		if err != nil {
			logs.Errorf("delete verification failed, key: %s, err: %v, rid: %s", key, err, cts.Kit.Rid)
		}
	}()

	return true, nil
}

// GetVerifyCode get verification code
func (svc *mailSvc) GetVerifyCode(key string) (string, error) {
	resp, err := svc.EtcdClient.Get(context.Background(), key)
	if err != nil {
		logs.Errorf("call etcd get failed, err: %v", err)
		return "", err
	}

	if len(resp.Kvs) == 0 {
		// 没有发送验证码，或已经过期了
		return "", nil
	}

	VerifyCode := string(resp.Kvs[0].Value)
	return VerifyCode, nil
}
