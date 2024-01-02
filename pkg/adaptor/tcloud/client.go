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

	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

const (
	ErrNotFound = "Code=ResourceNotFound"
)

type clientSet struct {
	credential *common.Credential
	profile    *profile.ClientProfile
}

func newClientSet(s *types.BaseSecret, profile *profile.ClientProfile) *clientSet {
	return &clientSet{
		credential: common.NewCredential(s.CloudSecretID, s.CloudSecretKey),
		profile:    profile,
	}
}

func (c *clientSet) camServiceClient(region string) (*cam.Client, error) {
	client, err := cam.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) cvmClient(region string) (*cvm.Client, error) {
	client, err := cvm.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) cbsClient(region string) (*cbs.Client, error) {
	client, err := cbs.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) vpcClient(region string) (*vpc.Client, error) {
	client, err := vpc.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) billClient() (*billing.Client, error) {
	client, err := billing.NewClient(c.credential, "", c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) clbClient(region string) (*clb.Client, error) {
	client, err := clb.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}

	return client, nil
}
