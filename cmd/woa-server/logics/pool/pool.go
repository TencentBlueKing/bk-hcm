/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pool

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/pool/dao"
	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/cmd/woa-server/logics/pool/classifier"
	"hcm/cmd/woa-server/logics/pool/launcher"
	"hcm/cmd/woa-server/logics/pool/recaller"
	"hcm/cmd/woa-server/logics/pool/recycler"
	types "hcm/cmd/woa-server/types/pool"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
)

// PoolIf provides management interface for operations of resource pool
type PoolIf interface {
	// CreateLaunchTask create resource launch task
	CreateLaunchTask(kt *kit.Kit, param *types.LaunchReq) (mapstr.MapStr, error)
	// CreateRecallTask create resource recall task
	CreateRecallTask(kt *kit.Kit, param *types.RecallReq) (mapstr.MapStr, error)
	// GetLaunchTask gets pool launch task
	GetLaunchTask(kt *kit.Kit, param *types.GetLaunchTaskReq) (*types.GetLaunchTaskRst, error)
	// GetRecallTask gets pool recall task
	GetRecallTask(kt *kit.Kit, param *types.GetRecallTaskReq) (*types.GetRecallTaskRst, error)
	// GetLaunchHost gets pool launch host
	GetLaunchHost(kt *kit.Kit, param *types.GetLaunchHostReq) (*types.GetLaunchHostRst, error)
	// GetRecallHost gets pool recall host
	GetRecallHost(kt *kit.Kit, param *types.GetRecallHostReq) (*types.GetRecallHostRst, error)
	// GetPoolHost gets resource pool host
	GetPoolHost(kt *kit.Kit, param *types.GetPoolHostReq) (*types.GetPoolHostRst, error)
	// DrawHost draw hosts from resource pool
	DrawHost(kt *kit.Kit, param *types.DrawHostReq) error
	// ReturnHost return hosts to resource pool
	ReturnHost(kt *kit.Kit, param *types.ReturnHostReq) error
	// CreateRecallOrder create resource recall task
	CreateRecallOrder(kt *kit.Kit, param *types.CreateRecallOrderReq) (mapstr.MapStr, error)
	// GetRecallOrder gets pool recall order
	GetRecallOrder(kt *kit.Kit, param *types.GetRecallOrderReq) (*types.GetRecallOrderRst, error)
	// GetRecalledInstance gets pool recalled instance
	GetRecalledInstance(kt *kit.Kit, param *types.GetRecalledInstReq) (*types.GetRecalledInstRst, error)
	// GetLaunchMatchDevice get resource launch match devices
	GetLaunchMatchDevice(kt *kit.Kit, param *types.GetLaunchMatchDeviceReq) (*types.GetLaunchMatchDeviceRst, error)
	// GetRecallMatchDevice get resource recall match devices
	GetRecallMatchDevice(kt *kit.Kit, param *types.GetRecallMatchDeviceReq) (*types.GetRecallMatchDeviceRst, error)
	// GetRecallDetail gets resource pool recall task execution detail info
	GetRecallDetail(kt *kit.Kit, param *types.GetRecallDetailReq) (*types.GetRecallDetailRst, error)
	// ResumeRecycleTask resumes resource recycle task
	ResumeRecycleTask(kt *kit.Kit, param *types.ResumeRecycleTaskReq) error
	// CreateGradeCfg creates pool grade config
	CreateGradeCfg(kt *kit.Kit, param *table.GradeCfg) (mapstr.MapStr, error)
	// GetGradeCfg get pool grade config
	GetGradeCfg(kt *kit.Kit) (*types.GetGradeCfgRst, error)
	// GetDeviceType get pool supported device type list
	GetDeviceType(kt *kit.Kit) (*types.GetDeviceTypeRst, error)
}

// NewPoolIf creates a pool interface
func NewPoolIf(ctx context.Context, cliConf cc.ClientConfig, thirdCli *thirdparty.Client, cmdbCli cmdb.Client) PoolIf {
	recycle := recycler.New(ctx, cliConf, thirdCli, cmdbCli)
	recall := recaller.New(ctx, cmdbCli)
	recall.SetRecycler(recycle)

	return &pool{
		cmdbCli:  cmdbCli,
		launcher: launcher.New(ctx, cmdbCli),
		recaller: recall,
		recycler: recycle,
	}
}

type pool struct {
	cmdbCli  cmdb.Client
	launcher *launcher.Launcher
	recaller *recaller.Recaller
	recycler *recycler.Recycler
}

// CreateLaunchTask create resource launch task
func (p *pool) CreateLaunchTask(kt *kit.Kit, param *types.LaunchReq) (mapstr.MapStr, error) {
	// 1. get hosts info
	hosts, err := p.getHostDetailInfo(nil, nil, param.HostIDs)
	if err != nil {
		logs.Errorf("failed to get host info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(hosts) == 0 {
		logs.Errorf("get no valid host to create launch task")
		return nil, errors.New("get no valid host to create launch task")
	}

	// 2. verification
	// TODO

	// 3. init and save launch task
	task, err := p.initAndSaveLaunchTask(kt, hosts)
	if err != nil {
		logs.Errorf("failed to init launch task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	p.startLaunchTask(task)

	rst := mapstr.MapStr{
		"id": task.ID,
	}

	return rst, nil
}

// getHostBaseInfo get host detail info for recycle
func (p *pool) getHostDetailInfo(ips, assetIds []string, hostIds []int64) ([]*table.PoolHost, error) {
	// 1. get host base info
	hostBase, err := p.getHostBaseInfo(ips, assetIds, hostIds)
	if err != nil {
		logs.Errorf("failed to get host detail info, for list host err: %v", err)
		return nil, err
	}

	if len(hostBase) == 0 {
		return make([]*table.PoolHost, 0), nil
	}

	gradeConfigs, err := p.getGradeCfg()
	if err != nil {
		logs.Errorf("failed to get device grade config, err: %v", err)
		return nil, err
	}

	// 3. fill host info
	hostDetails := make([]*table.PoolHost, 0)
	for _, host := range hostBase {
		gradeTag := ""
		if cfg, ok := gradeConfigs[host.SvrDeviceClass]; ok {
			gradeTag = cfg.GradeTag
		}
		hostDetail := &table.PoolHost{
			HostID: host.BkHostID,
			Labels: map[string]string{
				table.IPKey:           host.GetUniqIp(),
				table.AssetIDKey:      host.BkAssetID,
				table.ResourceTypeKey: string(classifier.GetResType(host.BkAssetID)),
				table.DeviceTypeKey:   host.SvrDeviceClass,
				table.RegionKey:       host.BkZoneName,
				table.ZoneKey:         host.SubZone,
				table.GradeTagKey:     gradeTag,
				table.CloudRegionKey:  host.BkCloudRegion,
				table.CloudZoneKey:    host.BkCloudZone,
			},
		}

		hostDetails = append(hostDetails, hostDetail)
	}

	return hostDetails, nil
}

// getHostBaseInfo get host base info in cc 3.0
func (p *pool) getHostBaseInfo(ips, assetIds []string, hostIds []int64) ([]*cmdb.Host, error) {
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
			"bk_cloud_region",
			"bk_cloud_zone",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	kt := core.NewBackendKit()
	resp, err := p.cmdbCli.ListHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	return converter.SliceToPtr(resp.Info), nil
}

func (p *pool) getGradeCfg() (map[string]*table.GradeCfg, error) {
	filter := map[string]interface{}{}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	insts, err := dao.Set().GradeCfg().FindManyGradeCfg(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get grade configs, err: %v", err)
		return nil, err
	}

	deviceTypeToGrade := make(map[string]*table.GradeCfg)
	for _, inst := range insts {
		deviceTypeToGrade[inst.DeviceType] = inst
	}

	return deviceTypeToGrade, nil
}

func (p *pool) initAndSaveLaunchTask(kt *kit.Kit, hosts []*table.PoolHost) (*table.LaunchTask, error) {
	id, err := dao.Set().LaunchTask().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create launch task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	now := time.Now()
	num := uint(len(hosts))
	task := &table.LaunchTask{
		ID:   id,
		User: kt.User,
		Status: &table.PoolTaskStatus{
			TotalNum:   num,
			SuccessNum: 0,
			PendingNum: num,
			Phase:      table.OpTaskPhaseInit,
			Message:    "",
		},
		CreateAt: now,
		UpdateAt: now,
	}

	// create and save launch op records
	if err := p.initAndSaveOpRecords(kt, task, hosts); err != nil {
		logs.Errorf("failed to create launch task for save op record err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("failed to create launch task for save op record err: %v", err)
	}

	if err := dao.Set().LaunchTask().CreateLaunchTask(kt.Ctx, task); err != nil {
		logs.Errorf("failed to create launch task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return task, nil
}

func (p *pool) initAndSaveOpRecords(kt *kit.Kit, task *table.LaunchTask, hosts []*table.PoolHost) error {
	now := time.Now()
	for _, host := range hosts {
		id, err := dao.Set().OpRecord().NextSequence(kt.Ctx)
		if err != nil {
			logs.Errorf("failed to create op record, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		record := &table.OpRecord{
			ID:       id,
			HostID:   host.HostID,
			Labels:   host.Labels,
			OpType:   table.OpTypeLaunch,
			TaskID:   task.ID,
			Phase:    table.OpTaskPhaseInit,
			Message:  "",
			Operator: kt.User,
			CreateAt: now,
			UpdateAt: now,
		}

		if err := dao.Set().OpRecord().CreateOpRecord(context.Background(), record); err != nil {
			logs.Errorf("failed to save op record, host id: %d, err: %v", host.HostID, err)
			return fmt.Errorf("failed to save op record, host id: %d, err: %v", host.HostID, err)
		}
	}

	return nil
}

func (p *pool) startLaunchTask(task *table.LaunchTask) {
	// add task to launcher dispatch queue
	p.launcher.Add(task.ID)
}

// HostBizRelReq
// getHostTopoInfo get host topo info in cc 3.0
func (p *pool) getHostTopoInfo(hostIds []int64) ([]*cmdb.HostTopoRelation, error) {
	req := &cmdb.HostModuleRelationParams{
		HostID: hostIds,
	}

	kt := core.NewBackendKit()
	resp, err := p.cmdbCli.FindHostBizRelations(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host topo info, err: %v", err)
		return nil, err
	}

	return converter.SliceToPtr(converter.PtrToVal(resp)), nil
}

// getBizInfo get business info in cc 3.0
func (p *pool) getBizInfo(bizIds []int64) ([]*cmdb.Biz, error) {
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

	kt := core.NewBackendKit()
	resp, err := p.cmdbCli.SearchBusiness(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc business info, err: %v", err)
		return nil, err
	}

	return converter.SliceToPtr(resp.Info), nil
}

// getModuleInfo get module info in cc 3.0
func (p *pool) getModuleInfo(kt *kit.Kit, bizId int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error) {
	req := &cmdb.SearchModuleParams{
		BizID: bizId,
		Condition: mapstr.MapStr{
			"bk_module_id": mapstr.MapStr{
				pkg.BKDBIN: moduleIds,
			},
		},
		Fields: []string{"bk_module_id", "bk_module_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}

	resp, err := p.cmdbCli.SearchModule(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc module info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Info, nil
}

// CreateRecallTask create resource recall task
func (p *pool) CreateRecallTask(kt *kit.Kit, param *types.RecallReq) (mapstr.MapStr, error) {
	// 1. get hosts info

	// 2. verification

	// 3. init and save recall task
	task, err := p.initAndSaveRecallTask(kt, param)
	if err != nil {
		logs.Errorf("failed to init and save recall task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	p.startRecallTask(task)

	rst := mapstr.MapStr{
		"id": task.ID,
	}

	return rst, nil
}

func (p *pool) initAndSaveRecallTask(kt *kit.Kit, param *types.RecallReq) (*table.RecallTask, error) {
	id, err := dao.Set().RecallTask().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create recall task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	filter := &mapstr.MapStr{
		"device_type": param.DeviceType,
	}
	gradeCfg, err := dao.Set().GradeCfg().GetGradeCfg(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool grade cfg by device type: %s, err: %v", param.DeviceType, err)
	}

	now := time.Now()
	num := param.Replicas
	spec := &table.RecallTaskSpec{
		Selector: []*table.Selector{
			{
				Key:      table.ResourceTypeKey,
				Operator: table.SelectOpEqual,
				Value:    gradeCfg.ResourceType,
			},
			{
				Key:      table.DeviceTypeKey,
				Operator: table.SelectOpEqual,
				Value:    param.DeviceType,
			},
			{
				Key:      table.GradeTagKey,
				Operator: table.SelectOpEqual,
				Value:    gradeCfg.GradeTag,
			},
		},
		Replicas: num,
	}
	if len(param.AssetIDs) > 0 {
		spec.Selector = append(spec.Selector, &table.Selector{
			Key:      table.AssetIDKey,
			Operator: table.SelectOpIn,
			Value:    param.AssetIDs,
		})
	}
	if param.BkCloudRegion != "" {
		spec.Selector = append(spec.Selector, &table.Selector{
			Key:      table.CloudRegionKey,
			Operator: table.SelectOpEqual,
			Value:    param.BkCloudRegion,
		})
	}
	if param.BkCloudZone != "" {
		spec.Selector = append(spec.Selector, &table.Selector{
			Key:      table.CloudZoneKey,
			Operator: table.SelectOpEqual,
			Value:    param.BkCloudZone,
		})
	}
	task := &table.RecallTask{
		ID:   id,
		User: kt.User,
		Spec: spec,
		Status: &table.PoolTaskStatus{
			Phase:      table.OpTaskPhaseInit,
			Message:    "",
			TotalNum:   num,
			SuccessNum: 0,
			PendingNum: num,
			FailedNum:  0,
		},
		CreateAt: now,
		UpdateAt: now,
	}

	if err = dao.Set().RecallTask().CreateRecallTask(kt.Ctx, task); err != nil {
		logs.Errorf("failed to create recall task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return task, nil
}

func (p *pool) startRecallTask(task *table.RecallTask) {
	// add task to recaller dispatch queue
	p.recaller.Add(task.ID)
}

// GetLaunchTask gets pool launch task
func (p *pool) GetLaunchTask(kt *kit.Kit, param *types.GetLaunchTaskReq) (*types.GetLaunchTaskRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get launch task, for get filter err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetLaunchTaskRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().LaunchTask().CountLaunchTask(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get launch task count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.LaunchTask, 0)
		return rst, nil
	}

	insts, err := dao.Set().LaunchTask().FindManyLaunchTask(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get launch task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetRecallTask gets pool recall task
func (p *pool) GetRecallTask(kt *kit.Kit, param *types.GetRecallTaskReq) (*types.GetRecallTaskRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recall task, for get filter err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetRecallTaskRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().RecallTask().CountRecallTask(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recall task count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.RecallTask, 0)
		return rst, nil
	}

	insts, err := dao.Set().RecallTask().FindManyRecallTask(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recall task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetLaunchHost gets pool launch host
func (p *pool) GetLaunchHost(kt *kit.Kit, param *types.GetLaunchHostReq) (*types.GetLaunchHostRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get filter, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetLaunchHostRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().OpRecord().CountOpRecord(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get launch host count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.OpRecord, 0)
		return rst, nil
	}

	insts, err := dao.Set().OpRecord().FindManyOpRecord(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get launch host, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetRecallHost gets pool recall host
func (p *pool) GetRecallHost(kt *kit.Kit, param *types.GetRecallHostReq) (*types.GetRecallHostRst, error) {
	filter := map[string]interface{}{
		"op_type": table.OpTypeRecall,
		"task_id": param.ID,
	}

	rst := &types.GetRecallHostRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().OpRecord().CountOpRecord(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get launch host count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.OpRecord, 0)
		return rst, nil
	}

	insts, err := dao.Set().OpRecord().FindManyOpRecord(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get launch host, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetPoolHost gets resource pool host
func (p *pool) GetPoolHost(kt *kit.Kit, param *types.GetPoolHostReq) (*types.GetPoolHostRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get pool host, for get filter err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetPoolHostRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().PoolHost().CountPoolHost(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get pool host count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.PoolHost, 0)
		return rst, nil
	}

	insts, err := dao.Set().PoolHost().FindManyPoolHost(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get pool host, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// DrawHost 从资源池借出主机到指定业务下：
func (p *pool) DrawHost(kt *kit.Kit, param *types.DrawHostReq) error {
	// try
	// lock host
	filter := map[string]interface{}{
		"bk_host_id": map[string]interface{}{
			pkg.BKDBIN: param.HostIDs,
		},
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}
	hosts, err := dao.Set().PoolHost().FindManyPoolHost(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get pool host, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	hostIDToHost := make(map[int64]*table.PoolHost)
	for _, host := range hosts {
		hostIDToHost[host.HostID] = host
	}

	// confirm
	for _, hostID := range param.HostIDs {
		host, ok := hostIDToHost[hostID]
		if !ok {
			logs.Errorf("invalid bk_host_id %d, for not exist in resource pool", hostID)
			return fmt.Errorf("invalid bk_host_id %d, not exist in resource pool", hostID)
		}

		if host.Status.Phase != table.PoolHostPhaseIdle {
			logs.Errorf("invalid bk_host_id %d, for phase %s != %s", hostID, host.Status.Phase, table.PoolHostPhaseIdle)
			return fmt.Errorf("invalid bk_host_id %d, for phase %s != %s", hostID, host.Status.Phase,
				table.PoolHostPhaseIdle)
		}
	}

	// transfer host to destination business
	for _, hostID := range param.HostIDs {
		// transfer hosts from 资源运营服务-CR资源池 to destination business
		if err := p.transferHost(kt, hostID, types.BizIDPool, param.ToBizID, 0); err != nil {
			logs.Errorf("failed to transfer host %d, err: %v, rid: %s", hostID, err, kt.Rid)
			return err
		}

		// update pool host status
		if err := p.updateHostStatus(hostID, table.PoolHostPhaseInUse); err != nil {
			logs.Errorf("failed to update host %d status, err: %v, rid: %s", hostID, err, kt.Rid)
			return err
		}
	}

	// cancel
	// unlock host

	return nil
}

// ReturnHost return hosts to resource pool
func (p *pool) ReturnHost(kt *kit.Kit, param *types.ReturnHostReq) error {
	task, err := p.getRecallTaskByID(param.RecallID)
	if err != nil {
		logs.Errorf("failed to get recall task by id %d, err: %v", param.RecallID, err)
		return err
	}

	// 1. check task status
	if task.Status.Phase == table.OpTaskPhaseSuccess {
		logs.Errorf("need not return host for recall task %d, for its phase is %s", param.RecallID,
			table.OpTaskPhaseSuccess)
		return fmt.Errorf("need not return host for recall task %d, for its phase is %s", param.RecallID,
			table.OpTaskPhaseSuccess)
	}

	// try
	// lock host

	// 2. get host info
	hosts, err := p.getPoolHostInfo(param.HostIDs)
	if err != nil {
		logs.Errorf("failed to get pool host info: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 3. check host status
	if err := p.checkHostStatus(hosts, param.HostIDs); err != nil {
		logs.Errorf("failed to check host status: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, hostID := range param.HostIDs {
		// transfer hosts to 资源运营服务-CR资源下架中
		if err := p.transferHost(kt, hostID, param.FromBizID, types.BizIDPool,
			types.ModuleIDPoolRecalling); err != nil {
			logs.Errorf("failed to transfer host %d, err: %v, rid: %s", hostID, err, kt.Rid)
			return err
		}

		// update pool host status
		if err := p.updateHostStatus(hostID, table.PoolHostPhaseForRecall); err != nil {
			logs.Errorf("failed to update host %d status, err: %v, rid: %s", hostID, err, kt.Rid)
			return err
		}
	}

	// update op record
	if err := p.createRecallOpRecords(kt, task, hosts); err != nil {
		logs.Errorf("failed to create recall op record, err: %v", err)
		return err
	}

	// update recall detail
	if err := p.createRecallDetail(kt, task, hosts); err != nil {
		logs.Errorf("failed to create recall detail, err: %v", err)
		return err
	}

	// update task status
	task.Status.Phase = table.OpTaskPhaseRunning
	task.Status.SuccessNum = task.Status.SuccessNum + uint(len(hosts))
	task.Status.PendingNum = task.Status.TotalNum - task.Status.SuccessNum
	task.Status.FailedNum = 0
	if task.Status.SuccessNum >= task.Status.TotalNum {
		task.Status.Phase = table.OpTaskPhaseSuccess
	}

	if err := p.updateRecallTaskStatus(task); err != nil {
		logs.Errorf("failed to update recall task status, id: %d, err: %v", task.ID, err)
		return err
	}

	return nil
}

func (p *pool) getPoolHostInfo(hostIDs []int64) ([]*table.PoolHost, error) {
	filter := map[string]interface{}{
		"bk_host_id": map[string]interface{}{
			pkg.BKDBIN: hostIDs,
		},
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	hosts, err := dao.Set().PoolHost().FindManyPoolHost(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get pool host, err: %v", err)
		return nil, err
	}

	return hosts, nil
}

// checkHostStatus check whether hosts can be returned or not
func (p *pool) checkHostStatus(hosts []*table.PoolHost, hostIDs []int64) error {
	hostIDToHost := make(map[int64]*table.PoolHost)
	for _, host := range hosts {
		hostIDToHost[host.HostID] = host
	}

	for _, hostID := range hostIDs {
		host, ok := hostIDToHost[hostID]
		if !ok {
			logs.Errorf("invalid bk_host_id %d, for not exist in resource pool", hostID)
			return fmt.Errorf("invalid bk_host_id %d, not exist in resource pool", hostID)
		}

		if host.Status.Phase != table.PoolHostPhaseInUse {
			logs.Errorf("invalid bk_host_id %d, for phase %s != %s", hostID, host.Status.Phase,
				table.PoolHostPhaseInUse)
			return fmt.Errorf("invalid bk_host_id %d, for phase %s != %s", hostID, host.Status.Phase,
				table.PoolHostPhaseInUse)
		}
	}

	return nil
}

// transferHost transfer host to target business in cc 3.0
func (p *pool) transferHost(kt *kit.Kit, hostID, fromBizID, toBizID, toModuleId int64) error {
	transferReq := &cmdb.TransferHostReq{
		From: cmdb.TransferHostSrcInfo{
			FromBizID: fromBizID,
			HostIDs:   []int64{hostID},
		},
		To: cmdb.TransferHostDstInfo{
			ToBizID: toBizID,
		},
	}

	// if destination module id is 0, transfer host to idle module of business
	// otherwise, transfer host to input module
	if toModuleId > 0 {
		transferReq.To.ToModuleID = toModuleId
	}

	err := p.cmdbCli.TransferHost(kt, transferReq)
	if err != nil {
		return err
	}

	return nil
}

func (p *pool) getRecallTaskByID(id uint64) (*table.RecallTask, error) {
	filter := &mapstr.MapStr{
		"id": id,
	}
	task, err := dao.Set().RecallTask().GetRecallTask(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (p *pool) updateHostStatus(hostID int64, phase table.PoolHostPhase) error {
	filter := map[string]interface{}{
		"bk_host_id": hostID,
	}

	now := time.Now()
	update := map[string]interface{}{
		"status.phase": phase,
		"update_at":    now,
	}

	if err := dao.Set().PoolHost().UpdatePoolHost(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}

func (p *pool) createRecallOpRecords(kt *kit.Kit, task *table.RecallTask, hosts []*table.PoolHost) error {
	now := time.Now()
	for _, host := range hosts {
		id, err := dao.Set().OpRecord().NextSequence(kt.Ctx)
		if err != nil {
			logs.Errorf("failed to create op record, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		record := &table.OpRecord{
			ID:       id,
			HostID:   host.HostID,
			Labels:   host.Labels,
			OpType:   table.OpTypeRecall,
			TaskID:   task.ID,
			Phase:    table.OpTaskPhaseSuccess,
			Message:  "",
			Operator: kt.User,
			CreateAt: now,
			UpdateAt: now,
		}

		if err := dao.Set().OpRecord().CreateOpRecord(context.Background(), record); err != nil {
			logs.Errorf("failed to save op record, host id: %d, err: %v", host.HostID, err)
			return fmt.Errorf("failed to save op record, host id: %d, err: %v", host.HostID, err)
		}
	}

	return nil
}

func (p *pool) createRecallDetail(kt *kit.Kit, task *table.RecallTask, hosts []*table.PoolHost) error {
	now := time.Now()
	for _, host := range hosts {
		detail := &table.RecallDetail{
			ID:            fmt.Sprintf("%d-%d", task.ID, host.HostID),
			RecallID:      task.ID,
			HostID:        host.HostID,
			Labels:        host.Labels,
			Status:        table.RecallStatusReturned,
			Message:       "",
			ReinstallID:   "",
			ReinstallLink: "",
			ConfCheckID:   "",
			ConfCheckLink: "",
			Operator:      kt.User,
			CreateAt:      now,
			UpdateAt:      now,
		}

		if err := dao.Set().RecallDetail().CreateRecallDetail(context.Background(), detail); err != nil {
			logs.Errorf("failed to save recall detail, host id: %d, err: %v", host.HostID, err)
			return fmt.Errorf("failed to save recall detail, host id: %d, err: %v", host.HostID, err)
		}

		// add recall task to dispatch queue
		p.recycler.Add(detail.ID)
	}

	return nil
}

// updateRecallTaskStatus update recall task status
func (p *pool) updateRecallTaskStatus(task *table.RecallTask) error {
	filter := map[string]interface{}{
		"id": task.ID,
	}

	doc := map[string]interface{}{
		"status.phase":       task.Status.Phase,
		"status.success_num": task.Status.SuccessNum,
		"status.pending_num": task.Status.PendingNum,
		"status.failed_num":  task.Status.FailedNum,
		"update_at":          time.Now(),
	}

	if err := dao.Set().RecallTask().UpdateRecallTask(context.Background(), filter, doc); err != nil {
		return err
	}

	return nil
}

// CreateRecallOrder create resource recall task
func (p *pool) CreateRecallOrder(kt *kit.Kit, param *types.CreateRecallOrderReq) (mapstr.MapStr, error) {
	// 1. get hosts info

	// 2. verification

	// 3. init and save recall task
	taskParam := &types.RecallReq{
		DeviceType:    param.DeviceType,
		AssetIDs:      param.AssetIDs,
		Replicas:      param.Replicas,
		BkCloudRegion: param.BkCloudRegion,
		BkCloudZone:   param.BkCloudZone,
	}
	task, err := p.initAndSaveRecallTask(kt, taskParam)
	if err != nil {
		logs.Errorf("failed to init and save recall task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// init and save recall order
	if err := p.initAndSaveRecallOrder(kt, task, param); err != nil {
		logs.Errorf("failed to init and save recall order, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	p.startRecallTask(task)

	rst := mapstr.MapStr{
		"id": task.ID,
	}

	return rst, nil
}

func (p *pool) initAndSaveRecallOrder(kt *kit.Kit, task *table.RecallTask, param *types.CreateRecallOrderReq) error {

	order := &table.RecallOrder{
		ID:   task.ID,
		User: task.User,
		Spec: task.Spec,
		RecyclePolicy: &table.RecyclePolicy{
			ImageID: param.ImageID,
			OsType:  param.OsType,
		},
		Status:   task.Status,
		CreateAt: task.CreateAt,
		UpdateAt: task.UpdateAt,
	}

	if err := dao.Set().RecallOrder().CreateRecallOrder(kt.Ctx, order); err != nil {
		logs.Errorf("failed to create recall order, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// GetRecallOrder gets pool recall order
func (p *pool) GetRecallOrder(kt *kit.Kit, param *types.GetRecallOrderReq) (*types.GetRecallOrderRst, error) {
	filter := &mapstr.MapStr{
		"id": param.ID,
	}

	inst, err := dao.Set().RecallOrder().GetRecallOrder(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get recall task, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	detailFilter := map[string]interface{}{
		"recall_id": inst.ID,
		"status":    table.RecallStatusDone,
	}
	succCnt, err := dao.Set().RecallDetail().CountRecallDetail(kt.Ctx, detailFilter)
	if err != nil {
		logs.Errorf("failed to get recalled host count, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	inst.Status.SuccessNum = uint(succCnt)
	inst.Status.PendingNum = inst.Status.TotalNum - inst.Status.SuccessNum
	inst.Status.Phase = table.OpTaskPhaseRunning
	if inst.Status.SuccessNum >= inst.Status.TotalNum {
		inst.Status.Phase = table.OpTaskPhaseSuccess
	}

	rst := &types.GetRecallOrderRst{
		Count: 1,
		Info:  []*table.RecallOrder{inst},
	}

	return rst, nil
}

// GetRecalledInstance gets pool recalled instance
func (p *pool) GetRecalledInstance(kt *kit.Kit, param *types.GetRecalledInstReq) (*types.GetRecalledInstRst, error) {
	filter := map[string]interface{}{
		"recall_id": param.ID,
		"status":    table.RecallStatusDone,
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	insts, err := dao.Set().RecallDetail().FindManyRecallDetail(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recalled instance, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetRecalledInstRst{
		Count: int64(len(insts)),
		Info:  insts,
	}

	return rst, nil
}

// GetLaunchMatchDevice get resource launch match devices
func (p *pool) GetLaunchMatchDevice(kt *kit.Kit, param *types.GetLaunchMatchDeviceReq) (
	*types.GetLaunchMatchDeviceRst, error) {

	req, err := p.createListMatchDeviceReq(param)
	if err != nil {
		logs.Errorf("failed to create get cc host req, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, err := p.cmdbCli.ListBizHost(kt, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetLaunchMatchDeviceRst{
		Count: 0,
		Info:  make([]*types.MatchDevice, 0),
	}

	for _, host := range resp.Info {
		rackId, err := strconv.Atoi(host.RackId)
		if err != nil {
			logs.Warnf("failed to convert host %d rack_id %s to int", host.BkHostID, host.RackId)
			rackId = 0
		}
		device := &types.MatchDevice{
			BkHostId:     host.BkHostID,
			AssetId:      host.BkAssetID,
			Ip:           host.GetUniqIp(),
			OuterIp:      host.BkHostOuterIP,
			Isp:          host.BkIpOerName,
			DeviceType:   host.SvrDeviceClass,
			OsType:       host.BkOSName,
			Region:       host.BkZoneName,
			Zone:         host.SubZone,
			Module:       host.ModuleName,
			Equipment:    int64(rackId),
			IdcUnit:      host.IdcUnitName,
			IdcLogicArea: host.LogicDomain,
			RaidType:     host.RaidName,
			InputTime:    host.SvrInputTime,
		}

		rst.Info = append(rst.Info, device)
	}
	rst.Count = int64(len(rst.Info))

	return rst, nil
}

func (p *pool) createListMatchDeviceReq(param *types.GetLaunchMatchDeviceReq) (*cmdb.ListBizHostParams, error) {
	req := &cmdb.ListBizHostParams{
		BizID:       types.BizIDMatch,
		BkModuleIDs: []int64{types.ModuleIDPoolMatch},
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
			// 可用区
			"sub_zone",
			"module_name",
			// 机架号，string类型
			"rack_id",
			"idc_unit_name",
			// 逻辑区域
			"logic_domain",
			"raid_name",
			"svr_input_time",
		},
		Page: &cmdb.BasePage{
			Start: 0,
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	rule, err := p.createListMatchDeviceReqRule(param)
	if err != nil {
		return nil, err
	}

	if len(rule.Rules) > 0 {
		req.HostPropertyFilter = &cmdb.QueryFilter{
			Rule: rule,
		}
	}

	return req, nil
}

func (p *pool) createListMatchDeviceReqRule(param *types.GetLaunchMatchDeviceReq) (*querybuilder.CombinedRule, error) {
	rule := &querybuilder.CombinedRule{
		Condition: querybuilder.ConditionAnd,
		Rules:     make([]querybuilder.Rule, 0),
	}

	if len(param.Ips) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_host_innerip",
			Operator: querybuilder.OperatorIn,
			Value:    param.Ips,
		})
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_cloud_id",
			Operator: querybuilder.OperatorEqual,
			Value:    0,
		})
	}

	if len(param.AssetIDs) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_asset_id",
			Operator: querybuilder.OperatorIn,
			Value:    param.AssetIDs,
		})
	}

	if param.Spec == nil {
		return rule, nil
	}

	if err := p.setListMatchDeviceZoneFilter(param, rule); err != nil {
		return nil, err
	}

	if len(param.Spec.DeviceType) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "svr_device_class",
			Operator: querybuilder.OperatorIn,
			Value:    param.Spec.DeviceType,
		})
	}

	if len(param.Spec.OsType) != 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_os_name",
			Operator: querybuilder.OperatorIn,
			Value:    param.Spec.OsType,
		})
	}

	if len(param.Spec.RaidType) != 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "raid_name",
			Operator: querybuilder.OperatorIn,
			Value:    param.Spec.RaidType,
		})
	}

	if len(param.Spec.Isp) != 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_ip_oper_name",
			Operator: querybuilder.OperatorIn,
			Value:    param.Spec.Isp,
		})
	}

	return rule, nil
}

// setListMatchDeviceZoneFilter ...
// Deprecated: 上下架功能已弃用
func (p *pool) setListMatchDeviceZoneFilter(param *types.GetLaunchMatchDeviceReq,
	rule *querybuilder.CombinedRule) error {

	if param == nil || param.Spec == nil {
		return errors.New("invalid param or param.spec is nil")
	}

	if param.ResourceType != types.ResourceTypeCvm {
		if len(param.Spec.Region) != 0 {
			rule.Rules = append(rule.Rules, querybuilder.AtomRule{
				Field:    "bk_zone_name",
				Operator: querybuilder.OperatorIn,
				Value:    param.Spec.Region,
			})
		}
		if len(param.Spec.Zone) != 0 {
			rule.Rules = append(rule.Rules, querybuilder.AtomRule{
				Field:    "sub_zone",
				Operator: querybuilder.OperatorIn,
				Value:    param.Spec.Zone,
			})
		}
		return nil
	}

	return nil
}

// GetRecallMatchDevice get resource recall match devices
func (p *pool) GetRecallMatchDevice(kt *kit.Kit, param *types.GetRecallMatchDeviceReq) (
	*types.GetRecallMatchDeviceRst, error) {

	filter, err := p.getRecallMatchDeviceFilter(param)
	if err != nil {
		logs.Errorf("failed to get resource recall match device filter, err: %v", err)
		return nil, err
	}

	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	insts, err := dao.Set().PoolHost().FindManyPoolHost(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get pool host, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	recallItems := make(map[string]*types.RecallMatchDevice)
	for _, inst := range insts {
		key := strings.Join([]string{inst.Labels[table.DeviceTypeKey], inst.Labels[table.CloudRegionKey],
			inst.Labels[table.CloudZoneKey]}, ".")
		if _, ok := recallItems[key]; !ok {
			recallItems[key] = &types.RecallMatchDevice{
				DeviceType:    inst.Labels[table.DeviceTypeKey],
				Amount:        0,
				BkCloudRegion: inst.Labels[table.CloudRegionKey],
				BkCloudZone:   inst.Labels[table.CloudZoneKey],
			}
		}
		recallItems[key].Amount++
	}

	rst := new(types.GetRecallMatchDeviceRst)

	for _, val := range recallItems {
		rst.Info = append(rst.Info, val)
	}

	rst.Count = int64(len(rst.Info))

	return rst, nil
}

// getRecallMatchDeviceFilter get resource recall match devices filter
func (p *pool) getRecallMatchDeviceFilter(param *types.GetRecallMatchDeviceReq) (map[string]interface{}, error) {
	filter := map[string]interface{}{}

	if param.Spec != nil {
		if len(param.Spec.DeviceType) > 0 {
			filter["labels.device_type"] = map[string]interface{}{
				pkg.BKDBIN: param.Spec.DeviceType,
			}
		}
		if len(param.Spec.BkCloudRegions) != 0 {
			filter["labels.bk_cloud_region"] = map[string]interface{}{
				pkg.BKDBIN: param.Spec.BkCloudRegions,
			}
		}
		if len(param.Spec.BkCloudZones) != 0 {
			filter["labels.bk_cloud_zone"] = map[string]interface{}{
				pkg.BKDBIN: param.Spec.BkCloudZones,
			}
		}
	}

	// only return idle or in use device
	filter["status.phase"] = map[string]interface{}{
		pkg.BKDBIN: []table.PoolHostPhase{table.PoolHostPhaseIdle, table.PoolHostPhaseInUse},
	}

	return filter, nil
}

// GetRecallDetail gets resource pool recall task execution detail info
func (p *pool) GetRecallDetail(kt *kit.Kit, param *types.GetRecallDetailReq) (*types.GetRecallDetailRst, error) {
	filter := map[string]interface{}{
		"recall_id": param.ID,
	}

	rst := &types.GetRecallDetailRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().RecallDetail().CountRecallDetail(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recall detail count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.RecallDetail, 0)
		return rst, nil
	}

	insts, err := dao.Set().RecallDetail().FindManyRecallDetail(kt.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recall detail, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// ResumeRecycleTask resumes resource recycle task
func (p *pool) ResumeRecycleTask(kt *kit.Kit, param *types.ResumeRecycleTaskReq) error {
	filter := map[string]interface{}{
		"id": mapstr.MapStr{
			pkg.BKDBIN: param.ID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecallDetail().FindManyRecallDetail(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle task, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("found no recycle task to resume, rid: %s", cnt, kt.Rid)
		return fmt.Errorf("found no recycle task to resume")
	}

	// add order to dispatch queue
	for _, order := range insts {
		p.recycler.Add(order.ID)
	}

	return nil
}

// CreateGradeCfg create pool grade config
func (p *pool) CreateGradeCfg(kt *kit.Kit, param *table.GradeCfg) (mapstr.MapStr, error) {
	id, err := dao.Set().GradeCfg().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create plan config, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	now := time.Now()
	param.ID = id
	param.CreateAt = now
	param.UpdateAt = now

	if err := dao.Set().GradeCfg().CreateGradeCfg(kt.Ctx, param); err != nil {
		logs.Errorf("failed to create grade config, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := mapstr.MapStr{
		"id": id,
	}

	return rst, nil
}

// GetGradeCfg get pool grade config
func (p *pool) GetGradeCfg(kt *kit.Kit) (*types.GetGradeCfgRst, error) {
	filter := map[string]interface{}{}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	insts, err := dao.Set().GradeCfg().FindManyGradeCfg(kt.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get pool grade config, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetGradeCfgRst{
		Info: insts,
	}

	return rst, nil
}

// GetDeviceType get pool supported device type list
func (p *pool) GetDeviceType(kt *kit.Kit) (*types.GetDeviceTypeRst, error) {
	filter := map[string]interface{}{}
	insts, err := dao.Set().GradeCfg().Distinct(kt.Ctx, "device_type", filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetDeviceTypeRst{
		Info: insts,
	}

	return rst, nil
}
