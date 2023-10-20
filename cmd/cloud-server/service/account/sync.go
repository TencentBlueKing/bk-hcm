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
	"errors"
	"fmt"
	"strings"

	"hcm/cmd/cloud-server/service/sync/aws"
	"hcm/cmd/cloud-server/service/sync/azure"
	"hcm/cmd/cloud-server/service/sync/gcp"
	"hcm/cmd/cloud-server/service/sync/huawei"
	"hcm/cmd/cloud-server/service/sync/lock"
	"hcm/cmd/cloud-server/service/sync/tcloud"
	"hcm/pkg/api/core"
	imagecloud "hcm/pkg/api/data-service/cloud/image"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// SyncCloudResource ...
func (a *accountSvc) SyncCloudResource(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	// 校验用户有该账号的更新权限
	if err := a.checkPermission(cts, meta.Update, accountID); err != nil {
		return nil, err
	}

	// 查询该账号对应的Vendor
	baseInfo, err := a.client.DataService().Global.Cloud.GetResourceBasicInfo(
		cts.Kit.Ctx,
		cts.Kit.Header(),
		enumor.AccountCloudResType,
		accountID,
	)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	isNeedSyncPublicResFlag, err := isNeedSyncPublicResource(cts.Kit, a.client.DataService(), baseInfo.Vendor)
	if err != nil {
		logs.Errorf("is need sync public resource failed, err: %v, vendor: %s, rid: %s", err,
			baseInfo.Vendor, cts.Kit.Rid)
		return nil, err
	}

	leaseID, err := lock.Manager.TryLock(lock.Key(accountID))
	if err != nil {
		if err == lock.ErrLockFailed {
			return nil, errors.New("synchronization is in progress")
		}

		return nil, err
	}

	go func(leaseID etcd3.LeaseID) {
		defer func() {
			if err := lock.Manager.UnLock(leaseID); err != nil {
				// 锁已经超时释放了
				if strings.Contains(err.Error(), "requested lease not found") {
					return
				}

				logs.Errorf("%s: unlock account sync lock failed, err: %v, accountID: %s, leaseID: %d, rid: %s",
					constant.AccountSyncFailed, err, accountID, leaseID, cts.Kit.Rid)
			}
		}()

		a.syncAllResourceByVendor(cts, baseInfo, accountID, isNeedSyncPublicResFlag)
	}(leaseID)

	return nil, nil
}

func (a *accountSvc) syncAllResourceByVendor(cts *rest.Contexts, baseInfo *types.CloudResourceBasicInfo,
	accountID string, isNeedSyncPublicResFlag bool) {

	switch baseInfo.Vendor {
	case enumor.TCloud:
		opt := &tcloud.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeedSyncPublicResFlag,
		}
		tcloud.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.Aws:
		opt := &aws.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeedSyncPublicResFlag,
		}
		aws.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.HuaWei:
		opt := &huawei.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeedSyncPublicResFlag,
		}
		huawei.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.Gcp:
		opt := &gcp.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeedSyncPublicResFlag,
		}
		gcp.SyncAllResource(cts.Kit, a.client, opt)

	case enumor.Azure:
		opt := &azure.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeedSyncPublicResFlag,
		}
		azure.SyncAllResource(cts.Kit, a.client, opt)

	default:
		logs.Errorf("account: %s's vendor not support, vendor: %s", accountID, baseInfo.Vendor)
	}
}

func isNeedSyncPublicResource(kt *kit.Kit, dataCli *dataservice.Client, vendor enumor.Vendor) (
	bool, error) {

	need, err := isNeedSyncRegion(kt, dataCli, vendor)
	if err != nil {
		return false, err
	}

	if need {
		return true, nil
	}

	switch vendor {
	case enumor.Aws, enumor.TCloud, enumor.HuaWei, enumor.Gcp:
		listZoneReq := &protocloud.ZoneListReq{
			Filter: tools.EqualExpression("vendor", vendor),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.Global.Zone.ListZone(kt.Ctx, kt.Header(), listZoneReq)
		if err != nil {
			return false, err
		}

		if result.Count == 0 {
			return true, nil
		}

	case enumor.Azure:
		// azure没有可用区
	default:
		return false, fmt.Errorf("vendor: %s not support", vendor)
	}

	switch vendor {
	case enumor.Aws, enumor.TCloud, enumor.HuaWei, enumor.Gcp, enumor.Azure:
		listZoneReq := &imagecloud.ImageListReq{
			Filter: tools.EqualExpression("vendor", vendor),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.Global.ListImage(kt.Ctx, kt.Header(), listZoneReq)
		if err != nil {
			return false, err
		}

		if converter.PtrToVal(result.Count) == 0 {
			return true, nil
		}

	default:
		return false, fmt.Errorf("vendor: %s not support", vendor)
	}

	return false, nil
}

func isNeedSyncRegion(kt *kit.Kit, dataCli *dataservice.Client, vendor enumor.Vendor) (bool, error) {
	regionCount := uint64(0)
	switch vendor {
	case enumor.TCloud:
		listReq := &protoregion.TCloudRegionListReq{
			Filter: tools.AllExpression(),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.TCloud.Region.ListRegion(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			return false, err
		}
		regionCount = result.Count

	case enumor.Aws:
		listReq := &protoregion.AwsRegionListReq{
			Filter: tools.AllExpression(),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.Aws.Region.ListRegion(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			return false, err
		}
		regionCount = result.Count

	case enumor.HuaWei:
		listReq := &protoregion.HuaWeiRegionListReq{
			Filter: tools.AllExpression(),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.HuaWei.Region.ListRegion(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			return false, err
		}
		regionCount = result.Count

	case enumor.Gcp:
		listReq := &protoregion.GcpRegionListReq{
			Filter: tools.AllExpression(),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.Gcp.Region.ListRegion(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			return false, err
		}
		regionCount = result.Count

	case enumor.Azure:
		listReq := &protoregion.AzureRegionListReq{
			Filter: tools.AllExpression(),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.Azure.Region.ListRegion(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			return false, err
		}
		regionCount = result.Count

	default:
		return false, fmt.Errorf("vendor: %s not support", vendor)
	}

	if regionCount == 0 {
		return true, nil
	}

	return false, nil
}
