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
	"encoding/json"
	"fmt"
	"regexp"

	"hcm/pkg/criteria/validator"
)

// SendVerifyCodeReq send verify code req
type SendVerifyCodeReq struct {
	Mail  string          `json:"mail" validate:"required"`
	Scene Scene           `json:"scene" validate:"required,max=64"`
	Info  json.RawMessage `json:"info"`
}

// Validate validate req
func (req *SendVerifyCodeReq) Validate() error {
	if _, ok := supportScene[req.Scene]; !ok {
		return fmt.Errorf("unsupported secne: %s", req.Scene)
	}
	return validator.Validate.Struct(req)
}

// SecondAccountApplicationInfo second account application info
type SecondAccountApplicationInfo struct {
	Vendor      string `json:"vendor" validate:"required"`
	AccountName string `json:"account_name" validate:"required"`
}

// Validate validate req
func (info *SecondAccountApplicationInfo) Validate() error {
	return validator.Validate.Struct(info)
}

// VerificationReq verification code req
type VerificationReq struct {
	Mail              string `json:"mail" validate:"required"`
	Scene             Scene  `json:"scene" validate:"required,max=64"`
	VerifyCode        string `json:"verify_code" validate:"required"`
	DeleteAfterVerify bool   `json:"delete_after_verify"`
}

// Validate validate req
func (req *VerificationReq) Validate() error {
	if _, ok := supportScene[req.Scene]; !ok {
		return fmt.Errorf("unsupported secne: %s", req.Scene)
	}
	re := regexp.MustCompile(`^\d{6}$`)
	if !re.MatchString(req.VerifyCode) {
		return fmt.Errorf("verification code must be 6 digits, but: %s", req.VerifyCode)
	}

	return validator.Validate.Struct(req)
}

const (
	// VerificationCodeKeyTemplate verification code key template
	VerificationCodeKeyTemplate string = "verification-code-%s-%s"
)

// Scene verify type
type Scene string

const (
	// VerifySceneSecAccountApp secondary account application verification
	VerifySceneSecAccountApp Scene = "SecondAccountApplication"
)

var supportScene = map[Scene]struct{}{
	VerifySceneSecAccountApp: {},
}

const (
	// SendVerifyCodeMailLimitInterval sen verification code mail limit interval
	SendVerifyCodeMailLimitInterval = 1 * 60
	// SecAccountAppCodeTTL second account application verification code ttl
	SecAccountAppCodeTTL = 10 * 60
	// SecAccountAppMailTitle second account application verification mail title
	SecAccountAppMailTitle = "HCM二级账号申请邮箱验证码"
	// SecAccountAppMailTemplate second account application verification mail template
	SecAccountAppMailTemplate = `<!DOCTYPE html>
		<html lang="en">
		  <head>
			<meta charset="UTF-8" />
			<meta name="viewport" content="width=device-width, initial-scale=1.0" />
			<meta http-equiv="X-UA-Compatible" content="ie=edge" />
			<title>HCM二级账号申请邮箱验证码</title>
			<style>
				.title {
				  margin-bottom: 2em;
				}
				.content {
				  padding: 0 2em;
				}
				.content-body {
				  margin: 2em 0;
				}
			</style>
		  </head>
		  <body>
			<article>
			  <div class="content">
				<p class="content-header">尊敬的用户您好！</p>
				<div class="content-body">
				  <p>您正在申请 %s 云的二级账号</p>
				  <p><strong>云账号名称</strong>： %s </p>
				  <p><strong>云账号邮箱</strong>： %s </p>
				  <p><strong>本次验证码</strong>： %s </p>
				  <p><strong>验证码有效时间</strong>：%d分钟</p>
				</div>
				<p class="content-footer">请勿向他人提供验证码，使用专用邮箱，账号申请后邮箱不可更改。</p>
			  </div>
			</article>
		  </body>
		</html>
		`
)
