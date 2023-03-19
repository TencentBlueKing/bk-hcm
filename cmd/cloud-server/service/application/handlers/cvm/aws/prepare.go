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

package aws

import (
	"fmt"

	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/enumor"
)

// PrepareReq 预处理请求参数，比如敏感数据加密
func (a *ApplicationOfCreateAwsCvm) PrepareReq() error {
	// 密码加密
	encryptedPassword := a.Cipher.EncryptToBase64(a.req.Password)
	a.req.Password = encryptedPassword
	a.req.ConfirmedPassword = encryptedPassword

	return nil
}

// GenerateApplicationContent 获取预处理过的数据，以interface格式
func (a *ApplicationOfCreateAwsCvm) GenerateApplicationContent() interface{} {
	// 需要将Vendor也存储进去
	return &struct {
		*proto.AwsCvmCreateReq `json:",inline"`
		Vendor                 enumor.Vendor `json:"vendor"`
	}{
		AwsCvmCreateReq: a.req,
		Vendor:          a.vendor,
	}
}

// PrepareReqFromContent 预处理请求参数，对于申请内容来着DB，其实入库前是加密了的
func (a *ApplicationOfCreateAwsCvm) PrepareReqFromContent() error {
	// 解密密码
	password, err := a.Cipher.DecryptFromBase64(a.req.Password)
	if err != nil {
		return fmt.Errorf("decrypt password failed, err: %w", err)
	}
	a.req.Password = password
	a.req.ConfirmedPassword = password

	return nil
}
