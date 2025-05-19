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

package enumor

import (
	"errors"
)

// Vendor defines the cloud type where the hybrid cloud service is supported.
type Vendor string

// Validate the vendor is valid or not
func (v Vendor) Validate() error {
	if _, ok := vendorInfoMap[v]; !ok {
		return errors.New("unsupported cloud vendor: " + string(v))
	}
	return nil
}

const (
	// TCloud is tencent cloud
	TCloud Vendor = "tcloud"
	// Aws is amazon cloud
	Aws Vendor = "aws"
	// Gcp is the Google Cloud Platform
	Gcp Vendor = "gcp"
	// Azure is microsoft azure cloud.
	Azure Vendor = "azure"
	// HuaWei is hua wei cloud.
	HuaWei Vendor = "huawei"
	// Zenlayer is zenlayer cloud.
	Zenlayer Vendor = "zenlayer"
	// Kaopu is kaopu cloud.
	Kaopu Vendor = "kaopu"
	// Other 其他云厂商
	Other Vendor = "other"
)

// VendorInfo 厂商信息
type VendorInfo struct {
	NameEn string
	// 中文名
	NameZh string
	// main account id for duplication check. 云厂商对应的主账号属性名，用户主账号冲突检测
	MainAccountIDField string
	// secret key field, use to remove sensitive info
	SecretKeyField string
}

var (
	vendorInfoMap = map[Vendor]VendorInfo{
		TCloud: {
			NameEn:             "Tencent Cloud",
			NameZh:             "腾讯云",
			MainAccountIDField: "cloud_main_account_id",
			SecretKeyField:     "cloud_secret_key",
		},
		Aws: {
			NameEn:             "Amazon Web Services",
			NameZh:             "亚马逊云",
			MainAccountIDField: "cloud_account_id",
			SecretKeyField:     "cloud_secret_key",
		},
		HuaWei: {
			NameEn:             "Huawei Cloud",
			NameZh:             "华为云",
			MainAccountIDField: "cloud_sub_account_id",
			SecretKeyField:     "cloud_secret_key",
		},
		Gcp: {
			NameEn:             "Google Cloud",
			NameZh:             "谷歌云",
			MainAccountIDField: "cloud_project_id",
			SecretKeyField:     "cloud_service_secret_key",
		},
		Azure: {
			NameEn:             "Microsoft Azure",
			NameZh:             "微软云",
			MainAccountIDField: "cloud_subscription_id",
			SecretKeyField:     "cloud_client_secret_key",
		},
		Zenlayer: {
			NameEn:             "Zenlayer",
			NameZh:             "Zenlayer",
			MainAccountIDField: "cloud_main_account_id",
			SecretKeyField:     "cloud_secret_key",
		},
		Kaopu: {
			NameEn:             "Kaopu",
			NameZh:             "靠谱云",
			MainAccountIDField: "cloud_main_account_id",
			SecretKeyField:     "cloud_secret_key",
		},
		Other: {
			NameEn: "Other",
			NameZh: "其他云厂商",
		},
	}
)

// RegisterVendor 注册支持的厂商
func RegisterVendor(vendor Vendor, info VendorInfo) {
	if _, ok := vendorInfoMap[vendor]; ok {
		panic("vendor already registered")
	}
	vendorInfoMap[vendor] = info
}

// GetVendorInfo 获取厂商信息
func GetVendorInfo(vendor Vendor) VendorInfo {
	return vendorInfoMap[vendor]
}

// GetNameZh 返回中文名
func (v Vendor) GetNameZh() string {
	return vendorInfoMap[v].NameZh
}

// GetMainAccountIDField  返回云厂商对应的主账号字段名
func (v Vendor) GetMainAccountIDField() string {
	return vendorInfoMap[v].MainAccountIDField
}

// GetSecretField  返回云厂商对应的秘钥字段名
func (v Vendor) GetSecretField() string {
	return vendorInfoMap[v].SecretKeyField
}

// GetMainAccountIDFields 返回全部主账号字段
func GetMainAccountIDFields() []string {
	fields := make([]string, 0, len(vendorInfoMap))
	for _, info := range vendorInfoMap {
		fields = append(fields, info.MainAccountIDField)
	}
	return fields
}
