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

// Package config config
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	devicecapacity "hcm/pkg/api/data-service/device-capacity"
	woaserver "hcm/pkg/api/woa-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal/dao/tools"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	hcmjson "hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
	arrayutil "hcm/pkg/tools/util"
)

// CapacityIf provides management interface for operations of resource apply capacity
type CapacityIf interface {
	// GetCapacity gets resource apply capacity info
	GetCapacity(kt *kit.Kit, input *types.GetCapacityParam) (*types.GetCapacityRst, error)
	// UpsertCapacity upsert resource apply capacity info
	UpsertCapacity(kt *kit.Kit, input *types.UpdateCapacityParam) error
	// BatchGetCapacity 批量获取资源申请容量信息
	BatchGetCapacity(kt *kit.Kit, input *types.BatchGetCapacityParam) (*types.BatchGetCapacityRst, error)
	// ListCapacityWithDeviceInfo 查询设备库存及其机型详细信息
	ListCapacityWithDeviceInfo(kt *kit.Kit, input *devicecapacity.ListCapacityWithDeviceInfoReq) (
		*woaserver.ListCapacityWithDeviceInfoResult, error)
}

// NewCapacityOp creates a capacity interface
func NewCapacityOp(client *client.ClientSet, subnet SubnetIf, vpc VpcIf, thirdCli *thirdparty.Client, cmdbCli cmdb.Client) CapacityIf {
	return &capacity{
		client:  client,
		cvm:     thirdCli.OldCVM,
		vpc:     vpc,
		subnet:  subnet,
		cmdbCli: cmdbCli,
	}
}

type capacity struct {
	client  *client.ClientSet
	cvm     cvmapi.CVMClientInterface
	vpc     VpcIf
	subnet  SubnetIf
	cmdbCli cmdb.Client
}

func (c *capacity) GetCapacity(kt *kit.Kit, input *types.GetCapacityParam) (*types.GetCapacityRst, error) {
	// 1. query subnet from db
	subnetReq := &types.GetAllSubnetReq{Region: input.Region}
	if input.Zone != "" && input.Zone != cvmapi.CvmSeparateCampus {
		subnetReq.Zones = []string{input.Zone}
	}
	vpcID := input.Vpc
	if vpcID == "" {
		dftVpc, err := c.vpc.GetRegionDftVpc(kt, input.Region)
		if err != nil {
			logs.Errorf("failed to get default vpc for err: %v, region: %s, rid: %s", err, input.Region, kt.Rid)
			return nil, err
		}
		vpcID = dftVpc
	}
	subnetReq.CloudVpcID = vpcID

	if input.Subnet != "" {
		subnetReq.CloudID = input.Subnet
	} else {
		isDftRegionVpc, err := c.vpc.IsRegionDftVpc(kt, vpcID)
		if err != nil {
			logs.Errorf("failed to determine whether it is the default vpc, err: %v, region: %s, vpc id: %s, rid: %s",
				err, input.Region, vpcID, kt.Rid)
			return nil, err
		}
		if isDftRegionVpc {
			subnetReq.Name = "cvm_use_"
		}
	}

	subnetList, err := c.subnet.GetAllSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to find subnet with subnetReq: %+v, err: %v, rid: %s", subnetReq, err, kt.Rid)
		return nil, err
	}
	zoneToVpc := make(map[string][]string)
	vpcToSubnet := make(map[string][]string)

	for _, subnetItem := range subnetList.Info {
		zoneToVpc[subnetItem.Zone] = append(zoneToVpc[subnetItem.Zone], subnetItem.VpcId)
		vpcToSubnet[subnetItem.VpcId] = append(vpcToSubnet[subnetItem.VpcId], subnetItem.SubnetId)
	}

	// 2. query apply capacity
	zoneToCapacity := make(map[string]*types.CapacityInfo)
	for zoneID, vpcList := range zoneToVpc {
		vpcUniq := arrayutil.StrArrayUnique(vpcList)
		capa := c.getZoneCapacity(kt, input, zoneID, vpcUniq, vpcToSubnet, input.IgnorePrediction)
		if capa != nil {
			zoneToCapacity[zoneID] = capa
		}
	}
	rst := &types.GetCapacityRst{}
	upsertItems := make([]types.UpsertDeviceCapacityItem, 0)
	for _, capInfo := range zoneToCapacity {
		rst.Info = append(rst.Info, capInfo)
		upsertItems = append(upsertItems, types.UpsertDeviceCapacityItem{
			RequireType: input.RequireType,
			DeviceType:  input.DeviceType,
			Region:      capInfo.Region,
			Zone:        capInfo.Zone,
			MaxNum:      capInfo.MaxNum,
			MaxInfo:     capInfo.MaxInfo,
		})
	}
	rst.Count = int64(len(rst.Info))
	// 为方便排查问题，增加日志记录
	jsonRst, err := json.Marshal(rst)
	if err != nil {
		logs.Errorf("cvm apply order get capacity failed to marshal capacityRst, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	logs.Infof("get capacity, input: %+v, zoneInfo: %s, rid: %s", cvt.PtrToVal(input), string(jsonRst), kt.Rid)

	if input.DisableUpsertDB {
		return rst, nil
	}
	for _, batch := range slice.Split(upsertItems, constant.BatchOperationMaxLimit) {
		// 失败不影响正常返回，只打印日志
		if err = c.upsertDeviceCapacity(kt, batch); err != nil {
			logs.Errorf("batch get capacity failed to upsert device capacity, err: %v, input: %v, rid: %s", err,
				cvt.PtrToVal(input), kt.Rid)
		}
	}
	return rst, nil
}

// BatchGetCapacity 批量获取资源申请容量信息
func (c *capacity) BatchGetCapacity(kt *kit.Kit, input *types.BatchGetCapacityParam) (*types.BatchGetCapacityRst,
	error) {
	// 1. query subnet from db
	subnetReq := &types.GetAllSubnetReq{Region: input.Region}
	if len(input.Zones) > 0 {
		subnetReq.Zones = input.Zones
	}
	vpcID := input.Vpc
	if vpcID == "" {
		dftVpc, err := c.vpc.GetRegionDftVpc(kt, input.Region)
		if err != nil {
			logs.Errorf("failed to get default vpc for err: %v, region: %s, rid: %s", err, input.Region, kt.Rid)
			return nil, err
		}
		vpcID = dftVpc
	}
	subnetReq.CloudVpcID = vpcID

	if input.Subnet != "" {
		subnetReq.CloudID = input.Subnet
	} else {
		isDftRegionVpc, err := c.vpc.IsRegionDftVpc(kt, vpcID)
		if err != nil {
			logs.Errorf("failed to determine whether it is the default vpc, err: %v, region: %s, vpc id: %s, rid: %s",
				err, input.Region, vpcID, kt.Rid)
			return nil, err
		}
		if isDftRegionVpc {
			subnetReq.Name = "cvm_use_"
		}
	}

	subnetList, err := c.subnet.GetAllSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to find subnet with subnetReq: %+v, err: %v, rid: %s", subnetReq, err, kt.Rid)
		return nil, err
	}
	zoneToVpc := make(map[string][]string)
	vpcToSubnet := make(map[string][]string)

	for _, subnetItem := range subnetList.Info {
		zoneToVpc[subnetItem.Zone] = append(zoneToVpc[subnetItem.Zone], subnetItem.VpcId)
		vpcToSubnet[subnetItem.VpcId] = append(vpcToSubnet[subnetItem.VpcId], subnetItem.SubnetId)
	}

	// 2.query apply capacity concurrently
	deviceZoneToCapacity := c.queryBatchCapacityConcurrent(kt, input, zoneToVpc, vpcToSubnet)
	rst := &types.BatchGetCapacityRst{}
	upsertItems := make([]types.UpsertDeviceCapacityItem, 0)
	for _, capacityInfo := range deviceZoneToCapacity {
		rst.Info = append(rst.Info, capacityInfo)
		upsertItems = append(upsertItems, types.UpsertDeviceCapacityItem{
			RequireType: input.RequireType,
			DeviceType:  capacityInfo.DeviceType,
			Region:      capacityInfo.Region,
			Zone:        capacityInfo.Zone,
			MaxNum:      capacityInfo.MaxNum,
			MaxInfo:     capacityInfo.MaxInfo,
		})
	}
	rst.Count = int64(len(rst.Info))
	// 为方便排查问题，增加日志记录
	jsonRst, err := json.Marshal(rst)
	if err != nil {
		logs.Errorf("batch get capacity failed to marshal capacityRst, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	logs.Infof("batch get capacity, input: %+v, result: %s, rid: %s", cvt.PtrToVal(input), string(jsonRst), kt.Rid)
	if input.DisableUpsertDB {
		return rst, nil
	}
	for _, batch := range slice.Split(upsertItems, constant.BatchOperationMaxLimit) {
		// 失败不影响正常返回，只打印日志
		if err = c.upsertDeviceCapacity(kt, batch); err != nil {
			logs.Errorf("batch get capacity failed to upsert device capacity, err: %v, input: %v, rid: %s", err,
				cvt.PtrToVal(input), kt.Rid)
		}
	}
	return rst, nil
}

// queryBatchCapacityConcurrent 并发查询多机型、多可用区的容量信息
func (c *capacity) queryBatchCapacityConcurrent(kt *kit.Kit, input *types.BatchGetCapacityParam,
	zoneToVpc map[string][]string, vpcToSubnet map[string][]string) []*types.BatchCapacityInfo {

	deviceZoneToCapacity := make([]*types.BatchCapacityInfo, 0)
	if len(zoneToVpc) == 0 {
		return deviceZoneToCapacity
	}

	// 生成所有设备类型×可用区的组合任务
	tasks := make([]struct {
		deviceType string
		zone       string
		vpcList    []string
	}, 0, len(input.DeviceTypes)*len(zoneToVpc))

	for _, deviceType := range input.DeviceTypes {
		for zoneID, vpcList := range zoneToVpc {
			tasks = append(tasks, struct {
				deviceType string
				zone       string
				vpcList    []string
			}{deviceType: deviceType, zone: zoneID, vpcList: vpcList})
		}
	}

	resultChan := make(chan *types.BatchCapacityInfo, len(tasks))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5)

	for _, task := range tasks {
		wg.Add(1)
		go func(deviceType, zone string, vpcs []string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			vpcUniq := arrayutil.StrArrayUnique(vpcs)
			capacityInfo := c.getZoneCapacityForBatch(kt, input, deviceType, zone, vpcUniq, vpcToSubnet,
				input.IgnorePrediction)

			if capacityInfo != nil {
				resultChan <- capacityInfo
			}
		}(task.deviceType, task.zone, task.vpcList)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for result := range resultChan {
		deviceZoneToCapacity = append(deviceZoneToCapacity, result)
	}

	return deviceZoneToCapacity
}

// getZoneCapacityForBatch 获取单个设备类型+可用区的容量信息
func (c *capacity) getZoneCapacityForBatch(kt *kit.Kit, input *types.BatchGetCapacityParam, deviceType, zone string,
	vpcList []string,
	vpcToSubnet map[string][]string, ignorePrediction bool) *types.BatchCapacityInfo {

	capacityInfo := c.createEmptyCapacityInfo(deviceType, input.Region, zone)
	if len(vpcList) == 0 {
		return capacityInfo
	}

	// 1. 查询CVM容量
	capacityItem, req, resp := c.queryCvmCapacityForBatch(kt, input, deviceType, zone, vpcList, vpcToSubnet)
	if capacityItem == nil {
		return capacityInfo
	}

	// 2. 查询子网信息并更新容量
	c.updateCapacityWithSubnetInfo(kt, input, zone, vpcList, vpcToSubnet, capacityItem, ignorePrediction)

	// 3. 记录日志
	c.logCapacityInfo(kt, input, deviceType, zone, req, resp, capacityItem)

	return capacityItem
}

// createEmptyCapacityInfo 创建空的容量信息
func (c *capacity) createEmptyCapacityInfo(deviceType, region, zone string) *types.BatchCapacityInfo {
	return &types.BatchCapacityInfo{
		DeviceType: deviceType,
		Region:     region,
		Zone:       zone,
		Vpc:        "",
		Subnet:     "",
		MaxNum:     0,
		MaxInfo:    make([]*types.CapacityMaxInfo, 0),
	}
}

// queryCvmCapacityForBatch 查询CVM容量信息
func (c *capacity) queryCvmCapacityForBatch(kt *kit.Kit, input *types.BatchGetCapacityParam, deviceType, zone string,
	vpcList []string, vpcToSubnet map[string][]string) (*types.BatchCapacityInfo, interface{}, interface{}) {

	req, err := c.createCapacityReqForBatch(kt, input, deviceType, zone, vpcList, vpcToSubnet)
	if err != nil {
		logs.Errorf("failed to create cvm capacity req, err: %v, input: %+v, rid: %s", err, cvt.PtrToVal(input), kt.Rid)
		return nil, nil, nil
	}

	resp, err := c.cvm.QueryCvmCapacity(nil, nil, req)
	if err != nil {
		logs.ErrorJson("failed to get cvm apply capacity, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil, nil, nil
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get cvm apply capacity, code: %d, msg: %s, crpTraceID: %s, rid: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId, kt.Rid)
		return nil, nil, nil
	}

	if resp.Result == nil {
		logs.Errorf("failed to get cvm apply capacity, for result is nil, crpTraceID: %s, rid: %s",
			resp.TraceId, kt.Rid)
		return nil, nil, nil
	}

	capacityItem := &types.BatchCapacityInfo{
		DeviceType: deviceType,
		Region:     input.Region,
		Zone:       zone,
		Vpc:        "",
		Subnet:     "",
		MaxNum:     int64(resp.Result.MaxNum),
		MaxInfo:    make([]*types.CapacityMaxInfo, 0),
	}

	for _, info := range resp.Result.MaxInfo {
		capacityItem.MaxInfo = append(capacityItem.MaxInfo, &types.CapacityMaxInfo{
			Key:   c.translateCapacityKey(info.Key),
			Value: int64(info.Value),
		})
	}

	return capacityItem, req, resp
}

// updateCapacityWithSubnetInfo 更新容量信息中的子网信息
func (c *capacity) updateCapacityWithSubnetInfo(kt *kit.Kit, input *types.BatchGetCapacityParam, zone string,
	vpcList []string, vpcToSubnet map[string][]string, capacityItem *types.BatchCapacityInfo, ignorePrediction bool) {

	subnetToLeftIp := make(map[string]*cvmapi.SubnetInfo)
	for _, vpcItem := range vpcList {
		subnetList, err := c.querySubnet(kt, input.Region, zone, vpcItem)
		if err != nil {
			logs.Errorf("failed to get cvm subnet info, err: %v, rid: %s", err, kt.Rid)
			return
		}
		for _, subnetItem := range subnetList {
			subnetToLeftIp[subnetItem.Id] = subnetItem
		}
	}

	totalLeftIp := c.sumLeftIp(subnetToLeftIp, vpcList, vpcToSubnet)
	c.updateCapacityMaxInfoForBatch(capacityItem, totalLeftIp, ignorePrediction)
}

// logCapacityInfo 记录容量信息日志
func (c *capacity) logCapacityInfo(kt *kit.Kit, input *types.BatchGetCapacityParam, deviceType, zone string,
	req, resp interface{}, capacityItem *types.BatchCapacityInfo) {

	jsonReq, err := json.Marshal(req)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityReq, err: %v, rid: %s", err, kt.Rid)
		return
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityResp, err: %v, rid: %s", err, kt.Rid)
		return
	}

	jsonCapacityItem, err := json.Marshal(capacityItem)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityItem, err: %v, rid: %s", err, kt.Rid)
		return
	}

	logs.Infof("get zone capacity info, input: %+v, zone: %s, deviceType: %s, capacityReq: %s, capacityResp: %s, "+
		"capacityItem: %s, rid: %s", cvt.PtrToVal(input), zone, deviceType, string(jsonReq), string(jsonResp),
		string(jsonCapacityItem), kt.Rid)
}

// createCapacityReqForBatch 创建容量查询请求
func (c *capacity) createCapacityReqForBatch(kt *kit.Kit, input *types.BatchGetCapacityParam, deviceType, zone string,
	vpcList []string, vpcToSubnet map[string][]string) (*cvmapi.CapacityReq, error) {

	tempParam := &types.GetCapacityParam{
		RequireType:      input.RequireType,
		DeviceType:       deviceType,
		Region:           input.Region,
		Zone:             zone,
		Vpc:              input.Vpc,
		Subnet:           input.Subnet,
		ChargeType:       input.ChargeType,
		IgnorePrediction: input.IgnorePrediction,
		BizID:            input.BizID,
	}

	return c.createCapacityReq(kt, tempParam, zone, vpcList, vpcToSubnet)
}

// updateCapacityMaxInfoForBatch 更新容量最大信息
func (c *capacity) updateCapacityMaxInfoForBatch(capacity *types.BatchCapacityInfo, leftIp int64,
	ignorePrediction bool) {
	maxNum := leftIp
	for _, maxInfo := range capacity.MaxInfo {
		key := maxInfo.Key
		if key == hcmKeyIPCap {
			maxInfo.Value = leftIp
		}

		// 所有key的最小值，为可申请的最大值；当忽略预测时，只需要关心所选VPC子网可用IP数和云梯系统单次最大申请量。
		if maxInfo.Value < maxNum && (!ignorePrediction || key == hcmKeyIPCap || key == hcmKeyApplyLimit) {
			maxNum = maxInfo.Value
		}
	}

	capacity.MaxNum = maxNum
}

// UpsertCapacity upsert resource apply capacity info
func (c *capacity) UpsertCapacity(kt *kit.Kit, input *types.UpdateCapacityParam) error {
	// 1. get capacity
	param := &types.GetCapacityParam{
		RequireType:     input.RequireType,
		DeviceType:      input.DeviceType,
		Region:          input.Region,
		Zone:            input.Zone,
		DisableUpsertDB: true,
	}

	rst, err := c.GetCapacity(kt, param)
	if err != nil {
		logs.Errorf("failed to get capacity, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	count := len(rst.Info)
	if count != 1 {
		logs.Errorf("get invalid capacity info num %d not equal 1, input: %+v, rid: %s",
			count, cvt.PtrToVal(input), kt.Rid)
		return fmt.Errorf("get invalid capacity info num %d not equal 1", count)
	}

	if rst.Info[0] == nil {
		logs.Errorf("get invalid null capacity info, rid: %s", kt.Rid)
		return errors.New("get invalid null capacity info")
	}

	// 2. upsert device capacity
	upsertItem := []types.UpsertDeviceCapacityItem{{
		RequireType: input.RequireType,
		DeviceType:  input.DeviceType,
		Region:      input.Region,
		Zone:        input.Zone,
		MaxNum:      rst.Info[0].MaxNum,
		MaxInfo:     rst.Info[0].MaxInfo,
	}}
	if err = c.upsertDeviceCapacity(kt, upsertItem); err != nil {
		logs.Errorf("batch get capacity failed to upsert device capacity, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (c *capacity) getZoneCapacity(kt *kit.Kit, input *types.GetCapacityParam, zone string, vpcList []string,
	vpcToSubnet map[string][]string, ignorePrediction bool) *types.CapacityInfo {

	// 1. query cvm capacity
	if len(vpcList) == 0 {
		capacityInfo := &types.CapacityInfo{
			Region:  input.Region,
			Zone:    zone,
			Vpc:     "",
			Subnet:  "",
			MaxNum:  0,
			MaxInfo: make([]*types.CapacityMaxInfo, 0),
		}
		return capacityInfo
	}

	req, err := c.createCapacityReq(kt, input, zone, vpcList, vpcToSubnet)
	if err != nil {
		logs.Errorf("failed to create cvm capacity req, err: %v, input: %+v, rid: %s", err, cvt.PtrToVal(input), kt.Rid)
		return nil
	}
	resp, err := c.cvm.QueryCvmCapacity(nil, nil, req)
	if err != nil {
		logs.ErrorJson("failed to get cvm apply capacity, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return nil
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to get cvm apply capacity, code: %d, msg: %s, crpTraceID: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, resp.TraceId, kt.Rid)
		return nil
	}

	if resp.Result == nil {
		logs.Errorf("failed to get cvm apply capacity, for result is nil, crpTraceID: %s, rid: %s",
			resp.TraceId, kt.Rid)
		return nil
	}

	capacityItem := &types.CapacityInfo{
		Region:  input.Region,
		Zone:    zone,
		MaxNum:  int64(resp.Result.MaxNum),
		MaxInfo: make([]*types.CapacityMaxInfo, 0),
	}

	for _, info := range resp.Result.MaxInfo {
		capacityItem.MaxInfo = append(capacityItem.MaxInfo, &types.CapacityMaxInfo{
			Key:   c.translateCapacityKey(info.Key),
			Value: int64(info.Value),
		})
	}

	// 2. query all subnet info for left ip number
	subnetToLeftIp := make(map[string]*cvmapi.SubnetInfo)
	for _, vpcItem := range vpcList {
		subnetList, err := c.querySubnet(kt, input.Region, zone, vpcItem)
		if err != nil {
			logs.Errorf("failed to get cvm subnet info, err: %v, rid: %s", err, kt.Rid)
			return nil
		}
		for _, subnetItem := range subnetList {
			subnetToLeftIp[subnetItem.Id] = subnetItem
		}
	}

	// 3. sum up total left ip number
	totalLeftIp := c.sumLeftIp(subnetToLeftIp, vpcList, vpcToSubnet)

	// 4. update max info
	c.updateCapacityMaxInfo(capacityItem, totalLeftIp, ignorePrediction)

	jsonReq, err := json.Marshal(req)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityReq, err: %v, rid: %s", err, kt.Rid)
		return nil
	}
	// 需要记录crp返回的所有结果包括traceid
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityResp, err: %v, rid: %s", err, kt.Rid)
		return nil
	}
	jsonCapacityItem, err := json.Marshal(capacityItem)
	if err != nil {
		logs.Errorf("get zone capacity failed to marshal capacityItem, err: %v, rid: %s", err, kt.Rid)
		return nil
	}
	logs.Infof("get zone capacity info, input: %+v, zone: %s, capacityReq: %s, capacityResp: %s, capacityItem: %s, "+
		"vpcList: %v, rid: %s", cvt.PtrToVal(input), zone, string(jsonReq), string(jsonResp), jsonCapacityItem,
		vpcList, kt.Rid)

	return capacityItem
}

func (c *capacity) createCapacityReq(kt *kit.Kit, input *types.GetCapacityParam, zone string, vpcList []string,
	vpcToSubnet map[string][]string) (*cvmapi.CapacityReq, error) {

	projectName := input.RequireType.ToObsProject()
	business3ID, err := c.getBusiness3ID(kt, input)
	if err != nil {
		logs.Errorf("failed to get business3 id, err: %v, input: %+v, rid: %s", err, cvt.PtrToVal(input), kt.Rid)
		return nil, err
	}

	req := &cvmapi.CapacityReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmCapacityMethod,
		},
		Params: &cvmapi.CapacityParam{
			DeptId:       cvmapi.CvmDeptId,
			Business3Id:  business3ID,
			CloudCampus:  zone,
			InstanceType: input.DeviceType,
			VpcId:        vpcList[0],
			SubnetId:     vpcToSubnet[vpcList[0]][0],
			ProjectName:  string(projectName),
		},
	}
	// 计费模式,默认包年包月
	if len(input.ChargeType) > 0 {
		req.Params.ChargeType = input.ChargeType
	}

	return req, nil
}

func (c *capacity) getBusiness3ID(kt *kit.Kit, input *types.GetCapacityParam) (int, error) {
	business3ID := cvmapi.CvmLaunchBiz3Id
	if input == nil || input.BizID == 0 {
		logs.Warnf("can not find input bizID, use default val: %d, rid: %s", business3ID, kt.Rid)
		return business3ID, nil
	}

	// 获取业务空闲机模块的三级业务id
	req := &cmdb.SearchModuleParams{
		BizID:     input.BizID,
		Condition: mapstr.MapStr{"default": cmdb.DftModuleIdle},
		Fields:    []string{"bs3_name_id"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	module, err := c.cmdbCli.SearchModule(kt, req)
	if err != nil {
		logs.Errorf("failed to search module from cc, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(req), kt.Rid)
		return 0, err
	}
	if len(module.Info) == 0 {
		logs.Errorf("can not find idle module from cc, req: %+v, rid: %s", cvt.PtrToVal(req), kt.Rid)
		return 0, fmt.Errorf("can not find idle module from cc, bizID: %d", input.BizID)
	}

	if module.Info[0].Bs3NameID != 0 {
		return module.Info[0].Bs3NameID, nil
	}

	return business3ID, nil
}

func (c *capacity) querySubnet(kt *kit.Kit, region, zone, vpc string) ([]*cvmapi.SubnetInfo, error) {
	subnetReq := cvmapi.SubnetRealParam{
		Region:      region,
		CloudCampus: zone,
		VpcId:       vpc,
	}
	resp, err := c.cvm.QueryRealCvmSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Result, nil
}

func (c *capacity) sumLeftIp(subnetToLeftIp map[string]*cvmapi.SubnetInfo, vpcList []string,
	vpcToSubnet map[string][]string) int64 {

	subnetIdList := make([]string, 0)
	for _, vpc := range vpcList {
		subnetIdList = append(subnetIdList, vpcToSubnet[vpc]...)
	}

	subnetIdList = arrayutil.StrArrayUnique(subnetIdList)

	total := 0
	for _, subnetId := range subnetIdList {
		if subnetToLeftIp[subnetId] != nil {
			total = total + subnetToLeftIp[subnetId].LeftIpNum
		}
	}

	return int64(total)
}

func (c *capacity) updateCapacityMaxInfo(capacity *types.CapacityInfo, leftIp int64, ignorePrediction bool) {
	maxNum := leftIp
	for _, maxInfo := range capacity.MaxInfo {
		key := maxInfo.Key
		if key == hcmKeyIPCap {
			maxInfo.Value = leftIp
		}

		// 所有key的最小值，为可申请的最大值；当忽略预测时，只需要关心所选VPC子网可用IP数和云梯系统单次最大申请量。
		if maxInfo.Value < maxNum && (!ignorePrediction || key == hcmKeyIPCap || key == crpKeyApplyLimit) {
			maxNum = maxInfo.Value
		}
	}

	capacity.MaxNum = maxNum
}

func (c *capacity) getCapacityFlag(num int) int {
	flag := types.CapLevelEmpty
	if num <= 10 {
		flag = types.CapLevelLow
	} else if num <= 50 {
		flag = types.CapLevelMedium
	} else {
		flag = types.CapLevelHigh
	}

	return flag
}

const (
	crpKeyCBSCap        = "云后端CBS容量计算可申领量"
	crpKeyCVMCap        = "云后端CVM容量计算可申领量"
	crpKeyIPCap         = "所选VPC子网可用IP数"
	crpKeyPredictionCap = "未执行需求预测的可申领量"
	crpKeyApplyLimit    = "云梯系统单次提单最大量"

	hcmKeyCBSCap        = "云后端CBS库存可申请量"
	hcmKeyCVMCap        = "云后端CVM库存可申请量"
	hcmKeyIPCap         = "所选VPC子网可用IP数"
	hcmKeyPredictionCap = "未执行需求预测的可申请量"
	hcmKeyApplyLimit    = "云梯系统单次最大申请量"
)

// translateCapacityKey translate yunti capacity info key to cr capacity info key
func (c *capacity) translateCapacityKey(key string) string {
	switch key {
	case crpKeyCBSCap:
		return hcmKeyCBSCap
	case crpKeyCVMCap:
		return hcmKeyCVMCap
	case crpKeyIPCap:
		return hcmKeyIPCap
	case crpKeyPredictionCap:
		return hcmKeyPredictionCap
	case crpKeyApplyLimit:
		return hcmKeyApplyLimit
	default:
		return key
	}
}

// upsertDeviceCapacity 更新或创建库存信息
func (c *capacity) upsertDeviceCapacity(kt *kit.Kit, items []types.UpsertDeviceCapacityItem) error {
	if len(items) == 0 {
		return nil
	}
	if len(items) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("items exceed limit %d", constant.BatchOperationMaxLimit)
	}

	// 验证所有输入
	for i, item := range items {
		if err := item.Validate(); err != nil {
			logs.Errorf("input is invalid at index %d, err: %v, rid: %s", i, err, kt.Rid)
			return fmt.Errorf("input is invalid at index %d: %w", i, err)
		}
	}

	// 构建批量查询条件,查询所有可能存在的记录
	// 构建 OR 表达式,每个分支是一个 AND 表达式
	orRules := make([]filter.RuleFactory, 0, len(items))
	for _, item := range items {
		andExpressions := []*filter.AtomRule{
			tools.RuleEqual("require_type", item.RequireType),
			tools.RuleEqual("region", item.Region),
			tools.RuleEqual("zone", item.Zone),
			tools.RuleEqual("device_type", item.DeviceType),
		}
		orRules = append(orRules, tools.ExpressionAnd(andExpressions...))
	}

	// 构建已存在记录的映射表,key为 "require_type:region:zone:device_type"
	existingMap := make(map[string]string)
	listReq := &core.ListReq{
		Filter: &filter.Expression{Op: filter.Or, Rules: orRules},
		Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	for {
		listRst, err := c.client.DataService().Global.DeviceCapacity.List(kt, listReq)
		if err != nil {
			logs.Errorf("failed to list device capacity, err: %v, req: %+v, rid: %s", err, cvt.PtrToVal(listReq),
				kt.Rid)
			return err
		}
		for _, detail := range listRst.Details {
			key := fmt.Sprintf("%d:%s:%s:%s", int64(detail.RequireType), detail.Region, detail.Zone, detail.DeviceType)
			existingMap[key] = detail.ID
		}

		if len(listRst.Details) < int(listReq.Page.Limit) {
			break
		}
		listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return c.upsertDeviceCapacityByExistMap(kt, items, existingMap)
}

func (c *capacity) upsertDeviceCapacityByExistMap(kt *kit.Kit, items []types.UpsertDeviceCapacityItem,
	existingMap map[string]string) error {

	// 分离需要创建和更新的记录
	createItems := make([]devicecapacity.CreateDeviceCapacityField, 0)
	updateItems := make([]devicecapacity.UpdateDeviceCapacityField, 0)
	for _, item := range items {
		extension, err := hcmjson.MarshalToString(item.MaxInfo)
		if err != nil {
			logs.Errorf("fail to marshal capacity max info to string, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		key := fmt.Sprintf("%d:%s:%s:%s", int64(item.RequireType), item.Region, item.Zone, item.DeviceType)
		id, exists := existingMap[key]
		if exists {
			// 需要更新
			updateItems = append(updateItems, devicecapacity.UpdateDeviceCapacityField{
				ID:          id,
				RequireType: cvt.ValToPtr(item.RequireType),
				Region:      cvt.ValToPtr(item.Region),
				Zone:        cvt.ValToPtr(item.Zone),
				DeviceType:  cvt.ValToPtr(item.DeviceType),
				Capacity:    cvt.ValToPtr(item.MaxNum),
				Extension:   cvt.ValToPtr(tabletype.JsonField(extension)),
			})
			continue
		}
		// 需要创建
		createItems = append(createItems, devicecapacity.CreateDeviceCapacityField{
			RequireType: item.RequireType,
			Region:      item.Region,
			Zone:        item.Zone,
			DeviceType:  item.DeviceType,
			Capacity:    cvt.ValToPtr(item.MaxNum),
			Extension:   tabletype.JsonField(extension),
		})
	}

	// 分批处理创建操作
	for _, batch := range slice.Split(createItems, constant.BatchOperationMaxLimit) {
		_, err := c.client.DataService().Global.DeviceCapacity.Create(kt, &devicecapacity.CreateDeviceCapacityReq{
			Items: batch,
		})
		if err != nil {
			logs.Errorf("failed to create device capacity batch, err: %v, batch: %v, rid: %s", err, batch, kt.Rid)
			return err
		}
	}

	// 分批处理更新操作
	for _, batch := range slice.Split(updateItems, constant.BatchOperationMaxLimit) {
		err := c.client.DataService().Global.DeviceCapacity.Update(kt, &devicecapacity.UpdateDeviceCapacityReq{
			Items: batch,
		})
		if err != nil {
			logs.Errorf("failed to update device capacity batch, err: %v, batch: %v, rid: %s", err, batch, kt.Rid)
			return err
		}
	}
	return nil
}

// ListCapacityWithDeviceInfo 查询设备库存及其机型详细信息
func (c *capacity) ListCapacityWithDeviceInfo(kt *kit.Kit, input *devicecapacity.ListCapacityWithDeviceInfoReq) (
	*woaserver.ListCapacityWithDeviceInfoResult, error) {

	if err := input.Validate(); err != nil {
		logs.Errorf("failed to validate list capacity with device info request, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, err := c.client.DataService().Global.DeviceCapacity.ListCapacityWithDeviceInfo(kt, input)
	if err != nil {
		logs.Errorf("failed to list capacity with device info, err: %v, req: %+v, rid: %s", err, input, kt.Rid)
		return nil, err
	}
	if input.Page.Count {
		return &woaserver.ListCapacityWithDeviceInfoResult{Count: resp.Count}, nil
	}

	details := make([]woaserver.CapacityWithDeviceInfo, 0, len(resp.Details))
	for _, one := range resp.Details {
		details = append(details, woaserver.CapacityWithDeviceInfo{
			CapacityWithDeviceInfo: one,
			CapacityFlag:           c.getCapacityFlag(int(one.Capacity)),
		})
	}
	return &woaserver.ListCapacityWithDeviceInfoResult{Details: details}, nil
}
