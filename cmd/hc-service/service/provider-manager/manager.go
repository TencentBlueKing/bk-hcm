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

package providermgr

import (
	"fmt"
	"sync"

	"hcm/pkg/ad/provider"
	"hcm/pkg/ad/tcloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// NewManager new manager.
func NewManager(dataCli *dataservice.Client) Manager {
	return &manager{
		dvLock:        new(sync.Mutex),
		disableVendor: make(map[enumor.Vendor]struct{}, 0),
		secretCli:     NewSecretClient(dataCli),
		dsCli:         dataCli,
	}
}

// Manager provider manger support interface.
type Manager interface {
	Provider(kt *kit.Kit, accountID string) (provider.Provider, error)
}

var _ Manager = new(manager)

// manager define manager.
type manager struct {
	dvLock        *sync.Mutex
	disableVendor map[enumor.Vendor]struct{}
	secretCli     *SecretClient
	dsCli         *dataservice.Client
}

// Provider get account vendor's provider.
func (mgr *manager) Provider(kt *kit.Kit, accountID string) (provider.Provider, error) {

	info, err := mgr.dsCli.Global.Cloud.GetResourceBasicInfo(kt.Ctx, kt.Header(), enumor.AccountCloudResType, accountID)
	if err != nil {
		logs.Errorf("get resource basic info failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if _, exist := mgr.disableVendor[info.Vendor]; exist {
		return nil, fmt.Errorf("vendor: %s is disable", info.Vendor)
	}

	switch info.Vendor {
	case enumor.TCloud:
		secret, err := mgr.secretCli.TCloudSecret(kt, accountID)
		if err != nil {
			return nil, err
		}

		return tcloud.NewProvider(secret)

	default:
		return nil, fmt.Errorf("provider: %s not support", info.Vendor)
	}
}

// DisableVendor define disable vendor func.
func (mgr *manager) DisableVendor(vendor enumor.Vendor) {
	mgr.dvLock.Lock()
	defer mgr.dvLock.Unlock()

	mgr.disableVendor[vendor] = struct{}{}
	return
}

// EnableVendor define enable vendor func.
func (mgr *manager) EnableVendor(vendor enumor.Vendor) {
	mgr.dvLock.Lock()
	defer mgr.dvLock.Unlock()

	delete(mgr.disableVendor, vendor)
	return
}
