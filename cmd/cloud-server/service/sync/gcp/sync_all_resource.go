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
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// gcp list每分钟限制1500次请求
// TODO: 使用限频工具控制请求
var syncConcurrencyCount = 1

// SyncAllResourceOption ...
type SyncAllResourceOption struct {
	AccountID string `json:"account_id" validate:"required"`
	// SyncPublicResource 是否同步公共资源
	SyncPublicResource bool `json:"sync_public_resource" validate:"omitempty"`
}

// Validate SyncAllResourceOption
func (opt *SyncAllResourceOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SyncAllResource sync resource.
func SyncAllResource(kt *kit.Kit, cliSet *client.ClientSet,
	opt *SyncAllResourceOption) (enumor.CloudResourceType, error) {

	if err := opt.Validate(); err != nil {
		return "", err
	}

	start := time.Now()
	logs.V(3).Infof("gcp account[%s] sync all resource start, time: %v, opt: %v, rid: %s", opt.AccountID,
		start, opt, kt.Rid)

	var hitErr error
	defer func() {
		if hitErr != nil {
			logs.Errorf("%s: sync all resource failed, err: %v, account: %s, rid: %s", constant.AccountSyncFailed,
				hitErr, opt.AccountID, kt.Rid)
			return
		}

		logs.V(3).Infof("gcp account[%s] sync all resource end, cost: %v, opt: %v, rid: %s", opt.AccountID,
			time.Since(start), opt, kt.Rid)
	}()

	if opt.SyncPublicResource {
		syncOpt := &SyncPublicResourceOption{AccountID: opt.AccountID}
		if hitErr = SyncPublicResource(kt, cliSet, syncOpt); hitErr != nil {
			logs.Errorf("sync public resource failed, err: %v, opt: %v, rid: %s", hitErr, opt, kt.Rid)
			return "", hitErr
		}
	}

	regions, hitErr := ListRegion(kt, cliSet.DataService())
	if hitErr != nil {
		return "", hitErr
	}

	regionZoneMap, hitErr := GetRegionZoneMap(kt, cliSet.DataService())
	if hitErr != nil {
		return "", hitErr
	}

	sd := &detail.SyncDetail{
		Kt:        kt,
		DataCli:   cliSet.DataService(),
		AccountID: opt.AccountID,
		Vendor:    string(enumor.Gcp),
	}

	if hitErr = SyncDisk(kt, cliSet, opt.AccountID, regionZoneMap, sd); hitErr != nil {
		return enumor.DiskCloudResType, hitErr
	}

	if hitErr = SyncVpc(kt, cliSet, opt.AccountID, sd); hitErr != nil {
		return enumor.VpcCloudResType, hitErr
	}

	if hitErr = SyncSubnet(kt, cliSet, opt.AccountID, regions, sd); hitErr != nil {
		return enumor.SubnetCloudResType, hitErr
	}

	if hitErr = SyncEip(kt, cliSet, opt.AccountID, regions, sd); hitErr != nil {
		return enumor.EipCloudResType, hitErr
	}

	if hitErr = SyncFireWall(kt, cliSet, opt.AccountID, sd); hitErr != nil {
		return enumor.GcpFirewallRuleCloudResType, hitErr
	}

	if hitErr = SyncCvm(kt, cliSet, opt.AccountID, regionZoneMap, sd); hitErr != nil {
		return enumor.CvmCloudResType, hitErr
	}

	if hitErr = SyncRoute(kt, cliSet, opt.AccountID, sd); hitErr != nil {
		return enumor.RouteTableCloudResType, hitErr
	}

	if hitErr = SyncSubAccount(kt, cliSet, opt.AccountID, sd); hitErr != nil {
		return enumor.SubAccountCloudResType, hitErr
	}

	if hitErr = SyncCvmCCHostInfo(kt, cliSet, opt.AccountID, regions, sd); hitErr != nil {
		return enumor.CvmCCInfoResType, hitErr
	}

	return "", nil
}
