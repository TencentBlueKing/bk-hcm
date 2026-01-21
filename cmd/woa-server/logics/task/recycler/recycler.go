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

// Package recycler ...
package recycler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/biz"
	configLogics "hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/dissolve"
	"hcm/cmd/woa-server/logics/plan"
	rslogics "hcm/cmd/woa-server/logics/rolling-server"
	srlogics "hcm/cmd/woa-server/logics/short-rental"
	"hcm/cmd/woa-server/logics/task/recycler/classifier"
	"hcm/cmd/woa-server/logics/task/recycler/detector"
	"hcm/cmd/woa-server/logics/task/recycler/dispatcher"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/cmd/woa-server/logics/task/recycler/notifier"
	"hcm/cmd/woa-server/logics/task/recycler/returner"
	"hcm/cmd/woa-server/logics/task/recycler/transit"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/api-gateway/cmsi"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/language"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/util"

	"go.mongodb.org/mongo-driver/mongo"
)

// Interface recycler interface
type Interface interface {
	// RecycleCheck check whether hosts can be recycled or not
	RecycleCheck(kit *kit.Kit, param *types.RecycleCheckReq, bkBizIDMap map[int64]struct{}, resType meta.ResourceType,
		action meta.Action) (*types.RecycleCheckRst, error)
	// PreviewRecycleOrder preview resource recycle order
	PreviewRecycleOrder(kit *kit.Kit, param *types.PreviewRecycleReq, bkBizIDMap map[int64]struct{}) (
		*types.PreviewRecycleOrderCpuRst, error)
	// AuditRecycleOrder audit resource recycle orders
	AuditRecycleOrder(kit *kit.Kit, param *types.AuditRecycleReq) error
	// CreateRecycleOrder 创建回收单，并启动回收流程，Preview+SetNextState， 直接到 COMMITTED状态
	CreateRecycleOrder(kit *kit.Kit, param *types.CreateRecycleReq, bkBizIDMap map[int64]struct{},
		resType meta.ResourceType, action meta.Action) (*types.CreateRecycleOrderRst, error)
	// GetRecycleOrder gets resource recycle order info
	GetRecycleOrder(kit *kit.Kit, param *types.GetRecycleOrderReq) (*types.GetRecycleOrderRst, error)
	// GetRecycleDetect gets resource recycle detection task info
	GetRecycleDetect(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.GetDetectTaskRst, error)
	// ListDetectHost gets recycle detection host list
	ListDetectHost(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.ListDetectHostRst, error)
	// GetRecycleDetectStep gets resource recycle detection step info
	GetRecycleDetectStep(kit *kit.Kit, param *types.GetDetectStepReq) (*types.GetDetectStepRst, error)

	// StartRecycleOrder 提交回收单（转到 committed)
	StartRecycleOrder(kit *kit.Kit, param *types.StartRecycleOrderReq) error
	// StartRecycleOrderByRecycleType starts resource recycle order by recycle type
	StartRecycleOrderByRecycleType(kit *kit.Kit, param *types.StartRecycleOrderByRecycleTypeReq) error
	// StartDetectTask starts resource detection task
	StartDetectTask(kit *kit.Kit, param *types.StartDetectTaskReq) error
	// ReviseRecycleOrder revise recycle orders to remove detection failed hosts
	ReviseRecycleOrder(kit *kit.Kit, param *types.ReviseRecycleOrderReq) error
	// PauseRecycleOrder pauses resource recycle order
	PauseRecycleOrder(kit *kit.Kit, param mapstr.MapStr) error
	// ResumeRecycleOrder resumes resource recycle order
	ResumeRecycleOrder(kit *kit.Kit, param *types.ResumeRecycleOrderReq) error
	// TerminateRecycleOrder terminates resource recycle order
	TerminateRecycleOrder(kit *kit.Kit, param *types.TerminateRecycleOrderReq) error

	// GetRecycleHost gets resource recycle host info
	GetRecycleHost(kit *kit.Kit, param *types.GetRecycleHostReq) (*types.GetRecycleHostRst, error)
	// GetRecycleRecordDeviceType gets resource recycle record device type list
	GetRecycleRecordDeviceType(kit *kit.Kit) (*types.GetRecycleRecordDevTypeRst, error)
	// GetRecycleRecordRegion gets resource recycle record region list
	GetRecycleRecordRegion(kit *kit.Kit) (*types.GetRecycleRecordRegionRst, error)
	// GetRecycleRecordZone gets resource recycle record zone list
	GetRecycleRecordZone(kit *kit.Kit) (*types.GetRecycleRecordZoneRst, error)
	// GetRecycleBizHost gets business hosts in recycle module
	GetRecycleBizHost(kit *kit.Kit, param *types.GetRecycleBizHostReq) (*types.GetRecycleBizHostRst, error)
	// GetDetectStepCfg gets resource recycle step config info
	GetDetectStepCfg(kit *kit.Kit) (*types.GetDetectStepCfgRst, error)
	// GetDispatcher gets dispatcher instance
	GetDispatcher() *dispatcher.Dispatcher
	// GetShortRentalLogic get short rental logic
	GetShortRentalLogic() srlogics.Logics

	// RunRecycleTask run resource recycle detect task
	// RunRecycleTask(task *table.DetectTask, startStep uint)

	// CheckDetectStatus check whether detection is finished or not
	CheckDetectStatus(kt *kit.Kit, orderId string) error
	// CheckUworkOpenTicket check whether uwork has open ticket
	CheckUworkOpenTicket(kt *kit.Kit, assetID string) ([]string, error)
	// TransitCvm transit CVM resource
	TransitCvm(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event
	// DealTransitTask2Pool transit regular Pm resource
	DealTransitTask2Pool(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event
	// UpdateHostInfo update recycle host info
	UpdateHostInfo(order *table.RecycleOrder, stage table.RecycleStage, status table.RecycleStatus) error
	// DealTransitTask2Transit transit Pm resource which is dissolved or expired
	DealTransitTask2Transit(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event
	// TransferHost2BizTransit transfer host to business module
	TransferHost2BizTransit(kt *kit.Kit, hosts []*table.RecycleHost, srcBizID, srcModuleID, destBizId int64) error
	// RecoverReturnCvm recover retrun CVM resource without yunti orderId
	RecoverReturnCvm(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event
	// QueryReturnStatus query return status
	QueryReturnStatus(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event
	// UpdateOrderInfo update recycle order info
	UpdateOrderInfo(kt *kit.Kit, orderId, handler string, success, failed, pending uint, msg string) error
	// UpdateReturnTaskStatus update return task status
	UpdateReturnTaskStatus(kt *kit.Kit, suborderID string, status table.ReturnStatus, msg string) error

	StartStuckCheckLoop(kt *kit.Kit)
	StartFailedNoticeScanLoop(kt *kit.Kit)
}

// recycler provides resource recycle service
type recycler struct {
	lang language.CCLanguageIf
	// clientSet apimachinery.ClientSetInterface
	cc  cmdb.Client
	cvm cvmapi.CVMClientInterface

	dispatcher    *dispatcher.Dispatcher
	authorizer    auth.Authorizer
	configLogics  configLogics.Logics
	rsLogic       rslogics.Logics
	srLogic       srlogics.Logics
	bizLogic      biz.Logics
	dissolveLogic dissolve.Logics
	cmsiClient    cmsi.Client
	thirdCli      *thirdparty.Client
	notifier      *notifier.Notifier
}

// New create a recycler
func New(ctx context.Context, thirdCli *thirdparty.Client, bizLogic biz.Logics, cmdbCli cmdb.Client,
	authorizer auth.Authorizer, rsLogic rslogics.Logics, srLogics srlogics.Logics, dissolveLogic dissolve.Logics,
	cliSet *client.ClientSet, configLogics configLogics.Logics, planLogic plan.Logics, cmsiClient cmsi.Client) (
	Interface, error) {

	// new detector
	moduleDetector, err := detector.New(ctx, thirdCli, cmdbCli, cliSet)
	if err != nil {
		return nil, err
	}

	// new returner
	moduleReturner, err := returner.New(ctx, thirdCli, cmdbCli, planLogic)
	if err != nil {
		return nil, err
	}

	// new transit
	moduleTransit, err := transit.New(ctx, thirdCli, cmdbCli, rsLogic, cliSet)
	if err != nil {
		return nil, err
	}

	// new notifier
	notice, err := notifier.New(cmsiClient, thirdCli)
	if err != nil {
		return nil, err
	}

	// new dispatcher
	dispatch, err := dispatcher.New(ctx, moduleDetector, moduleReturner, moduleTransit, rsLogic, srLogics, notice)
	if err != nil {
		return nil, err
	}

	recycler := &recycler{
		lang:          language.NewFromCtx(language.EmptyLanguageSetting),
		cc:            cmdbCli,
		cvm:           thirdCli.CVM,
		dispatcher:    dispatch,
		authorizer:    authorizer,
		configLogics:  configLogics,
		rsLogic:       rsLogic,
		srLogic:       srLogics,
		bizLogic:      bizLogic,
		dissolveLogic: dissolveLogic,
		cmsiClient:    cmsiClient,
		thirdCli:      thirdCli,
		notifier:      notice,
	}

	return recycler, nil
}

// RecoverReturnCvm recover return CVM resource which return task is int
func (r *recycler) RecoverReturnCvm(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event {
	return r.dispatcher.GetReturn().RecoverReturnCvm(kt, task, hosts)
}

// GetDispatcher get dispatcher
func (r *recycler) GetDispatcher() *dispatcher.Dispatcher {
	return r.dispatcher
}

// GetShortRentalLogic get short rental logic
func (r *recycler) GetShortRentalLogic() srlogics.Logics {
	return r.srLogic
}

// UpdateReturnTaskStatus update return task status
func (r *recycler) UpdateReturnTaskStatus(kt *kit.Kit, suborderID string, status table.ReturnStatus, msg string) error {
	return r.dispatcher.GetReturn().UpdateReturnTaskStatus(kt, suborderID, status, msg)
}

// RecycleCheck check whether hosts can be recycled or not
func (r *recycler) RecycleCheck(kt *kit.Kit, param *types.RecycleCheckReq, bkBizIDMap map[int64]struct{},
	resType meta.ResourceType, action meta.Action) (*types.RecycleCheckRst, error) {

	if kt.User == "" {
		logs.Errorf("failed to recycle check, for invalid user is empty, rid: %s", kt.Rid)
		return nil, errors.New("failed to recycle check, for invalid user is empty")
	}

	// 1. get host base info
	hostBase, err := r.getHostBaseInfo(kt, param.IPs, param.AssetIDs, param.HostIDs)
	if err != nil {
		logs.Errorf("failed to recycle check, for list host err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(hostBase) == 0 {
		return nil, errf.New(errf.RecordNotFound, "host not found")
	}

	hostIds := make([]int64, 0)
	for _, host := range hostBase {
		hostIds = append(hostIds, host.BkHostID)
	}

	// 2. get host topo info
	relations, err := r.getHostTopoInfo(kt, hostIds)
	if err != nil {
		logs.Errorf("failed to recycle check, for list host err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	bizIds := make([]int64, 0)
	mapBizToModule := make(map[int64][]int64)
	mapHostToRel := make(map[int64]*cmdb.HostTopoRelation)
	for _, rel := range relations {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[rel.BizID]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID:%d is located is not in "+
				"the bizIDMap:%+v passed in", rel.BizID, rel.HostID, bkBizIDMap)
		}

		mapHostToRel[rel.HostID] = rel
		if _, ok := mapBizToModule[rel.BizID]; !ok {
			mapBizToModule[rel.BizID] = []int64{rel.BkModuleID}
			bizIds = append(bizIds, rel.BizID)
		} else {
			mapBizToModule[rel.BizID] = append(mapBizToModule[rel.BizID], rel.BkModuleID)
		}
	}

	mapBizIdToBiz, err := r.getBizInfo(kt, bizIds)
	if err != nil {
		logs.Errorf("failed to recycle check, for get business info err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	mapModuleIdToModule, err := r.getModuleInfoMap(kt, mapBizToModule)
	if err != nil {
		logs.Errorf("failed to recycle check, for get module info err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 3. check recycle permissions
	mapBizPermission := r.checkRecyclePermissions(kt, resType, action, bizIds)

	// 4. check recyclability and create check result
	rst := r.checkRecycleAbility(kt, hostBase, mapHostToRel, mapBizIdToBiz, mapBizPermission, mapModuleIdToModule)

	return rst, nil
}

func (r *recycler) checkRecycleAbility(kt *kit.Kit, hostBase []*cmdb.Host,
	mapHostToRel map[int64]*cmdb.HostTopoRelation, mapBizIdToBiz map[int64]*cmdb.Biz, mapBizPermission map[int64]bool,
	mapModuleIdToModule map[int64]*cmdb.ModuleInfo) *types.RecycleCheckRst {

	checkInfos := make([]*types.RecycleCheckInfo, 0)
	for _, host := range hostBase {
		bizId := int64(0)
		moduleId := int64(0)
		if rel, ok := mapHostToRel[host.BkHostID]; ok {
			bizId = rel.BizID
			moduleId = rel.BkModuleID
		}
		bizName := ""
		if biz, ok := mapBizIdToBiz[bizId]; ok {
			bizName = biz.BizName
		} else if bizId == 5000008 {
			// 业务资源池
			bizName = "业务资源池"
		}
		var moduleDefaultVal int64
		var topoModule string
		if module, ok := mapModuleIdToModule[moduleId]; ok {
			moduleDefaultVal = module.Default
			topoModule = module.BkModuleName
		}
		hasPermission := false
		if permission, ok := mapBizPermission[bizId]; ok {
			hasPermission = permission
		}
		checkInfo := &types.RecycleCheckInfo{
			HostID:           host.BkHostID,
			AssetID:          host.BkAssetID,
			IP:               host.GetUniqIp(),
			BkHostOuterIP:    host.GetUniqOuterIp(),
			BizID:            bizId,
			BizName:          bizName,
			ModuleDefaultVal: moduleDefaultVal,
			TopoModule:       topoModule,
			Operator:         host.Operator,
			BakOperator:      host.BkBakOperator,
			DeviceType:       host.SvrDeviceClass,
			State:            host.SrvStatus,
			InputTime:        host.SvrInputTime,
		}
		r.fillCheckInfo(checkInfo, kt.User, hasPermission)
		checkInfos = append(checkInfos, checkInfo)
	}

	rst := &types.RecycleCheckRst{
		Count: int64(len(checkInfos)),
		Info:  checkInfos,
	}
	return rst
}

func (r *recycler) checkRecyclePermissions(kt *kit.Kit, resType meta.ResourceType, action meta.Action,
	bizIds []int64) map[int64]bool {

	mapBizPermission := make(map[int64]bool)
	for _, bizId := range bizIds {
		err := r.authorizer.AuthorizeWithPerm(kt, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Warnf("failed to check recycle permission, bizID: %d, err: %v", bizId, err)
			mapBizPermission[bizId] = false
			continue
		}
		mapBizPermission[bizId] = true
	}
	return mapBizPermission
}

func (r *recycler) getModuleInfoMap(kt *kit.Kit, mapBizToModule map[int64][]int64) (
	map[int64]*cmdb.ModuleInfo, error) {

	mapModuleIdToModule := make(map[int64]*cmdb.ModuleInfo)
	for bizId, moduleIds := range mapBizToModule {
		moduleIdUniq := util.IntArrayUnique(moduleIds)
		moduleList, err := r.getModuleInfo(kt, bizId, moduleIdUniq)
		if err != nil {
			logs.Errorf("failed to recycle check, for get module info err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, module := range moduleList {
			mapModuleIdToModule[module.BkModuleId] = module
		}
	}
	return mapModuleIdToModule, nil
}

// fillCheckInfo fill host with recyclability check info
func (r *recycler) fillCheckInfo(host *types.RecycleCheckInfo, user string, hasPermission bool) {
	if !hasPermission {
		host.Recyclable = false
		host.Message = "无该业务资源回收权限，请至权限中心申请"
	} else if classifier.IsUnsupportedDevice(host.AssetID, host.IP) {
		host.Recyclable = false
		host.Message = "非YUNTI/SCR平台申请设备，请至具体申领平台回收"
	} else if host.ModuleDefaultVal != cmdb.DftModuleRecycle {
		host.Recyclable = false
		host.Message = "主机不在空闲机池下的待回收模块中"
	} else if strings.Contains(host.Operator, user) == false && strings.Contains(host.BakOperator, user) == false {
		host.Recyclable = false
		host.Message = "必须为主机负责人或备份负责人"
	} else if !r.hasInnnerIP(host.IP) {
		host.Recyclable = false
		host.Message = "主机没有内网IP，不可回收"
	} else {
		host.Recyclable = true
		host.Message = "可回收"
	}
}

// hasInnnerIP 检查主机是否有内网IP
func (r *recycler) hasInnnerIP(ip string) bool {
	return strings.TrimSpace(ip) != ""
}

// getHostBaseInfo get host detail info for recycle
func (r *recycler) getHostDetailInfo(kt *kit.Kit, ips, assetIds []string, hostIds []int64) (
	[]*table.RecycleHost, error) {

	// 1. get host base info
	hostBase, err := r.getHostBaseInfo(kt, ips, assetIds, hostIds)
	if err != nil {
		logs.Errorf("failed to get host detail info, for list host err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(hostBase) == 0 {
		return make([]*table.RecycleHost, 0), nil
	}

	bkHostIds := make([]int64, 0)
	for _, host := range hostBase {
		bkHostIds = append(bkHostIds, host.BkHostID)
	}

	// 2. get host biz info
	mapHostToRel, mapBizIdToBiz, mapModuleIdToModule, err := r.getBizModuleRelByHostIDs(kt, bkHostIds)
	if err != nil {
		return nil, err
	}

	// 3. fill host info
	hostDetails := r.getHostDetails(kt, hostBase, mapHostToRel, mapBizIdToBiz, mapModuleIdToModule)

	// 4. fill cvm info
	if err = r.fillCvmInfo(kt, hostDetails); err != nil {
		logs.Errorf("failed to fill cvm info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return hostDetails, nil
}

func (r *recycler) getBizModuleRelByHostIDs(kt *kit.Kit, bkHostIds []int64) (map[int64]*cmdb.HostTopoRelation,
	map[int64]*cmdb.Biz, map[int64]*cmdb.ModuleInfo, error) {

	relations, err := r.getHostTopoInfo(kt, bkHostIds)
	if err != nil {
		logs.Errorf("failed to get host detail info, for list host err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	bizIds := make([]int64, 0)
	mapHostToRel := make(map[int64]*cmdb.HostTopoRelation)
	mapBizToModule := make(map[int64][]int64)
	for _, rel := range relations {
		bizIds = append(bizIds, rel.BizID)
		mapHostToRel[rel.HostID] = rel

		// 记录业务ID跟模块ID的映射
		if _, ok := mapBizToModule[rel.BizID]; !ok {
			mapBizToModule[rel.BizID] = make([]int64, 0)
		}
		mapBizToModule[rel.BizID] = append(mapBizToModule[rel.BizID], rel.BkModuleID)
	}
	bizIds = util.IntArrayUnique(bizIds)

	mapBizIdToBiz, err := r.getBizInfo(kt, bizIds)
	if err != nil {
		logs.Errorf("failed to get host detail info, for get business info err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, err
	}

	mapModuleIdToModule := make(map[int64]*cmdb.ModuleInfo)
	for bizId, moduleIds := range mapBizToModule {
		moduleIdUniq := slice.Unique(moduleIds)
		moduleList, err := r.getModuleInfo(kt, bizId, moduleIdUniq)
		if err != nil {
			logs.Errorf("failed to recycle check, for get module info err: %v, bizId: %d, moduleIds: %v, "+
				"rid: %s", err, bizId, moduleIds, kt.Rid)
			return nil, nil, nil, err
		}
		for _, module := range moduleList {
			mapModuleIdToModule[module.BkModuleId] = module
		}
	}
	return mapHostToRel, mapBizIdToBiz, mapModuleIdToModule, nil
}

func (r *recycler) getHostDetails(kt *kit.Kit, hostBase []*cmdb.Host,
	mapHostToRel map[int64]*cmdb.HostTopoRelation, mapBizIdToBiz map[int64]*cmdb.Biz,
	mapModuleIdToModule map[int64]*cmdb.ModuleInfo) []*table.RecycleHost {

	hostDetails := make([]*table.RecycleHost, 0)
	for _, host := range hostBase {
		bizId := int64(0)
		moduleId := int64(0)
		if rel, ok := mapHostToRel[host.BkHostID]; ok {
			bizId = rel.BizID
			moduleId = rel.BkModuleID
		}
		bizName := ""
		if biz, ok := mapBizIdToBiz[bizId]; ok {
			bizName = biz.BizName
		} else if bizId == 5000008 {
			// 业务资源池
			bizName = "业务资源池"
		}
		var moduleDefaultVal int64
		if module, ok := mapModuleIdToModule[moduleId]; ok {
			moduleDefaultVal = module.Default
		}

		hostDetail := &table.RecycleHost{
			BizID:           bizId,
			BizName:         bizName,
			HostID:          host.BkHostID,
			AssetID:         host.BkAssetID,
			IP:              host.GetUniqIp(),
			BkHostOuterIP:   host.GetUniqOuterIp(),
			DeviceType:      host.SvrDeviceClass,
			Zone:            host.BkZoneName,
			SubZone:         host.SubZone,
			ModuleName:      host.ModuleName,
			Operator:        host.Operator,
			BakOperator:     host.BkBakOperator,
			InputTime:       host.SvrInputTime,
			SvrSourceTypeID: host.SvrSourceTypeID,
		}
		// 检查该主机是否可回收
		checkInfo := &types.RecycleCheckInfo{
			AssetID:          hostDetail.AssetID,
			IP:               hostDetail.IP,
			ModuleDefaultVal: moduleDefaultVal,
			Operator:         hostDetail.Operator,
			BakOperator:      hostDetail.BakOperator,
		}
		r.fillCheckInfo(checkInfo, kt.User, true)
		hostDetail.Recyclable = checkInfo.Recyclable
		hostDetail.RecycleMessage = checkInfo.Message

		hostDetails = append(hostDetails, hostDetail)
	}
	return hostDetails
}

// getHostBaseInfo get host base info in cc 3.0
func (r *recycler) getHostBaseInfo(kt *kit.Kit, ips, assetIds []string, hostIds []int64) ([]*cmdb.Host, error) {
	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionOr,
		Rules:     make([]querybuilder.Rule, 0),
	}
	if len(ips) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.CombinedRule{
			Condition: querybuilder.ConditionAnd,
			Rules: []querybuilder.Rule{
				querybuilder.AtomRule{
					Field:    "bk_host_innerip",
					Operator: querybuilder.OperatorIn,
					Value:    ips,
				},
				// support bk_cloud_id 0 only
				querybuilder.AtomRule{
					Field:    "bk_cloud_id",
					Operator: querybuilder.OperatorEqual,
					Value:    0,
				},
			},
		})
	}
	if len(assetIds) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_asset_id",
			Operator: querybuilder.OperatorIn,
			Value:    assetIds,
		})
	}
	if len(hostIds) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_host_id",
			Operator: querybuilder.OperatorIn,
			Value:    hostIds,
		})
	}

	req := &cmdb.ListHostReq{
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: rule,
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_host_outerip",
			// 机型
			"svr_device_class",
			// 逻辑区域
			"logic_domain",
			"bk_zone_name",
			"sub_zone",
			"module_name",
			"raid_name",
			"svr_input_time",
			"operator",
			"bk_bak_operator",
			"srv_status",
			"bk_svr_source_type_id",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	resp, err := r.cc.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return cvt.SliceToPtr(resp.Info), nil
}

// getHostTopoInfo get host topo info in cc 3.0
func (r *recycler) getHostTopoInfo(kt *kit.Kit, hostIds []int64) ([]*cmdb.HostTopoRelation, error) {
	req := &cmdb.HostModuleRelationParams{
		HostID: hostIds,
	}

	resp, err := r.cc.FindHostBizRelations(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host topo info, err: %v", err)
		return nil, err
	}

	return cvt.SliceToPtr(cvt.PtrToVal(resp)), nil
}

// getBizInfo get business info in cc 3.0
func (r *recycler) getBizInfo(kt *kit.Kit, bizIds []int64) (map[int64]*cmdb.Biz, error) {
	req := &cmdb.SearchBizParams{
		BizPropertyFilter: &cmdb.QueryFilter{
			Rule: cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					cmdb.AtomRule{
						Field:    "bk_biz_id",
						Operator: cmdb.OperatorIn,
						Value:    bizIds,
					},
				},
			},
		},
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}

	resp, err := r.cc.SearchBusiness(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc business info, err: %v", err)
		return nil, err
	}

	mapBizIdToBiz := make(map[int64]*cmdb.Biz)
	for _, biz := range resp.Info {
		mapBizIdToBiz[biz.BizID] = &biz
	}
	return mapBizIdToBiz, nil
}

// getModuleInfo get module info in cc 3.0
func (r *recycler) getModuleInfo(kit *kit.Kit, bizId int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error) {
	req := &cmdb.SearchModuleParams{
		BizID: bizId,
		Condition: mapstr.MapStr{
			"bk_module_id": mapstr.MapStr{
				pkg.BKDBIN: moduleIds,
			},
		},
		Fields: []string{"bk_module_id", "bk_module_name", "default"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}

	resp, err := r.cc.SearchModule(kit, req)
	if err != nil {
		logs.Errorf("failed to get cc module info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	return resp.Info, nil
}

// PreviewRecycleOrder preview resource recycle order
func (r *recycler) PreviewRecycleOrder(kt *kit.Kit, param *types.PreviewRecycleReq, bkBizIDMap map[int64]struct{}) (
	*types.PreviewRecycleOrderCpuRst, error) {

	// 1. get hosts info
	hosts, err := r.getHostDetailInfo(kt, param.IPs, param.AssetIDs, param.HostIDs)
	if err != nil {
		logs.Errorf("failed to preview recycle order, for list host err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 2. fill host resource type and return plan
	// Must be done before filling recycle type, as the caiche recycle type depends on the host's resource type
	classifier.FillClassifyInfo(hosts, param.ReturnPlan)

	// 分业务处理
	bkBizHostsMap := make(map[int64][]*table.RecycleHost)
	for _, host := range hosts {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[host.BizID]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID:%d is located is not in "+
				"the bizIDMap:%+v passed in", host.BizID, host.HostID, bkBizIDMap)
		}
		bkBizHostsMap[host.BizID] = append(bkBizHostsMap[host.BizID], host)
	}

	for bkBizID, bizHosts := range bkBizHostsMap {
		// 3. fill host recycle type info
		bkBizHostsMap[bkBizID], err = r.fillHostRecycleType(kt, bkBizID, bizHosts, param.RecycleTypeSequence)
		if err != nil {
			logs.Errorf("failed to fill host recycle type, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	// 4. classify hosts into groups with different recycle strategies
	groups, err := classifier.ClassifyRecycleGroups(bkBizHostsMap, param.ReturnPlan)
	if err != nil {
		logs.Errorf("failed to preview recycle order, for classify hosts err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 5. 查询每个回收子订单的机型对应的CPU核数并填充到orders里面
	orders, err := r.createAndSaveRecycleOrders(kt, param.SkipConfirm, param.Remark, groups)
	if err != nil {
		logs.Errorf("failed to preview recycle order, create and save recycle orders, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.PreviewRecycleOrderCpuRst{
		Info: orders,
	}

	return rst, nil
}

// fillRecycleOrderDeviceList 创建初始化回收Order及主机Hosts并将机型对应的CPU核数填充到orders里面
func (r *recycler) createAndSaveRecycleOrders(kt *kit.Kit, skipConfirm bool, remark string,
	bizGroups map[int64]classifier.RecycleGroup) ([]*types.RecycleOrderCpuInfo, error) {

	// init and save recycle orders
	orders, subOrderIDDeviceTypes, err := r.initAndSaveRecycleOrders(kt, skipConfirm, remark,
		bizGroups)
	if err != nil {
		logs.Errorf("failed to preview recycle order, init and save orders err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	subOrdersMap := make(map[string]*table.RecycleOrder, 0)
	for _, subOrderItem := range orders {
		subOrdersMap[subOrderItem.SuborderID] = subOrderItem
	}

	orderList := make([]*types.RecycleOrderCpuInfo, 0)
	// 根据CVM机型列表获取CPU核数
	for subOrderID, deviceTypeMap := range subOrderIDDeviceTypes {
		deviceTypes := maps.Keys(deviceTypeMap)
		deviceTypesMap, err := r.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
		if err != nil {
			logs.Errorf("failed to preview recycle order, get cvm instance by device type failed, err: %v, rid: %s",
				err, kt.Rid)
			return nil, err
		}

		// 汇总CPU核数
		sumCpuCore := int64(0)
		for devideType, deviceItem := range deviceTypesMap {
			// 机型对应的CPU核数 * 该机型的回收数量
			sumCpuCore += deviceItem.CPUAmount * deviceTypeMap[devideType]
		}

		orderList = append(orderList, &types.RecycleOrderCpuInfo{
			RecycleOrder: subOrdersMap[subOrderID],
			SumCpuCore:   sumCpuCore,
		})
	}
	return orderList, nil
}

// initRecycleOrder init and save recycle orders
func (r *recycler) initAndSaveRecycleOrders(kt *kit.Kit, skipConfirm bool, remark string,
	bizGroups map[int64]classifier.RecycleGroup) ([]*table.RecycleOrder, map[string]map[string]int64, error) {

	subOrderIDDeviceTypes := make(map[string]map[string]int64, 0)
	orders := make([]*table.RecycleOrder, 0)
	txnErr := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		for biz, groups := range bizGroups {
			id, err := dao.Set().RecycleOrder().NextSequence(sc)
			if err != nil {
				return errf.New(pkg.CCErrObjectDBOpErrno, err.Error())
			}
			index := 1
			for grpType, group := range groups {
				if len(group) <= 0 {
					continue
				}
				// 1. init recycle order
				bizName := group[0].BizName
				order := newRecycleOrder(kt, id, biz, bizName, grpType, skipConfirm, index, len(group), remark)
				// 记录回收订单日志，方便排查问题
				logs.Infof("start to create recycle order for save recycle host, orderInfo: %+v, group: %+v, rid: %s",
					order, cvt.PtrToSlice(group), kt.Rid)
				// 2. create and save recycle hosts
				if err = r.initAndSaveHosts(sc, order, group); err != nil {
					logs.Errorf("failed to create recycle order for save recycle host err: %v, rid: %s", err, kt.Rid)
					return fmt.Errorf("failed to create recycle order for save recycle host err: %v", err)
				}
				// 3. save recycle order
				if err = dao.Set().RecycleOrder().CreateRecycleOrder(sc, order); err != nil {
					logs.Errorf("failed to create recycle order for save recycle order err: %v, rid: %s", err, kt.Rid)
					return fmt.Errorf("failed to create recycle order for save recycle order err: %v", err)
				}
				// 4.插入退还主机的滚服匹配记录
				if grpType.IsRollingServerType() {
					if err = r.rsLogic.InsertReturnedHostMatched(kt, biz, order.OrderID, order.SuborderID,
						group, enumor.LockedStatus); err != nil {
						logs.Errorf("failed to create returned host matched for save recycle order err: %v, "+
							"subOrderID: %s, rid: %s",
							err, order.SuborderID, kt.Rid)
						return fmt.Errorf("failed to create returned host matched for save recycle order err: %v", err)
					}
				}
				// 5. 记录子订单跟机型的关系
				for _, groupItem := range group {
					if _, ok := subOrderIDDeviceTypes[order.SuborderID]; !ok {
						subOrderIDDeviceTypes[order.SuborderID] = make(map[string]int64, 0)
					}
					if _, ok := subOrderIDDeviceTypes[order.SuborderID][groupItem.DeviceType]; !ok {
						subOrderIDDeviceTypes[order.SuborderID][groupItem.DeviceType] = 0
					}
					subOrderIDDeviceTypes[order.SuborderID][groupItem.DeviceType]++
				}
				orders = append(orders, order)
				index++
			}
		}
		return nil
	})
	if txnErr != nil {
		return nil, nil, fmt.Errorf("failed to init and save recycle orders, err: %v, rid: %s", txnErr, kt.Rid)
	}
	return orders, subOrderIDDeviceTypes, nil
}

func newRecycleOrder(kt *kit.Kit, id uint64, biz int64, bizName string, grpType classifier.RecycleGrpType,
	skipConfirm bool, index, groupLen int, remark string) *table.RecycleOrder {

	now := time.Now()
	return &table.RecycleOrder{
		OrderID:            id,
		SuborderID:         fmt.Sprintf("%d-%d", id, index),
		BizID:              biz,
		BizName:            bizName,
		User:               kt.User,
		ResourceType:       classifier.MapGroupProperty[grpType].ResourceType,
		RecycleType:        classifier.MapGroupProperty[grpType].RecycleType,
		ReturnPlan:         classifier.MapGroupProperty[grpType].ReturnType,
		CostConcerned:      classifier.MapGroupProperty[grpType].CostConcerned,
		SkipConfirm:        skipConfirm,
		Stage:              table.RecycleStageCommit,
		Status:             table.RecycleStatusUncommit,
		Handler:            "AUTO",
		TotalNum:           uint(groupLen),
		SuccessNum:         0,
		PendingNum:         0,
		FailedNum:          0,
		Remark:             remark,
		CreateAt:           now,
		UpdateAt:           now,
		ReturnForecast:     false,
		ReturnForecastTime: "",
	}
}

// initAndSaveHosts inits and saves recycle hosts
func (r *recycler) initAndSaveHosts(ctx context.Context, order *table.RecycleOrder, hosts []*table.RecycleHost) error {
	now := time.Now()
	costRate := 0.0
	if order.ResourceType == table.ResourceTypePm && order.RecycleType == table.RecycleTypeRegular {
		costRate = 0.6
	}
	for _, host := range hosts {
		host.OrderID = order.OrderID
		host.SuborderID = order.SuborderID
		host.User = order.User
		host.Stage = order.Stage
		host.Status = order.Status
		host.ReturnCostRate = costRate
		host.CreateAt = now
		host.UpdateAt = now

		if err := dao.Set().RecycleHost().CreateRecycleHost(ctx, host); err != nil {
			logs.Errorf("failed to save recycle host, ip: %s, err: %v, subOrderId: %s", host.IP, err, order.SuborderID)
			return fmt.Errorf("failed to save recycle host, ip: %s, err: %v, subOrderId: %s", host.IP, err,
				order.SuborderID)
		}
	}

	return nil
}

func (r *recycler) fillCvmInfo(kt *kit.Kit, hostDetails []*table.RecycleHost) error {
	hosts := make([]*table.RecycleHost, 0)
	for _, host := range hostDetails {
		if classifier.IsQcloudCvm(host.AssetID) {
			hosts = append(hosts, host)
		}
	}

	// skip query cvm instance when input no hosts
	if len(hosts) == 0 {
		return nil
	}

	ipList := make([]string, 0)
	mapIp2Host := make(map[string]*table.RecycleHost)
	for _, host := range hosts {
		ipList = append(ipList, host.IP)
		mapIp2Host[host.IP] = host
	}

	req := &cvmapi.InstanceQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmInstanceStatusMethod,
		},
		Params: &cvmapi.InstanceQueryParam{
			LanIp: ipList,
		},
	}

	resp, err := r.cvm.QueryCvmInstances(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to query cvm instance, err: %v", err)
		return err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("query cvm failed, code: %d, msg: %s, crpTraceID: %s, ipList: %v, rid: %s", resp.Error.Code,
			resp.Error.Message, resp.TraceId, ipList, kt.Rid)
		return fmt.Errorf("query cvm failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	for _, inst := range resp.Result.Data {
		if host, ok := mapIp2Host[inst.LanIp]; ok {
			host.InstID = inst.InstanceId
			host.ObsProject = inst.ObsProject
			host.CloudRegionID = inst.CloudRegion
			if inst.Pool == 1 {
				host.Pool = table.PoolPublic
			} else {
				host.Pool = table.PoolPrivate
			}
		}
	}

	return nil
}

// fillHostRecycleType fill host recycle type
func (r *recycler) fillHostRecycleType(kt *kit.Kit, bkBizID int64, hosts []*table.RecycleHost,
	recycleTypeSeq []table.RecycleType) ([]*table.RecycleHost, error) {

	assetIDs := make([]string, 0, len(hosts))
	for _, host := range hosts {
		assetIDs = append(assetIDs, host.AssetID)
	}

	// 1. 根据资产编号获取裁撤主机列表
	dissolveHostMap, err := r.dissolveLogic.RecycledHost().IsDissolveHost(kt, assetIDs)
	if err != nil {
		logs.Errorf("failed to check if host is dissolve host, err: %v, assetIDs: %v, rid: %s", err, assetIDs, kt.Rid)
		return nil, err
	}

	// 2. 识别固定的回收类型：机房裁撤，春节保障；其中机房裁撤 > 春节保障
	// 过滤出物理机和其他类型机器
	pmHosts := make([]*table.RecycleHost, 0)
	otherHosts := make([]*table.RecycleHost, 0)
	for _, host := range hosts {
		recycleType := classifier.GetFixedRecycleType(host, dissolveHostMap[host.AssetID])
		if host.RecycleType.CanUpdateRecycleType(recycleTypeSeq, recycleType) {
			host.RecycleType = recycleType
		}

		if cmdb.IsPhysicalMachine(host.SvrSourceTypeID) {
			pmHosts = append(pmHosts, host)
			continue
		}
		otherHosts = append(otherHosts, host)
	}

	// 3. 识别根据退回计划匹配的回收类型
	// 3.1. 根据顺序匹配回收类型
	for _, rType := range recycleTypeSeq {
		switch rType {
		case table.RecycleTypeRollServer:
			otherHosts, err = r.matchRollingServer(kt, bkBizID, otherHosts, recycleTypeSeq)
			if err != nil {
				logs.Errorf("failed to match rolling server, err: %v, hosts: %v, rid: %s", err, otherHosts, kt.Rid)
				return nil, err
			}
		case table.RecycleTypeShortRental:
			otherHosts, err = r.matchShortRental(kt, bkBizID, otherHosts, recycleTypeSeq)
			if err != nil {
				logs.Errorf("failed to match short rental, err: %v, hosts: %v, rid: %s", err, otherHosts, kt.Rid)
				return nil, err
			}
		default:
			continue
		}
	}
	hosts = append(pmHosts, otherHosts...)

	return hosts, nil
}

// AuditRecycleOrder audit resource recycle orders
func (r *recycler) AuditRecycleOrder(kit *kit.Kit, param *types.AuditRecycleReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)

	if cnt != 1 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("get invalid recycle order count %d != 1", cnt)
	}

	// verify order status
	order := insts[0]
	if order.Status != table.RecycleStatusAudit {
		logs.Errorf("failed to audit recycle order, for order status %s is not %s", order.Status,
			table.RecycleStatusAudit)
		return fmt.Errorf("failed to audit recycle order, for order status %s is not %s", order.Status,
			table.RecycleStatusAudit)
	}

	// verify operator permission
	operator := kit.User
	if operator != "dommyzhang" && operator != "forestchen" {
		logs.Errorf("failed to audit recycle order, for operator has no permission")
		return errors.New("failed to audit recycle order, for operator has no permission")
	}

	task := dispatcher.NewTask(order.Status, r.srLogic)
	taskCtx := &dispatcher.AuditContext{
		Order:      order,
		Dispatcher: r.dispatcher,
		Operator:   operator,
		Approval:   param.Approval,
		Remark:     param.Remark,
	}
	if err := task.State.Execute(taskCtx); err != nil {
		logs.Errorf("failed to execute audit task, err: %v, order id: %s, state: %s", err, order.SuborderID,
			task.State.Name())
		return err
	}

	return nil
}

// CreateRecycleOrder 创建回收单，并启动回收流程，Preview+SetNextState， 直接到COMMITTED状态
func (r *recycler) CreateRecycleOrder(kt *kit.Kit, param *types.CreateRecycleReq, bkBizIDMap map[int64]struct{},
	resType meta.ResourceType, action meta.Action) (*types.CreateRecycleOrderRst, error) {

	// 1. get hosts info
	hosts, err := r.getHostDetailInfo(kt, param.IPs, param.AssetIDs, param.HostIDs)
	if err != nil {
		logs.Errorf("failed to create recycle order, for list host err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(hosts) == 0 {
		logs.Errorf("get no valid host to create recycle order")
		return nil, errors.New("get no valid host to create recycle order")
	}

	// 2. check permission
	bizIds := make([]int64, 0)
	for _, host := range hosts {
		// 如果访问的是业务下的接口，但是查出来的业务不属于当前业务，需要报错或过滤掉
		if _, ok := bkBizIDMap[host.BizID]; !ok && len(bkBizIDMap) > 0 {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d where the hostID:%d is located is not in "+
				"the bizIDMap:%+v passed in", host.BizID, host.HostID, bkBizIDMap)
		}
		// 校验该主机是否可回收
		if !host.Recyclable {
			return nil, errf.Newf(errf.InvalidParameter, "bizID:%d, hostID:%d, IP: %s, can not be recycled, "+
				"message: %s", host.BizID, host.HostID, host.IP, host.RecycleMessage)
		}

		bizIds = append(bizIds, host.BizID)
	}
	bizIds = util.IntArrayUnique(bizIds)

	for _, bizId := range bizIds {
		err = r.authorizer.AuthorizeWithPerm(kt, meta.ResourceAttribute{
			Basic: &meta.Basic{Type: resType, Action: action}, BizID: bizId,
		})
		if err != nil {
			logs.Errorf("failed to check recycle permission, bizID: %d, bkBizIDMap: %+v, err: %v",
				bizId, bkBizIDMap, err)
			return nil, err
		}
	}

	previewResult, err := r.PreviewRecycleOrder(kt, param.ToPreviewParam(), bkBizIDMap)
	if err != nil {
		logs.Errorf("failed to create recycle order, for preview err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	orders := make([]*table.RecycleOrder, len(previewResult.Info))
	for i, info := range previewResult.Info {
		if info == nil || info.RecycleOrder == nil {
			logs.Errorf("failed to create recycle order, for preview result info is nil, rid: %s", kt.Rid)
			return nil, errors.New("failed to create recycle order, for preview result info is nil")
		}

		orders[i] = info.RecycleOrder
	}

	r.setOrderNextStatus(kt, orders)

	rst := &types.CreateRecycleOrderRst{
		Info: orders,
	}

	return rst, nil
}

// GetRecycleOrder gets resource recycle order info
func (r *recycler) GetRecycleOrder(kit *kit.Kit, param *types.GetRecycleOrderReq) (*types.GetRecycleOrderRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle order, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetRecycleOrderRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().RecycleOrder().CountRecycleOrder(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle order count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.RecycleOrder, 0)
		return rst, nil
	}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetRecycleDetect gets resource recycle detection task info
func (r *recycler) GetRecycleDetect(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.GetDetectTaskRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle detection task, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetDetectTaskRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().DetectTask().CountDetectTask(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle detection task count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.DetectTask, 0)
		return rst, nil
	}

	insts, err := dao.Set().DetectTask().FindManyDetectTask(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection task, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// ListDetectHost gets recycle detection host list
func (r *recycler) ListDetectHost(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.ListDetectHostRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle detection task, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	insts, err := dao.Set().DetectTask().GetRecycleHostList(kit.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection task, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.ListDetectHostRst{
		Info: insts,
	}

	return rst, nil
}

// GetRecycleDetectStep gets resource recycle detection step info
func (r *recycler) GetRecycleDetectStep(kit *kit.Kit, param *types.GetDetectStepReq) (*types.GetDetectStepRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle detection step, for get filter err: %v, param: %+v, rid: %s",
			err, cvt.PtrToVal(param), kit.Rid)
		return nil, err
	}

	rst := &types.GetDetectStepRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().DetectStep().CountDetectStep(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle detection step count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.DetectStep, 0)
		return rst, nil
	}

	insts, err := dao.Set().DetectStep().FindManyDetectStep(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection step, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// StartRecycleOrder 提交回收单（转到 committed)
func (r *recycler) StartRecycleOrder(kt *kit.Kit, param *types.StartRecycleOrderReq) error {
	filter := map[string]interface{}{}
	if len(param.OrderID) > 0 {
		filter["order_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.OrderID,
		}
	}

	if len(param.SuborderID) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		}
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kt.Rid)
		return fmt.Errorf("found no recycle order to start")
	}

	// 校验回收订单是否有滚服剩余额度
	err = r.rsLogic.CheckReturnedStatusBySubOrderID(kt, insts)
	if err != nil {
		logs.Errorf("failed to check recycle order status, err: %v, param: %+v, rid: %s",
			err, cvt.PtrToVal(param), kt.Rid)
		return err
	}

	err = r.updateOrderReturnForecastConfig(kt, insts, param.ReturnForecastConfigs)
	if err != nil {
		logs.Errorf("failed to update recycle order return forecast config, err: %v, param: %+v, rid: %s",
			err, cvt.PtrToVal(param), kt.Rid)
		return err
	}

	r.setOrderNextStatus(kt, insts)

	return nil
}

func (r *recycler) setOrderNextStatus(kt *kit.Kit, orders []*table.RecycleOrder) {
	now := time.Now()
	for _, order := range orders {
		nextStatus := order.Status
		failedNum := uint(0)
		switch order.Status {
		case table.RecycleStatusUncommit:
			nextStatus = table.RecycleStatusCommitted
			// 根据回收子订单ID解锁滚服回收的状态(仅限未提交状态)
			err := r.rsLogic.UpdateReturnedStatusBySubOrderID(kt, order.BizID, order.SuborderID, enumor.NormalStatus)
			if err != nil {
				logs.Errorf("failed to set order %s to next status, update returned status failed err: %v, rid: %s",
					order.SuborderID, err, kt.Rid)
				continue
			}
			if order.RecycleType == table.RecycleTypeShortRental {
				// 根据回收子订单ID创建短租退回记录
				err = r.srLogic.CreateReturnedHostRecord(kt, order.BizID, order.OrderID, order.SuborderID,
					enumor.ShortRentalStatusReturning)
				if err != nil {
					logs.Errorf("failed to create short rental returned record: %v, suborderID: %s, rid: %s",
						err, order.SuborderID, kt.Rid)
					continue
				}
			}
		case table.RecycleStatusDetectFailed:
			nextStatus = table.RecycleStatusDetecting
		case table.RecycleStatusTransitFailed:
			nextStatus = table.RecycleStatusTransiting
		case table.RecycleStatusReturnFailed:
			nextStatus = table.RecycleStatusReturning
			failedNum = order.FailedNum
		case table.RecycleStatusReturnPlanFailed:
			nextStatus = table.RecycleStatusReturningPlan
		default:
			logs.Warnf("failed to set order %s to next status, for unsupported status %s", order.SuborderID,
				order.Status)
			continue
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"failed_num":   failedNum,
			"status":       nextStatus,
			"recycle_type": order.RecycleType,
			"update_at":    now,
		}
		if nextStatus == table.RecycleStatusCommitted {
			(*update)["committed_at"] = now
		}

		// do not dispatch order to start if set next status failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s committed, err: %v", order.SuborderID, err)
			continue
		}

		// add order to dispatch queue
		r.dispatcher.Add(order.SuborderID)
	}
}

// StartDetectTask starts resource detection task
func (r *recycler) StartDetectTask(kit *kit.Kit, param *types.StartDetectTaskReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to start")
	}

	// check status
	for _, order := range insts {
		// cannot restart detection task if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
			return fmt.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
		}
	}

	// set order status detecting
	r.setOrderDetecting(insts)

	return nil
}

func (r *recycler) setOrderDetecting(orders []*table.RecycleOrder) {
	now := time.Now()
	for _, order := range orders {
		// need not set order detecting if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Warnf("failed to set order %s committed, for invalid status %s", order.SuborderID, order.Status)
			continue
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"failed_num": 0,
			"status":     table.RecycleStatusDetecting,
			"update_at":  now,
		}

		// do not dispatch order to start if set committed failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s detecting, err: %v", order.SuborderID, err)
			continue
		}

		// add order to dispatch queue
		r.dispatcher.Add(order.SuborderID)
	}
}

// ReviseRecycleOrder revise recycle orders to remove detection failed hosts
func (r *recycler) ReviseRecycleOrder(kit *kit.Kit, param *types.ReviseRecycleOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to start")
	}

	// check status
	for _, order := range insts {
		// cannot restart detection task if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
			return fmt.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
		}
	}

	// remove detection failed hosts and set order status detecting
	if err := r.reviseOrder(kit, insts); err != nil {
		logs.Errorf("failed to revise recycle order, err: %v", err)
		return fmt.Errorf("failed to revise recycle order, err: %v", err)
	}

	return nil
}

func (r *recycler) reviseOrder(kt *kit.Kit, orders []*table.RecycleOrder) error {
	now := time.Now()
	for _, order := range orders {
		// need not set order detecting if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Warnf("failed to set order %s committed, for invalid status %s", order.SuborderID, order.Status)
			continue
		}

		// get detection failed ips
		filter := map[string]interface{}{
			"suborder_id": order.SuborderID,
			"status":      table.DetectStatusFailed,
		}
		ips, err := dao.Set().DetectTask().GetRecycleHostList(context.Background(), filter)
		if err != nil {
			logs.Errorf("failed to get order %s detection failed host list, err: %v", order.SuborderID, err)
			return fmt.Errorf("failed to get order %s detection failed host list, err: %v", order.SuborderID, err)
		}

		failedNum := uint(len(ips))
		if failedNum >= order.TotalNum {
			logs.Errorf("cannot revise order %s, for all detection task failed", order.SuborderID)
			return fmt.Errorf("cannot revise order %s, for all detection task failed", order.SuborderID)
		}

		// remove detection failed hosts
		if err := r.removeDetectFailedHosts(order.SuborderID, ips); err != nil {
			logs.Errorf("failed to remove detection failed hosts, err: %v", err)
			return fmt.Errorf("failed to remove detection failed hosts, err: %v", err)
		}

		leftNum := order.TotalNum - failedNum
		orderFilter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"total_num":   leftNum,
			"success_num": 0,
			"pending_num": leftNum,
			"failed_num":  0,
			"status":      table.RecycleStatusDetecting,
			"update_at":   now,
		}

		// do not dispatch order to start if set committed failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), orderFilter, update); err != nil {
			logs.Errorf("failed to set order %s detecting, err: %v", order.SuborderID, err)
			return fmt.Errorf("failed to set order detecting, err: %v", err)
		}

		// add order to dispatch queue
		r.dispatcher.Add(order.SuborderID)
	}

	return nil
}

func (r *recycler) removeDetectFailedHosts(orderID string, ips []interface{}) error {
	filter := map[string]interface{}{
		"suborder_id": orderID,
		"ip": map[string]interface{}{
			pkg.BKDBIN: ips,
		},
	}

	if _, err := dao.Set().RecycleHost().DeleteRecycleHost(context.Background(), filter); err != nil {
		logs.Errorf("failed to delete recycle host, err: %v", err)
		return err
	}

	if _, err := dao.Set().DetectTask().DeleteDetectTask(context.Background(), filter); err != nil {
		logs.Errorf("failed to delete detection task, err: %v", err)
		return err
	}

	if _, err := dao.Set().DetectStep().DeleteDetectStep(context.Background(), filter); err != nil {
		logs.Errorf("failed to delete detection task step, err: %v", err)
		return err
	}

	return nil
}

func (r *recycler) getRecycleTaskById(kit *kit.Kit, taskId string) (*table.DetectTask, error) {
	filter := map[string]interface{}{
		"task_id": taskId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: 1,
	}

	insts, err := dao.Set().DetectTask().FindManyDetectTask(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle task, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	cnt := len(insts)
	if cnt != 1 {
		logs.Errorf("get invalid recycle task count %d != 1, rid: %s", cnt, kit.Rid)
		return nil, fmt.Errorf("get invalid recycle task count %d != 1, rid: %s", cnt, kit.Rid)
	}

	return insts[0], nil
}

// PauseRecycleOrder pauses resource recycle order
func (r *recycler) PauseRecycleOrder(_ *kit.Kit, _ mapstr.MapStr) error {
	// TODO
	return nil
}

// ResumeRecycleOrder resumes resource recycle order
func (r *recycler) ResumeRecycleOrder(kit *kit.Kit, param *types.ResumeRecycleOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to terminate")
	}

	// add order to dispatch queue
	for _, order := range insts {
		r.dispatcher.Add(order.SuborderID)
	}

	return nil
}

// TerminateRecycleOrder terminates resource recycle order
func (r *recycler) TerminateRecycleOrder(kt *kit.Kit, param *types.TerminateRecycleOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			pkg.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kt.Rid)
		return fmt.Errorf("found no recycle order to terminate")
	}

	// check status
	for _, order := range insts {
		// cannot terminate detection task if it's not detect failed
		switch order.Status {
		case table.RecycleStatusDone, table.RecycleStatusTerminate:
			logs.Errorf("need not terminate order %s, for its status %s, rid: %s",
				order.SuborderID, order.Status, kt.Rid)
			return fmt.Errorf("need not terminate order %s, for its status %s", order.SuborderID, order.Status)
		case table.RecycleStatusTransiting, table.RecycleStatusReturning:
			logs.Errorf("cannot terminate order %s, for its status %s, rid: %s", order.SuborderID, order.Status, kt.Rid)
			return fmt.Errorf("cannot terminate order %s, for its status %s", order.SuborderID, order.Status)
		}
	}

	// set order status terminate
	if err = r.terminateOrder(kt, insts); err != nil {
		logs.Errorf("failed to revise recycle order, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("failed to revise recycle order, err: %v", err)
	}

	return nil
}

func (r *recycler) terminateOrder(kt *kit.Kit, orders []*table.RecycleOrder) error {
	now := time.Now()
	for _, order := range orders {
		switch order.Status {
		case table.RecycleStatusDone, table.RecycleStatusTerminate:
			logs.Errorf("need not terminate order %s, for its status %s, rid: %s",
				order.SuborderID, order.Status, kt.Rid)
			return fmt.Errorf("need not terminate order %s, for its status %s", order.SuborderID, order.Status)
		case table.RecycleStatusTransiting, table.RecycleStatusReturning:
			logs.Errorf("cannot terminate order %s, for its status %s, rid: %s", order.SuborderID, order.Status, kt.Rid)
			return fmt.Errorf("cannot terminate order %s, for its status %s", order.SuborderID, order.Status)
		case table.RecycleStatusDetecting:
			if err := r.dispatcher.Cancel(kt, order.SuborderID); err != nil {
				logs.Errorf("fail to call dispatcher to cancel detecting order %s, err: %s, rid: %s",
					order.SuborderID, err, kt.Rid)
				return err
			}
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"stage":     table.RecycleStageTerminate,
			"status":    table.RecycleStatusTerminate,
			"update_at": now,
		}

		// do not dispatch order to start if set committed failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s detecting, err: %v, rid: %s", order.SuborderID, err, kt.Rid)
			return fmt.Errorf("failed to terminate order %s, err:%v", order.SuborderID, err)
		}

		// 根据回收子订单ID更新滚服回收的状态
		if err := r.rsLogic.UpdateReturnedStatusBySubOrderID(kt, order.BizID, order.SuborderID,
			enumor.TerminateStatus); err != nil {
			logs.Errorf("failed to update rolling server returned record status, subOrderID: %s, err: %v, rid: %s",
				order.SuborderID, err, kt.Rid)
			return fmt.Errorf("failed to terminate order %s, err:%v", order.SuborderID, err)
		}
		// 根据回收子订单ID更新短租回收的状态
		if err := r.srLogic.UpdateReturnedStatusBySubOrderID(kt, order.SuborderID,
			enumor.ShortRentalStatusTerminate); err != nil {
			logs.Errorf("failed to update short rental returned record status, subOrderID: %s, err: %v, rid: %s",
				order.SuborderID, err, kt.Rid)
			return fmt.Errorf("failed to terminate order %s, err:%v", order.SuborderID, err)
		}
	}

	return nil
}

// GetRecycleHost gets resource recycle host info
func (r *recycler) GetRecycleHost(kit *kit.Kit, param *types.GetRecycleHostReq) (*types.GetRecycleHostRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle host, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetRecycleHostRst{}

	if param.Page.EnableCount {
		cnt, err := dao.Set().RecycleHost().CountRecycleHost(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle host count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.RecycleHost, 0)
		return rst, nil
	}

	insts, err := dao.Set().RecycleHost().FindManyRecycleHost(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle host, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetRecycleRecordDeviceType gets resource recycle record device type list
func (r *recycler) GetRecycleRecordDeviceType(kit *kit.Kit) (*types.GetRecycleRecordDevTypeRst, error) {
	filter := map[string]interface{}{}
	insts, err := dao.Set().RecycleHost().Distinct(kit.Ctx, "device_type", filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRecycleRecordDevTypeRst{
		Info: insts,
	}

	return rst, nil
}

// GetRecycleRecordRegion gets resource recycle record region list
func (r *recycler) GetRecycleRecordRegion(kit *kit.Kit) (*types.GetRecycleRecordRegionRst, error) {
	filter := map[string]interface{}{}
	insts, err := dao.Set().RecycleHost().Distinct(kit.Ctx, "bk_zone_name", filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRecycleRecordRegionRst{
		Info: make([]interface{}, 0),
	}
	for _, tmpZoneName := range insts {
		if len(metadata.GetString(tmpZoneName)) == 0 {
			continue
		}
		rst.Info = append(rst.Info, tmpZoneName)
	}

	return rst, nil
}

// GetRecycleRecordZone gets resource recycle record zone list
func (r *recycler) GetRecycleRecordZone(kit *kit.Kit) (*types.GetRecycleRecordZoneRst, error) {
	filter := map[string]interface{}{}
	insts, err := dao.Set().RecycleHost().Distinct(kit.Ctx, "sub_zone", filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRecycleRecordZoneRst{
		Info: make([]interface{}, 0),
	}

	for _, tmpSubZone := range insts {
		if len(metadata.GetString(tmpSubZone)) == 0 {
			continue
		}
		rst.Info = append(rst.Info, tmpSubZone)
	}

	return rst, nil
}

// GetRecycleBizHost gets business hosts in recycle module
func (r *recycler) GetRecycleBizHost(kit *kit.Kit, param *types.GetRecycleBizHostReq) (*types.GetRecycleBizHostRst,
	error) {

	req := &cmdb.ListBizHostParams{
		BizID: param.BizID,
		ModuleCond: []cmdb.ConditionItem{
			// recycle module's default is 3
			{
				Field:    "default",
				Operator: "$eq",
				Value:    3,
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"operator",
			"bk_bak_operator",
			"svr_device_class",
			"sub_zone",
			"svr_input_time",
			"srv_status",
		},
		Page: &cmdb.BasePage{
			Start: int64(param.Page.Start),
			Limit: int64(param.Page.Limit),
		},
	}

	if param.Page.EnableCount {
		req.Page.Start = 0
		req.Page.Limit = 1
	}

	resp, err := r.cc.ListBizHost(kit, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	hostList := make([]*types.RecycleBizHost, 0)
	rst := new(types.GetRecycleBizHostRst)
	if param.Page.EnableCount {
		rst.Count = int64(resp.Count)
		return rst, nil
	}

	for _, ccHost := range resp.Info {
		host := &types.RecycleBizHost{
			HostID:      ccHost.BkHostID,
			AssetID:     ccHost.BkAssetID,
			IP:          ccHost.GetUniqIp(),
			Operator:    ccHost.Operator,
			BakOperator: ccHost.BkBakOperator,
			DeviceType:  ccHost.SvrDeviceClass,
			SubZone:     ccHost.SubZone,
			State:       ccHost.SrvStatus,
			InputTime:   ccHost.SvrInputTime,
		}
		hostList = append(hostList, host)
	}

	rst.Info = hostList
	return rst, nil
}

// GetDetectStepCfg gets resource recycle detection step config info
func (r *recycler) GetDetectStepCfg(kit *kit.Kit) (*types.GetDetectStepCfgRst, error) {
	filter := map[string]interface{}{
		"enable": true,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	insts, err := dao.Set().DetectStepCfg().GetDetectStepConfig(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection step, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetDetectStepCfgRst{
		Info: insts,
	}

	return rst, nil
}

// CheckDetectStatus check recycle task info
func (r *recycler) CheckDetectStatus(kt *kit.Kit, orderId string) error {
	return r.dispatcher.GetDetector().CheckDetectStatus(kt, orderId)
}

// CheckUworkOpenTicket ckeck host uwork ticket status
func (r *recycler) CheckUworkOpenTicket(kt *kit.Kit, assetID string) ([]string, error) {
	return r.dispatcher.GetDetector().GetUworkOpenTicketByAssetID(kt, assetID)
}

// DealTransitTask2Pool deal recycle task info
func (r *recycler) DealTransitTask2Pool(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	return r.dispatcher.GetTransit().DealTransitTask2Pool(order, hosts)
}

// TransitCvm transit cvm
func (r *recycler) TransitCvm(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	return r.dispatcher.GetTransit().TransitCvm(order, hosts)
}

// UpdateHostInfo update host info
func (r *recycler) UpdateHostInfo(order *table.RecycleOrder, stage table.RecycleStage,
	status table.RecycleStatus) error {
	return r.dispatcher.GetTransit().UpdateHostInfo(order, stage, status)
}

// DealTransitTask2Transit deal recycle task info
func (r *recycler) DealTransitTask2Transit(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	return r.dispatcher.GetTransit().DealTransitTask2Transit(order, hosts)
}

// TransferHost2BizTransit transit host to biz
func (r *recycler) TransferHost2BizTransit(kt *kit.Kit, hosts []*table.RecycleHost,
	srcBizID, srcModuleID, destBizId int64) error {

	return r.dispatcher.GetTransit().TransferHost2BizTransit(kt, hosts, srcBizID, srcModuleID, destBizId)
}

// QueryReturnStatus transit host to biz
func (r *recycler) QueryReturnStatus(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event {
	return r.dispatcher.GetReturn().QueryReturnStatus(kt, task, hosts)
}

// UpdateOrderInfo update order info
func (r *recycler) UpdateOrderInfo(kt *kit.Kit, orderId, handler string, success, failed, pending uint,
	msg string) error {

	return r.dispatcher.GetReturn().UpdateOrderInfo(kt.Ctx, orderId, handler, success, failed, pending,
		msg)
}

// StartRecycleOrderByRecycleType starts resource recycle order by recycle type
func (r *recycler) StartRecycleOrderByRecycleType(kt *kit.Kit, param *types.StartRecycleOrderByRecycleTypeReq) error {
	subOrderIDs := make([]string, 0)
	subOrderIDTypeMap := make(map[string]table.RecycleType, 0)
	configs := make([]*types.RecycleOrderReturnForecastConfig, 0)
	for _, item := range param.SubOrderIDTypes {
		subOrderIDs = append(subOrderIDs, item.SuborderID)
		subOrderIDTypeMap[item.SuborderID] = item.RecycleType
		configs = append(configs, &types.RecycleOrderReturnForecastConfig{
			SuborderID:         item.SuborderID,
			ReturnForecast:     item.ReturnForecast,
			ReturnForecastTime: item.ReturnForecastTime,
		})
	}
	subOrderIDs = slice.Unique(subOrderIDs)

	if len(subOrderIDs) == 0 {
		return fmt.Errorf("suborder_ids can not be empty")
	}

	filter := map[string]interface{}{}
	filter["suborder_id"] = mapstr.MapStr{
		pkg.BKDBIN: subOrderIDs,
	}
	page := metadata.BasePage{}
	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order by recycle type, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order by recycle type count %d == 0, rid: %s", cnt, kt.Rid)
		return fmt.Errorf("found no recycle order by recycle type to start")
	}

	// 设置滚服项目的回收Host表的回收记录及退还方式等
	if err = r.setRollingServerRecycleHost(kt, insts, subOrderIDTypeMap, enumor.CrpReturnedWay); err != nil {
		logs.Errorf("failed to get recycle order by recycle type, insert rolling server returned record failed, "+
			"err: %v, insts: %+v, rid: %s", err, cvt.PtrToSlice(insts), kt.Rid)
		return err
	}

	err = r.updateOrderReturnForecastConfig(kt, insts, configs)
	if err != nil {
		logs.Errorf("failed to update recycle order return forecast config, err: %v, param: %+v, rid: %s",
			err, cvt.PtrToVal(param), kt.Rid)
		return err
	}
	r.setOrderNextStatus(kt, insts)

	return nil
}

// setRollingServerRecycleHost 设置滚服项目的回收Host表的回收记录及退还方式等
func (r *recycler) setRollingServerRecycleHost(kt *kit.Kit, orders []*table.RecycleOrder,
	subOrderIDTypeMap map[string]table.RecycleType, returnedWay enumor.ReturnedWay) error {

	// 把符合条件的主机回收子订单里面的“回收类型”置为传入的回收类型
	for _, item := range orders {
		if recycleType, ok := subOrderIDTypeMap[item.SuborderID]; ok {
			item.RecycleType = recycleType
		}

		// 查询该回收子订单对应的回收Host列表
		hostReq := &types.GetRecycleHostReq{
			SuborderID: []string{item.SuborderID},
			BizID:      []int64{item.BizID},
			Page:       metadata.BasePage{Limit: pkg.BKNoLimit, Start: 0},
		}
		hostList, err := r.GetRecycleHost(kt, hostReq)
		if err != nil {
			logs.Errorf("failed to get recycle order host list, err: %v, order: %+v, rid: %s",
				err, cvt.PtrToVal(item), kt.Rid)
			return err
		}

		for _, hostItem := range hostList.Info {
			// 管理员指定回收类型的话，回收方式为CRP回收
			hostItem.ReturnedWay = enumor.CrpReturnedWay
			hostItem.IsMatchRollServer = true
		}
		// 插入需要退还的主机匹配记录
		if err = r.rsLogic.InsertReturnedHostMatched(kt, item.BizID, item.OrderID, item.SuborderID,
			hostList.Info, enumor.NormalStatus); err != nil {
			logs.Errorf("create returned host matched for save recycle order failed, err: %v, "+
				"subOrderID: %s, bkBizIDHost: %v, rid: %s",
				err, item.SuborderID, hostList.Info, kt.Rid)
			return fmt.Errorf("failed to create returned host matched for save recycle order err: %v", err)
		}
		logs.Infof("create returned host matched for save recycle order success, subOrderID: %s, "+
			"bkBizIDHost: %v, rid: %s", item.SuborderID, hostList.Info, kt.Rid)
	}

	now := time.Now()
	for _, order := range orders {
		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"recycle_type": order.RecycleType,
			"returned_way": returnedWay,
			"update_at":    now,
		}

		// do not dispatch order to start if set next status failed
		if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), filter, update); err != nil {
			logs.Errorf("failed to update recycle host status, subOrderID: %s, err: %v, returnedWay: %s, order: %+v, "+
				"rid: %s", order.SuborderID, err, returnedWay, cvt.PtrToVal(order), kt.Rid)
			return err
		}
	}
	return nil
}

// setRollingServerRecycleHost 设置滚服项目的回收Host表的回收记录及退还方式等
func (r *recycler) updateOrderReturnForecastConfig(kt *kit.Kit, orders []*table.RecycleOrder,
	configs []*types.RecycleOrderReturnForecastConfig) error {

	configMap := cvt.SliceToMap(configs, func(item *types.RecycleOrderReturnForecastConfig) (
		string, *types.RecycleOrderReturnForecastConfig) {
		return item.SuborderID, item
	})
	// 把符合条件的主机回收子订单里面的预测配置更新
	for _, item := range orders {
		config, ok := configMap[item.SuborderID]
		if !ok {
			continue
		}
		if item.RecycleType != table.RecycleTypeRegular {
			logs.Warnf("order(%s) is not a regular order, can not return forecast, rid: %s",
				item.SuborderID, kt.Rid)
			continue
		}
		item.ReturnForecastTime = config.ReturnForecastTime
		item.ReturnForecast = config.ReturnForecast
	}

	now := time.Now()
	for _, order := range orders {
		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"return_forecast_time": order.ReturnForecastTime,
			"return_forecast":      order.ReturnForecast,
			"update_at":            now,
		}

		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Errorf("failed to update recycle order, subOrderID: %s, err: %v, order: %+v, "+
				"rid: %s", order.SuborderID, err, cvt.PtrToVal(order), kt.Rid)
			return err
		}
	}
	return nil
}
