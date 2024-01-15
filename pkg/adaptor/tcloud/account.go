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

	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// ListAccount 查询账号列表.
// reference: https://cloud.tencent.com/document/api/598/34587
func (t *TCloudImpl) ListAccount(kt *kit.Kit) ([]typeaccount.TCloudAccount, error) {

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewListUsersRequest()
	resp, err := camClient.ListUsersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list users failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list users failed, err: %v", err)
	}

	list := make([]typeaccount.TCloudAccount, 0, len(resp.Response.Data))
	for _, one := range resp.Response.Data {
		list = append(list, typeaccount.TCloudAccount{
			Uin:          one.Uin,
			Name:         one.Name,
			Uid:          one.Uid,
			Remark:       one.Remark,
			ConsoleLogin: one.ConsoleLogin,
			PhoneNum:     one.PhoneNum,
			CountryCode:  one.CountryCode,
			Email:        one.Email,
			CreateTime:   one.CreateTime,
			NickName:     one.NickName,
		})
	}

	return list, nil
}

// CountAccount 查询账号数量，基于ListUsersWithContext
// reference: https://cloud.tencent.com/document/api/598/34587
func (t *TCloudImpl) CountAccount(kt *kit.Kit) (int32, error) {

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return 0, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewListUsersRequest()
	resp, err := camClient.ListUsersWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("count users failed, err: %v,  rid: %s", err, kt.Rid)
		return 0, fmt.Errorf("list users failed, err: %v", err)
	}

	return int32(len(resp.Response.Data)), nil
}

// GetAccountZoneQuota 获取账号配额信息.
// reference: https://cloud.tencent.com/document/api/213/55628
func (t *TCloudImpl) GetAccountZoneQuota(kt *kit.Kit, opt *typeaccount.GetTCloudAccountZoneQuotaOption) (
	*typeaccount.TCloudAccountQuota, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "account check option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := t.clientSet.CvmClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud client failed, err: %v", err)
	}

	req := cvm.NewDescribeAccountQuotaRequest()
	req.Filters = []*cvm.Filter{{Name: common.StringPtr("zone"), Values: []*string{&opt.Zone}}}

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

// GetAccountInfoBySecret 根据秘钥获取云上获取账号信息
// reference: https://cloud.tencent.com/document/api/598/70416
func (t *TCloudImpl) GetAccountInfoBySecret(kt *kit.Kit) (*cloud.TCloudInfoBySecret, error) {

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewGetUserAppIdRequest()
	resp, err := camClient.GetUserAppIdWithContext(kt.Ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get user app id failed, err: %v", err)
	}

	if resp.Response.Uin == nil {
		return nil, errors.New("user uin is empty")
	}

	if resp.Response.OwnerUin == nil {
		return nil, errors.New("user owner uin is empty")
	}
	return &cloud.TCloudInfoBySecret{
		CloudSubAccountID:  converter.PtrToVal(resp.Response.Uin),
		CloudMainAccountID: converter.PtrToVal(resp.Response.OwnerUin),
	}, nil
}
