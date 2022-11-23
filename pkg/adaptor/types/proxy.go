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

package types

import "errors"

// TCloudProxy holds all the tencent cloud proxy operation on the cloud resources.
type TCloudProxy interface {
}

// AwsProxy holds all the Aws cloud proxy operation on the cloud resources.
type AwsProxy interface {
}

// GcpProxy holds all the Gcp cloud proxy operation on the cloud resources.
type GcpProxy interface {
}

// HuaWeiProxy holds all the HuaWei cloud proxy operation on the cloud resources.
type HuaWeiProxy interface {
}

// AzureProxy holds all the Azure cloud proxy operation on the cloud resources.
type AzureProxy interface {
}

// Proxy holds all the proxy operation on the cloud resources.
type Proxy interface {
	TCloud() TCloudProxy
	Aws() AwsProxy
	Gcp() GcpProxy
	HuaWei() HuaWeiProxy
	Azure() AzureProxy
}

// NewProxyManagerOption new proxy manager option.
type NewProxyManagerOption struct {
	TCloud TCloudProxy
	Aws    AwsProxy
	Gcp    GcpProxy
	HuaWei HuaWeiProxy
	Azure  AzureProxy
}

// NewProxyManager new cloud proxy manager.
func NewProxyManager(opt *NewProxyManagerOption) (*ProxyManager, error) {
	if opt == nil {
		return nil, errors.New("opt is nil")
	}

	if opt.TCloud == nil {
		return nil, errors.New("option.TCloud is required")
	}

	// if opt.Aws == nil {
	// 	return nil, errors.New("option.Aws is required")
	// }
	//
	// if opt.Gcp == nil {
	// 	return nil, errors.New("option.Gcp is required")
	// }
	//
	// if opt.HuaWei == nil {
	// 	return nil, errors.New("option.HuaWei is required")
	// }
	//
	// if opt.Azure == nil {
	// 	return nil, errors.New("option.Azure is required")
	// }

	pm := &ProxyManager{
		tcloud: opt.TCloud,
		aws:    opt.Aws,
		gcp:    opt.Gcp,
		huawei: opt.HuaWei,
		azure:  opt.Azure,
	}

	return pm, nil
}

// ProxyManager manage all cloud proxy.
type ProxyManager struct {
	tcloud TCloudProxy
	aws    AwsProxy
	gcp    GcpProxy
	huawei HuaWeiProxy
	azure  AzureProxy
}

// TCloud returns tencent cloud proxy
func (p *ProxyManager) TCloud() TCloudProxy {
	return p.tcloud
}

// Aws returns Aws proxy
func (p *ProxyManager) Aws() AwsProxy {
	return p.aws
}

// Gcp returns Gcp proxy
func (p *ProxyManager) Gcp() GcpProxy {
	return p.gcp
}

// HuaWei returns HuaWei proxy
func (p *ProxyManager) HuaWei() HuaWeiProxy {
	return p.huawei
}

// Azure returns Azure proxy
func (p *ProxyManager) Azure() AzureProxy {
	return p.azure
}
