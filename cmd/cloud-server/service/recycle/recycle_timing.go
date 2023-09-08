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
	"hcm/pkg/api/core"
	recyclerecord "hcm/pkg/api/core/recycle-record"
	dataproto "hcm/pkg/api/data-service/cloud"
	rr "hcm/pkg/api/data-service/recycle-record"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/retry"
)

type recycle struct {
	client *client.ClientSet
	logics *logics.Logics
	state  serviced.State
}

// RecycleTiming timing recycle all resource.
func RecycleTiming(c *client.ClientSet, state serviced.State, conf cc.Recycle) {
	r := &recycle{
		client: c,
		state:  state,
		logics: logics.NewLogics(c),
	}

	go r.recycleTiming(enumor.DiskCloudResType, r.recycleDisk, conf)
	go r.recycleTiming(enumor.CvmCloudResType, r.recycleCvm, conf)
}

type recycleWorker func(kt *kit.Kit, info *types.CloudResourceBasicInfo) error

func (r *recycle) recycleTiming(resType enumor.CloudResourceType, worker recycleWorker, conf cc.Recycle) {
	for {
		kt := kit.New()
		kt.User = constant.RecycleTimingUserKey
		kt.AppCode = constant.RecycleTimingAppCodeKey

		if !r.state.IsMaster() {
			logs.Infof("recycle %s, but is not master, skip", resType)
			time.Sleep(time.Minute)
			continue
		}

		logs.Infof("start recycle %s, rid: %s", resType, kt.Rid)
		// get need recycled resource records
		expr, err := tools.And(tools.EqualWithOpExpression(filter.And,
			map[string]interface{}{"res_type": resType, "status": enumor.WaitingRecycleRecordStatus}),
			&filter.AtomRule{Field: "created_at", Op: filter.LessThanEqual.Factory(),
				Value: time.Now().Add(-time.Hour * time.Duration(conf.AutoDeleteTime)).Format(constant.TimeStdFormat)})
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}
		listReq := &core.ListReq{
			Filter: expr,
			Page:   core.NewDefaultBasePage(),
			Fields: []string{"id", "res_id"},
		}
		recordRes, err := r.client.DataService().Global.RecycleRecord.ListRecycleRecord(kt.Ctx, kt.Header(), listReq)
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
		basicInfoMap, err := r.client.DataService().Global.Cloud.ListResourceBasicInfo(kt.Ctx, kt.Header(), infoReq)
		if err != nil {
			if ef := errf.Error(err); ef.Code == errf.RecordNotFound {
				recordIDs := make([]string, 0, len(recordRes.Details))
				for _, record := range recordRes.Details {
					recordIDs = append(recordIDs, record.ID)
				}
				logs.Errorf("recycle %s res(ids: %+v) all don't exist, mark all as fail, reason: %v, rid: %s",
					resType, ids, err, kt.Rid)
				r.markfail(kt, err, recordIDs)
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
		r.markfail(kt, errf.New(errf.RecordNotFound, "Recourse Not Found"), []string{record.ID})
		return
	}

	rty := retry.NewRetryPolicy(maxRetryCount, [2]uint{500, 15000})
	var err error
	err = rty.BaseExec(kt, func() error {
		return worker(kt, &basicInfo)
	})
	if err != nil {
		// Failed after retry
		r.markfail(kt, err, []string{record.ID})
		return
	}
	// Success
	logs.V(3).Infof("[%s]recycle res(id: %s) success,  rid: %s", record.ResType, record.ID, kt.Rid)
	req := &rr.BatchUpdateReq{Data: []rr.UpdateReq{{
		ID:     record.ID,
		Status: enumor.RecycledRecycleRecordStatus,
	}}}
	err = r.client.DataService().Global.RecycleRecord.BatchUpdateRecycleRecord(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] update recycle record %d failed, err: %v, rid: %s",
			constant.RecycleUpdateRecordFailed, record.ID, err, kt.Rid)
	}
}

func (r *recycle) recycleDisk(kt *kit.Kit, info *types.CloudResourceBasicInfo) error {
	res, err := r.logics.Disk.DeleteRecycledDisk(kt, map[string]types.CloudResourceBasicInfo{info.ID: *info})
	if err != nil {
		logs.Errorf("delete disk failed, err: %v, res: %+v, disk: %s, rid: %s", err, res, info.ID, kt.Rid)
		return err
	}
	return nil
}

func (r *recycle) recycleCvm(kt *kit.Kit, info *types.CloudResourceBasicInfo) error {
	res, err := r.logics.Cvm.DeleteRecycledCvm(kt, map[string]types.CloudResourceBasicInfo{info.ID: *info})
	if err != nil {
		logs.Errorf("delete cvm failed, err: %v, res: %+v, cvm: %s, rid: %s", err, res, info.ID, kt.Rid)
		return err
	}
	return nil
}

func (r *recycle) markfail(kt *kit.Kit, err error, recordIDs []string) {
	updateReq := make([]rr.UpdateReq, len(recordIDs))
	for i, id := range recordIDs {
		updateReq[i] = rr.UpdateReq{
			ID:     id,
			Status: enumor.FailedRecycleRecordStatus,
			Detail: recyclerecord.BaseRecycleDetail{ErrorMessage: err.Error()},
		}
	}
	req := &rr.BatchUpdateReq{Data: updateReq}
	err = r.client.DataService().Global.RecycleRecord.BatchUpdateRecycleRecord(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] update recycle record (%+v) failed, err: %v, rid: %s",
			constant.RecycleUpdateRecordFailed, recordIDs, err, kt.Rid)
	}
}
