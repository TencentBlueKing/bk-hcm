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

// Package cert ...
package cert

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/rest"
)

// BaseCert define base cert.
type BaseCert struct {
	ID               string          `json:"id"`
	CloudID          string          `json:"cloud_id"`
	Name             string          `json:"name"`
	Vendor           enumor.Vendor   `json:"vendor"`
	BkBizID          int64           `json:"bk_biz_id"`
	AccountID        string          `json:"account_id"`
	Domain           []*string       `json:"domain"`
	CertType         enumor.CertType `json:"cert_type"`
	EncryptAlgorithm string          `json:"encrypt_algorithm"`
	CloudCreatedTime string          `json:"cloud_created_time"`
	CloudExpiredTime string          `json:"cloud_expired_time"`
	/*
		tcloud: 0：审核中，1：已通过，2：审核失败，3：已过期，4：验证方式为 DNS_AUTO 类型的证书， 已添加DNS记录，5：企业证书，待提交
			6：订单取消中，7：已取消，8：已提交资料， 待上传确认函，9：证书吊销中，10：已吊销，11：重颁发中，12：待上传吊销确认函
			13：免费证书待提交资料状态，14：已退款。
	*/
	CertStatus string `json:"cert_status"`

	Memo           *string `json:"memo"`
	*core.Revision `json:",inline"`
}

// Cert define cert.
type Cert[Ext Extension] struct {
	BaseCert  `json:",inline"`
	Extension *Ext `json:"extension"`
}

// GetID ...
func (cert Cert[T]) GetID() string {
	return cert.BaseCert.ID
}

// GetCloudID ...
func (cert Cert[T]) GetCloudID() string {
	return cert.BaseCert.CloudID
}

// Extension extension.
type Extension interface {
	TCloudCertExtension
}

// CertCreateResp ...
type CertCreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CertCreateResult `json:"data"`
}

// CertCreateResult ...
type CertCreateResult struct {
	ID string `json:"id"`
}
