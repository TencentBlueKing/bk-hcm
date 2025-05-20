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
	"hcm/cmd/cloud-server/service/sync/other"
	"hcm/cmd/cloud-server/service/sync/tcloud"
	"hcm/pkg/api/core"
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

	syncer, ok := vendorSyncerMap[vendor]
	if !ok {
		return fmt.Errorf("vendor: %s not support", vendor)
	}

	isNeedSyncPublicResFlag, err := isNeedSyncPublicResource(kt, cli.DataService(), syncer)
	if err != nil {
		logs.Errorf("is need sync public resource failed, err: %v, vendor: %s, rid: %s",
			err, vendor, kt.Rid)
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

		resType, err := syncer.SyncAllResource(kt, cli, accountID, isNeedSyncPublicResFlag)
		if err != nil {
			logs.Errorf("[%s] sync account %s failed on %s, err: %v, rid: %s", vendor, accountID, resType, err, kt.Rid)
		}

	}(leaseID)

	return nil
}

// check is there any tree types of public resources, if one of that type does not exist, we sync all public resources
func isNeedSyncPublicResource(kt *kit.Kit, dataCli *dataservice.Client, syncer VendorSyncer) (
	bool, error) {

	// 1. check region count
	regionNum, err := syncer.CountRegion(kt, dataCli)
	if err != nil {
		return false, err
	}
	// need sync if no region found
	if regionNum == 0 {
		return true, nil
	}

	// 2. check zone list
	zoneNum, err := syncer.CountZone(kt, dataCli)
	if err != nil {
		return false, err
	}
	if zoneNum == 0 {
		return true, nil
	}

	// 3. check image list
	imageNum, err := syncer.CountImage(kt, dataCli)
	if err != nil {
		return false, err
	}
	if imageNum == 0 {
		return true, nil
	}

	return false, nil
}

// VendorSyncer ...
type VendorSyncer interface {
	Vendor() enumor.Vendor
	CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error)
	CountZone(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error)
	CountImage(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error)
	SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string,
		syncPubRes bool) (resType enumor.CloudResourceType, err error)
}

type generalSyncer struct {
	vendor enumor.Vendor
}

func newAzureSyncer() azureSyncer {
	return azureSyncer{generalSyncer{vendor: enumor.Azure}}
}

func newGcpSyncer() gcpSyncer {
	return gcpSyncer{generalSyncer{vendor: enumor.Gcp}}
}

func newHuaweiSyncer() huaweiSyncer {
	return huaweiSyncer{generalSyncer{vendor: enumor.HuaWei}}
}

func newAwsSyncer() awsSyncer {
	return awsSyncer{generalSyncer{vendor: enumor.Aws}}
}

func newTCloudSyncer() tcloudSyncer {
	return tcloudSyncer{generalSyncer{vendor: enumor.TCloud}}
}

func newOtherSyncer() otherSyncer {
	return otherSyncer{generalSyncer{vendor: enumor.Other}}
}

// GetAvailableVendorSyncers ...
func GetAvailableVendorSyncers() []VendorSyncer {
	return availableVendorSyncer
}

var availableVendorSyncer = []VendorSyncer{
	newTCloudSyncer(),
	newAwsSyncer(),
	newHuaweiSyncer(),
	newGcpSyncer(),
	newAzureSyncer(),
	newOtherSyncer(),
}

// vendorSyncerMap
var vendorSyncerMap = map[enumor.Vendor]VendorSyncer{
	enumor.TCloud: newTCloudSyncer(),
	enumor.Aws:    newAwsSyncer(),
	enumor.HuaWei: newHuaweiSyncer(),
	enumor.Gcp:    newGcpSyncer(),
	enumor.Azure:  newAzureSyncer(),
	enumor.Other:  newOtherSyncer(),
}

// CountZone ...
func (c generalSyncer) CountZone(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	listZoneReq := &protocloud.ZoneListReq{
		Filter: tools.EqualExpression("vendor", c.vendor),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.Global.Zone.ListZone(kt.Ctx, kt.Header(), listZoneReq)
	if err != nil {
		return 0, err
	}

	return result.Count, nil
}

// CountImage ...
func (c generalSyncer) CountImage(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	listZoneReq := &core.ListReq{
		Filter: tools.EqualExpression("vendor", c.Vendor()),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.Global.ListImage(kt, listZoneReq)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// Vendor ...
func (c generalSyncer) Vendor() enumor.Vendor {
	return c.vendor
}

// tcloudSyncer ...
type tcloudSyncer struct {
	generalSyncer
}

// CountRegion ...
func (t tcloudSyncer) CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.TCloud.Region.ListRegion(kt.Ctx, kt.Header(), req)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// SyncAllResource ...
func (t tcloudSyncer) SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string,
	syncPubRes bool) (reType enumor.CloudResourceType, err error) {

	opt := &tcloud.SyncAllResourceOption{
		AccountID:          account,
		SyncPublicResource: syncPubRes,
	}
	return tcloud.SyncAllResource(kt, cli, opt)
}

// awsSyncer ...
type awsSyncer struct {
	generalSyncer
}

// CountRegion ...
func (t awsSyncer) CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.Aws.Region.ListRegion(kt.Ctx, kt.Header(), req)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// SyncAllResource ...
func (t awsSyncer) SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string,
	syncPubRes bool) (reType enumor.CloudResourceType, err error) {

	opt := &aws.SyncAllResourceOption{
		AccountID:          account,
		SyncPublicResource: syncPubRes,
	}
	return aws.SyncAllResource(kt, cli, opt)
}

// huaweiSyncer ...
type huaweiSyncer struct {
	generalSyncer
}

// CountRegion ...
func (t huaweiSyncer) CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.HuaWei.Region.ListRegion(kt.Ctx, kt.Header(), req)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// SyncAllResource ...
func (t huaweiSyncer) SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string,
	syncPubRes bool) (reType enumor.CloudResourceType, err error) {

	opt := &huawei.SyncAllResourceOption{
		AccountID:          account,
		SyncPublicResource: syncPubRes,
	}
	return huawei.SyncAllResource(kt, cli, opt)
}

// gcpSyncer ...
type gcpSyncer struct {
	generalSyncer
}

// CountRegion ...
func (t gcpSyncer) CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.Gcp.Region.ListRegion(kt.Ctx, kt.Header(), req)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// SyncAllResource ...
func (t gcpSyncer) SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string,
	syncPubRes bool) (reType enumor.CloudResourceType, err error) {

	opt := &gcp.SyncAllResourceOption{
		AccountID:          account,
		SyncPublicResource: syncPubRes,
	}
	return gcp.SyncAllResource(kt, cli, opt)
}

// azureSyncer ...
type azureSyncer struct {
	generalSyncer
}

// CountRegion ...
func (t azureSyncer) CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	req := &core.ListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewCountPage(),
	}
	result, err := dataCli.Azure.Region.ListRegion(kt.Ctx, kt.Header(), req)
	if err != nil {
		return 0, err
	}
	return result.Count, nil
}

// CountZone ...
func (t azureSyncer) CountZone(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	// azure没有可用区, 按有资源处理
	return 1, nil
}

// SyncAllResource ...
func (t azureSyncer) SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string,
	syncPubRes bool) (reType enumor.CloudResourceType, err error) {

	opt := &azure.SyncAllResourceOption{
		AccountID:          account,
		SyncPublicResource: syncPubRes,
	}
	return azure.SyncAllResource(kt, cli, opt)
}

// otherSyncer ...
type otherSyncer struct {
	generalSyncer
}

// CountRegion ...
func (t otherSyncer) CountRegion(kt *kit.Kit, dataCli *dataservice.Client) (uint64, error) {
	// 其他云厂商暂没有地域, 按有资源处理
	return 1, nil
}

// SyncAllResource ...
func (t otherSyncer) SyncAllResource(kt *kit.Kit, cli *client.ClientSet, account string, syncPubRes bool) (
	reType enumor.CloudResourceType, err error) {

	opt := &other.SyncAllResourceOption{
		AccountID: account,
	}
	return other.SyncAllResource(kt, cli, opt)
}
