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
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"hcm/pkg/adaptor/types"
)

type clientSet struct {
	profile *profile.ClientProfile
}

func newClientSet(profile *profile.ClientProfile) *clientSet {
	return &clientSet{
		profile: profile,
	}
}

func (c *clientSet) camServiceClient(secret *types.BaseSecret, region string) (*cam.Client, error) {
	credential := common.NewCredential(secret.CloudSecretID, secret.CloudSecretKey)
	client, err := cam.NewClient(credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) cvmClient(secret *types.BaseSecret, region string) (*cvm.Client, error) {
	credential := common.NewCredential(secret.CloudSecretKey, secret.CloudSecretID)
	client, err := cvm.NewClient(credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}
