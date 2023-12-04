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

package detail

import (
	"fmt"
	"time"

	"hcm/pkg/api/core"
	dssync "hcm/pkg/api/data-service/cloud/sync"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	ttimes "hcm/pkg/tools/times"
)

// SyncDetail ...
type SyncDetail struct {
	Kt        *kit.Kit
	DataCli   *dataservice.Client
	AccountID string
	Vendor    string
	ResStatus string
}

// ResSyncStatusFailed ...
func (s *SyncDetail) ResSyncStatusFailed(resName enumor.CloudResourceType, failedErr error) error {

	s.ResStatus = string(enumor.SyncFailed)
	if err := s.changeResSyncStatus(resName, failedErr); err != nil {
		return err
	}

	return nil
}

// ResSyncStatusSuccess ...
func (s *SyncDetail) ResSyncStatusSuccess(resName enumor.CloudResourceType) error {

	s.ResStatus = string(enumor.SyncSuccess)
	if err := s.changeResSyncStatus(resName, nil); err != nil {
		return err
	}

	return nil
}

// ResSyncStatusSyncing ...
func (s *SyncDetail) ResSyncStatusSyncing(resName enumor.CloudResourceType) error {

	s.ResStatus = string(enumor.Syncing)
	if err := s.changeResSyncStatus(resName, nil); err != nil {
		return err
	}

	return nil
}

func (s *SyncDetail) changeResSyncStatus(resName enumor.CloudResourceType, failedErr error) error {

	failedString := types.JsonField("")
	if failedErr != nil {
		if ef := errf.Error(failedErr); ef != nil && ef.Code == errf.Unknown {
			// 对于未知错误，直接给Message
			failedString, _ = types.NewJsonField(ef.Message)
		} else {
			failedString, _ = types.NewJsonField(failedErr)
		}
	}

	listReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: s.AccountID,
				},
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: s.Vendor,
				},
				&filter.AtomRule{
					Field: "res_name",
					Op:    filter.Equal.Factory(),
					Value: resName,
				},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	accountSyncDetail, err := s.DataCli.Global.AccountSyncDetail.List(s.Kt, listReq)
	if err != nil {
		return err
	}

	if len(accountSyncDetail.Details) > 1 {
		return fmt.Errorf("%s sync detail can not big than 1", s.AccountID)
	}

	if len(accountSyncDetail.Details) == 0 {
		// 不存在则新增
		createReq := &dssync.CreateReq{
			Items: []dssync.CreateField{
				{
					Vendor:          enumor.Vendor(s.Vendor),
					AccountID:       s.AccountID,
					ResName:         string(resName),
					ResStatus:       s.ResStatus,
					ResEndTime:      ttimes.ConvStdTimeFormat(time.Now()),
					ResFailedReason: failedString,
				},
			},
		}
		_, err := s.DataCli.Global.AccountSyncDetail.BatchCreate(s.Kt, createReq)
		if err != nil {
			return err
		}
	} else {
		// 存在则更新
		updateReq := &dssync.UpdateReq{
			Items: []dssync.UpdateField{
				{
					ID:              accountSyncDetail.Details[0].ID,
					ResStatus:       s.ResStatus,
					ResEndTime:      ttimes.ConvStdTimeFormat(time.Now()),
					ResFailedReason: failedString,
				},
			},
		}
		err := s.DataCli.Global.AccountSyncDetail.BatchUpdate(s.Kt, updateReq)
		if err != nil {
			return err
		}
	}

	return nil
}
