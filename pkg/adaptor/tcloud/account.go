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
	"hcm/pkg/criteria/constant"
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

// DescribeSubAccounts query sub accounts by UIN list.
// reference: https://cloud.tencent.com/document/api/598/53486
func (t *TCloudImpl) DescribeSubAccounts(kt *kit.Kit, opt *typeaccount.DescribeSubAccountsOption) (
	[]typeaccount.TCloudSubAccountUser, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "describe sub accounts option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewDescribeSubAccountsRequest()
	req.FilterSubAccountUin = converter.SliceToPtr(opt.SubUin)

	rangeMS := [2]uint{constant.TCloudRetryDelayMinMS, constant.TCloudRetryDelayMaxMS}
	resp, err := LimitExceededErrRetry(camClient.DescribeSubAccountsWithContext, kt, req,
		constant.TCloudClientErrRetryTimes, rangeMS)
	if err != nil {
		logs.Errorf("describe sub accounts failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("describe sub accounts failed, err: %v", err)
	}

	list := make([]typeaccount.TCloudSubAccountUser, 0, len(resp.Response.SubAccounts))
	for _, one := range resp.Response.SubAccounts {
		list = append(list, typeaccount.TCloudSubAccountUser{
			Uin:           one.Uin,
			Name:          one.Name,
			Uid:           one.Uid,
			Remark:        one.Remark,
			CreateTime:    one.CreateTime,
			UserType:      one.UserType,
			LastLoginIp:   one.LastLoginIp,
			LastLoginTime: one.LastLoginTime,
		})
	}

	return list, nil
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

// AddUser 创建子用户
// reference: https://cloud.tencent.com/document/product/598/34595
func (t *TCloudImpl) AddUser(kt *kit.Kit, opt *typeaccount.AddUserOption) (*typeaccount.AddUserResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "add user option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewAddUserRequest()
	req.Name = converter.ValToPtr(opt.Name)
	req.UseApi = converter.ValToPtr(opt.UseAPI)

	if opt.Remark != "" {
		req.Remark = converter.ValToPtr(opt.Remark)
	}

	req.ConsoleLogin = converter.ValToPtr(opt.ConsoleLogin)

	if opt.Password != "" {
		req.Password = converter.ValToPtr(opt.Password)
	}
	if opt.NeedResetPassword > 0 {
		req.NeedResetPassword = converter.ValToPtr(opt.NeedResetPassword)
	}
	if opt.PhoneNum != "" {
		req.PhoneNum = converter.ValToPtr(opt.PhoneNum)
	}
	if opt.CountryCode != "" {
		req.CountryCode = converter.ValToPtr(opt.CountryCode)
	}
	if opt.Email != "" {
		req.Email = converter.ValToPtr(opt.Email)
	}

	resp, err := camClient.AddUserWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("add user failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("add user failed, err: %v", err)
	}

	return &typeaccount.AddUserResult{
		Uin:       resp.Response.Uin,
		Name:      resp.Response.Name,
		UID:       resp.Response.Uid,
		SecretID:  converter.PtrToVal(resp.Response.SecretId),
		SecretKey: converter.PtrToVal(resp.Response.SecretKey),
		Password:  converter.PtrToVal(resp.Response.Password),
	}, nil
}

// DeleteUser 删除子用户
// reference: https://cloud.tencent.com/document/product/598/34592
func (t *TCloudImpl) DeleteUser(kt *kit.Kit, name string) error {
	if name == "" {
		return errf.New(errf.InvalidParameter, "user name is required")
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewDeleteUserRequest()
	req.Name = converter.ValToPtr(name)

	// Force为0：若该用户存在未删除API密钥，则不删除用户；
	// Force为1：若该用户存在未删除API密钥，则先删除密钥再删除用户。
	req.Force = converter.ValToPtr(uint64(0))

	_, err = camClient.DeleteUserWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete user failed, name: %s, err: %v, rid: %s", name, err, kt.Rid)
		return fmt.Errorf("delete user failed, err: %v", err)
	}

	return nil
}

// UpdateUser 更新子用户
// reference: https://cloud.tencent.com/document/product/598/34583
func (t *TCloudImpl) UpdateUser(kt *kit.Kit, opt *typeaccount.UpdateUserOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "update user option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewUpdateUserRequest()
	req.Name = common.StringPtr(opt.Name)
	req.Remark = opt.Remark
	req.ConsoleLogin = opt.ConsoleLogin
	req.Password = opt.Password
	req.NeedResetPassword = opt.NeedResetPassword
	req.PhoneNum = opt.PhoneNum
	req.CountryCode = opt.CountryCode
	req.Email = opt.Email

	_, err = camClient.UpdateUserWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("update user failed, name: %s, err: %v, rid: %s", opt.Name, err, kt.Rid)
		return fmt.Errorf("update user failed, err: %v", err)
	}

	return nil
}

// DescribeSafeAuthFlagColl get sub-account safe auth flag settings (CAM DescribeSafeAuthFlagColl).
// reference: https://cloud.tencent.com/document/product/598/48602
func (t *TCloudImpl) DescribeSafeAuthFlagColl(kt *kit.Kit, opt *typeaccount.DescribeSafeAuthFlagCollOption) (
	[]typeaccount.SafeAuthFlagCollResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "describe safe auth flag coll option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	results := make([]typeaccount.SafeAuthFlagCollResult, 0, len(opt.SubUins))

	// CAM DescribeSafeAuthFlagColl API only supports querying one SubUins at a time internally
	// Iterate through all SubUins and make individual API calls
	for _, subUin := range opt.SubUins {
		req := cam.NewDescribeSafeAuthFlagCollRequest()
		req.SubUin = converter.ValToPtr(subUin)

		rangeMS := [2]uint{constant.TCloudRetryDelayMinMS, constant.TCloudRetryDelayMaxMS}
		resp, err := LimitExceededErrRetry(camClient.DescribeSafeAuthFlagCollWithContext, kt, req,
			constant.TCloudClientErrRetryTimes, rangeMS)
		if err != nil {
			logs.Errorf("describe safe auth flag coll failed, sub_uin: %d, err: %v, rid: %s", subUin, err, kt.Rid)
			return nil, fmt.Errorf("describe safe auth flag coll failed for sub_uin %d, err: %v", subUin, err)
		}

		result := typeaccount.SafeAuthFlagCollResult{
			SubUin:      subUin,
			PromptTrust: resp.Response.PromptTrust,
		}

		if resp.Response.LoginFlag != nil {
			result.LoginFlag = &typeaccount.LoginActionFlag{
				Phone:    resp.Response.LoginFlag.Phone,
				Token:    resp.Response.LoginFlag.Token,
				Stoken:   resp.Response.LoginFlag.Stoken,
				Wechat:   resp.Response.LoginFlag.Wechat,
				Custom:   resp.Response.LoginFlag.Custom,
				Mail:     resp.Response.LoginFlag.Mail,
				U2FToken: resp.Response.LoginFlag.U2FToken,
			}
		}

		if resp.Response.ActionFlag != nil {
			result.ActionFlag = &typeaccount.LoginActionFlag{
				Phone:    resp.Response.ActionFlag.Phone,
				Token:    resp.Response.ActionFlag.Token,
				Stoken:   resp.Response.ActionFlag.Stoken,
				Wechat:   resp.Response.ActionFlag.Wechat,
				Custom:   resp.Response.ActionFlag.Custom,
				Mail:     resp.Response.ActionFlag.Mail,
				U2FToken: resp.Response.ActionFlag.U2FToken,
			}
		}

		if resp.Response.OffsiteFlag != nil {
			result.OffsiteFlag = &typeaccount.OffsiteFlag{
				VerifyFlag:   resp.Response.OffsiteFlag.VerifyFlag,
				NotifyPhone:  resp.Response.OffsiteFlag.NotifyPhone,
				NotifyEmail:  resp.Response.OffsiteFlag.NotifyEmail,
				NotifyWechat: resp.Response.OffsiteFlag.NotifyWechat,
				Tips:         resp.Response.OffsiteFlag.Tips,
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// SetMfaFlag set sub-account login protection and sensitive operation protection.
// reference: https://cloud.tencent.com/document/product/598/36227
func (t *TCloudImpl) SetMfaFlag(kt *kit.Kit, opt *typeaccount.SetMfaFlagOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "set mfa flag option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewSetMfaFlagRequest()
	req.OpUin = converter.ValToPtr(opt.OpUin)

	if opt.LoginFlag != nil {
		req.LoginFlag = &cam.LoginActionMfaFlag{
			Phone:  opt.LoginFlag.Phone,
			Stoken: opt.LoginFlag.Stoken,
			Wechat: opt.LoginFlag.Wechat,
		}
	}

	if opt.ActionFlag != nil {
		req.ActionFlag = &cam.LoginActionMfaFlag{
			Phone:  opt.ActionFlag.Phone,
			Stoken: opt.ActionFlag.Stoken,
			Wechat: opt.ActionFlag.Wechat,
		}
	}

	_, err = camClient.SetMfaFlagWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("set mfa flag failed, op_uin: %d, err: %v, rid: %s", opt.OpUin, err, kt.Rid)
		return fmt.Errorf("set mfa flag failed, err: %v", err)
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
		AppID:              converter.PtrToVal(resp.Response.AppId),
	}, nil
}

// CreateAccessKey create access key for CAM sub-user.
// reference: https://cloud.tencent.com/document/product/598/82370
func (t *TCloudImpl) CreateAccessKey(kt *kit.Kit,
	opt *typeaccount.CreateAccessKeyOption) (*typeaccount.CreateAccessKeyResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create access key option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewCreateAccessKeyRequest()
	req.TargetUin = converter.ValToPtr(opt.TargetUin)
	if opt.Description != nil {
		req.Description = opt.Description
	}

	resp, err := camClient.CreateAccessKeyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create access key failed, target_uin: %d, err: %v, rid: %s", opt.TargetUin, err, kt.Rid)
		return nil, fmt.Errorf("create access key failed, err: %v", err)
	}

	if resp.Response == nil || resp.Response.AccessKey == nil {
		return nil, errors.New("create access key returned nil response")
	}

	ak := resp.Response.AccessKey
	return &typeaccount.CreateAccessKeyResult{
		AccessKeyID:     converter.PtrToVal(ak.AccessKeyId),
		SecretAccessKey: converter.PtrToVal(ak.SecretAccessKey),
		Status:          converter.PtrToVal(ak.Status),
		CreateTime:      ak.CreateTime,
	}, nil
}

// DeleteAccessKey delete access key for CAM sub-user.
// reference: https://cloud.tencent.com/document/product/598/82369
func (t *TCloudImpl) DeleteAccessKey(kt *kit.Kit, opt *typeaccount.DeleteAccessKeyOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete access key option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewDeleteAccessKeyRequest()
	req.AccessKeyId = converter.ValToPtr(opt.AccessKeyID)
	req.TargetUin = converter.ValToPtr(opt.TargetUin)

	_, err = camClient.DeleteAccessKeyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete access key failed, access_key_id: %s, target_uin: %d, err: %v, rid: %s",
			opt.AccessKeyID, opt.TargetUin, err, kt.Rid)
		return fmt.Errorf("delete access key failed, err: %v", err)
	}

	return nil
}

// UpdateAccessKey update access key status (Active/Inactive) for CAM sub-user.
// reference: https://cloud.tencent.com/document/product/598/82368
func (t *TCloudImpl) UpdateAccessKey(kt *kit.Kit, opt *typeaccount.UpdateAccessKeyOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "update access key option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewUpdateAccessKeyRequest()
	req.AccessKeyId = converter.ValToPtr(opt.AccessKeyID)
	req.Status = converter.ValToPtr(opt.Status)
	req.TargetUin = converter.ValToPtr(opt.TargetUin)

	_, err = camClient.UpdateAccessKeyWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("update access key failed, access_key_id: %s, target_uin: %d, err: %v, rid: %s",
			opt.AccessKeyID, opt.TargetUin, err, kt.Rid)
		return fmt.Errorf("update access key failed, err: %v", err)
	}

	return nil
}

// ListAccessKeys list access keys for CAM user.
// reference: https://cloud.tencent.com/document/product/598/45156
func (t *TCloudImpl) ListAccessKeys(kt *kit.Kit,
	opt *typeaccount.ListAccessKeysOption) ([]typeaccount.AccessKeyInfo, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list access keys option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewListAccessKeysRequest()
	req.TargetUin = converter.ValToPtr(opt.TargetUin)

	rangeMS := [2]uint{constant.TCloudRetryDelayMinMS, constant.TCloudRetryDelayMaxMS}
	resp, err := LimitExceededErrRetry(camClient.ListAccessKeysWithContext, kt, req,
		constant.TCloudClientErrRetryTimes, rangeMS)
	if err != nil {
		logs.Errorf("list access keys failed, target_uin: %d, err: %v, rid: %s",
			opt.TargetUin, err, kt.Rid)
		return nil, fmt.Errorf("list access keys failed, err: %v", err)
	}

	list := make([]typeaccount.AccessKeyInfo, 0, len(resp.Response.AccessKeys))
	for _, one := range resp.Response.AccessKeys {
		list = append(list, typeaccount.AccessKeyInfo{
			AccessKeyID: converter.PtrToVal(one.AccessKeyId),
			Status:      converter.PtrToVal(one.Status),
			CreateTime:  converter.PtrToVal(one.CreateTime),
			Description: one.Description,
		})
	}

	return list, nil
}

// GetSecurityLastUsed get access key last usage info.
// reference: https://cloud.tencent.com/document/product/598/58230
func (t *TCloudImpl) GetSecurityLastUsed(kt *kit.Kit,
	opt *typeaccount.GetSecurityLastUsedOption) ([]typeaccount.SecretIdLastUsed, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "get security last used option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	camClient, err := t.clientSet.CamServiceClient("")
	if err != nil {
		return nil, fmt.Errorf("new cam client failed, err: %v", err)
	}

	req := cam.NewGetSecurityLastUsedRequest()
	secretIds := make([]*string, 0, len(opt.SecretIdList))
	for i := range opt.SecretIdList {
		secretIds = append(secretIds, &opt.SecretIdList[i])
	}
	req.SecretIdList = secretIds

	resp, err := camClient.GetSecurityLastUsedWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("get security last used failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("get security last used failed, err: %v", err)
	}

	list := make([]typeaccount.SecretIdLastUsed, 0, len(resp.Response.SecretIdLastUsedRows))
	for _, one := range resp.Response.SecretIdLastUsedRows {
		list = append(list, typeaccount.SecretIdLastUsed{
			SecretId:           converter.PtrToVal(one.SecretId),
			LastUsedDate:       one.LastUsedDate,
			LastSecretUsedDate: one.LastSecretUsedDate,
		})
	}

	return list, nil
}
