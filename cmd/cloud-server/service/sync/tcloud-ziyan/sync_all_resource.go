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

package tziyan

import (
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"golang.org/x/sync/errgroup"
)

var syncConcurrencyCount = 10

// SyncAllResourceOption ...
type SyncAllResourceOption struct {
	AccountID string `json:"account_id" validate:"required"`
	// SyncPublicResource 是否同步公共资源
	SyncPublicResource bool `json:"sync_public_resource" validate:"omitempty"`
}

// ResSyncFunc 资源同步函数
type ResSyncFunc func(kt *kit.Kit, cliSet *client.ClientSet, accountID string, regions []string,
	sd *detail.SyncDetail) error

// Validate SyncAllResourceOption
func (opt *SyncAllResourceOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SyncAllResource sync resource.
func SyncAllResource(kt *kit.Kit, cliSet *client.ClientSet, opt *SyncAllResourceOption) (
	failedRes enumor.CloudResourceType, hitErr error) {

	if err := opt.Validate(); err != nil {
		return "", err
	}

	start := time.Now()
	logs.V(3).Infof("tcloud ziyan account[%s] sync all resource start, time: %v, opt: %v, rid: %s", opt.AccountID,
		start, opt, kt.Rid)

	defer func() {
		if hitErr != nil {
			logs.Errorf("%s: sync all resource failed on %s(%s), err: %v, rid: %s",
				constant.AccountSyncFailed, opt.AccountID, failedRes, hitErr, kt.Rid)
			return
		}

		logs.V(3).Infof("tcloud ziyan account(%s) sync all resource end, cost: %v, opt: %v, rid: %s", opt.AccountID,
			time.Since(start), opt, kt.Rid)
	}()

	if opt.SyncPublicResource {
		syncOpt := &SyncPublicResourceOption{
			AccountID: opt.AccountID,
		}
		if failedRes, hitErr = SyncPublicResource(kt, cliSet, syncOpt); hitErr != nil {
			logs.Errorf("sync public resource failed, err: %v, opt: %v, rid: %s", hitErr, opt, kt.Rid)
			return failedRes, hitErr
		}
	}

	regions, hitErr := ListRegion(kt, cliSet.DataService())
	if hitErr != nil {
		return "", hitErr
	}

	sd := &detail.SyncDetail{
		Kt:        kt,
		DataCli:   cliSet.DataService(),
		AccountID: opt.AccountID,
		Vendor:    string(enumor.TCloudZiyan),
	}

	var eg, _ = errgroup.WithContext(kt.Ctx)
	var resType enumor.CloudResourceType
	// 单独开协程处理，自研云主机同步和其他资源的同步不相互影响
	eg.Go(func() error {
		if syncErr := SyncHost(kt, cliSet, opt.AccountID, sd); syncErr != nil {
			// 由于自研云主机和其他流程是独立的，所以这里失败时单独更新
			if err := sd.ResSyncStatusFailed(enumor.CvmCloudResType, syncErr); err != nil {
				logs.Errorf("sync res detail failed, vendor: %s, res: %s, err: %v, accountID: %s, rid: %s",
					enumor.TCloudZiyan, enumor.CvmCloudResType, err, sd.AccountID, kt.Rid)
				return err
			}
			return nil
		}
		return nil
	})

	eg.Go(func() error {
		for _, syncer := range syncOrder {
			if hitErr = syncer.ResSyncFunc(kt, cliSet, opt.AccountID, regions, sd); hitErr != nil {
				resType = syncer.ResType
				return hitErr
			}
		}

		return nil
	})

	if hitErr = eg.Wait(); hitErr != nil {
		return resType, hitErr
	}

	return "", nil
}

type syncItem struct {
	ResType enumor.CloudResourceType
	ResSyncFunc
}

var syncOrder = []syncItem{
	{enumor.VpcCloudResType, SyncVpc},
	{enumor.SubnetCloudResType, SyncSubnet},
	{enumor.ArgumentTemplateResType, SyncArgsTpl},
	{enumor.SecurityGroupCloudResType, SyncSG},
	{enumor.CertCloudResType, SyncCert},
	{enumor.LoadBalancerCloudResType, SyncLoadBalancer},
	{enumor.SecurityGroupUsageBizRelResType, SyncSGUsageBizRel},
}
