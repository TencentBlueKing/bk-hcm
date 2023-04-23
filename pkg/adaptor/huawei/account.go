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
	"errors"

	"hcm/pkg/adaptor/types"
	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
)

// AccountCheck check account authentication information and permissions.
// KeystoneListAuthDomains: https://support.huaweicloud.com/intl/zh-cn/api-iam/iam_07_0001.html
// 账号概念 https://support.huaweicloud.com/intl/zh-cn/api-iam/iam_17_0002.html
func (h *HuaWei) AccountCheck(kt *kit.Kit, opt *types.HuaWeiAccountInfo) error {
	client, err := h.clientSet.iamClient(region.AP_SOUTHEAST_1)
	if err != nil {
		logs.Errorf("init huawei client failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	domainsResp, err := client.KeystoneListAuthDomains(nil)
	if err != nil {
		logs.Errorf("KeystoneListAuthDomains failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	domains := domainsResp.Domains
	if domains == nil {
		return errors.New("KeystoneListAuthDomains failed, err: no auth domains")
	}

	for _, domain := range *domains {
		if domain.Id == opt.CloudSubAccountID && domain.Name == opt.CloudSubAccountName {
			return nil
		}
	}

	return errors.New("SubAccount does not match SecretId/SecretKey")
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
