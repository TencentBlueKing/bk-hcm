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

package logicsrecycle

import (
	corerr "hcm/pkg/api/core/recycle-record"
	"hcm/pkg/api/data-service/recycle-record"
	"hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// MarkRecordFailed 标记回收失败
func MarkRecordFailed(kt *kit.Kit, ds *dataservice.Client, err error, recordIDs []string) {

	updateReq := &recyclerecord.BatchUpdateReq{Data: slice.Map(recordIDs, func(r string) recyclerecord.UpdateReq {
		return recyclerecord.UpdateReq{
			ID:     r,
			Status: enumor.FailedRecycleRecordStatus,
			Detail: corerr.BaseRecycleDetail{ErrorMessage: err.Error()},
		}
	})}

	err = ds.Global.RecycleRecord.BatchUpdateRecycleRecord(kt, updateReq)
	if err != nil {
		logs.Errorf("[%s] fail to update recycle record (%+v) status to failed, err: %v, rid: %s",
			constant.RecycleUpdateRecordFailed, recordIDs, err, kt.Rid)
	}
}

// MarkRecordSuccess 标记回收成功
func MarkRecordSuccess(kt *kit.Kit, ds *dataservice.Client, recordIDs []string) {

	updateReq := &recyclerecord.BatchUpdateReq{Data: slice.Map(recordIDs, func(r string) recyclerecord.UpdateReq {
		return recyclerecord.UpdateReq{
			ID:     r,
			Status: enumor.RecycledRecycleRecordStatus,
		}
	})}

	if err := ds.Global.RecycleRecord.BatchUpdateRecycleRecord(kt, updateReq); err != nil {
		logs.Errorf("[%s] fail to update recycle records(%+v) status to success, err: %v, rid: %s",
			constant.RecycleUpdateRecordFailed, recordIDs, err, kt.Rid)
	}
}
