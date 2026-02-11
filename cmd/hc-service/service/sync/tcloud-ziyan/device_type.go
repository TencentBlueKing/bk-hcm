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

package ziyan

import (
	"fmt"

	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/ziyan"
	"hcm/cmd/hc-service/service/sync/handler"
	"hcm/pkg/api/core"
	coredevicetype "hcm/pkg/api/core/cloud/device-type"
	"hcm/pkg/api/data-service/cloud/zone"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/cc"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/cvmapi"
)

// SyncDeviceType 同步机型接口
func (svc *service) SyncDeviceType(cts *rest.Contexts) (interface{}, error) {
	return nil, handler.ResourceSyncV2[coredevicetype.DeviceType](cts, &deviceTypeHandler{
		cli:     svc.syncCli,
		dataCli: svc.dataCli,
		crpCli:  svc.crpCli,
	})
}

// deviceTypeHandler device type sync handler.
type deviceTypeHandler struct {
	cli ressync.Interface

	// Prepare 构建参数
	request *sync.TCloudSyncReq
	syncCli ziyan.Interface
	dataCli *dataservice.Client
	crpCli  cvmapi.CVMClientInterface

	// 是否已经获取所有机型数据
	isGetAllDeviceTypes bool
}

var _ handler.HandlerV2[coredevicetype.DeviceType] = new(deviceTypeHandler)

// Prepare ...
func (hd *deviceTypeHandler) Prepare(cts *rest.Contexts) error {
	req := new(sync.TCloudSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := hd.cli.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return err
	}

	hd.request = req
	hd.syncCli = syncCli

	return nil
}

// Next 返回需要同步的机型实例列表
func (hd *deviceTypeHandler) Next(kt *kit.Kit) ([]coredevicetype.DeviceType, error) {
	if hd.isGetAllDeviceTypes {
		return make([]coredevicetype.DeviceType, 0), nil
	}

	region := hd.request.Region

	zones, err := hd.listZones(kt, region)
	if err != nil {
		logs.Errorf("list zones failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return nil, err
	}

	instances, err := hd.listDeviceTypeFromCloud(kt, region, zones)
	if err != nil {
		logs.Errorf("list device type from cloud failed, err: %v, region: %s, zones: %v, rid: %s",
			err, region, zones, kt.Rid)
		return nil, err
	}

	hd.isGetAllDeviceTypes = true

	return instances, nil
}

// listZones 分页获取指定地域的可用区
func (hd *deviceTypeHandler) listZones(kt *kit.Kit, region string) ([]string, error) {
	req := &zone.ZoneListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan), tools.RuleEqual("region", region)),
		Page:   core.NewDefaultBasePage(),
	}
	zones := make([]string, 0)

	for {
		result, err := hd.dataCli.Global.Zone.ListZone(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list zones failed, err: %v, region: %s, rid: %s", err, region, kt.Rid)
			return nil, err
		}
		for _, zone := range result.Details {
			zones = append(zones, zone.Name)
		}
		if len(result.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return zones, nil
}

// listDeviceTypeFromCloud 从云上获取机型数据
func (hd *deviceTypeHandler) listDeviceTypeFromCloud(kt *kit.Kit, region string, zones []string) (
	[]coredevicetype.DeviceType, error) {

	deviceTypes := make([]coredevicetype.DeviceType, 0)
	for _, zoneName := range zones {
		// 1. 调用 GetInstanceTypeInfo 获取可用区下的机型列表
		getInfoParams := &cvmapi.GetInstanceTypeInfoParams{
			DeptId: cvmapi.CvmDeptId,
			Zone:   zoneName,
		}
		infoResp, err := hd.crpCli.GetInstanceTypeInfo(kt, getInfoParams)
		if err != nil {
			logs.Errorf("get instance type info failed, err: %v, zone: %s, rid: %s", err, zoneName, kt.Rid)
			return nil, err
		}
		if infoResp == nil || infoResp.Result == nil {
			logs.Errorf("get instance type info result is nil, zone: %s, rid: %s", zoneName, kt.Rid)
			return nil, fmt.Errorf("get instance type info result is nil, zone: %s, rid: %s", zoneName, kt.Rid)
		}
		// 收集所有机型名称
		instanceTypes := make([]string, 0)
		for _, item := range infoResp.Result.InstanceTypes {
			if item.CvmInstanceModel != "" {
				instanceTypes = append(instanceTypes, item.CvmInstanceModel)
			}
		}
		if len(instanceTypes) == 0 {
			continue
		}

		// 2. 调用 QueryCvmInstanceType 获取机型详细信息
		queryParams := &cvmapi.QueryCvmInstanceTypeParams{
			InstanceType: instanceTypes,
		}
		queryResp, err := hd.crpCli.QueryCvmInstanceType(kt, queryParams)
		if err != nil {
			logs.Errorf("query cvm instance type failed, err: %v, zone: %s, rid: %s", err, zoneName, kt.Rid)
			return nil, err
		}
		if queryResp == nil || queryResp.Result == nil {
			logs.Errorf("query cvm instance type result is nil, zone: %s, rid: %s", zoneName, kt.Rid)
			return nil, fmt.Errorf("query cvm instance type result is nil, zone: %s, rid: %s", zoneName, kt.Rid)
		}

		// 3. 构建机型数据，构造完整的 device_type 表字段
		for _, item := range queryResp.Result.Data {
			deviceTypes = append(deviceTypes, coredevicetype.DeviceType{
				Vendor:          enumor.TCloudZiyan,
				Region:          region,
				Zone:            zoneName,
				DeviceType:      item.InstanceType,
				DeviceTypeClass: item.InstanceTypeClass,
				DeviceClass:     item.InstanceClassDesc,
				DeviceFamily:    item.InstanceGroup,
				CoreType:        enumor.GetCoreTypeByCRPCoreTypeID(item.CoreType),
				CpuCore:         int64(item.CPUAmount),
				Memory:          int64(item.RamAmount),
				TechnicalClass:  item.CvmInstanceTypeClass,
				Source:          enumor.DeviceTypeSourceSync,
			})
		}
	}

	return deviceTypes, nil
}

// Sync 同步机型数据
func (hd *deviceTypeHandler) Sync(kt *kit.Kit, instances []coredevicetype.DeviceType) error {
	if len(instances) == 0 {
		return nil
	}

	params := &ziyan.SyncDeviceTypeParams{
		Region:      hd.request.Region,
		DeviceTypes: instances,
	}
	if _, err := hd.syncCli.DeviceType(kt, params); err != nil {
		logs.Errorf("sync tcloud ziyan device type failed, err: %v, region: %s, rid: %s", err, hd.request.Region,
			kt.Rid)
		return err
	}

	return nil
}

// RemoveDeletedFromCloud 删除云上已不存在的机型
func (hd *deviceTypeHandler) RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error {
	params := &ziyan.SyncRemovedParams{
		AccountID: hd.request.AccountID,
		Region:    hd.request.Region,
	}
	if err := hd.syncCli.RemoveDeviceTypeDeleteFromCloud(kt, params, allCloudIDMap); err != nil {
		logs.Errorf("remove tcloud ziyan device type delete from cloud failed, err: %v, region: %s, rid: %s", err,
			hd.request.Region, kt.Rid)
		return err
	}
	return nil
}

// SyncConcurrent 支持的并发数
func (hd *deviceTypeHandler) SyncConcurrent() uint {
	if hd.request != nil && hd.request.Concurrent != 0 {
		return hd.request.Concurrent
	}
	// read from config file
	_, syncing := cc.HCService().SyncConfig.GetSyncConcurrent(enumor.Ziyan, enumor.DeviceType, cc.ConcurrentWildcard)
	return max(syncing, 1)
}

// Describe 描述信息
func (hd *deviceTypeHandler) Describe() string {
	if hd.request == nil {
		return fmt.Sprintf("ziyan %s(-)", hd.Resource())
	}
	return fmt.Sprintf("ziyan %s(region=%s,account=%s)", hd.Resource(), hd.request.Region, hd.request.AccountID)
}

// Resource ...
func (hd *deviceTypeHandler) Resource() enumor.CloudResourceType {
	return enumor.DeviceType
}
