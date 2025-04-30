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

// Package cmsi ...
package cmsi

import (
	"fmt"
	
	"hcm/pkg/kit"
	"hcm/pkg/rest"
	apigateway "hcm/pkg/thirdparty/api-gateway"
)

// CmsiMailParams ...
type CmsiMailParams struct {
	Receiver         []string             `json:"receiver,omitempty"`
	ReceiverUserName []string             `json:"receiver__username,omitempty"`
	Sender           string               `json:"sender,omitempty"`
	Title            string               `json:"title"`
	Content          string               `json:"content"`
	Cc               []string             `json:"cc,omitempty"`
	CcUserName       []string             `json:"cc__username,omitempty"`
	BodyFormat       string               `json:"body_format,omitempty"`
	IsContentBase64  bool                 `json:"is_content_base64,omitempty"`
	Attachments      []CmsiMailAttachment `json:"attachments,omitempty"`
}

// CmsiMailAttachment ...
type CmsiMailAttachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Type        string `json:"type,omitempty"`
	Disposition string `json:"disposition,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
}

// SendMail ...
func (c *cmsi) SendMail(kt *kit.Kit, param *CmsiMailParams) error {
	// 可以自定义发送人，未自定义则使用配置默认
	if param.Sender == "" {
		param.Sender = c.sender
	}

	// 邮件默认抄送给平台管理员
	if len(param.Cc) == 0 && len(param.CcUserName) == 0 {
		param.Cc = append(param.Cc, c.cc...)
	}

	_, err := apigateway.ApiGatewayCallWithRichError[CmsiMailParams, CmsiMailResult](
		c.client, c.config, rest.POST, kt, param, "/send_mail")

	if err != nil {
		return fmt.Errorf("send mail failed: %s", err)
	}
	return nil
}
