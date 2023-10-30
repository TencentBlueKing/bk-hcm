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
	protoregion "hcm/pkg/api/data-service/cloud/region"
	protocloud "hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// Sync 账号同步。该操作同一账号不可并行执行，且是异步同步。
func Sync(kt *kit.Kit, cli *client.ClientSet, vendor enumor.Vendor, accountID string) error {

	isNeedSyncPublicResFlag, err := isNeedSyncPublicResource(kt, cli.DataService(), vendor)
	if err != nil {
		logs.Errorf("is need sync public resource failed, err: %v, vendor: %s, rid: %s", err,
			vendor, kt.Rid)
		return err
	}

	leaseID, err := lock.Manager.TryLock(lock.Key(accountID))
	if err != nil {
		if err == lock.ErrLockFailed {
			return errors.New("synchronization is in progress")
		}

		return err
	}

	go func(leaseID etcd3.LeaseID) {
		defer func() {
			if err := lock.Manager.UnLock(leaseID); err != nil {
				// 锁已经超时释放了
				if strings.Contains(err.Error(), "requested lease not found") {
					return
				}

				logs.Errorf("%s: unlock account sync lock failed, err: %v, accountID: %s, leaseID: %d, rid: %s",
					constant.AccountSyncFailed, err, accountID, leaseID, kt.Rid)
			}
		}()

		err = SyncAllResource(kt, cli, vendor, accountID, isNeedSyncPublicResFlag)
		if err != nil {
			logs.Errorf("sync account: %s failed, err: %v, rid: %s", accountID, err, kt.Rid)
		}

	}(leaseID)

	return nil
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
		listZoneReq := &core.ListReq{
			Filter: tools.EqualExpression("vendor", vendor),
			Page:   core.NewCountPage(),
		}
		result, err := dataCli.Global.ListImage(kt, listZoneReq)
		if err != nil {
			return false, err
		}

		if result.Count == 0 {
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

// SyncAllResource sync all resource.
func SyncAllResource(kt *kit.Kit, cli *client.ClientSet, vendor enumor.Vendor,
	accountID string, isNeed bool) error {

	var resType enumor.CloudResourceType
	var err error
	switch vendor {
	case enumor.TCloud:
		opt := &tcloud.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeed,
		}
		resType, err = tcloud.SyncAllResource(kt, cli, opt)

	case enumor.Aws:
		opt := &aws.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeed,
		}
		resType, err = aws.SyncAllResource(kt, cli, opt)

	case enumor.HuaWei:
		opt := &huawei.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeed,
		}
		resType, err = huawei.SyncAllResource(kt, cli, opt)

	case enumor.Gcp:
		opt := &gcp.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeed,
		}
		resType, err = gcp.SyncAllResource(kt, cli, opt)

	case enumor.Azure:
		opt := &azure.SyncAllResourceOption{
			AccountID:          accountID,
			SyncPublicResource: isNeed,
		}
		resType, err = azure.SyncAllResource(kt, cli, opt)

	default:
		logs.Errorf("account: %s's vendor not support, vendor: %s, rid: %s", accountID, vendor, kt.Rid)
		return fmt.Errorf("account: %s's vendor not support, vendor: %s", accountID, vendor)
	}
	if err != nil {
		logs.Errorf("sync %s failed, err: %v, rid: %s", resType, err, kt.Rid)
		return fmt.Errorf("sync %s failed, err: %v", resType, err)
	}

	return nil
}
