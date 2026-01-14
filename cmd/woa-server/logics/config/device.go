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
	"errors"

	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
	utils "hcm/pkg/tools/util"
)

// DeviceIf provides management interface for operations of device config
type DeviceIf interface {
	// GetDeviceWithCapacity get device config list with enable_capacity
	GetDeviceWithCapacity(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceInfoResult, error)
	// GetDevice get device config list
	GetDevice(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceInfoResult, error)
	// GetCvmDeviceDetail get cvm device detail config list
	GetCvmDeviceDetail(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceInfoResult, error)
	// GetDeviceType gets config device type list
	GetDeviceType(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceTypeResult, error)
	// GetDeviceTypeDetail gets config device type with detail info
	GetDeviceTypeDetail(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceTypeDetailResult, error)
	// CreateDevice creates device config
	CreateDevice(kt *kit.Kit, input *types.DeviceInfo) (mapstr.MapStr, error)
	// CreateManyDevice creates device config in batch
	CreateManyDevice(kt *kit.Kit, input *types.CreateManyDeviceParam) error
	// UpdateDevice updates device config
	UpdateDevice(kt *kit.Kit, instId int64, input map[string]interface{}) error
	// UpdateDeviceBatch updates device config in batch
	UpdateDeviceBatch(kt *kit.Kit, cond, update map[string]interface{}) error
	// DeleteDevice deletes device config
	DeleteDevice(kt *kit.Kit, instId int64) error

	// GetDvmDeviceType gets config dvm device type list
	GetDvmDeviceType(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDvmDeviceRst, error)
	// CreateDvmDevice creates config dvm device type
	CreateDvmDevice(kt *kit.Kit, input *types.DvmDeviceInfo) (mapstr.MapStr, error)

	// GetPmDeviceType gets config physical machine device type list
	GetPmDeviceType(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetPmDeviceRst, error)
	// CreatePmDevice creates config physical machine device type
	CreatePmDevice(kt *kit.Kit, input *types.PmDeviceInfo) (mapstr.MapStr, error)

	// ListCvmInstanceInfoByDeviceTypes list cvm instance info by device types
	ListCvmInstanceInfoByDeviceTypes(kt *kit.Kit, deviceTypes []string) (map[string]types.DeviceTypeCpuItem, error)
	ListInstanceGroup(kt *kit.Kit, deviceTypes []string) (map[string]string, error)

	ListDeviceTypeInfoFromCrp(kt *kit.Kit, deviceTypes []string) (map[string]cvmapi.QueryCvmInstanceTypeItem, error)
}

// NewDeviceOp creates a device interface
func NewDeviceOp(thirdCli *thirdparty.Client, zoneOp ZoneIf) DeviceIf {
	return &device{
		cvm:    thirdCli.CVM,
		zoneOp: zoneOp,
	}
}

type device struct {
	cvm    cvmapi.CVMClientInterface
	zoneOp ZoneIf
}

// GetDeviceWithCapacity get device config list with enable_capacity
func (d *device) GetDeviceWithCapacity(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceInfoResult, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config device detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// get devices with enable capacity only
	filter["enable_capacity"] = true

	count, err := config.Operation().CvmDevice().CountDevice(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := config.Operation().CvmDevice().FindManyDevice(kt.Ctx, input.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetDeviceInfoResult{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetDevice get device config list
func (d *device) GetDevice(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceInfoResult, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config device detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// get devices with enable apply only
	filter["enable_apply"] = true

	count, err := config.Operation().CvmDevice().CountDevice(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := config.Operation().CvmDevice().FindManyDevice(kt.Ctx, input.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetDeviceInfoResult{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// GetCvmDeviceDetail get cvm device detail config list
func (d *device) GetCvmDeviceDetail(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceInfoResult, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config device detail failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetDeviceInfoResult{}
	if input.Page.EnableCount {
		cnt, err := config.Operation().CvmDevice().CountDevice(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get device detail count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*types.DeviceInfo, 0)
		return rst, nil
	}

	insts, err := config.Operation().CvmDevice().FindManyDevice(kt.Ctx, input.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetDeviceType gets config device type list
func (d *device) GetDeviceType(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceTypeResult, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config device type failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// get devices with enable apply only
	filter["enable_apply"] = true

	insts, err := config.Operation().CvmDevice().FindManyDeviceType(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	// 如果没有数据，直接返回
	if len(insts) == 0 {
		return &types.GetDeviceTypeResult{}, nil
	}

	instTypes := make([]string, 0)
	for _, inst := range insts {
		instStr := utils.GetStrByInterface(inst)
		instTypes = append(instTypes, instStr)
	}
	// instTypes为空时，CRP接口会返回全部机型列表，导致前端传入不符合条件的机型，也会返回全部机型
	req := &cvmapi.QueryCvmInstanceTypeReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.QueryCvmInstanceType,
		},
		Params: &cvmapi.QueryCvmInstanceTypeParams{InstanceType: instTypes},
	}

	resp, err := d.cvm.QueryCvmInstanceType(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("query cvm instance type failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	// 记录日志
	logs.Infof("get config device type, QueryCvmInstanceType, instTypes: %+v, rid: %s", instTypes, kt.Rid)

	infos := make([]types.DeviceTypeItem, 0)
	for _, item := range resp.Result.Data {
		infos = append(infos, types.DeviceTypeItem{
			DeviceType:      item.InstanceType,
			DeviceTypeClass: item.InstanceTypeClass,
			DeviceGroup:     item.InstanceGroup,
			CPUAmount:       item.CPUAmount,
			RamAmount:       item.RamAmount,
			CoreType:        item.CoreType,
			DeviceClass:     item.InstanceClassDesc,
			TechnicalClass:  item.CvmInstanceTypeClass,
		})
	}

	rst := &types.GetDeviceTypeResult{
		Count: int64(len(infos)),
		Info:  infos,
	}

	return rst, nil
}

// GetDeviceTypeDetail gets config device type with detail info
func (d *device) GetDeviceTypeDetail(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetDeviceTypeDetailResult,
	error) {

	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config device type failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// get devices with enable apply only
	filter["enable_apply"] = true

	insts, err := config.Operation().CvmDevice().FindManyDevice(kt.Ctx, input.Page, filter)
	if err != nil {
		return nil, err
	}

	deviceTypeMap := make(map[string]*types.DeviceInfo)
	for _, inst := range insts {
		deviceTypeMap[inst.DeviceType] = inst
	}

	instDetails := make([]*types.DeviceTypeInfo, 0)
	for _, deviceType := range deviceTypeMap {
		deviceGroup := deviceType.Label["device_group"]
		deviceGroupStr, ok := deviceGroup.(string)
		if !ok {
			deviceGroupStr = ""
		}

		instDetails = append(instDetails, &types.DeviceTypeInfo{
			DeviceType:  deviceType.DeviceType,
			DeviceGroup: deviceGroupStr,
		})
	}

	rst := &types.GetDeviceTypeDetailResult{
		Count: int64(len(instDetails)),
		Info:  instDetails,
	}

	return rst, nil
}

// CreateDevice creates device config
func (d *device) CreateDevice(kt *kit.Kit, input *types.DeviceInfo) (mapstr.MapStr, error) {
	// uniqueness check
	filter := map[string]interface{}{
		"require_type": input.RequireType,
		"region":       input.Region,
		"zone":         input.Zone,
		"device_type":  input.DeviceType,
	}

	cnt, err := config.Operation().CvmDevice().CountDevice(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to count device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if cnt != 0 {
		logs.Errorf("device exist locally, need not create again, rid: %s", kt.Rid)
		return nil, errf.NewFromErr(errf.DeviceTypeExistLocally,
			errors.New("device exist locally, need not create again"))
	}

	// create instance
	id, err := config.Operation().CvmDevice().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().CvmDevice().CreateDevice(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// CreateManyDevice creates device config in batch
func (d *device) CreateManyDevice(kt *kit.Kit, input *types.CreateManyDeviceParam) error {
	zoneResult, err := d.zoneOp.GetZone(kt, &types.GetZoneParam{
		Zone: input.Zone,
	})
	if err != nil {
		logs.Errorf("failed to get zones, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	zones := zoneResult.Info

	for _, requireType := range input.RequireType {
		param := &types.DeviceInfo{
			DeviceType: input.DeviceType,
			Cpu:        input.Cpu,
			Mem:        input.Mem,
			// set disk 100 as default
			Disk:   100,
			Remark: input.Remark,
			Label: mapstr.MapStr{
				"device_group":    input.DeviceGroup,
				"device_size":     input.DeviceSize,
				"technical_class": input.TechnicalClass,
			},
			EnableCapacity:  true,
			EnableApply:     true,
			Score:           0,
			Comment:         "",
			DeviceTypeClass: input.DeviceTypeClass,
			RequireType:     requireType,
		}

		if err = d.batchCreateDeviceForZones(kt, param, zones); err != nil {
			logs.Errorf("failed to create device, err: %v, param: %+v, rid: %s", err, cvt.PtrToVal(param), kt.Rid)
			return err
		}
	}

	return nil
}

// batchCreateDeviceForZones batch creates device config for zone
func (d *device) batchCreateDeviceForZones(kt *kit.Kit, input *types.DeviceInfo, zones []*types.Zone) error {

	// uniqueness check
	zoneConditions := make([]string, 0, len(zones))
	for _, item := range zones {
		zoneConditions = append(zoneConditions, item.Zone)
	}
	filter := map[string]interface{}{
		"require_type": input.RequireType,
		"device_type":  input.DeviceType,
		"zone": &mapstr.MapStr{
			pkg.BKDBIN: zoneConditions,
		},
	}

	page := metadata.BasePage{Limit: pkg.BKNoLimit, Start: 0}
	resp, err := config.Operation().CvmDevice().FindManyDevice(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to find device, err: %v, filter: %v, rid: %s", err, filter, kt.Rid)
		return err
	}
	// 删除db中已有的数据，并在后面重新创建
	if len(resp) > 0 {
		logs.Warnf("the device info exist more than one, device info: %+v, rid: %s", cvt.PtrToSlice(resp), kt.Rid)
		ids := make([]int64, 0, len(resp))
		for _, item := range resp {
			ids = append(ids, item.BkInstId)
		}
		delFilter := &mapstr.MapStr{
			"id": &mapstr.MapStr{
				pkg.BKDBIN: ids,
			},
		}
		if err = config.Operation().CvmDevice().DeleteDevice(kt.Ctx, delFilter); err != nil {
			logs.Errorf("failed to delete device, err: %v, filter: %v, rid: %s", err, filter, kt.Rid)
			return err
		}
	}

	// create instance
	devices := make([]types.DeviceInfo, 0, len(zones))
	ids, err := config.Operation().CvmDevice().NextSequences(kt.Ctx, len(zones))
	if err != nil {
		logs.Errorf("failed to create device, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	for i, item := range zones {
		input.Zone = item.Zone
		input.Region = item.Region
		input.BkInstId = int64(ids[i])
		devices = append(devices, *input)
	}

	if err := config.Operation().CvmDevice().BatchCreateDevices(kt.Ctx, cvt.SliceToPtr(devices)); err != nil {
		logs.Errorf("failed to create device, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UpdateDevice updates device config
func (d *device) UpdateDevice(kt *kit.Kit, instId int64, input map[string]interface{}) error {
	filter := map[string]interface{}{
		"id": instId,
	}

	if err := config.Operation().CvmDevice().UpdateDevice(kt.Ctx, filter, input); err != nil {
		logs.Errorf("failed to update device, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UpdateDeviceBatch updates device config in batch
func (d *device) UpdateDeviceBatch(kt *kit.Kit, cond, update map[string]interface{}) error {
	if err := config.Operation().CvmDevice().UpdateDevice(kt.Ctx, cond, update); err != nil {
		logs.Errorf("failed to update device, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteDevice deletes device config
func (d *device) DeleteDevice(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().CvmDevice().DeleteDevice(kt.Ctx, filter); err != nil {
		logs.Errorf("failed to delete device, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
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
func (d *device) GetPmDeviceType(kt *kit.Kit, input *types.GetDeviceParam) (*types.GetPmDeviceRst, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("get config physical machine device type failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	count, err := config.Operation().PmDevice().CountDevice(kt.Ctx, filter)
	if err != nil {
		return nil, err
	}

	insts, err := config.Operation().PmDevice().FindManyDevice(kt.Ctx, input.Page, filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetPmDeviceRst{
		Count: int64(count),
		Info:  insts,
	}

	return rst, nil
}

// CreatePmDevice creates config physical machine device type
func (d *device) CreatePmDevice(kt *kit.Kit, input *types.PmDeviceInfo) (mapstr.MapStr, error) {
	id, err := config.Operation().PmDevice().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create physical machine device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().PmDevice().CreateDevice(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create physical machine device, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// ListCvmInstanceInfoByDeviceTypes list cvm instance info by device types
func (d *device) ListCvmInstanceInfoByDeviceTypes(kt *kit.Kit, deviceTypes []string) (
	map[string]types.DeviceTypeCpuItem, error) {

	deviceReq := &types.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "device_type",
						Operator: querybuilder.OperatorIn,
						Value:    deviceTypes,
					}},
			},
		},
		Page: metadata.BasePage{Limit: pkg.BKNoLimit, Start: 0},
	}
	deviceList, err := d.GetDevice(kt, deviceReq)
	if err != nil {
		logs.Errorf("get device list from mongo failed, err: %v, deviceTypes: %v, rid: %s", err, deviceTypes, kt.Rid)
		return nil, err
	}

	deviceTypeMap := make(map[string]types.DeviceTypeCpuItem, 0)
	existDeviceTypes := make([]string, 0)
	for _, deviceItem := range deviceList.Info {
		deviceType := deviceItem.DeviceType

		// 机型族
		deviceGroup, ok := deviceItem.Label["device_group"]
		if !ok {
			return nil, errors.New("get invalid empty device group")
		}
		deviceGroupStr, ok := deviceGroup.(string)
		if !ok {
			return nil, errors.New("get invalid non-string device group")
		}

		// 机型核心类型
		deviceSize, ok := deviceItem.Label["device_size"]
		if !ok {
			return nil, errors.New("get invalid empty device size")
		}
		deviceSizeStr, ok := deviceSize.(string)
		if !ok {
			return nil, errors.New("get invalid non-string device size")
		}

		// 技术分类
		technicalClass, ok := deviceItem.Label["technical_class"].(string)
		if !ok {
			// TODO 由于存量的机型可能没有技术分类，所以这里只打印错误日志，防止影响上层调用逻辑
			logs.Errorf("get invalid technical class, deviceType: %s, rid: %s", deviceType, kt.Rid)
		}

		deviceTypeMap[deviceType] = types.DeviceTypeCpuItem{
			DeviceType:     deviceItem.DeviceType,
			CPUAmount:      deviceItem.Cpu,
			DeviceGroup:    deviceGroupStr,
			CoreType:       enumor.CoreType(deviceSizeStr),
			TechnicalClass: technicalClass,
		}
		existDeviceTypes = append(existDeviceTypes, deviceType)
	}

	// 如果查到了全部的DeviceType，则直接返回
	if len(deviceList.Info) == len(deviceTypes) {
		logs.Infof("get cvm instance info from mongo by device types, deviceTypes: %v, deviceTypeMap: %+v, rid: %s",
			deviceTypes, deviceTypeMap, kt.Rid)
		return deviceTypeMap, nil
	}

	notExistDevice := make([]string, 0)
	for _, dtype := range deviceTypes {
		if !slice.IsItemInSlice(existDeviceTypes, dtype) {
			notExistDevice = append(notExistDevice, dtype)
		}
	}
	if len(notExistDevice) == 0 {
		logs.Infof("get cvm instance info from params by device types, deviceTypes: %v, deviceTypeMap: %+v, rid: %s",
			deviceTypes, deviceTypeMap, kt.Rid)
		return deviceTypeMap, nil
	}

	deviceTypeMapFromCrp, err := d.listCvmInstanceTypeFromCrp(kt, notExistDevice)
	if err != nil {
		logs.Errorf("list cvm instance type from crp failed, err: %v, deviceTypes: %v, rid: %s",
			err, notExistDevice, kt.Rid)
		return nil, err
	}
	for dtype, item := range deviceTypeMapFromCrp {
		deviceTypeMap[dtype] = item
	}

	// 记录日志
	logs.Infof("get cvm instance info from crp by device types, deviceTypes: %v, notExistDevice: %v, "+
		"deviceTypeMap: %+v, rid: %s", deviceTypes, notExistDevice, deviceTypeMap, kt.Rid)

	return deviceTypeMap, nil
}

// listCvmInstanceTypeFromCrp 从Crp平台获取实例信息
func (d *device) listCvmInstanceTypeFromCrp(kt *kit.Kit, deviceTypes []string) (map[string]types.DeviceTypeCpuItem,
	error) {

	req := &cvmapi.QueryCvmInstanceTypeReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.QueryCvmInstanceType,
		},
		Params: &cvmapi.QueryCvmInstanceTypeParams{InstanceType: deviceTypes},
	}

	resp, err := d.cvm.QueryCvmInstanceType(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("query cvm instance type failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if resp.Result == nil {
		logs.Errorf("query cvm instance type error, deviceTypes: %v, resp: %+v, rid: %s", deviceTypes, resp, kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "query cvm instance type failed, resp:[%+v] is nil", resp)
	}

	deviceTypeMap := make(map[string]types.DeviceTypeCpuItem, 0)
	for _, item := range resp.Result.Data {
		if _, ok := deviceTypeMap[item.InstanceType]; ok {
			continue
		}
		deviceTypeMap[item.InstanceType] = types.DeviceTypeCpuItem{
			DeviceType:  item.InstanceType,
			CPUAmount:   int64(item.CPUAmount),
			DeviceGroup: item.InstanceGroup,
			CoreType:    enumor.GetCoreTypeByCRPCoreTypeID(item.CoreType),
		}
	}

	// 记录日志
	logs.Infof("get yunti crp device type, QueryCvmInstanceType, instTypes: %v, deviceTypeMap; %+v, rid: %s",
		deviceTypes, deviceTypeMap, kt.Rid)

	return deviceTypeMap, nil
}

// ListInstanceGroup 获取设备机型族
func (d *device) ListInstanceGroup(kt *kit.Kit, deviceTypes []string) (map[string]string, error) {
	req := &cvmapi.QueryCvmInstanceTypeReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.QueryCvmInstanceType,
		},
		Params: &cvmapi.QueryCvmInstanceTypeParams{InstanceType: deviceTypes},
	}

	resp, err := d.cvm.QueryCvmInstanceType(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("query cvm instance group failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if resp.Result == nil {
		logs.Errorf("query cvm instance group error, deviceTypes: %v, resp: %+v, rid: %s", deviceTypes, resp, kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "query cvm instance group failed, resp:[%+v] is nil", resp)
	}

	instGroupMap := make(map[string]string, 0)
	for _, item := range resp.Result.Data {
		instGroupMap[item.InstanceType] = item.InstanceGroup
	}

	// 记录日志
	logs.Infof("get yunti crp instance group type, QueryCvmInstanceType, instTypes: %v, instGroupMap; %+v, rid: %s",
		deviceTypes, instGroupMap, kt.Rid)

	return instGroupMap, nil
}

// ListDeviceTypeInfoFromCrp 从crp获取设备机型信息
func (d *device) ListDeviceTypeInfoFromCrp(kt *kit.Kit, deviceTypes []string) (
	map[string]cvmapi.QueryCvmInstanceTypeItem, error) {

	req := &cvmapi.QueryCvmInstanceTypeReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.QueryCvmInstanceType,
		},
		Params: &cvmapi.QueryCvmInstanceTypeParams{InstanceType: deviceTypes},
	}

	resp, err := d.cvm.QueryCvmInstanceType(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("query cvm instance type failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, err
	}
	if resp.Result == nil {
		logs.Errorf("query cvm instance type error, deviceTypes: %v, resp: %+v, rid: %s", deviceTypes, resp, kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "query cvm instance type failed, resp:[%+v]", resp)
	}

	deviceTypeInfos := make(map[string]cvmapi.QueryCvmInstanceTypeItem, len(resp.Result.Data))
	for _, item := range resp.Result.Data {
		deviceTypeInfos[item.InstanceType] = item
	}

	return deviceTypeInfos, nil
}
