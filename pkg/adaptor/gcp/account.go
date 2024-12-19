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
	"strings"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"cloud.google.com/go/asset/apiv1/assetpb"
)

// CountAccount count account.
// reference: https://cloud.google.com/asset-inventory/docs/reference/rest/v1/TopLevel/analyzeIamPolicy
func (g *Gcp) CountAccount(kt *kit.Kit) (int32, error) {

	client, err := g.clientSet.assetClient(kt)
	if err != nil {
		return 0, fmt.Errorf("new asset client failed, err: %v", err)
	}

	req := &assetpb.AnalyzeIamPolicyRequest{
		// See https://pkg.go.dev/cloud.google.com/go/asset/apiv1/assetpb#AnalyzeIamPolicyRequest.
		AnalysisQuery: &assetpb.IamPolicyAnalysisQuery{
			Scope: fmt.Sprintf("projects/%s", g.CloudProjectID()),
			AccessSelector: &assetpb.IamPolicyAnalysisQuery_AccessSelector{
				Roles: []string{"roles/owner"},
			},
		},
	}
	resp, err := client.AnalyzeIamPolicy(kt.Ctx, req)
	if err != nil {
		logs.Errorf("analyze iam policy failed, err: %v, rid: %s", err, kt.Rid)
		return 0, err
	}

	var count int32
	if resp.MainAnalysis != nil && resp.MainAnalysis.AnalysisResults != nil &&
		len(resp.MainAnalysis.AnalysisResults) != 0 {
		for _, item := range resp.MainAnalysis.AnalysisResults {
			if item.IamBinding != nil && len(item.IamBinding.Members) != 0 {
				for _, member := range item.IamBinding.Members {
					if strings.HasPrefix(member, "user:") {
						count++
					}
				}
			}
		}
	}

	return count, nil
}

// ListAccount list account.
// reference: https://cloud.google.com/asset-inventory/docs/reference/rest/v1/TopLevel/analyzeIamPolicy
func (g *Gcp) ListAccount(kt *kit.Kit) ([]typeaccount.GcpAccount, error) {

	client, err := g.clientSet.assetClient(kt)
	if err != nil {
		return nil, fmt.Errorf("new asset client failed, err: %v", err)
	}

	req := &assetpb.AnalyzeIamPolicyRequest{
		// See https://pkg.go.dev/cloud.google.com/go/asset/apiv1/assetpb#AnalyzeIamPolicyRequest.
		AnalysisQuery: &assetpb.IamPolicyAnalysisQuery{
			Scope:            fmt.Sprintf("projects/%s", g.CloudProjectID()),
			ResourceSelector: nil,
			IdentitySelector: nil,
			AccessSelector: &assetpb.IamPolicyAnalysisQuery_AccessSelector{
				Roles:       []string{"roles/owner"},
				Permissions: nil,
			},
			Options:          nil,
			ConditionContext: nil,
		},
	}
	resp, err := client.AnalyzeIamPolicy(kt.Ctx, req)
	if err != nil {
		logs.Errorf("analyze iam policy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	list := make([]typeaccount.GcpAccount, 0)
	if resp.MainAnalysis != nil && resp.MainAnalysis.AnalysisResults != nil &&
		len(resp.MainAnalysis.AnalysisResults) != 0 {
		for _, item := range resp.MainAnalysis.AnalysisResults {
			if item.IamBinding != nil && len(item.IamBinding.Members) != 0 {
				for _, member := range item.IamBinding.Members {
					if strings.HasPrefix(member, "user:") {
						list = append(list, typeaccount.GcpAccount{
							Name: member[5:],
						})
					}
				}
			}
		}
	}

	return list, nil
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

// GetAccountInfoBySecret 根据秘钥获取账号信息
// reference:
// 1. https://cloud.google.com/resource-manager/reference/rest/v3/projects/search
// 2. https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/get
func (g *Gcp) GetAccountInfoBySecret(kit *kit.Kit, cloudSecretKeyString string) (*cloud.GcpInfoBySecret, error) /**/ {
	client, err := g.clientSet.resClient(kit)
	if err != nil {
		return nil, err
	}

	// 1. 获取该账号可以访问的项目 https://cloud.google.com/resource-manager/reference/rest/v3/projects/search
	projectList, err := client.Projects.Search().Do()
	if err != nil {
		logs.Errorf("search project failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	if len(projectList.Projects) == 0 {
		return nil, fmt.Errorf("not project avaiable, please check the permission of given screct")
	}

	iamClient, err := g.clientSet.iamServiceClient(kit)
	if err != nil {
		return nil, err
	}

	// 2. 根据秘钥信息获取服务账号信息
	// https://cloud.google.com/iam/docs/reference/rest/v1/projects.serviceAccounts/get
	sk, err := account.DecodeGcpSecretKey(cloudSecretKeyString)
	if err != nil {
		return nil, err
	}

	projectInfos := make([]cloud.GcpProjectInfo, 0)
	for _, project := range projectList.Projects {
		serviceAccount, err := iamClient.Projects.ServiceAccounts.Get(
			fmt.Sprintf("projects/%s/serviceAccounts/%s", project.ProjectId, sk.ClientEmail),
		).Do()
		if err != nil {
			return nil, err
		}
		projectInfos = append(projectInfos, cloud.GcpProjectInfo{
			Email:                   serviceAccount.Email,
			CloudProjectID:          project.ProjectId,
			CloudProjectName:        project.DisplayName,
			CloudServiceAccountID:   serviceAccount.UniqueId,
			CloudServiceAccountName: serviceAccount.DisplayName,
			CloudServiceSecretID:    sk.PrivateKeyID,
		})
	}

	accountInfo := &cloud.GcpInfoBySecret{
		CloudProjectInfos: projectInfos,
	}
	return accountInfo, nil
}
