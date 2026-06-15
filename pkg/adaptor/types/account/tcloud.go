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

package account

import (
	"fmt"
	"strconv"

	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"
)

// TCloudAccount define tcloud sub account.
type TCloudAccount struct {
	// 子用户用户 ID
	Uin *uint64 `json:"uin"`

	// 子用户用户名
	Name *string `json:"name"`

	// 子用户 UID
	Uid *uint64 `json:"uid"`

	// 子用户备注
	Remark *string `json:"remark"`

	// 子用户能否登录控制台
	ConsoleLogin *uint64 `json:"console_login"`

	// 手机号
	PhoneNum *string `json:"phone_num"`

	// 区号
	CountryCode *string `json:"country_code"`

	// 邮箱
	Email *string `json:"email"`

	// 创建时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreateTime *string `json:"create_time"`

	// 昵称
	// 注意：此字段可能返回 null，表示取不到有效值。
	NickName *string `json:"nick_name"`
}

// GetCloudID ...
func (account TCloudAccount) GetCloudID() string {
	return strconv.FormatUint(converter.PtrToVal(account.Uin), 10)
}

// TCloudAccountWithExt 腾讯云子账号完整信息（包含基础信息和扩展信息）
type TCloudAccountWithExt struct {
	TCloudAccount `json:",inline"`
	Extension     *coresubaccount.TCloudExtension `json:"extension"`
}

// GetCloudID 实现 CloudResType 接口，用于 Diff 对比
func (a TCloudAccountWithExt) GetCloudID() string {
	return a.TCloudAccount.GetCloudID()
}

// AddUserOption define tcloud add user option.
type AddUserOption struct {
	Name              string `json:"name" validate:"required"`
	Remark            string `json:"remark" validate:"omitempty"`
	ConsoleLogin      uint64 `json:"console_login" validate:"omitempty"`
	UseAPI            uint64 `json:"use_api" validate:"omitempty"`
	Password          string `json:"password" validate:"omitempty"`
	NeedResetPassword uint64 `json:"need_reset_password" validate:"omitempty"`
	PhoneNum          string `json:"phone_num" validate:"omitempty"`
	CountryCode       string `json:"country_code" validate:"omitempty"`
	Email             string `json:"email" validate:"omitempty"`
}

// Validate add user option.
func (opt AddUserOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AddUserResult define tcloud AddUser API response fields.
// Only contains fields actually returned by the API, not the full TCloudAccount.
type AddUserResult struct {
	Uin       *uint64 `json:"uin"`
	Name      *string `json:"name"`
	UID       *uint64 `json:"uid"`
	SecretID  string  `json:"secret_id"`
	SecretKey string  `json:"secret_key"`
	Password  string  `json:"password,omitempty"`
}

// UpdateUserOption define tcloud update user option.
// reference: https://cloud.tencent.com/document/product/598/34583
// Pointer fields use nil to indicate "no change".
type UpdateUserOption struct {
	Name              string  `json:"name" validate:"required"`
	Remark            *string `json:"remark" validate:"omitempty"`
	ConsoleLogin      *uint64 `json:"console_login" validate:"omitempty"`
	Password          *string `json:"password" validate:"omitempty"`
	NeedResetPassword *uint64 `json:"need_reset_password" validate:"omitempty"`
	PhoneNum          *string `json:"phone_num" validate:"omitempty"`
	CountryCode       *string `json:"country_code" validate:"omitempty"`
	Email             *string `json:"email" validate:"omitempty"`
}

// Validate update user option.
func (opt UpdateUserOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudListPolicyOption define tcloud list policy option.
type TCloudListPolicyOption struct {
	Uin         uint64  `json:"uin" validate:"required"`
	ServiceType *string `json:"service_type" validate:"omitempty"`
}

// Validate define tcloud list policy option.
func (opt TCloudListPolicyOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// DescribeSafeAuthFlagCollMaxUIN is the max number of UINs per DescribeSafeAuthFlagColl API call.
const DescribeSafeAuthFlagCollMaxUIN = 10

// DescribeSafeAuthFlagCollOption define tcloud describe sub-account safe auth flag option.
// reference: https://cloud.tencent.com/document/product/598/48602
type DescribeSafeAuthFlagCollOption struct {
	// SubUins is the sub-account UIN list.
	SubUins []uint64 `json:"sub_uins" validate:"required,min=1,max=10"`
}

// Validate DescribeSafeAuthFlagCollOption.
func (opt DescribeSafeAuthFlagCollOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}
	if len(opt.SubUins) > DescribeSafeAuthFlagCollMaxUIN {
		return fmt.Errorf("sub_uin count %d exceeds max %d", len(opt.SubUins), DescribeSafeAuthFlagCollMaxUIN)
	}
	return nil
}

// DescribeSafeAuthFlagOption define tcloud describe user's safe auth flag option.
// reference: https://cloud.tencent.com/document/product/598/48426
type DescribeSafeAuthFlagOption struct {
	// Uin is the main account UIN, optional. If not provided, queries current user.
	Uin *uint64 `json:"uin,omitempty" validate:"omitempty"`
}

// Validate DescribeSafeAuthFlagOption.
func (opt DescribeSafeAuthFlagOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// LoginActionFlag define login or sensitive operation protection flag.
type LoginActionFlag struct {
	// Phone indicates whether phone verification is enabled (1: enabled, 0: disabled).
	Phone *uint64 `json:"phone"`
	// Token indicates whether hard token verification is enabled.
	Token *uint64 `json:"token"`
	// Stoken indicates whether soft token verification is enabled.
	Stoken *uint64 `json:"stoken"`
	// Wechat indicates whether WeChat verification is enabled.
	Wechat *uint64 `json:"wechat"`
	// Custom indicates whether custom verification is enabled.
	Custom *uint64 `json:"custom"`
	// Mail indicates whether email verification is enabled.
	Mail *uint64 `json:"mail"`
	// U2FToken indicates whether U2F hardware token verification is enabled.
	U2FToken *uint64 `json:"u2f_token"`
}

// ToProtectionFlag maps enabled fields (value == 1) to AccountProtectionFlag using priority:
// Phone > Token > Stoken > Wechat > Custom > Mail > U2FToken. Returns nil if flag is nil or none enabled.
func (flag LoginActionFlag) ToProtectionFlag() *enumor.AccountProtectionFlag {
	if flag.Phone != nil && converter.PtrToVal(flag.Phone) == 1 {
		return converter.ValToPtr(enumor.PhoneProtection)
	}
	if flag.Token != nil && converter.PtrToVal(flag.Token) == 1 {
		return converter.ValToPtr(enumor.TokenProtection)
	}
	if flag.Stoken != nil && converter.PtrToVal(flag.Stoken) == 1 {
		return converter.ValToPtr(enumor.StokenProtection)
	}
	if flag.Wechat != nil && converter.PtrToVal(flag.Wechat) == 1 {
		return converter.ValToPtr(enumor.WechatProtection)
	}
	if flag.Custom != nil && converter.PtrToVal(flag.Custom) == 1 {
		return converter.ValToPtr(enumor.CustomProtection)
	}
	if flag.Mail != nil && converter.PtrToVal(flag.Mail) == 1 {
		return converter.ValToPtr(enumor.MailProtection)
	}
	if flag.U2FToken != nil && converter.PtrToVal(flag.U2FToken) == 1 {
		return converter.ValToPtr(enumor.U2FTokenProtection)
	}

	return nil
}

// OffsiteFlag define offsite login protection settings.
type OffsiteFlag struct {
	// VerifyFlag indicates whether verification is required for offsite login.
	VerifyFlag *uint64 `json:"verify_flag"`
	// NotifyPhone indicates whether phone notification is enabled.
	NotifyPhone *uint64 `json:"notify_phone"`
	// NotifyEmail indicates whether email notification is enabled.
	NotifyEmail *int64 `json:"notify_email"`
	// NotifyWechat indicates whether WeChat notification is enabled.
	NotifyWechat *uint64 `json:"notify_wechat"`
	// Tips indicates tip settings.
	Tips *uint64 `json:"tips"`
}

// SetMfaFlagOption define tcloud set sub-account login protection and sensitive operation protection option.
// reference: https://cloud.tencent.com/document/product/598/36227
type SetMfaFlagOption struct {
	// OpUin is the sub-account UIN to set MFA flag for.
	OpUin uint64 `json:"op_uin" validate:"required"`
	// LoginFlag is the login protection settings, nil means no change.
	LoginFlag *LoginActionFlag `json:"login_flag" validate:"omitempty"`
	// ActionFlag is the sensitive operation protection settings, nil means no change.
	ActionFlag *LoginActionFlag `json:"action_flag" validate:"omitempty"`
}

// Validate SetMfaFlagOption.
func (opt SetMfaFlagOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// DescribeSubAccountsMaxUIN is the max number of UINs per DescribeSubAccounts API call.
const DescribeSubAccountsMaxUIN = 50

// DescribeSubAccountsOption define tcloud DescribeSubAccounts option.
// reference: https://cloud.tencent.com/document/api/598/53486
type DescribeSubAccountsOption struct {
	SubUin []uint64 `json:"sub_uin" validate:"required"`
}

// Validate DescribeSubAccountsOption.
func (opt DescribeSubAccountsOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.SubUin) < 1 {
		return fmt.Errorf("sub_uin count %d is less than 1", len(opt.SubUin))
	}

	if len(opt.SubUin) > DescribeSubAccountsMaxUIN {
		return fmt.Errorf("sub_uin count %d exceeds max %d",
			len(opt.SubUin), DescribeSubAccountsMaxUIN)
	}

	return nil
}

// TCloudSubAccountUser define tcloud DescribeSubAccounts API result item.
type TCloudSubAccountUser struct {
	Uin           *uint64 `json:"uin"`
	Name          *string `json:"name"`
	Uid           *uint64 `json:"uid"`
	Remark        *string `json:"remark"`
	CreateTime    *string `json:"create_time"`
	UserType      *uint64 `json:"user_type"`
	LastLoginIp   *string `json:"last_login_ip"`
	LastLoginTime *string `json:"last_login_time"`
}

// SafeAuthFlagCollResult define tcloud DescribeSafeAuthFlagColl API result item.
type SafeAuthFlagCollResult struct {
	// SubUin is the sub-account UIN.
	SubUin uint64 `json:"sub_uin"`
	// LoginFlag is the login protection settings.
	LoginFlag *LoginActionFlag `json:"login_flag"`
	// ActionFlag is the sensitive operation protection settings.
	ActionFlag *LoginActionFlag `json:"action_flag"`
	// OffsiteFlag is the offsite login protection settings.
	OffsiteFlag *OffsiteFlag `json:"offsite_flag"`
	// PromptTrust indicates whether to prompt the user to trust the device (1: prompt, 0: no prompt).
	PromptTrust *int64 `json:"prompt_trust"`
}

// SafeAuthFlagResult define tcloud DescribeSafeAuthFlag API result.
type SafeAuthFlagResult struct {
	// LoginFlag is the login protection settings.
	LoginFlag *LoginActionFlag `json:"login_flag"`
	// ActionFlag is the sensitive operation protection settings.
	ActionFlag *LoginActionFlag `json:"action_flag"`
	// OffsiteFlag is the offsite login protection settings.
	OffsiteFlag *OffsiteFlag `json:"offsite_flag"`
	// PromptTrust indicates whether to prompt the user to trust the device (1: prompt, 0: no prompt).
	PromptTrust *int64 `json:"prompt_trust"`
}

// CreateAccessKeyOption define tcloud CreateAccessKey option.
// reference: https://cloud.tencent.com/document/product/598/82370
type CreateAccessKeyOption struct {
	TargetUin   uint64  `json:"target_uin" validate:"required"`
	Description *string `json:"description" validate:"omitempty"`
}

// Validate CreateAccessKeyOption.
func (opt CreateAccessKeyOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// CreateAccessKeyResult define tcloud CreateAccessKey API result.
type CreateAccessKeyResult struct {
	AccessKeyID     string  `json:"access_key_id"`
	SecretAccessKey string  `json:"secret_access_key"`
	Status          string  `json:"status"`
	CreateTime      *string `json:"create_time"`
}

// DeleteAccessKeyOption define tcloud DeleteAccessKey option.
// reference: https://cloud.tencent.com/document/product/598/82369
type DeleteAccessKeyOption struct {
	AccessKeyID string `json:"access_key_id" validate:"required"`
	TargetUin   uint64 `json:"target_uin" validate:"required"`
}

// Validate DeleteAccessKeyOption.
func (opt DeleteAccessKeyOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloud CAM access key status constants.
// reference: https://cloud.tencent.com/document/product/598/82368
const (
	// TCloudAccessKeyStatusActive represents an active access key.
	TCloudAccessKeyStatusActive = "Active"
	// TCloudAccessKeyStatusInactive represents an inactive access key.
	TCloudAccessKeyStatusInactive = "Inactive"
)

// UpdateAccessKeyOption define tcloud UpdateAccessKey option.
// reference: https://cloud.tencent.com/document/product/598/82368
type UpdateAccessKeyOption struct {
	AccessKeyID string `json:"access_key_id" validate:"required"`
	Status      string `json:"status" validate:"required"`
	TargetUin   uint64 `json:"target_uin" validate:"required"`
}

// Validate UpdateAccessKeyOption.
func (opt UpdateAccessKeyOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ListAccessKeysOption define tcloud ListAccessKeys option.
// reference: https://cloud.tencent.com/document/product/598/45156
type ListAccessKeysOption struct {
	TargetUin uint64 `json:"target_uin" validate:"required"`
}

// Validate ListAccessKeysOption.
func (opt ListAccessKeysOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AccessKeyInfo define tcloud ListAccessKeys API result item.
type AccessKeyInfo struct {
	AccessKeyID string  `json:"access_key_id"`
	Status      string  `json:"status"`
	CreateTime  string  `json:"create_time"`
	Description *string `json:"description"`
}

// GetSecurityLastUsedMaxKeys is the max number of secret IDs per GetSecurityLastUsed API call.
const GetSecurityLastUsedMaxKeys = 10

// GetSecurityLastUsedOption define tcloud GetSecurityLastUsed option.
// reference: https://cloud.tencent.com/document/product/598/58230
type GetSecurityLastUsedOption struct {
	SecretIdList []string `json:"secret_id_list" validate:"required,min=1,max=10"`
}

// Validate GetSecurityLastUsedOption.
func (opt GetSecurityLastUsedOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SecretIdLastUsed define tcloud GetSecurityLastUsed API result item.
type SecretIdLastUsed struct {
	SecretId           string  `json:"secret_id"`
	LastUsedDate       *string `json:"last_used_date"`
	LastSecretUsedDate *uint64 `json:"last_secret_used_date"`
}

// TCloudSubAccountSecret holds cloud access key data enriched with sub-account context for DB operations.
type TCloudSubAccountSecret struct {
	AccountID          string  `json:"account_id"`
	SubAccountID       string  `json:"sub_account_id"`
	CloudMainAccountID string  `json:"cloud_main_account_id"`
	CloudSubAccountID  string  `json:"cloud_sub_account_id"`
	AccessKeyID        string  `json:"access_key_id"`
	Status             string  `json:"status"`
	CreateTime         string  `json:"create_time"`
	LastUsedTime       *string `json:"last_used_time"`
}

// GetCloudID returns the cloud-side unique identifier for the secret (AccessKeyID).
func (s TCloudSubAccountSecret) GetCloudID() string {
	return s.AccessKeyID
}

// TCloudListAttachedUserAllPoliciesOption defines options for listing all policies attached to a sub-user.
// reference: https://cloud.tencent.com/document/product/598/67728
type TCloudListAttachedUserAllPoliciesOption struct {
	// TargetUin 目标子用户 Uin
	TargetUin uint64 `json:"target_uin" validate:"required"`
	// Page 页码，从 1 开始，不能大于 200
	Page uint64 `json:"page" validate:"required,min=1,max=200"`
	// Rp 每页数量，必须大于 0 且小于等于 200
	Rp uint64 `json:"rp" validate:"required,min=1,max=200"`
	// AttachType 关联类型，可选值为："User"、"Group"，默认 "User"
	AttachType *uint64 `json:"attach_type" validate:"required"`
}

// Validate TCloudListAttachedUserAllPoliciesOption.
func (opt *TCloudListAttachedUserAllPoliciesOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudAttachedPolicy defines a single policy attached to a sub-user.
type TCloudAttachedPolicy struct {
	// PolicyID 策略 ID（字符串形式）
	PolicyID string `json:"policy_id"`
	// PolicyName 策略名称
	PolicyName string `json:"policy_name"`
	// Description 策略描述
	Description string `json:"description"`
	// AddTime 绑定时间
	AddTime string `json:"add_time"`
	// StrategyType 策略类型（"1" 表示自定义策略，"2" 表示预设策略）
	StrategyType string `json:"strategy_type"`
}

// TCloudListAttachedUserAllPoliciesResult defines the result of listing all policies attached to a sub-user.
type TCloudListAttachedUserAllPoliciesResult struct {
	// PolicyList 策略列表
	PolicyList []TCloudAttachedPolicy `json:"policy_list"`
	// TotalNum 策略总数
	TotalNum uint64 `json:"total_num"`
}

// TCloudAttachUserPolicyOption defines options for attaching a policy to a sub-user.
// reference: https://cloud.tencent.com/document/product/598/34579
type TCloudAttachUserPolicyOption struct {
	// TargetUin is the target sub-account UIN.
	TargetUin uint64 `json:"target_uin" validate:"required"`
	// PolicyId is the cloud policy ID to attach.
	PolicyId uint64 `json:"policy_id" validate:"required"`
}

// Validate TCloudAttachUserPolicyOption.
func (opt *TCloudAttachUserPolicyOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudDetachUserPolicyOption defines options for detaching a policy from a sub-user.
// reference: https://cloud.tencent.com/document/product/598/34575
type TCloudDetachUserPolicyOption struct {
	// DetachUin is the target sub-account UIN.
	DetachUin uint64 `json:"detach_uin" validate:"required"`
	// PolicyId is the cloud policy ID to detach.
	PolicyId uint64 `json:"policy_id" validate:"required"`
}

// Validate TCloudDetachUserPolicyOption.
func (opt *TCloudDetachUserPolicyOption) Validate() error {
	return validator.Validate.Struct(opt)
}
