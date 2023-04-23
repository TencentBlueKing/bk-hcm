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
	typeaccount "hcm/pkg/adaptor/types/account"
	proto "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetTCloudAccountZoneQuota 获取腾讯云账号配额
func (svc *service) GetTCloudAccountZoneQuota(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GetTCloudAccountZoneQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeaccount.GetTCloudAccountZoneQuotaOption{
		Region: req.Region,
		Zone:   req.Zone,
	}
	quota, err := client.GetAccountZoneQuota(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor get account zone quota failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	return quota, nil
}

// GetHuaWeiAccountRegionQuota 获取华为云账号配额
func (svc *service) GetHuaWeiAccountRegionQuota(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GetHuaWeiAccountRegionQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeaccount.GetHuaWeiAccountZoneQuotaOption{
		Region: req.Region,
	}
	quota, err := client.GetAccountQuota(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor get account quota failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	return quota, nil
}

// GetGcpAccountRegionQuota ...
func (svc *service) GetGcpAccountRegionQuota(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.GetGcpAccountRegionQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.New(errf.DecodeRequestFailed, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, errf.Newf(errf.InvalidParameter, err.Error())
	}

	client, err := svc.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeaccount.GcpProjectRegionQuotaOption{
		Region: req.Region,
	}
	quota, err := client.GetProjectRegionQuota(cts.Kit, opt)
	if err != nil {
		logs.Errorf("request adaptor get project quota failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	return quota, nil
}
