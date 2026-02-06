/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	datapmdevicetype "hcm/pkg/api/data-service/tcloud-ziyan-pm-device-type"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/slice"
)

// DeviceIf provides management interface for operations of device config
type DeviceIf interface {
	// ListDeviceType list device type
	ListDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeListReq) (*protocloud.DeviceTypeListResult, error)
	// ListDistinctDeviceType list distinct device type
	ListDistinctDeviceType(kt *kit.Kit, req *protocloud.DistinctDeviceTypeListReq) (
		*protocloud.DistinctDeviceTypeListResult, error)
	// BatchCreateDeviceType batch create device type
	BatchCreateDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeBatchCreateReq) (*core.BatchCreateResult, error)
	// BatchUpdateDeviceType batch update device type
	BatchUpdateDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeBatchUpdateReq) error
	// BatchDeleteDeviceType batch delete device type
	BatchDeleteDeviceType(kt *kit.Kit, req *dataproto.BatchDeleteReq) error
	// GetDvmDeviceType gets config dvm device type list
	GetDvmDeviceType(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDvmDeviceRst, error)
	// CreateDvmDevice creates config dvm device type
	CreateDvmDevice(kt *kit.Kit, input *types.DvmDeviceInfo) (mapstr.MapStr, error)

	// GetPmDeviceType gets config physical machine device type list
	GetPmDeviceType(kt *kit.Kit, input *types.GetPmDeviceTypeParam) (*types.GetPmDeviceRst, error)
	// CreatePmDevice creates config physical machine device type
	CreatePmDevice(kt *kit.Kit, req *datapmdevicetype.CreateTCloudZiyanPmDeviceTypeReq) (mapstr.MapStr, error)

	// ListCvmInstanceInfoByDeviceTypes list cvm instance info by device types
	ListCvmInstanceInfoByDeviceTypes(kt *kit.Kit, deviceTypes []string) (map[string]types.DeviceTypeCpuItem, error)
}

// NewDeviceOp creates a device interface
func NewDeviceOp(thirdCli *thirdparty.Client, zoneOp ZoneIf, clientSet *client.ClientSet) DeviceIf {
	return &device{
		cvm:       thirdCli.CVM,
		zoneOp:    zoneOp,
		clientSet: clientSet,
	}
}

type device struct {
	cvm       cvmapi.CVMClientInterface
	zoneOp    ZoneIf
	clientSet *client.ClientSet
}

// ListDeviceType list device type
func (d *device) ListDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeListReq) (
	*protocloud.DeviceTypeListResult, error) {

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list device type parameter, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return d.clientSet.DataService().TCloudZiyan.DeviceType.ListDeviceType(kt, req)
}

// ListDistinctDeviceType list distinct device type
func (d *device) ListDistinctDeviceType(kt *kit.Kit, req *protocloud.DistinctDeviceTypeListReq) (
	*protocloud.DistinctDeviceTypeListResult, error) {

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate list device type parameter, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return d.clientSet.DataService().TCloudZiyan.DeviceType.ListDistinctDeviceType(kt, req)
}

// BatchCreateDeviceType batch creates device type
func (d *device) BatchCreateDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeBatchCreateReq) (
	*core.BatchCreateResult, error) {

	if err := req.Validate(); err != nil {
		logs.Errorf("failed to validate create device type parameter, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	for i := range req.DeviceTypes {
		req.DeviceTypes[i].Source = enumor.DeviceTypeSourceManually
	}

	return d.clientSet.DataService().TCloudZiyan.DeviceType.BatchCreateDeviceType(kt, req)
}

// BatchUpdateDeviceType batch updates device type properties
func (d *device) BatchUpdateDeviceType(kt *kit.Kit, req *protocloud.DeviceTypeBatchUpdateReq) error {
	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return d.clientSet.DataService().TCloudZiyan.DeviceType.BatchUpdateDeviceType(kt, req)
}

// BatchDeleteDeviceType batch deletes device type
func (d *device) BatchDeleteDeviceType(kt *kit.Kit, req *dataproto.BatchDeleteReq) error {
	if err := req.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}
	return d.clientSet.DataService().Global.DeviceType.BatchDeleteDeviceType(kt, req)
}

// GetDvmDeviceType get dvm device config list
func (d *device) GetDvmDeviceType(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDvmDeviceRst, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config dvm device type failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	count, err := config.Operation().DvmDevice().CountDevice(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := config.Operation().DvmDevice().FindManyDevice(kt.Ctx, input.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetDvmDeviceRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// CreateDvmDevice creates config dvm device type
func (d *device) CreateDvmDevice(kt *kit.Kit, input *types.DvmDeviceInfo) (mapstr.MapStr, error) {
	id, err := config.Operation().DvmDevice().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create dvm device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().DvmDevice().CreateDevice(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create dvm device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// GetPmDeviceType get physical machine device config list
func (d *device) GetPmDeviceType(kt *kit.Kit, input *types.GetPmDeviceTypeParam) (*types.GetPmDeviceRst, error) {
	countReq := &core.ListReq{
		Filter: input.Filter,
		Page: &core.BasePage{
			Count: true,
		},
	}
	countResult, err := d.clientSet.DataService().Global.TCloudZiyanPmDeviceType.List(kt, countReq)
	if err != nil {
		logs.Errorf("failed to count pm device type, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	dataReq := &core.ListReq{
		Filter: input.Filter,
		Page:   input.Page,
	}
	dataResult, err := d.clientSet.DataService().Global.TCloudZiyanPmDeviceType.List(kt, dataReq)
	if err != nil {
		logs.Errorf("failed to list pm device type, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetPmDeviceRst{
		Count: int64(countResult.Count),
		Info:  dataResult.Details,
	}

	return rst, nil
}

// CreatePmDevice creates config physical machine device type
func (d *device) CreatePmDevice(kt *kit.Kit, req *datapmdevicetype.CreateTCloudZiyanPmDeviceTypeReq) (
	mapstr.MapStr, error) {

	result, err := d.clientSet.DataService().Global.TCloudZiyanPmDeviceType.Create(kt, req)
	if err != nil {
		logs.Errorf("failed to create physical machine device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(result.IDs) == 0 {
		return nil, errf.New(errf.InvalidParameter, "failed to create pm device type, no id returned")
	}

	rst := mapstr.MapStr{
		"ids": result.IDs,
	}

	return rst, nil
}

// ListCvmInstanceInfoByDeviceTypes list cvm instance info by device types
func (d *device) ListCvmInstanceInfoByDeviceTypes(kt *kit.Kit, deviceTypes []string) (
	map[string]types.DeviceTypeCpuItem, error) {

	deviceTypeMap := make(map[string]types.DeviceTypeCpuItem)
	for _, batch := range slice.Split(deviceTypes, int(core.DefaultMaxPageLimit)) {
		req := core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleEqual("vendor", enumor.TCloudZiyan),
				tools.RuleIn("device_type", batch)),
			Page: core.NewDefaultBasePage(),
		}
		params := &protocloud.DistinctDeviceTypeListReq{ListReq: req}
		resp, err := d.clientSet.DataService().TCloudZiyan.DeviceType.ListDistinctDeviceType(kt, params)
		if err != nil {
			logs.Errorf("failed to list distinct device type, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			deviceTypeMap[detail.DeviceType] = types.DeviceTypeCpuItem{
				DeviceType:      detail.DeviceType,
				CPUAmount:       detail.CpuCore,
				DeviceGroup:     detail.DeviceFamily,
				CoreType:        detail.CoreType,
				TechnicalClass:  detail.TechnicalClass,
				DeviceTypeClass: detail.DeviceTypeClass,
			}
		}
		if len(resp.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return deviceTypeMap, nil
}
