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

// Package generator provides ...
package generator

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"hcm/cmd/woa-server/logics/config"
	poolLogics "hcm/cmd/woa-server/logics/pool"
	rollingserver "hcm/cmd/woa-server/logics/rolling-server"
	"hcm/cmd/woa-server/logics/task/scheduler/algorithm"
	"hcm/cmd/woa-server/model/task"
	cfgtypes "hcm/cmd/woa-server/types/config"
	poolTypes "hcm/cmd/woa-server/types/pool"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/dvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
	utils "hcm/pkg/tools/util"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

// Generator generates vm devices
type Generator struct {
	cvm          cvmapi.CVMClientInterface
	dvm          dvmapi.DVMClientInterface
	cc           cmdb.Client
	ctx          context.Context
	configLogics config.Logics
	poolLogics   poolLogics.Logics
	rsLogics     rollingserver.Logics
	clientConf   cc.ClientConfig

	predicateFuncs map[string]algorithm.FitPredicate
	priorityFuncs  []algorithm.PriorityConfig
}

// New creates a generator
func New(ctx context.Context, rsLogics rollingserver.Logics, thirdCli *thirdparty.Client, cmdbCli cmdb.Client,
	clientConf cc.ClientConfig, configLogics config.Logics) (*Generator, error) {

	predicateFuncs := initPredicateFuncs()
	priorityFuncs := initpriorityFuncs()

	generator := &Generator{
		cvm:            thirdCli.CVM,
		dvm:            thirdCli.DVM,
		cc:             cmdbCli,
		predicateFuncs: predicateFuncs,
		priorityFuncs:  priorityFuncs,
		ctx:            ctx,
		clientConf:     clientConf,
		configLogics:   configLogics,
		rsLogics:       rsLogics,
		poolLogics:     poolLogics.New(ctx, clientConf, thirdCli, cmdbCli),
	}

	return generator, nil
}

func initPredicateFuncs() map[string]algorithm.FitPredicate {
	predicateFuncs := map[string]algorithm.FitPredicate{
		"VMFitHostVirtualRatio": algorithm.VMFitHostVirtualRatio,
		"VMFitRegion":           algorithm.VMFitRegion,
		"VMFitCampus":           algorithm.VMFitCampus,
		"VMFitKernel":           algorithm.VMFitKernel,
		"VMFitCpuProvider":      algorithm.VMFitCpuProvider,
	}

	return predicateFuncs
}

func initpriorityFuncs() []algorithm.PriorityConfig {
	priorityFuncs := []algorithm.PriorityConfig{
		{
			Function: algorithm.CalculateBalancedResourceAllocation,
			Weight:   10,
		},
	}

	return priorityFuncs
}

// GenerateCVM generates cvm devices
func (g *Generator) GenerateCVM(kt *kit.Kit, order *types.ApplyOrder) error {
	// 1. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// check if need generate cvm
	existCount := uint(len(existDevices))
	if existCount >= order.TotalNum {
		logs.Infof("apply order %s has been scheduled %d cvm", order.SubOrderId, existCount)
		// check if need retry match task
		if err := g.retryMatchDevice(existDevices); err != nil {
			logs.Warnf("failed to retry match device, order id: %s, err: %v", order.SubOrderId, err)
		}
		return nil
	}

	logs.Infof("apply order %s existing device number: %d", order.SubOrderId, existCount)

	// 获取该申请单的可用区
	orderZones, err := g.getApplyOrderMultiZones(kt, order)
	if err != nil {
		logs.Errorf("failed to get apply order zone list, subOrderID: %s, err: %v, rid: %s",
			order.SubOrderId, err, kt.Rid)
		return err
	}

	// for given zone case
	if !order.Spec.IsCVMSeparateCampus() {
		if err = g.generateCVMConcentrate(kt, order, existDevices, orderZones); err != nil {
			logs.Errorf("failed to generate cvm in zone %s, subOrderID: %s, rid: %s",
				order.Spec.Zone, order.SubOrderId, kt.Rid)
			return err
		}
		return nil
	}

	// for cvm_separate_campus case
	if err = g.generateCVMSeparate(kt, order, existDevices, orderZones); err != nil {
		logs.Errorf("failed to generate cvm in separate zones in region %s, subOrderID: %s, rid: %s",
			order.Spec.Region, order.SubOrderId, kt.Rid)
		return err
	}

	return nil
}

// getApplyOrderMultiZones 获取多可用区
func (g *Generator) getApplyOrderMultiZones(kt *kit.Kit, order *types.ApplyOrder) ([]string, error) {
	if order.Spec == nil {
		return nil, fmt.Errorf("order spec is nil")
	}

	// V2版本选了“全部” 或者 V1版本选了“分Campus生产”，则需要后端获取该地域所有的可用区
	if (len(order.Spec.Zones) == 1 && order.Spec.Zones[0] == cvmapi.CvmZoneAll) ||
		(len(order.Spec.Zone) > 0 && order.Spec.Zone == cvmapi.CvmSeparateCampus) {

		// 选择了“全部”则需要后端获取该地域所有的可用区
		allZones, err := g.getZoneList(kt, order.Spec.Region)
		if err != nil {
			return nil, err
		}

		orderZones := make([]string, 0)
		for _, zone := range allZones {
			orderZones = append(orderZones, zone.Zone)
		}

		if len(orderZones) == 0 {
			return nil, fmt.Errorf("get order all zone is nil, region: %s", order.Spec.Region)
		}
		return orderZones, nil
	}

	// V2版本获取可用区的方式
	if len(order.Spec.Zones) > 0 {
		// 分Campus生产时，可用区数量需要大于1个
		if order.Spec.ResAssign == enumor.CampusResAssign && len(order.Spec.Zones) == 1 {
			return nil, fmt.Errorf("[分Campus生产]可用区数量需要大于1个")
		}
		return slice.Unique(order.Spec.Zones), nil
	}

	// V1版本获取可用区的方式
	if len(order.Spec.Zone) > 0 && order.Spec.Zone != cvmapi.CvmSeparateCampus {
		return []string{order.Spec.Zone}, nil
	}

	return nil, fmt.Errorf("both order spec zones and zone fields are empty or invalid")
}

// retryMatchDevice retry to match generated devices
func (g *Generator) retryMatchDevice(devices []*types.DeviceInfo) error {
	genIDs := make([]int64, 0)
	for _, device := range devices {
		if !device.IsDelivered {
			genIDs = append(genIDs, int64(device.GenerateId))
		}
	}

	genIDs = utils.IntArrayUnique(genIDs)
	// update generate record to unmatched
	for _, genID := range genIDs {
		filter := &mapstr.MapStr{
			"generate_id": genID,
		}

		doc := mapstr.MapStr{
			"is_matched": false,
			"update_at":  time.Now(),
		}

		err := model.Operation().GenerateRecord().UpdateGenerateRecord(context.Background(), filter, &doc)
		if err != nil {
			logs.Errorf("failed to update generate record, generate id: %d, update: %+v, err: %v", genID, doc, err)
			return err
		}
	}

	return nil
}

// generateCVMConcentrate generates cvm devices in certain zone
func (g *Generator) generateCVMConcentrate(kt *kit.Kit, order *types.ApplyOrder, existDevices []*types.DeviceInfo,
	orderZones []string) error {

	replicas := order.TotalNum - uint(len(existDevices))

	genRecordIds := make([]uint64, 0)
	errs := make([]error, 0)
	switch order.ResourceType {
	case types.ResourceTypeCvm:
		genRecordIds, errs = g.batchLaunchCvm(kt, order, orderZones, replicas)
	case types.ResourceTypeUpgradeCvm:
		genRecordId, _, err := g.batchUpgradeCvm(kt, order, replicas)
		if err != nil {
			errs = append(errs, err)
		}
		if genRecordId != 0 {
			genRecordIds = append(genRecordIds, genRecordId)
		}
	default:
		logs.Errorf("unsupported resource type: %s", order.ResourceType)
		return fmt.Errorf("unsupported resource type: %s", order.ResourceType)
	}

	return g.checkLaunchCvmResult(kt, order.ResourceType, order.SubOrderId, genRecordIds, errs)
}

// generateCVMSeparate generates cvm devices in separate zones
func (g *Generator) generateCVMSeparate(kt *kit.Kit, order *types.ApplyOrder, existDevices []*types.DeviceInfo,
	orderZones []string) error {

	// 1. sum up each zone created devices
	createdTotalCount := uint(0)
	zoneCreatedCount := make(map[string]uint, 0)
	for _, device := range existDevices {
		zoneCreatedCount[device.CloudZone]++
		createdTotalCount++
	}

	// 2. get available zones
	availZonesMap, err := g.getAvailableZoneInfo(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region)
	if err != nil {
		logs.Errorf("failed to generate cvm, for get available zones err: %v, order id: %s, rid: %s",
			err, order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for get available zones err: %v", err)
	}
	if len(availZonesMap) == 0 {
		logs.Errorf("failed to generate cvm, for get no available zones, order id: %s, rid: %s",
			order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for get no available zones")
	}

	// 3. get capacity
	zoneCapacity, err := g.getCapacity(kt, order, cvmapi.CvmSeparateCampus, "", "")
	if err != nil {
		logs.Errorf("failed to generate cvm, for get zone capacity err: %v, order id: %s, rid: %s",
			err, order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for get zone capacity err: %v", err)
	}

	// 检查是否有已经失败过的可用区，需要跳过这些失败的可用区
	failedZoneMap := make(map[string]struct{})
	if order.Spec != nil && len(order.Spec.FailedZoneIDs) > 0 {
		failedZoneMap = cvt.StringSliceToMap(order.Spec.FailedZoneIDs)
	}

	logs.Infof("generateCVMSeparate campus start, subOrderID: %s, createdTotalCount: %d, zoneCapacity: %+v, "+
		"zoneCreatedCount: %v, availZones: %+v, failedZoneMap: %+v, rid: %s", order.SubOrderId, createdTotalCount,
		zoneCapacity, zoneCreatedCount, availZonesMap, failedZoneMap, kt.Rid)

	// 4. 分Campus生产
	genRecordIds, errs := g.generateCvmAcrossCampus(kt, order, orderZones, availZonesMap, createdTotalCount,
		zoneCapacity, zoneCreatedCount, failedZoneMap)

	return g.checkLaunchCvmResult(kt, order.ResourceType, order.SubOrderId, genRecordIds, errs)
}

func (g *Generator) generateCvmAcrossCampus(kt *kit.Kit, order *types.ApplyOrder, orderZones []string,
	availZonesMap map[string]*cfgtypes.Zone, createdTotalCount uint, zoneCapacity map[string]int64,
	zoneCreatedCount map[string]uint, failedZoneMap map[string]struct{}) ([]uint64, []error) {

	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	genRecordIds := make([]uint64, 0)
	appendError := func(subErrs []error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, subErrs...)
	}
	appendGenRecord := func(ids []uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		genRecordIds = append(genRecordIds, ids...)
	}
	maxCount := math.Ceil(float64(order.TotalNum) / 2)
	failedSkipNum := 0
	var isSkip bool
	for _, zone := range orderZones {
		failedSkipNum, isSkip = checkZoneAvailability(kt, order.SubOrderId, availZonesMap,
			failedZoneMap, zone, failedSkipNum)
		if isSkip {
			continue
		}
		replicas := uint(0)
		// 该地域下可用区的数量
		if len(availZonesMap) > 1 {
			// 一个城市有大于一个campus的话，该campus最多只能生产需求数量的一半
			// 若单据无法完成，则剩余不生产，等人工介入处理
			campusMax := math.Max(maxCount-float64(zoneCreatedCount[zone]), 0)
			replicas = uint(math.Min(math.Min(float64(order.TotalNum-createdTotalCount), float64(zoneCapacity[zone])),
				campusMax))
		} else {
			// 一个城市只有一个campus的话，全部生产
			replicas = uint(math.Min(math.Min(float64(order.TotalNum-createdTotalCount), float64(zoneCapacity[zone])),
				maxCount))
		}

		logs.Infof("generateCVMSeparate campus loop, subOrderID: %s, maxCount: %d, createdTotalCount: %d, "+
			"zoneCapacity: %+v, zoneCreatedCount: %v, zoneInfo: %s, availZonesNum: %d, replicas: %d, rid: %s",
			order.SubOrderId, maxCount, createdTotalCount, zoneCapacity, zoneCreatedCount, zone,
			len(orderZones), replicas, kt.Rid)
		if replicas <= 0 {
			// 当某一个可用区交付超过一半后，不会被计入 failedSkipNum，导致失败的可用区无法清空
			failedSkipNum++
			continue
		}

		zoneCreatedCount[zone] += replicas
		createdTotalCount += replicas

		wg.Add(1)
		go func(order *types.ApplyOrder, zoneId string, replicas uint) {
			defer wg.Done()
			genIds, subErrs := g.batchLaunchCvm(kt, order, []string{zoneId}, replicas)
			if len(subErrs) != 0 {
				logs.Errorf("failed to launch cvm, subOrderID: %s, subErrs: %v, zoneId: %s, rid: %s", order.SubOrderId,
					subErrs, zoneId, kt.Rid)
				appendError(subErrs)
			}
			if len(genIds) > 0 {
				logs.Infof("success to launch cvm, subOrderID: %s, zone: %s, generate ids: %v, rid: %s",
					order.SubOrderId, zoneId, genIds, kt.Rid)
				appendGenRecord(genIds)
			}
		}(order, zone, replicas)
		if order.TotalNum <= createdTotalCount {
			break
		}
	}

	// 全部被跳过的情况下，清空所有失败的可用区
	if failedSkipNum == len(orderZones) {
		if err := g.updateOrderFailedZones(kt, order.SubOrderId, ""); err != nil {
			// 只记录日志，不应该影响主流程
			logs.Warnf("failed to update order failed zoneIDs, subOrderID: %s, err: %v, rid: %s",
				order.SubOrderId, err, kt.Rid)
		}
	}
	wg.Wait()
	return genRecordIds, errs
}

func checkZoneAvailability(kt *kit.Kit, subOrderID string, availZonesMap map[string]*cfgtypes.Zone,
	failedZoneMap map[string]struct{}, zone string, failedSkipNum int) (int, bool) {

	// 该可用区不在该机型的可用区列表中
	if _, ok := availZonesMap[zone]; !ok {
		logs.Warnf("generateCVMSeparate campus loop skip avail zone, subOrderID: %s, zone: %s, availZonesMap: %+v, "+
			"rid: %s", subOrderID, zone, availZonesMap, kt.Rid)
		return failedSkipNum, true
	}

	// 已经失败过的可用区，直接跳过
	if _, ok := failedZoneMap[zone]; ok {
		// 记录失败跳过的数量
		failedSkipNum++
		logs.Warnf("generateCVMSeparate campus loop skip has failed zone, subOrderID: %s, zone: %s, "+
			"failedZoneMap: %+v, rid: %s", subOrderID, zone, failedZoneMap, kt.Rid)
		return failedSkipNum, true
	}

	return failedSkipNum, false
}

func (g *Generator) checkLaunchCvmResult(kt *kit.Kit, resType types.ResourceType, subOrderID string,
	genRecordIds []uint64, errs []error) error {

	if len(genRecordIds) == 0 {
		logs.Errorf("failed to generate cvm, for no zone has generate record, subOrderID: %s, errs: %v, rid: %s",
			subOrderID, errs, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for no zone has generate record")
	}

	if len(errs) > 0 {
		logs.Errorf("failed to generate cvm, subOrderID: %s, errs: %v, rid: %s", subOrderID, errs, kt.Rid)

		// check all generate records and update apply order status
		if err := g.UpdateOrderStatus(resType, subOrderID); err != nil {
			logs.Errorf("failed to update order status, subOrderId: %s, err: %v, rid: %s", subOrderID, err, kt.Rid)
		}
	}

	return nil
}

// UpdateOrderStatus 更新订单状态
func (g *Generator) UpdateOrderStatus(resType types.ResourceType, suborderID string) error {
	genRecords, err := g.getOrderGenRecords(suborderID)
	if err != nil {
		logs.Errorf("failed to get generate records, order id: %s, err: %v", suborderID, err)
		return err
	}

	hasGenRecordMatching := false
	for _, record := range genRecords {
		if record.Status == types.GenerateStatusHandling ||
			record.Status == types.GenerateStatusSuccess && !record.IsMatched {
			hasGenRecordMatching = true
			break
		}
	}

	stage := types.TicketStageRunning
	status := types.ApplyStatusMatchedSome
	// TODO 临时，升降配order不进入matchedSome，直接失败
	if resType == types.ResourceTypeUpgradeCvm {
		stage = types.TicketStageSuspend
		status = types.ApplyStatusTerminate
	}
	if hasGenRecordMatching {
		stage = types.TicketStageRunning
		status = types.ApplyStatusMatching
	}

	// do update apply order status
	filter := &mapstr.MapStr{
		"suborder_id": suborderID,
	}

	doc := &mapstr.MapStr{
		"stage":     stage,
		"status":    status,
		"update_at": time.Now(),
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update apply order, id: %s, err: %v", suborderID, err)
		return err
	}

	return nil
}

// updateOrderFailedZones 更新订单失败的可用区
func (g *Generator) updateOrderFailedZones(kt *kit.Kit, suborderID string, zone string) error {
	// 如果传入的是分Campus的可用区，则不需要更新，直接返回
	if zone == cvmapi.CvmSeparateCampus {
		return nil
	}

	var failedZoneIDs []string
	if len(zone) > 0 {
		order, err := g.GetApplyOrder(suborderID)
		if err != nil {
			logs.Errorf("failed to get apply order, err: %v, subOrderID: %s, rid: %s", err, suborderID, kt.Rid)
			return fmt.Errorf("failed to get apply order, err: %v, order id: %s", err, suborderID)
		}

		if order.Spec != nil {
			failedZoneIDs = order.Spec.FailedZoneIDs
		}
		failedZoneIDs = append(failedZoneIDs, zone)
	}

	filter := &mapstr.MapStr{
		"suborder_id": suborderID,
	}

	doc := &mapstr.MapStr{
		"spec.failed_zone_ids": failedZoneIDs,
		"update_at":            time.Now(),
	}
	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update apply order failed zones, subOrderID: %s, err: %v, zone: %s, "+
			"failedZoneIDs: %v, rid: %s", suborderID, err, zone, failedZoneIDs, kt.Rid)
		return err
	}

	return nil
}

// GenerateDVM generates docker vm devices
func (g *Generator) GenerateDVM(kt *kit.Kit, order *types.ApplyOrder) error {
	// 1. 解析请求结构
	selector, err := g.parseDvmSelector(kt, order)
	if err != nil {
		logs.Errorf("failed to parse dvm selector, err: %v, order id: %s", err, order.SubOrderId)
		return err
	}

	// 2. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	existCount := uint(len(existDevices))
	if existCount >= order.TotalNum {
		logs.Infof("apply order %s has been scheduled %d docker vm", order.SubOrderId, existCount)
		return nil
	}
	logs.Infof("apply order %s existing device number: %d", order.SubOrderId, existCount)

	// 3. 初始化（存量设备）亲和性
	// 记录每类亲和维度的设备数
	// 如果亲和性有要求时，每个维度(campus\module...)的设备不能超过一半
	antiAffinityReplicas := make(map[string]uint)
	for _, device := range existDevices {
		antiAffinityReplicas[g.antiAffinityValue(order.AntiAffinityLevel, types.HostPriority{
			IP:         device.Ip,
			SZone:      device.ZoneName,
			Equipment:  device.Equipment,
			ModuleName: strings.ToLower(device.ModuleName),
		})]++
	}

	// 4. 计算(一个子单)虚拟比不超过1:3
	existingHostMap := make(map[string]*dvmapi.DockerHost)
	for _, host := range existDevices {
		hostAssetID := ""
		parts := strings.Split(host.AssetId, "-")
		if len(parts) > 1 {
			hostAssetID = parts[0]
		}
		if val, ok := existingHostMap[hostAssetID]; ok {
			val.ScheduledVMs++
			existingHostMap[hostAssetID] = val
		} else {
			existingHostMap[hostAssetID] = &dvmapi.DockerHost{
				ScheduledVMs: 1,
				AssetID:      hostAssetID,
			}
		}
	}

	// 5. get allocatable docker hosts
	allocatableHosts, err := g.getAllocatableHosts(kt, selector, order.ResourceType, existingHostMap)
	if err != nil {
		logs.Errorf("failed to get allocatable hosts, err: %v, order id: %s", err, order.SubOrderId)
		return err
	}
	if len(allocatableHosts) == 0 {
		logs.Errorf("get no allocatable hosts, order id: %s", order.SubOrderId)
		return fmt.Errorf("get no allocatable hosts")
	}

	// 6. sort allocatable docker hosts
	sortList := g.sortHosts(order.AntiAffinityLevel, allocatableHosts)
	logs.V(4).Infof("allocatable host list: %+v", sortList)

	// 7. try launch docker vm in order
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	genRecordIds := make([]uint64, 0)
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	appendGenRecord := func(id uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		genRecordIds = append(genRecordIds, id)
	}

	maxCount := order.TotalNum
	if order.AntiAffinityLevel != types.AntiNone {
		maxCount = uint(math.Ceil(float64(order.TotalNum) / 2))
		if maxCount == 0 {
			maxCount = 1
		}
	}
	for _, host := range sortList {
		// 计算最多可生产的容器数
		existNum := g.sumReplicas(antiAffinityReplicas)
		replicas := uint(math.Min(
			math.Min(
				math.Min(
					// 还需要生产的数量
					float64(order.TotalNum-existNum),
					// 每台母机剩余的可生产数
					float64(host.AllocatableCount)),
				// 最大虚拟比
				float64(maxVirtualRatio-host.ScheduledVMs)),
			// 亲和性最大可生产数
			float64(maxCount-antiAffinityReplicas[g.antiAffinityValue(order.AntiAffinityLevel, host)]),
		))

		logs.V(5).Infof("host %s, module name: %s, total: %d, created: %d, allocatable: %d, observed replicas: %d",
			host.IP, host.ModuleName, order.TotalNum, existNum, host.AllocatableCount, replicas)

		if replicas <= 0 {
			continue
		}

		antiAffinityReplicas[g.antiAffinityValue(order.AntiAffinityLevel, host)] += replicas

		// launch docker vm
		wg.Add(1)
		go func(order *types.ApplyOrder, selector *types.DVMSelector, host *types.HostPriority, replicas uint) {
			defer wg.Done()
			genId, err := g.launchDvm(kt, order, selector, host, replicas)
			if err != nil {
				logs.Errorf("failed to launch dvm, err: %v", err)
				appendError(err)
			} else {
				logs.Infof("success to launch dvm, host: %+v, replicas: %d, order id: %s, generate id: %d", host,
					replicas, order.SubOrderId, genId)
				appendGenRecord(genId)
			}
		}(order, selector, &host, replicas)

		if order.TotalNum <= g.sumReplicas(antiAffinityReplicas) {
			break
		}
	}

	wg.Wait()

	if len(errs) > 0 {
		logs.Warnf("failed to generate dvm, errs: %v", errs)
	}

	if len(genRecordIds) == 0 {
		logs.Errorf("failed to generate dvm, for no host has generate record")
		return fmt.Errorf("failed to generate dvm, for no host has generate record")
	}

	return nil
}

func (g *Generator) parseDvmSelector(kt *kit.Kit, order *types.ApplyOrder) (*types.DVMSelector, error) {
	req := &cfgtypes.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "device_type",
						Operator: querybuilder.OperatorEqual,
						Value:    order.Spec.DeviceType,
					}},
			},
		},
		Page: metadata.BasePage{
			Limit: 1,
			Start: 0,
		},
	}

	rst, err := g.configLogics.Device().GetDvmDeviceType(kt, req)
	if err != nil {
		logs.Errorf("failed to get dvm device info, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("failed to get device info, err: %v", err)
	}
	cnt := len(rst.Info)
	if cnt != 1 {
		logs.Errorf("failed to get dvm device info, for invalid info cnt %d != 1, rid: %s", cnt, kt.Rid)
		return nil, fmt.Errorf("failed to get dvm device info, for invalid info cnt %d != 1", cnt)
	}

	deviceGroup, err := rst.Info[0].Label.String("device_group")
	if err != nil {
		logs.Errorf("failed to get dvm device info, for invalid label.device_group %v is not string, err: %v, rid: %s",
			rst.Info[0].Label, err, kt.Rid)
		return nil, fmt.Errorf("failed to get dvm device info, for invalid label.device_group %v is not string",
			rst.Info[0].Label)
	}
	selector := &types.DVMSelector{
		Cores:             int(rst.Info[0].Cpu),
		Memory:            int(rst.Info[0].Mem),
		Disk:              int(rst.Info[0].Disk),
		DeviceClass:       rst.Info[0].DeviceType,
		Image:             order.Spec.Image,
		Kernel:            order.Spec.Kernel,
		DockerType:        deviceGroup,
		NetworkType:       rst.Info[0].NetWork,
		DataDiskMountPath: order.Spec.MountPath,
		DataDiskType:      order.Spec.DiskType,
		DataDiskRaid:      order.Spec.RaidType,
		Region:            order.Spec.Region,
		Zone:              order.Spec.Zone,
		ExtranetIsp:       order.Spec.Isp,
		CpuProvider:       rst.Info[0].CpuProvider,
	}

	// set network type TENTHOUSAND by default
	if selector.NetworkType == "" {
		selector.NetworkType = "TENTHOUSAND"
	}

	// get amd device pattern
	if selector.CpuProvider != "" {
		selector.AmdDevicePattern = g.getAMDDevicePattern()
	}
	selector.HostRole = g.getSpecialAppRole(strconv.Itoa(int(order.BkBizId)))

	return selector, nil
}

// getUnreleasedDevice gets unreleased devices bindings to current apply order
func (g *Generator) getUnreleasedDevice(orderId string) ([]*types.DeviceInfo, error) {
	filter := &mapstr.MapStr{
		"suborder_id": orderId,
		// "is_delivered": true,
	}

	devices, err := model.Operation().DeviceInfo().GetDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get binding devices to order %s, err: %v", orderId, err)
		return nil, err
	}

	return devices, nil
}

// batchLaunchCvm  batch creates cvm and return created device ips
func (g *Generator) batchLaunchCvm(kt *kit.Kit, order *types.ApplyOrder, orderZones []string, replicas uint) ([]uint64,
	[]error) {

	logs.Infof("start batch launch cvm, sub order id: %s, orderZones: %v, replicas: %d, rid: %s",
		order.SubOrderId, orderZones, replicas, kt.Rid)

	var requestNum uint
	excludeSubnetIDMap := make(map[string]struct{})
	generateIDs := make([]uint64, 0)
	errs := make([]error, 0)
	mutex := sync.Mutex{}
	appendGenRecord := func(id uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		generateIDs = append(generateIDs, id)
	}
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	eg := errgroup.Group{}
	eg.SetLimit(5)

	for replicas > requestNum {
		curRequiredNum := replicas - requestNum
		generateID, err := g.initGenerateRecord(kt.Ctx, order.ResourceType, order.SubOrderId, curRequiredNum, false)
		if err != nil {
			logs.Errorf("failed to launch cvm when init generate record, err: %v, sub order id: %s, rid: %s", err,
				order.SubOrderId, kt.Rid)
			appendError(fmt.Errorf("failed to launch cvm, sub order id: %s, err: %v", order.SubOrderId, err))
			break
		}

		createCvmReq, err := g.buildGenRecordCvmReq(
			kt, generateID, order, orderZones, curRequiredNum, excludeSubnetIDMap)
		if err != nil {
			logs.Errorf("failed to launch cvm when build cvm request, err: %v, generateID: %d, sub order id: %s, "+
				"rid: %s", err, generateID, order.SubOrderId, kt.Rid)
			appendError(fmt.Errorf("failed to launch cvm, sub order id: %s, err: %v", order.SubOrderId, err))
			break
		}
		excludeSubnetIDMap[createCvmReq.SubnetId] = struct{}{}
		requestNum += createCvmReq.ApplyNumber

		eg.Go(func() error {
			if err = g.launchCvm(kt, order, createCvmReq, generateID); err != nil {
				logs.Errorf("failed to launch cvm, err: %v, sub order id: %s, zone: %s, replicas: %d, generateID: %d,"+
					" rid: %s", err, order.SubOrderId, createCvmReq.Zone, createCvmReq.ApplyNumber, generateID, kt.Rid)
				appendError(err)
				return nil
			}
			logs.Infof("success to launch cvm, sub order id: %s, zone: %s, replicas: %d, generate id: %d, rid: %s",
				order.SubOrderId, createCvmReq.Zone, createCvmReq.ApplyNumber, generateID, kt.Rid)
			appendGenRecord(generateID)
			return nil
		})
	}

	_ = eg.Wait()

	return generateIDs, errs
}

func (g *Generator) buildGenRecordCvmReq(kt *kit.Kit, generateID uint64, order *types.ApplyOrder, orderZones []string,
	replicas uint, excludeSubnetIDMap map[string]struct{}) (*types.CVM, error) {

	createCvmReq, err := g.buildCvmReq(kt, order, orderZones, replicas, excludeSubnetIDMap)
	if err != nil {
		logs.Errorf("failed to launch cvm when build cvm request, err: %v, order id: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		// update generate record status to Done
		if subErr := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateID,
			types.GenerateStatusFailed, err.Error(), "", nil); subErr != nil {
			logs.Errorf("failed to create cvm when update generate record, err: %v, order id: %s, rid: %s",
				subErr, order.SubOrderId, kt.Rid)
			return nil, subErr
		}
		return nil, err
	}

	if createCvmReq.ApplyNumber == replicas {
		return createCvmReq, nil
	}

	filter := &mapstr.MapStr{
		"generate_id": generateID,
	}
	now := time.Now()
	doc := mapstr.MapStr{
		"total_num": createCvmReq.ApplyNumber,
		"update_at": now,
	}
	if err = model.Operation().GenerateRecord().UpdateGenerateRecord(kt.Ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update generate record, err: %v, generate id: %d, update: %+v, rid: %s", err, generateID,
			doc, kt.Rid)
		return nil, err
	}

	return createCvmReq, nil
}

// launchCvm creates cvm and return created device ips
func (g *Generator) launchCvm(kt *kit.Kit, order *types.ApplyOrder, createCvmReq *types.CVM, generateId uint64) error {
	taskId, err := g.createCVM(kt, createCvmReq, order)
	if err != nil {
		logs.Errorf("scheduler:logics:launch:cvm:failed, failed to launch cvm when create generate task, "+
			"order id: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, errRecord, kt.Rid)
			return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		// CRP库存不足时，更新失败的可用区到主机申请单（只有分Campus生产失败，才需要记录）
		if strings.Contains(err.Error(), crpCapacityLackMsg) && order.Spec.IsCVMSeparateCampus() {
			if orderErr := g.updateOrderFailedZones(kt, order.SubOrderId, createCvmReq.Zone); orderErr != nil {
				// 只记录日志，不应该影响主流程
				logs.Warnf("failed to update order failed zoneIDs, subOrderID: %s, zone: %s, orderErr: %v, err: %v, "+
					"rid: %s", order.SubOrderId, createCvmReq.Zone, orderErr, err, kt.Rid)
			}
		}

		return fmt.Errorf("failed to launch cvm, subOrderID: %s, zone: %s, err: %v",
			order.SubOrderId, createCvmReq.Zone, err)
	}

	// update generate record status to Query
	if err = g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId, types.GenerateStatusHandling,
		"handling", taskId, nil); err != nil {
		logs.Errorf("scheduler:logics:launch:cvm:failed, failed to launch cvm when update generate record, "+
			"order id: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, err)
	}
	// check cvm task result and update generate record
	return g.AddCvmDevices(kt, taskId, generateId, order, createCvmReq.Zone)
}

// AddCvmDevices check generated device, create device infos and update generate record status
func (g *Generator) AddCvmDevices(kt *kit.Kit, taskId string, generateId uint64,
	order *types.ApplyOrder, zone string) error {

	// 1. check cvm task result
	if err := g.CheckCVM(kt, taskId, order.SubOrderId, order.Spec.ChargeType); err != nil {
		logs.Errorf("scheduler:logics:launch:cvm:failed, failed to create cvm when check generate task, "+
			"order id: %s, task id: %s, err: %v, rid: %s", order.SubOrderId, taskId, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateId, types.GenerateStatusFailed,
			err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, errRecord, kt.Rid)
			return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		// 更新失败的可用区到主机申请单（只有分Campus生产失败，才需要记录）
		if strings.Contains(err.Error(), crpProductFailedMsg) && order.Spec.IsCVMSeparateCampus() {
			if orderErr := g.updateOrderFailedZones(kt, order.SubOrderId, zone); orderErr != nil {
				// 只记录日志，不应该影响主流程
				logs.Warnf("failed to update order failed zoneIDs, subOrderID: %s, zone: %s, orderErr: %v, err: %v, "+
					"rid: %s", order.SubOrderId, zone, orderErr, err, kt.Rid)
			}
		}

		return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, zone: %s, err: %v", order.SubOrderId,
			taskId, zone, err)
	}

	// 2. get generated cvm instances
	hosts, err := g.listCVM(taskId)
	if err != nil {
		logs.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v, rid: %s",
			order.SubOrderId, taskId, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateId, types.GenerateStatusFailed,
			err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, errRecord, kt.Rid)
			return fmt.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
		}

		// 更新失败的可用区到主机申请单（只有分Campus生产失败，才需要记录）
		if strings.Contains(err.Error(), crpProductZeroNumMsg) && order.Spec.Zone == cvmapi.CvmSeparateCampus {
			if orderErr := g.updateOrderFailedZones(kt, order.SubOrderId, zone); orderErr != nil {
				// 只记录日志，不应该影响主流程
				logs.Warnf("failed to update order failed zoneIDs, subOrderID: %s, zone: %s, orderErr: %v, err: %v, "+
					"rid: %s", order.SubOrderId, zone, orderErr, err, kt.Rid)
			}
		}

		return fmt.Errorf("failed to list created cvm, order id: %s, task id: %s, zone: %s, err: %v",
			order.SubOrderId, taskId, zone, err)
	}
	// 3. create device infos
	return g.createDeviceInfo(kt, order, generateId, hosts, taskId)
}

func (g *Generator) createDeviceInfo(kt *kit.Kit, order *types.ApplyOrder, generateId uint64,
	hosts []*cvmapi.InstanceItem, taskId string) error {

	deviceList := make([]*types.DeviceInfo, 0)
	successIps := make([]string, 0)
	for _, host := range hosts {
		deviceList = append(deviceList, &types.DeviceInfo{
			Ip:               host.LanIp,
			AssetId:          host.AssetId,
			GenerateTaskId:   taskId,
			GenerateTaskLink: cvmapi.CvmOrderLinkPrefix + taskId,
			Deliverer:        "icr",
			CloudRegion:      host.CloudRegion,
			CloudZone:        host.CloudCampus, // 记录当前主机所在可用区
		})
		successIps = append(successIps, host.LanIp)
	}

	// NOTE: sleep 15 seconds to wait for CMDB host sync.
	time.Sleep(15 * time.Second)

	txnErr := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		// 1. save generated cvm instances info
		sessionKit := &kit.Kit{Ctx: sc, Rid: kt.Rid}
		if err := g.createGeneratedDevices(sessionKit, order, generateId, deviceList); err != nil {
			logs.Errorf("failed to update generated device, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
				kt.Rid)
			// update generate record status to Done
			// 不参与回滚
			if err := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
				types.GenerateStatusFailed, err.Error(), "", nil); err != nil {
				logs.Errorf("failed to update generate record, generate id: %d, err: %v, rid: %s", generateId, err,
					kt.Rid)
				return err
			}

			return fmt.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
		}

		// 2. update generate record status to success
		if err := g.UpdateGenerateRecord(sc, order.ResourceType, generateId, types.GenerateStatusSuccess, "success",
			"", successIps); err != nil {
			logs.Errorf("failed to launch cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, err, kt.Rid)
			return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId, taskId, err)
		}

		return nil
	})

	if txnErr != nil {
		logs.Errorf("failed to launch cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
			order.SubOrderId, taskId, txnErr, kt.Rid)
		return fmt.Errorf("failed to launch cvm when update generate record, order id: %s, task id: %s, "+
			"err: %v", order.SubOrderId, taskId, txnErr)
	}
	return nil
}

// launchDvm creates docker vm and return created device ips
func (g *Generator) launchDvm(kt *kit.Kit, order *types.ApplyOrder, applyRequest *types.DVMSelector,
	host *types.HostPriority, replicas uint) (uint64, error) {

	// 1. init generate record
	generateId, err := g.initGenerateRecord(kt.Ctx, order.ResourceType, order.SubOrderId, replicas, false)
	if err != nil {
		logs.Errorf("failed to launch docker vm when init generate record, order id: %s, err: %v", order.SubOrderId,
			err)
		return 0, fmt.Errorf("failed to launch docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 2. launch dvm request
	taskId, err := g.createDVM(applyRequest, order, host, replicas)
	if err != nil {
		logs.Errorf("failed to create docker vm when launch generate task, order id: %s, err: %v",
			order.SubOrderId, err)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create dvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch dvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}
		return generateId, fmt.Errorf("failed to create docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 3. update generate record status to Query
	if err := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId, types.GenerateStatusHandling,
		"handling", taskId, nil); err != nil {
		logs.Errorf("failed to launch docker vm when update generate record, order id: %s, err: %v", order.SubOrderId,
			err)
		return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 4. check cvm task result
	if err = g.checkDVM(taskId); err != nil {
		logs.Errorf("failed to launch docker vm when check generate task, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to launch docker vm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}

	// 5. get generated cvm instances
	hosts, err := g.listDVM(taskId)
	if err != nil {
		logs.Errorf("failed to list created docker vm, order id: %s, task id: %s, err: %v", order.SubOrderId, taskId,
			err)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create dvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch dvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to list created docker vm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}

	if err = g.saveGenerateCvmInstancesInfo(kt, taskId, generateId, order, hosts); err != nil {
		return generateId, err
	}

	return generateId, nil
}

func (g *Generator) saveGenerateCvmInstancesInfo(kt *kit.Kit, taskId string, generateId uint64,
	order *types.ApplyOrder, hosts []dvmapi.TaskList) error {

	deviceList := make([]*types.DeviceInfo, 0)
	successIps := make([]string, 0)
	for _, host := range hosts {
		if len(host.IP) <= 0 {
			continue
		}
		deviceList = append(deviceList, &types.DeviceInfo{
			Ip:               host.IP,
			GenerateTaskId:   taskId,
			GenerateTaskLink: fmt.Sprintf(dvmapi.DvmOrderLinkFormat, taskId),
			Deliverer:        "icr",
		})
		successIps = append(successIps, host.IP)
	}

	// 6. save generated cvm instances info
	if err := g.createGeneratedDevices(kt, order, generateId, deviceList); err != nil {
		logs.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
		return fmt.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 7. update generate record status to WaitForMatch
	if err := g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateId, types.GenerateStatusSuccess,
		"success", "", successIps); err != nil {
		logs.Errorf("failed to launch docker vm when update generate record, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
		return fmt.Errorf("failed to launch docker vm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}
	return nil
}

// getOrderGenRecords gets all generate records related to given order
func (g *Generator) getOrderGenRecords(suborderID string) ([]*types.GenerateRecord, error) {
	filter := map[string]interface{}{
		"suborder_id": suborderID,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	records, err := model.Operation().GenerateRecord().FindManyGenerateRecord(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get generate record by order id: %s", suborderID)
		return nil, err
	}

	return records, nil
}

// initGenerateRecord creates generate record
func (g *Generator) initGenerateRecord(ctx context.Context, resourceType types.ResourceType, subOrderId string,
	total uint, isManualMatched bool) (uint64, error) {

	id, err := model.Operation().GenerateRecord().NextSequence(ctx)
	if err != nil {
		logs.Errorf("failed to get generate record next sequence id, subOrderId: %s, err: %v", subOrderId, err)
		return 0, err
	}

	now := time.Now()
	record := &types.GenerateRecord{
		SubOrderId:      subOrderId,
		GenerateId:      id,
		GenerateType:    string(resourceType),
		Status:          types.GenerateStatusInit,
		IsMatched:       false,
		TotalNum:        total,
		SuccessNum:      0,
		SuccessList:     make([]string, 0),
		CreateAt:        now,
		UpdateAt:        now,
		StartAt:         now,
		IsManualMatched: isManualMatched, // 是否手工匹配
	}

	if err = model.Operation().GenerateRecord().CreateGenerateRecord(ctx, record); err != nil {
		logs.Errorf("failed to init generate record, subOrderId: %s, err: %v", subOrderId, err)
		return 0, err
	}

	return id, nil
}

// UpdateGenerateRecord updates generate record
func (g *Generator) UpdateGenerateRecord(ctx context.Context, resourceType types.ResourceType,
	generateId uint64, status types.GenerateStepStatus, msg, vmTaskId string, ipList []string) error {

	// TODO: filter add last status
	filter := &mapstr.MapStr{
		"generate_id": generateId,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    status,
		"update_at": now,
	}

	if len(msg) != 0 {
		doc["message"] = msg
	}

	if len(vmTaskId) != 0 {
		link := ""
		switch resourceType {
		case types.ResourceTypePool:
			link = PoolOrderLinkPrefix + vmTaskId
		case types.ResourceTypeCvm:
			link = cvmapi.CvmOrderLinkPrefix + vmTaskId
		case types.ResourceTypeQcloudDvm, types.ResourceTypeIdcDvm:
			link = fmt.Sprintf(dvmapi.DvmOrderLinkFormat, vmTaskId)
		case types.ResourceTypeUpgradeCvm:
			link = cvmapi.CvmUpgradeLinkPrefix + vmTaskId
		}
		doc["task_id"] = vmTaskId
		doc["task_link"] = link
	}

	if ipList != nil && len(ipList) > 0 {
		doc["success_num"] = len(ipList)
		doc["success_list"] = ipList
	}

	if status == types.GenerateStatusFailed || status == types.GenerateStatusSuccess {
		doc["end_at"] = now
	}

	if err := model.Operation().GenerateRecord().UpdateGenerateRecord(ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update generate record, generate id: %d, update: %+v, err: %v", generateId, doc, err)
		return err
	}
	return nil
}

func (g *Generator) createGeneratedDevices(kt *kit.Kit, order *types.ApplyOrder, generateId uint64,
	items []*types.DeviceInfo) error {

	ips := make([]string, 0)
	assetIds := make([]string, 0)
	for _, item := range items {
		ips = append(ips, item.Ip)
		assetIds = append(assetIds, item.AssetId)
	}

	mapAssetIDToHost, err := g.syncHostToCMDB(kt, order, generateId, ips, assetIds)
	if err != nil {
		logs.Errorf("failed to syn to cmdb, order id: %s, generateId: %d, err: %v, rid: %s", order.SubOrderId,
			generateId, err, kt.Rid)
		return err
	}

	devices := g.buildDevicesInfo(kt, items, order, generateId, mapAssetIDToHost)

	// 统计有母机IP的设备数量
	ownerIPCount := 0
	for _, device := range devices {
		if device.OwnerIP != "" {
			ownerIPCount++
		}
	}

	if err = model.Operation().DeviceInfo().CreateDeviceInfos(kt.Ctx, devices); err != nil {
		logs.Errorf("[OWNER_IP] failed to save device info to db, subOrderID: %s, generateId: %d, err: %v, "+
			"devicesNum: %d, ownerIPCount: %d, rid: %s", order.SubOrderId, generateId, err, len(devices),
			ownerIPCount, kt.Rid)
		return err
	}

	return nil
}

func (g *Generator) syncHostToCMDB(kt *kit.Kit, order *types.ApplyOrder, generateId uint64, ips, assetIDs []string) (
	map[string]*cmdb.Host, error) {

	// 线上Bug，返回了空的DeviceInfo数组，导致mongo插入失败
	if len(ips) == 0 && len(assetIDs) == 0 {
		logs.Errorf("failed to sync device info to cc, ips and assetIds is empty, subOrderID: %s, generateId: %s, "+
			"ips: %v, assetIDs: %v, rid: %s", order.SubOrderId, generateId, ips, assetIDs, kt.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "failed to sync device info to cc, ips and assetIds is empty, "+
			"subOrderID: %s", order.SubOrderId)
	}

	// 1. sync device info to cc
	if order.ResourceType == types.ResourceTypeCvm ||
		order.ResourceType == types.ResourceTypeUpgradeCvm {
		if err := g.syncHostByAsset(kt, assetIDs); err != nil {
			logs.Errorf("failed to sync device info to cc, order id: %s, err: %v, rid: %s",
				order.SubOrderId, err, kt.Rid)
			return nil, err
		}
	} else {
		if err := g.syncHostByIp(ips); err != nil {
			logs.Errorf("failed to sync device info to cc, order id: %s, err: %v, rid: %s",
				order.SubOrderId, err, kt.Rid)
			return nil, err
		}
	}

	// 2. get cc host detail info
	// 新申领的CVM默认在931业务下
	bizID := poolTypes.BizIDMatch
	bkModuleIDs := []int64{
		poolTypes.ModuleIDPoolRA,
		poolTypes.ModuleIDPoolMatch,
		poolTypes.ModuleIDPoolSCR,
	}
	if order.ResourceType == types.ResourceTypeUpgradeCvm {
		bizID = order.BkBizId
		bkModuleIDs = make([]int64, 0)
	}
	// 由于会存在主机在cc，但是此时机器还没有ip, 所以需要通过固资号进行查询
	ccHosts, err := g.getHostDetail(bizID, bkModuleIDs, assetIDs)
	if err != nil {
		logs.Errorf("failed to get cc host info, order id: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return nil, err
	}
	mapAssetIDToHost := make(map[string]*cmdb.Host)
	for _, host := range ccHosts {
		mapAssetIDToHost[host.BkAssetID] = host
	}

	logs.Infof("successfully sync device info to cc, subOrderID: %s, ips: %+v, assets: %+v, ccHostNum: %d, rid: %s",
		order.SubOrderId, ips, assetIDs, len(ccHosts), kt.Rid)
	return mapAssetIDToHost, nil
}

func (g *Generator) buildDevicesInfo(kt *kit.Kit, items []*types.DeviceInfo, order *types.ApplyOrder, generateId uint64,
	mapAssetIDToHost map[string]*cmdb.Host) []*types.DeviceInfo {

	now := time.Now()
	ownerIPMap := g.batchQueryOwnerIPs(kt, items, order.SubOrderId, mapAssetIDToHost)
	devices := g.buildDeviceInfoList(kt, items, order, generateId, mapAssetIDToHost, ownerIPMap, now)
	return devices
}

// batchQueryOwnerIPs 批量查询母机IP
func (g *Generator) batchQueryOwnerIPs(kt *kit.Kit, items []*types.DeviceInfo, subOrderID string,
	mapAssetIDToHost map[string]*cmdb.Host) map[string]string {
	// 收集需要查询的母机固资号
	ownerAssetIDs := make([]string, 0)
	hostAssetIDToOwnerAssetID := make(map[string]string) // 主机固资号 -> 母机固资号映射

	for _, item := range items {
		if item.AssetId == "" {
			continue
		}
		// 从CC主机信息中获取母机固资号
		host, ok := mapAssetIDToHost[item.AssetId]
		if !ok {
			logs.Warnf("[OWNER_IP] failed to get host info for assetID: %s, subOrderID: %s, rid: %s",
				item.AssetId, subOrderID, kt.Rid)
			continue
		}
		// 获取母机固资号
		if host.BKSvrOwnerAssetID == "" {
			logs.Warnf("[OWNER_IP] owner assetID is empty for host assetID: %s, subOrderID: %s, rid: %s",
				item.AssetId, subOrderID, kt.Rid)
			continue
		}
		// 去重收集母机固资号
		if _, exists := hostAssetIDToOwnerAssetID[item.AssetId]; !exists {
			ownerAssetIDs = append(ownerAssetIDs, host.BKSvrOwnerAssetID)
			hostAssetIDToOwnerAssetID[item.AssetId] = host.BKSvrOwnerAssetID
		}
	}

	ownerIPMap := make(map[string]string)
	if len(ownerAssetIDs) == 0 {
		logs.Warnf("[OWNER_IP] no owner assetID to query owner IP, subOrderID: %s, totalItems: %d, rid: %s",
			subOrderID, len(items), kt.Rid)
		return ownerIPMap
	}

	// 批量查询母机IP
	batchSize := 200
	batches := slice.Split(ownerAssetIDs, batchSize)
	ownerAssetIDToIP := make(map[string]string) // 母机固资号 -> 母机IP映射

	for idx, batch := range batches {
		hosts, err := g.cc.FindManyHostsByAssetID(kt, batch)
		if err != nil {
			logs.Warnf("[OWNER_IP] failed to batch get owner IP from bkcc, subOrderID: %s, batchNum: %d/%d, "+
				"ownerAssetIDs: %v, err: %v, rid: %s", subOrderID, idx+1, len(batches), batch, err, kt.Rid)
			continue
		}

		for _, host := range hosts {
			if host.BkHostInnerIP == "" {
				logs.Warnf("[OWNER_IP] owner IP is empty from bkcc, subOrderID: %s, ownerAssetID: %s, rid: %s",
					subOrderID, host.BkAssetID, kt.Rid)
				continue
			}
			ownerAssetIDToIP[host.BkAssetID] = host.BkHostInnerIP
		}
	}

	for hostAssetID, ownerAssetID := range hostAssetIDToOwnerAssetID {
		ownerIP, ok := ownerAssetIDToIP[ownerAssetID]
		if !ok {
			logs.Warnf("[OWNER_IP] failed to get owner IP for ownerAssetID: %s, hostAssetID: %s, subOrderID: %s, rid: %s",
				ownerAssetID, hostAssetID, subOrderID, kt.Rid)
			continue
		}
		ownerIPMap[hostAssetID] = ownerIP
	}

	return ownerIPMap
}

// buildDeviceInfoList 构建设备信息列表
func (g *Generator) buildDeviceInfoList(kt *kit.Kit, items []*types.DeviceInfo, order *types.ApplyOrder,
	generateId uint64, mapAssetIDToHost map[string]*cmdb.Host, ownerIPMap map[string]string,
	now time.Time) []*types.DeviceInfo {

	devices := make([]*types.DeviceInfo, 0, len(items))
	for _, item := range items {
		if isDup, _ := g.isDuplicateHost(order.SubOrderId, item.AssetId); isDup {
			continue
		}

		device := g.buildSingleDeviceInfo(kt, item, order, generateId, mapAssetIDToHost, ownerIPMap, now)
		devices = append(devices, device)
	}
	return devices
}

// buildSingleDeviceInfo 构建单个设备信息
func (g *Generator) buildSingleDeviceInfo(kt *kit.Kit, item *types.DeviceInfo, order *types.ApplyOrder,
	generateId uint64, mapAssetIDToHost map[string]*cmdb.Host, ownerIPMap map[string]string,
	now time.Time) *types.DeviceInfo {

	device := &types.DeviceInfo{
		OrderId:      order.OrderId,
		SubOrderId:   order.SubOrderId,
		GenerateId:   generateId,
		BkBizId:      int(order.BkBizId),
		User:         order.User,
		AssetId:      item.AssetId,
		Ip:           item.Ip,
		RequireType:  order.RequireType,
		ResourceType: order.ResourceType,
		// set device type according to order specification by default
		DeviceType:        order.Spec.DeviceType,
		Description:       order.Description,
		Remark:            order.Remark,
		IsMatched:         false,
		IsChecked:         false,
		IsInited:          false,
		IsDiskChecked:     false,
		IsDelivered:       false,
		GenerateTaskId:    item.GenerateTaskId,
		GenerateTaskLink:  item.GenerateTaskLink,
		InitTaskId:        item.InitTaskId,
		InitTaskLink:      item.InitTaskLink,
		DiskCheckTaskId:   item.DiskCheckTaskId,
		DiskCheckTaskLink: item.DiskCheckTaskLink,
		Deliverer:         item.Deliverer,
		IsManualMatched:   item.IsManualMatched,
		CloudZone:         item.CloudZone,
		CreateAt:          now,
		UpdateAt:          now,
	}
	// add device detail info from cc
	g.enrichDeviceInfoFromCC(kt, device, item.AssetId, order.SubOrderId, mapAssetIDToHost)
	g.setOwnerIP(kt, device, item, order.SubOrderId, ownerIPMap)
	return device
}

// enrichDeviceInfoFromCC 从CC获取设备详细信息
func (g *Generator) enrichDeviceInfoFromCC(kt *kit.Kit, device *types.DeviceInfo, assetID, subOrderID string,
	mapAssetIDToHost map[string]*cmdb.Host) {

	host, ok := mapAssetIDToHost[assetID]
	if !ok {
		logs.Warnf("failed to get host detail info in cc, subOrderID: %s, assetID: %s, rid: %s",
			subOrderID, assetID, kt.Rid)
		return
	}

	device.DeviceType = host.SvrDeviceClass
	device.ZoneName = host.SubZone
	zoneId, err := strconv.Atoi(host.SubZoneId)
	if err != nil {
		logs.Warnf("failed to convert sub zone id %s to int, rid: %s", host.SubZoneId, kt.Rid)
		device.ZoneID = 0
	} else {
		device.ZoneID = zoneId
	}
	device.ModuleName = host.ModuleName
	device.Equipment = host.RackId
}

// setOwnerIP 设置母机IP
func (g *Generator) setOwnerIP(kt *kit.Kit, device *types.DeviceInfo, item *types.DeviceInfo,
	subOrderID string, ownerIPMap map[string]string) {

	if ownerIP, ok := ownerIPMap[item.AssetId]; ok {
		device.OwnerIP = ownerIP
		logs.Infof("[OWNER_IP] get owner IP success, ip: %s, assetID: %s, ownerIP: %s, subOrderID: %s, rid: %s",
			item.Ip, item.AssetId, ownerIP, subOrderID, kt.Rid)
	} else {
		logs.Warnf("[OWNER_IP] failed to get owner IP, ip: %s, assetID: %s, subOrderID: %s, rid: %s",
			item.Ip, item.AssetId, subOrderID, kt.Rid)
	}
}

// getOwnerIPs 批量获取母机IP
func (g *Generator) getOwnerIPs(kt *kit.Kit, ownerAssetIDs []string) (map[string]string, error) {
	ownerIPMap := make(map[string]string)
	if len(ownerAssetIDs) == 0 {
		return ownerIPMap, nil
	}

	assetIDMap := make(map[string]struct{}, len(ownerAssetIDs))
	for _, id := range ownerAssetIDs {
		if id == "" {
			continue
		}
		assetIDMap[id] = struct{}{}
	}

	uniqueOwnerAssetIDs := maps.Keys(assetIDMap)
	if len(uniqueOwnerAssetIDs) == 0 {
		return ownerIPMap, nil
	}

	batchSize := 200
	batches := slice.Split(uniqueOwnerAssetIDs, batchSize)

	for idx, batch := range batches {
		hosts, err := g.cc.FindManyHostsByAssetID(kt, batch)
		if err != nil {
			logs.Warnf("[OWNER_IP] failed to batch get owner IP from bkcc, batchNum: %d/%d, ownerAssetIDs: %v, "+
				"err: %v, rid: %s", idx+1, len(batches), batch, err, kt.Rid)
			return nil, err
		}

		for _, host := range hosts {
			if host.BkHostInnerIP == "" {
				logs.Warnf("[OWNER_IP] owner IP is empty from bkcc, ownerAssetID: %s, rid: %s", host.BkAssetID, kt.Rid)
				continue
			}
			ownerIPMap[host.BkAssetID] = host.BkHostInnerIP
		}
	}

	return ownerIPMap, nil
}

func (g *Generator) isDuplicateHost(suborderID, assetID string) (bool, error) {
	filter := map[string]interface{}{
		"suborder_id": suborderID,
		"asset_id":    assetID,
	}

	cnt, err := model.Operation().DeviceInfo().CountDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to count device info, suborderID: %s, assetID: %s, err: %v", suborderID, assetID, err)
		return false, err
	}

	if cnt >= 1 {
		return true, nil
	}

	return false, nil
}

func (g *Generator) getHostDetail(bizID int64, bkModuleIDs []int64, assetIds []string) ([]*cmdb.Host, error) {
	req := &cmdb.ListBizHostParams{
		BizID:       bizID,
		BkModuleIDs: bkModuleIDs,
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_asset_id",
						Operator: querybuilder.OperatorIn,
						Value:    assetIds,
					},
				},
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_host_outerip",
			// 外网运营商
			"bk_ip_oper_name",
			// 机型
			"svr_device_class",
			"bk_os_name",
			// 地域
			"bk_zone_name",
			// 可用区(子Zone)
			"sub_zone",
			// 子ZoneID
			"sub_zone_id",
			"module_name",
			// 机架号，string类型
			"rack_id",
			"idc_unit_name",
			// 逻辑区域
			"logic_domain",
			"raid_name",
			"svr_input_time",
			// 母机固资号
			"bk_svr_owner_asset_id",
		},
		Page: &cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	resp, err := g.cc.ListBizHost(kit.New(), req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	hosts := make([]*cmdb.Host, 0)
	for _, host := range resp.Info {
		hosts = append(hosts, cvt.ValToPtr(host))
	}

	return hosts, nil
}

// MatchCVM manual match cvm devices 手工匹配CVM
func (g *Generator) MatchCVM(kt *kit.Kit, param *types.MatchDeviceReq, order *types.ApplyOrder) error {
	// cannot match device if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot match device, for order %s stage %s != %s, rid: %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend, kt.Rid)
		return fmt.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
	}
	// 升降配暂不支持手工匹配
	if order.ResourceType == types.ResourceTypeUpgradeCvm {
		logs.Errorf("cannot match device, for order %s resource type is %s, rid: %s", order.SubOrderId,
			order.ResourceType, kt.Rid)
		return fmt.Errorf("cannot match device, for order %s resource type is %s", order.SubOrderId,
			order.ResourceType)
	}
	// set apply order status MATCHING
	if err := g.lockApplyOrder(order); err != nil {
		logs.Errorf("failed to match cvm when lock apply order, err: %v, order id: %s, rid: %s", err, param.SuborderId,
			kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, param.SuborderId)
	}

	replicas := uint(len(param.Device))
	// 2. init generate record
	generateId, err := g.initGenerateRecord(kt.Ctx, order.ResourceType, order.SubOrderId, replicas, true)
	if err != nil {
		logs.Errorf("failed to match cvm when init generate record, err: %v, order id: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, order.SubOrderId)
	}

	// 根据固资号，查询主机可用区
	assetIDs := make([]string, 0)
	for _, host := range param.Device {
		assetIDs = append(assetIDs, host.AssetId)
	}
	hostMap, err := g.getHostZoneInfoByAssetIDs(kt, assetIDs)
	if err != nil {
		return fmt.Errorf("failed to batch get host list by assetIDs, err: %v, subOrderID: %s, rid: %s",
			err, order.SubOrderId, kt.Rid)
	}

	// TODO: check whether device is locked by other orders
	deviceList := make([]*types.DeviceInfo, 0)
	successIps := make([]string, 0)
	for _, host := range param.Device {
		deviceInfo := &types.DeviceInfo{
			Ip:              host.Ip,
			AssetId:         host.AssetId,
			Deliverer:       param.Operator,
			IsManualMatched: true,
		}
		// 这里只是补充deviceInfo里面的zone信息，即使zone没有值也不影响后续逻辑
		if hostInfo, ok := hostMap[host.AssetId]; ok {
			deviceInfo.CloudZone = hostInfo.BkCloudZone
			deviceInfo.ZoneName = hostInfo.SubZone
			deviceInfo.ZoneID, err = strconv.Atoi(hostInfo.SubZoneId)
			if err != nil {
				logs.Warnf("failed to convert sub zone id %s to int, subOrderID: %s, err: %+v, rid: %s",
					hostInfo.SubZoneId, order.SubOrderId, err, kt.Rid)
			}
		}
		deviceList = append(deviceList, deviceInfo)
		successIps = append(successIps, host.Ip)
	}

	// 3. save generated cvm instances info
	if err = g.createGeneratedDevices(kt, order, generateId, deviceList); err != nil {
		logs.Errorf("failed to update generated device, err: %v, order id: %s", err, order.SubOrderId)
		return fmt.Errorf("failed to update generated device, err: %v, order id: %s, rid: %s", err, order.SubOrderId,
			kt.Rid)
	}
	// 4. update generate record status to success
	msg := fmt.Sprintf("manually matched by %s successfully", param.Operator)
	if err = g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId, types.GenerateStatusSuccess,
		msg, "", successIps); err != nil {
		logs.Errorf("failed to match cvm when update generate record, err: %v, order id: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, order.SubOrderId)
	}
	return nil
}

// GetApplyOrder gets apply order by order id
func (g *Generator) GetApplyOrder(key string) (*types.ApplyOrder, error) {
	filter := &mapstr.MapStr{
		"suborder_id": key,
	}
	order, err := model.Operation().ApplyOrder().GetApplyOrder(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get apply order by id: %s", key)
		return nil, err
	}

	return order, nil
}

// lockApplyOrder locks apply order to avoid order repeat dispatch
func (g *Generator) lockApplyOrder(order *types.ApplyOrder) error {
	filter := &mapstr.MapStr{
		"suborder_id": order.SubOrderId,
	}

	doc := &mapstr.MapStr{
		"stage":     types.TicketStageRunning,
		"status":    types.ApplyStatusMatching,
		"update_at": time.Now(),
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to lock apply order, id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	return nil
}

// MatchPM automatically match physical machine devices
func (g *Generator) MatchPM(kt *kit.Kit, order *types.ApplyOrder) error {
	// 1. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// 2. check if need generate device
	existCount := uint(len(existDevices))
	if existCount >= order.TotalNum {
		logs.Infof("apply order %s has been scheduled %d pm", order.SubOrderId, existCount)
		// check if need retry match task
		if err = g.retryMatchDevice(existDevices); err != nil {
			logs.Warnf("failed to retry match device, order id: %s, err: %v", order.SubOrderId, err)
		}
		return nil
	}

	logs.Infof("apply order %s existing device number: %d", order.SubOrderId, existCount)

	// 3. match pm
	if err := g.matchPM(kt, order, existDevices); err != nil {
		logs.Errorf("failed to match pm, suborder id: %s", order.SubOrderId)
		return err
	}

	return nil
}

// MatchPoolDevice manual match pool devices
func (g *Generator) MatchPoolDevice(param *types.MatchPoolDeviceReq) error {
	// 1. get order by suborder id
	order, err := g.GetApplyOrder(param.SuborderId)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, order id: %s", err, param.SuborderId)
		return fmt.Errorf("failed to get apply order, err: %v, order id: %s", err, param.SuborderId)
	}

	// cannot match device if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
		return fmt.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
	}

	// set apply order status MATCHING
	if err := g.lockApplyOrder(order); err != nil {
		logs.Errorf("failed to lock apply order, err: %v, order id: %s", err, param.SuborderId)
		return fmt.Errorf("failed to lock apply order, err: %v, order id: %s", err, param.SuborderId)
	}

	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	genRecordIds := make([]uint64, 0)
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	appendGenRecord := func(id uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		genRecordIds = append(genRecordIds, id)
	}

	for _, task := range param.Spec {
		wg.Add(1)
		go func(order *types.ApplyOrder, recall *types.MatchPoolSpec) {
			defer wg.Done()
			newKit := kit.New()
			genId, err := g.launchRecallHost(newKit, order, recall)
			if err != nil {
				logs.Errorf("failed to launch recall order, err: %v", err)
				appendError(err)
			} else {
				logs.Infof("success to launch recall order, replicas: %d, order id: %s, generate id: %d",
					recall.Replicas, order.SubOrderId, genId)
				appendGenRecord(genId)
			}
		}(order, task)
	}

	wg.Wait()

	if len(errs) > 0 {
		logs.Warnf("failed to generate pool recall device, errs: %v", errs)
	}

	if len(genRecordIds) == 0 {
		logs.Errorf("failed to generate pool recall device, for no generate record")
		return fmt.Errorf("failed to generate pool recall device, for no generate record")
	}

	return nil
}

func (g *Generator) getHostZoneInfoByAssetIDs(kt *kit.Kit, assetIDs []string) (map[string]cmdb.Host, error) {
	req := &cmdb.ListHostWithoutBizParams{
		HostPropertyFilter: &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: []cmdb.Rule{
			&cmdb.AtomRule{Field: "bk_asset_id", Operator: cmdb.OperatorIn, Value: assetIDs},
		}}},
		Fields: []string{"bk_host_id", "bk_asset_id", "bk_cloud_zone", "sub_zone", "sub_zone_id"},
		Page:   &cmdb.BasePage{Start: 0, Limit: pkg.BKMaxInstanceLimit},
	}

	resp, err := g.cc.ListHostWithoutBiz(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host info by asset ids, err: %v, assetIDs: %v, rid: %s", err, assetIDs, kt.Rid)
		return nil, err
	}

	hostsMap := make(map[string]cmdb.Host, 0)
	for _, host := range resp.Info {
		if _, ok := hostsMap[host.BkAssetID]; ok {
			continue
		}
		hostsMap[host.BkAssetID] = host
	}

	return hostsMap, nil
}
