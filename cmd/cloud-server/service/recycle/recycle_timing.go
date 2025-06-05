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

package recycle

import (
	"time"

	"hcm/cmd/cloud-server/logics"
	"hcm/cmd/cloud-server/logics/recycle"
	"hcm/pkg/api/core"
	recyclerecord "hcm/pkg/api/core/recycle-record"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"
)

type recycle struct {
	client *client.ClientSet
	logics *logics.Logics
	state  serviced.State
}

// RecycleTiming timing recycle all resource.
func RecycleTiming(c *client.ClientSet, state serviced.State, conf cc.Recycle, cmdbClient cmdb.Client) {
	r := &recycle{
		client: c,
		state:  state,
		logics: logics.NewLogics(c, cmdbClient, nil),
	}

	go r.recycleTiming(enumor.DiskCloudResType, r.recycleDiskWorker, conf)
	go r.recycleTiming(enumor.CvmCloudResType, r.recycleCvmWorker, conf)
}

type recycleWorker func(kt *kit.Kit, info *types.CloudResourceBasicInfo) error

func (r *recycle) recycleTiming(resType enumor.CloudResourceType, worker recycleWorker, conf cc.Recycle) {
	for {
		kt := core.NewBackendKit()

		if !r.state.IsMaster() {
			logs.Infof("recycle %s, but is not master, skip", resType)
			time.Sleep(time.Minute)
			continue
		}

		logs.Infof("start recycle %s, rid: %s", resType, kt.Rid)
		// get need recycled resource records
		expr, err := tools.And(tools.EqualWithOpExpression(filter.And,
			map[string]interface{}{"res_type": resType, "status": enumor.WaitingRecycleRecordStatus}),
			&filter.AtomRule{Field: "recycled_at", Op: filter.LessThanEqual.Factory(),
				Value: times.ConvStdTimeFormat(time.Now())},
			// 不处理关联资源回收任务
			&filter.AtomRule{Field: "recycle_type", Op: filter.NotEqual.Factory(), Value: enumor.RecycleTypeRelated},
		)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}
		listReq := &core.ListReq{
			Filter: expr,
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "res_id", "bk_biz_id"},
		}
		recordRes, err := r.client.DataService().Global.RecycleRecord.ListRecycleRecord(kt, listReq)
		if err != nil {
			logs.Errorf("list %s resource recycle record failed, err: %v, rid: %s", resType, err, kt.Rid)
			time.Sleep(time.Minute)
			continue
		}

		// sleep for a while if no resource needs recycling
		if len(recordRes.Details) == 0 {
			time.Sleep(time.Minute * 10)
			continue
		}

		// get need recycled resource basic info
		ids := make([]string, 0, len(recordRes.Details))
		for _, record := range recordRes.Details {
			ids = append(ids, record.ResID)
		}

		infoReq := dataproto.ListResourceBasicInfoReq{
			ResourceType: resType,
			IDs:          ids,
			Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
		}
		basicInfoMap, err := r.client.DataService().Global.Cloud.ListResBasicInfo(kt, infoReq)
		if err != nil {
			if ef := errf.Error(err); ef.Code == errf.RecordNotFound {
				recordIDs := slice.Map(recordRes.Details, func(r recyclerecord.RecycleRecord) string { return r.ID })
				logs.Errorf("recycle %s res(ids: %+v) all don't exist, mark all as fail, reason: %v, rid: %s",
					resType, ids, err, kt.Rid)
				logicsrecycle.MarkRecordFailed(kt, r.client.DataService(), err, recordIDs)
				continue
			}
			logs.Errorf("get recycle %s resource detail failed, err: %v, ids: %+v, rid: %s", resType, err, ids, kt.Rid)
			time.Sleep(time.Minute)
			continue
		}

		// recycle resources one by one
		for _, record := range recordRes.Details {
			if !r.state.IsMaster() {
				logs.Infof("recycle %s res(id: %s), but is not master, skip, rid: %s", resType, record.ResID, kt.Rid)
				time.Sleep(time.Minute)
				break
			}
			r.execWorker(kt, worker, record, basicInfoMap)
		}

		logs.Infof("finished recycle %s, count: %d, rid: %s", resType, len(recordRes.Details), kt.Rid)
	}
}

const maxRetryCount = 3

func (r *recycle) execWorker(kt *kit.Kit, worker recycleWorker, record recyclerecord.RecycleRecord,
	basicInfoMap map[string]types.CloudResourceBasicInfo) {

	basicInfo, exists := basicInfoMap[record.ResID]
	if !exists {
		logs.Errorf("recycle %s res(id: %s) doesn't exists, mark as failed, rid: %s", record.ResType, record.ResID,
			kt.Rid)
		logicsrecycle.MarkRecordFailed(kt, r.client.DataService(),
			errf.New(errf.RecordNotFound, "Recourse Not Found"), []string{record.ID})
		return
	}

	rty := retry.NewRetryPolicy(maxRetryCount, [2]uint{500, 15000})
	var err error

	// 类型为cvm且在业务下回收的，需要检查是否在cmdb 待回收模块中
	// 因为cvm记录中的BkBizID已经在加入业务的时候被清掉了，所以要以recycle_record中的为准
	basicInfo.BkBizID = record.BkBizID
	err = rty.BaseExec(kt, func() error {
		return worker(kt, &basicInfo)
	})
	if err != nil {
		// Failed after retry
		logicsrecycle.MarkRecordFailed(kt, r.client.DataService(), err, []string{record.ID})
		return
	}
	// Success
	logs.V(3).Infof("[%s]recycle res(id: %s) success,  rid: %s", record.ResType, record.ID, kt.Rid)

	logicsrecycle.MarkRecordSuccess(kt, r.client.DataService(), []string{record.ID})
}

func (r *recycle) recycleDiskWorker(kt *kit.Kit, info *types.CloudResourceBasicInfo) error {
	res, err := r.logics.Disk.DeleteRecycledDisk(kt, map[string]types.CloudResourceBasicInfo{info.ID: *info})
	if err != nil {
		logs.Errorf("delete disk failed, err: %v, res: %+v, disk: %s, rid: %s", err, res, info.ID, kt.Rid)
		return err
	}
	return nil
}

func (r *recycle) recycleCvmWorker(kt *kit.Kit, info *types.CloudResourceBasicInfo) error {
	// 实际销毁CVM
	res, err := r.logics.Cvm.DestroyRecycledCvm(kt, map[string]types.CloudResourceBasicInfo{info.ID: *info}, nil)
	if err != nil {
		logs.Errorf("delete cvm failed, err: %v, res: %+v, cvm: %s, rid: %s", err, res, info.ID, kt.Rid)
		return err
	}
	return nil
}
