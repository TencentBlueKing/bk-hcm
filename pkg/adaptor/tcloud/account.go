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
	"errors"
	"fmt"

	"hcm/pkg/adaptor/types"
	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// AccountCheck check account authentication information and permissions.
// reference: https://cloud.tencent.com/document/api/598/70416
func (t *TCloud) AccountCheck(kt *kit.Kit, opt *types.TCloudAccountInfo) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "account check option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	camClient, err := t.clientSet.camServiceClient("")
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewGetUserAppIdRequest()
	resp, err := camClient.GetUserAppIdWithContext(kt.Ctx, req)
	if err != nil {
		return fmt.Errorf("get user app id failed, err: %v", err)
	}

	if resp.Response.Uin == nil {
		return errors.New("user uin is empty")
	}

	if resp.Response.OwnerUin == nil {
		return errors.New("user owner uin is empty")
	}

	// check if cloud account info matches the hcm account detail.
	if *resp.Response.Uin != opt.CloudSubAccountID {
		return fmt.Errorf("account id does not match the account to which the secret belongs")
	}

	if *resp.Response.OwnerUin != opt.CloudMainAccountID {
		return fmt.Errorf("main account id does not match the account to which the secret belongs")
	}

	return nil
}

// GetAccountZoneQuota 获取账号配额信息.
// reference: https://cloud.tencent.com/document/api/213/55628
func (t *TCloud) GetAccountZoneQuota(kt *kit.Kit, opt *typeaccount.GetTCloudAccountZoneQuotaOption) (
	*typeaccount.TCloudAccountQuota, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "account check option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := t.clientSet.cvmClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}

	req := cvm.NewDescribeAccountQuotaRequest()
	req.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("zone"),
			Values: []*string{&opt.Zone},
		},
	}

	resp, err := client.DescribeAccountQuotaWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud account quota failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if err = validateDescribeAccountQuotaResp(resp); err != nil {
		return nil, err
	}

	result := new(typeaccount.TCloudAccountQuota)
	if len(resp.Response.AccountQuotaOverview.AccountQuota.PostPaidQuotaSet) == 1 {
		quota := resp.Response.AccountQuotaOverview.AccountQuota.PostPaidQuotaSet[0]
		result.PostPaidQuotaSet = &typeaccount.TCloudPostPaidQuota{
			UsedQuota:      quota.UsedQuota,
			RemainingQuota: quota.RemainingQuota,
			TotalQuota:     quota.TotalQuota,
		}
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.PrePaidQuotaSet) == 1 {
		quota := resp.Response.AccountQuotaOverview.AccountQuota.PrePaidQuotaSet[0]
		result.PrePaidQuota = &typeaccount.TCloudPrePaidQuota{
			UsedQuota:      quota.UsedQuota,
			OnceQuota:      quota.OnceQuota,
			RemainingQuota: quota.RemainingQuota,
			TotalQuota:     quota.TotalQuota,
		}
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.SpotPaidQuotaSet) == 1 {
		quota := resp.Response.AccountQuotaOverview.AccountQuota.SpotPaidQuotaSet[0]
		result.SpotPaidQuota = &typeaccount.TCloudSpotPaidQuota{
			UsedQuota:      quota.UsedQuota,
			RemainingQuota: quota.RemainingQuota,
			TotalQuota:     quota.TotalQuota,
		}
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.ImageQuotaSet) == 1 {
		quota := resp.Response.AccountQuotaOverview.AccountQuota.ImageQuotaSet[0]
		result.ImageQuota = &typeaccount.TCloudImageQuota{
			UsedQuota:  quota.UsedQuota,
			TotalQuota: quota.TotalQuota,
		}
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.DisasterRecoverGroupQuotaSet) == 1 {
		quota := resp.Response.AccountQuotaOverview.AccountQuota.DisasterRecoverGroupQuotaSet[0]
		result.DisasterRecoverGroupQuota = &typeaccount.TCloudDisasterRecoverGroupQuota{
			GroupQuota:            quota.GroupQuota,
			CurrentNum:            quota.CurrentNum,
			CvmInHostGroupQuota:   quota.CvmInRackGroupQuota,
			CvmInSwitchGroupQuota: quota.CvmInHostGroupQuota,
			CvmInRackGroupQuota:   quota.CvmInSwitchGroupQuota,
		}
	}

	return result, nil
}

func validateDescribeAccountQuotaResp(resp *cvm.DescribeAccountQuotaResponse) error {
	if resp.Response == nil || resp.Response.AccountQuotaOverview == nil ||
		resp.Response.AccountQuotaOverview.AccountQuota == nil {
		return errors.New("tcloud account quota api return nil response")
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.PostPaidQuotaSet) > 1 {
		return fmt.Errorf("tcloud account quota api return PostPaidQuotaSet > 1")
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.PrePaidQuotaSet) > 1 {
		return fmt.Errorf("tcloud account quota api return PrePaidQuotaSet > 1")
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.SpotPaidQuotaSet) > 1 {
		return fmt.Errorf("tcloud account quota api return SpotPaidQuotaSet > 1")
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.ImageQuotaSet) > 1 {
		return fmt.Errorf("tcloud account quota api return ImageQuotaSet > 1")
	}

	if len(resp.Response.AccountQuotaOverview.AccountQuota.DisasterRecoverGroupQuotaSet) > 1 {
		return fmt.Errorf("tcloud account quota api return DisasterRecoverGroupQuotaSet > 1")
	}
	return nil
}
