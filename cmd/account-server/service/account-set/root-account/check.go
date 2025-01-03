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

package rootaccount

import (
	"errors"
	"fmt"

	accountset "hcm/pkg/api/account-server/account-set"
	"hcm/pkg/api/cloud-server/account"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// CheckDuplicateRootAccount 检查主账号是否重复
func CheckDuplicateRootAccount(cts *rest.Contexts, client *client.ClientSet, vendor enumor.Vendor,
	mainAccountIDFieldValue string) error {

	// TODO: 后续需要解决并发问题
	// 后台查询是否主账号重复
	mainAccountIDFieldName := vendor.GetMainAccountIDField()
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", string(vendor)),
			tools.RuleJSONEqual(fmt.Sprintf("extension.%s", mainAccountIDFieldName), mainAccountIDFieldValue),
		),
		Page: core.NewCountPage(),
	}
	result, err := client.DataService().Global.RootAccount.List(cts.Kit, listReq)
	if err != nil {
		return err
	}

	if result.Count > 0 {
		return fmt.Errorf("%s[%s] should be not duplicate", mainAccountIDFieldName, mainAccountIDFieldValue)
	}

	return nil
}

// QueryRootAccountBySecret 根据秘钥获取账号信息
func (s *service) QueryRootAccountBySecret(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 校验用户有一级账号管理权限
	if err := s.checkPermission(cts, meta.RootAccount, meta.Find); err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.HuaWei:
		return s.getHuaWeiAccountInfo(cts)
	case enumor.Aws:
		return s.getAwsAccountInfo(cts)
	case enumor.Azure:
		return s.getAzureAccountInfo(cts)
	case enumor.Gcp:
		return s.getGcpAccountInfo(cts)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s, for get root account info", vendor)
	}
}

func (s *service) getHuaWeiAccountInfo(cts *rest.Contexts) (*cloud.HuaWeiInfoBySecret, error) {
	req := new(accountset.HuaWeiAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := s.client.HCService().HuaWei.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.HuaWeiSecret)
	if err != nil {
		logs.Errorf("fail to get huawei account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return info, nil
}

func (s *service) getAwsAccountInfo(cts *rest.Contexts) (*cloud.AwsInfoBySecret, error) {
	req := new(accountset.AwsAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := s.client.HCService().Aws.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.AwsSecret)
	if err != nil {
		logs.Errorf("fail to get aws account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return info, nil
}

func (s *service) getGcpAccountInfo(cts *rest.Contexts) ([]cloud.GcpProjectInfo, error) {
	req := new(accountset.GcpAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := s.client.HCService().Gcp.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.GcpSecret)
	if err != nil {
		logs.Errorf("fail to get gcp account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return info.CloudProjectInfos, nil
}

func (s *service) getAzureAccountInfo(cts *rest.Contexts) (*account.AzureAccountInfoBySecretResp, error) {
	req := new(accountset.AzureAccountInfoBySecretReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	info, err := s.client.HCService().Azure.Account.GetBySecret(cts.Kit.Ctx, cts.Kit.Header(), req.AzureSecret)
	if err != nil {
		logs.Errorf("fail to get azure account info, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(info.SubscriptionInfos) < 1 {
		logs.Errorf("get azure account info failed, no subscription found, rid: %s", cts.Kit.Rid)
		return nil, errors.New("no subscription found")
	}

	subscription := info.SubscriptionInfos[0]
	result := &account.AzureAccountInfoBySecretResp{
		CloudSubscriptionID:   subscription.CloudSubscriptionID,
		CloudSubscriptionName: subscription.CloudSubscriptionName,
	}
	// 补全ApplicationName
	for _, one := range info.ApplicationInfos {
		if one.CloudApplicationID == req.CloudApplicationID {
			result.CloudApplicationName = one.CloudApplicationName
			break
		}
	}
	// 没有拿到应用id的情况
	if len(result.CloudApplicationName) == 0 {
		logs.Errorf("failed to get application name, rid: %s", cts.Kit.Rid)
		return nil, fmt.Errorf("failed to get application name")
	}

	return result, nil
}
