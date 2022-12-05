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

package tcloud

import (
	"hcm/pkg/adaptor/types"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// NewTCloud new tcloud.
func NewTCloud() types.Factory {
	prof := profile.NewClientProfile()
	return &tcloud{profile: prof}
}

// NewTCloudProxy new tencent cloud proxy.
func NewTCloudProxy() types.TCloudProxy {
	prof := profile.NewClientProfile()
	return &tcloud{profile: prof}
}

var (
	_ types.Factory     = new(tcloud)
	_ types.TCloudProxy = new(tcloud)
)

type tcloud struct {
	profile *profile.ClientProfile
}

func (t *tcloud) cvmClient(secret *types.BaseSecret, region string) (*cvm.Client, error) {
	credential := common.NewCredential(secret.ID, secret.Key)
	client, err := cvm.NewClient(credential, region, t.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}
