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

package huawei

import (
	"fmt"

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
)

// ListAccount list account.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-iam/iam_08_0001.html
func (h *HuaWei) ListAccount(kt *kit.Kit) ([]typeaccount.HuaWeiAccount, error) {
	client, err := h.clientSet.iamGlobalClient(region.AP_SOUTHEAST_1)
	if err != nil {
		logs.Errorf("new iam client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	req := new(model.KeystoneListUsersRequest)
	resp, err := client.KeystoneListUsers(req)
	if err != nil {
		logs.Errorf("keystone list users failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("keystone list users failed, err: %v", err)
	}

	list := make([]typeaccount.HuaWeiAccount, 0)
	if resp.Users != nil {
		for _, one := range *resp.Users {
			list = append(list, typeaccount.HuaWeiAccount{
				PwdStatus:         one.PwdStatus,
				DomainID:          one.DomainId,
				LastProjectID:     one.LastProjectId,
				Name:              one.Name,
				Description:       one.Description,
				PasswordExpiresAt: one.PasswordExpiresAt,
				ID:                one.Id,
				Enabled:           one.Enabled,
				PwdStrength:       one.PwdStrength,
			})
		}
	}

	return list, nil
}

// GetAccountQuota get account quota.
// KeystoneListAuthDomains: https://support.huaweicloud.com/intl/zh-cn/api-ecs/ecs_02_0801.html
func (h *HuaWei) GetAccountQuota(kt *kit.Kit, opt *typeaccount.GetHuaWeiAccountZoneQuotaOption) (
	*typeaccount.HuaWeiAccountQuota, error) {

	client, err := h.clientSet.ecsClient(opt.Region)
	if err != nil {
		logs.Errorf("init huawei client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, err := client.ShowServerLimits(nil)
	if err != nil {
		logs.Errorf("show huawei server limit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	quota := &typeaccount.HuaWeiAccountQuota{
		MaxImageMeta:          resp.Absolute.MaxImageMeta,
		MaxPersonality:        resp.Absolute.MaxPersonality,
		MaxPersonalitySize:    resp.Absolute.MaxPersonalitySize,
		MaxSecurityGroupRules: resp.Absolute.MaxSecurityGroupRules,
		MaxSecurityGroups:     resp.Absolute.MaxSecurityGroups,
		MaxServerGroupMembers: resp.Absolute.MaxServerGroupMembers,
		MaxServerGroups:       resp.Absolute.MaxServerGroups,
		MaxServerMeta:         resp.Absolute.MaxServerMeta,
		MaxTotalCores:         resp.Absolute.MaxTotalCores,
		MaxTotalFloatingIps:   resp.Absolute.MaxTotalFloatingIps,
		MaxTotalInstances:     resp.Absolute.MaxTotalInstances,
		MaxTotalKeypairs:      resp.Absolute.MaxTotalKeypairs,
		MaxTotalRAMSize:       resp.Absolute.MaxTotalRAMSize,
		MaxTotalSpotInstances: resp.Absolute.MaxTotalSpotInstances,
		MaxTotalSpotCores:     resp.Absolute.MaxTotalSpotCores,
		MaxTotalSpotRAMSize:   resp.Absolute.MaxTotalSpotRAMSize,
	}
	return quota, nil
}

// GetAccountInfoBySecret 根据AccessKey 获取账号信息
// 1. https://console-intl.huaweicloud.com/apiexplorer/#/openapi/IAM/doc?api=ShowPermanentAccessKey
// 2. https://console-intl.huaweicloud.com/apiexplorer/#/openapi/IAM/debug?api=ShowUser
// 3. https://console-intl.huaweicloud.com/apiexplorer/#/openapi/IAM/doc?api=KeystoneListAuthDomains
func (h *HuaWei) GetAccountInfoBySecret(kt *kit.Kit, accessKeyID string) (*cloud.HuaWeiInfoBySecret, error) {

	client, err := h.clientSet.iamGlobalClient(region.AP_SOUTHEAST_1)
	if err != nil {
		logs.Errorf("new iam client failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// 1. 根据access key 获取iam用户id
	// https://console-intl.huaweicloud.com/apiexplorer/#/openapi/IAM/doc?api=ShowPermanentAccessKey
	akResp, err := client.ShowPermanentAccessKey(&model.ShowPermanentAccessKeyRequest{AccessKey: accessKeyID})
	if err != nil {
		logs.Errorf("ShowPermanentAccessKey failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("ShowPermanentAccessKey failed, err: %v", err)

	}
	accountInfo := &cloud.HuaWeiInfoBySecret{
		CloudIamUserID: akResp.Credential.UserId,
	}

	// 2. 根据iam用户id 获取iam用户名称和子账号id
	// https://console-intl.huaweicloud.com/apiexplorer/#/openapi/IAM/debug?api=ShowUser
	userResp, err := client.ShowUser(&model.ShowUserRequest{UserId: accountInfo.CloudIamUserID})
	if err != nil {
		logs.Errorf("ShowUser failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("ShowUser failed, err: %v", err)
	}
	accountInfo.CloudIamUsername = userResp.User.Name
	accountInfo.CloudSubAccountID = userResp.User.DomainId
	// 3. 遍历账号列表，根据子账号id 获取子账号名
	// https://console-intl.huaweicloud.com/apiexplorer/#/openapi/IAM/doc?api=KeystoneListAuthDomains
	domainResp, err := client.KeystoneListAuthDomains(new(model.KeystoneListAuthDomainsRequest))
	if err != nil {
		logs.Errorf("KeystoneListAuthDomainsRequest failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("KeystoneListAuthDomainsRequest failed, err: %v", err)
	}
	// 遍历寻找子账户名
	for _, one := range converter.PtrToVal(domainResp.Domains) {
		if one.Id == accountInfo.CloudSubAccountID {
			accountInfo.CloudSubAccountName = one.Name
			break
		}
	}
	// 没找到对应子账号id
	if len(accountInfo.CloudSubAccountName) == 0 {
		return nil, fmt.Errorf("KeystoneListAuthDomainsRequest not fount domain!, domains: %v", domainResp.Domains)
	}
	return accountInfo, nil

}
