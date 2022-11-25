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

package adaptor

import (
	"fmt"

	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/unknown"
	"hcm/pkg/criteria/enumor"
)

// Adaptor holds all the supported operations by the adaptor.
type Adaptor interface {
	// Vendor returns the according vendor's factory.
	Vendor(vendor enumor.Vendor) types.Factory
}

// NewAdaptor create a adaptor instance.
func NewAdaptor() (Adaptor, error) {
	fm := types.NewFactoryManager()
	if err := fm.RegisterVendor(enumor.TCloud, tcloud.NewTCloud()); err != nil {
		return nil, err
	}

	if err := fm.RegisterVendor(enumor.AWS, nil); err != nil {
		return nil, err
	}

	if err := fm.RegisterVendor(enumor.GCP, nil); err != nil {
		return nil, err
	}

	if err := fm.RegisterVendor(enumor.HuaWei, nil); err != nil {
		return nil, err
	}

	if err := fm.RegisterVendor(enumor.Azure, nil); err != nil {
		return nil, err
	}

	opt := &types.NewProxyManagerOption{
		TCloud: tcloud.NewTCloudProxy(),
		Aws:    nil,
		Gcp:    nil,
		HuaWei: nil,
		Azure:  nil,
	}
	pm, err := types.NewProxyManager(opt)
	if err != nil {
		return nil, fmt.Errorf("new proxy manger failed, err: %v", err)
	}

	ad := &adaptor{
		fm: fm,
		pm: pm,
	}

	return ad, nil
}

type adaptor struct {
	fm *types.FactoryManager
	pm types.Proxy
}

// Vendor returns the according vendor's factory.
func (ad *adaptor) Vendor(v enumor.Vendor) types.Factory {
	if err := v.Validate(); err != nil {
		return &unknown.Unknown{Vendor: v}
	}

	exist, f := ad.fm.Vendor(v)
	if !exist {
		return &unknown.Unknown{Vendor: v}
	}

	return f
}

// Proxy returns the cloud vendor proxy.
func (ad *adaptor) Proxy() types.Proxy {
	return ad.pm
}
