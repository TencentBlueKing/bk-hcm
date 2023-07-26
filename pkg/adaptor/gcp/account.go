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

package gcp

import (
	"fmt"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"google.golang.org/api/iam/v1"
)

// AccountCheck check account authentication information and permissions.
func (g *Gcp) AccountCheck(kt *kit.Kit) error {
	// 通过调用获取项目信息接口来验证账号有效性(账号需要有 compute.projects.get 权限)
	if _, err := g.getProject(kt); err != nil {
		return err
	}

	return nil
}

// GetProjectRegionQuota 获取项目地域配额
// reference: https://cloud.google.com/compute/docs/reference/rest/v1/regions/get
func (g *Gcp) GetProjectRegionQuota(kt *kit.Kit, opt *typeaccount.GcpProjectRegionQuotaOption) (
	*typeaccount.GcpProjectQuota, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	resp, err := client.Regions.Get(g.CloudProjectID(), opt.Region).Do()
	if err != nil {
		logs.Errorf("get gcp region failed, err: %v, region: %s, rid: %s", err, opt.Region, kt.Rid)
		return nil, err
	}

	for _, quota := range resp.Quotas {
		if quota.Metric == "INSTANCES" {
			return &typeaccount.GcpProjectQuota{
				Instance: &typeaccount.GcpResourceQuota{
					Limit: quota.Limit,
					Usage: quota.Usage,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("query project region: %s quota not match data", opt.Region)
}

// ListServiceAccounts list service accounts
// reference: https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/list
func (g *Gcp) ListServiceAccounts(kt *kit.Kit) ([]*iam.ServiceAccount, error) {
	client, err := g.clientSet.iamClient(kt)
	if err != nil {
		return nil, err
	}

	name := "projects/" + g.CloudProjectID()
	ret := make([]*iam.ServiceAccount, 0)

	req := client.Projects.ServiceAccounts.List(name)
	if err := req.Pages(kt.Ctx, func(page *iam.ListServiceAccountsResponse) error {
		ret = append(ret, page.Accounts...)
		return nil
	}); err != nil {
		return nil, err
	}

	return ret, nil
}
