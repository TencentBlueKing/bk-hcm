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

package quota

import (
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"

	"hcm/pkg/api/core"
	corequota "hcm/pkg/api/core/cloud/quota"
	dataservice "hcm/pkg/api/data-service"
	dsquota "hcm/pkg/api/data-service/cloud/quota"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablequota "hcm/pkg/dal/table/cloud/quota"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// BatchCreateBizQuota batch create biz quota.
func (svc *service) BatchCreateBizQuota(cts *rest.Contexts) (interface{}, error) {

	req := new(dsquota.CreateBizQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	quotaID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		// TODO: 补充判断云配额是否够分配和更新云配额已使用大小逻辑。

		model := &tablequota.BizQuotaTable{
			CloudQuotaID: req.CloudQuotaID,
			AccountID:    req.AccountID,
			BkBizID:      req.BkBizID,
			ResType:      req.ResType,
			Vendor:       req.Vendor,
			Region:       req.Region,
			Zone:         req.Zone,
			Levels:       req.Levels,
			Dimension:    req.Dimensions,
			Memo:         req.Memo,
			Creator:      cts.Kit.User,
			Reviser:      cts.Kit.User,
		}

		quotaID, err := svc.dao.BizQuota().CreateWithTx(cts.Kit, txn, model)
		if err != nil {
			logs.Errorf("create biz quota failed, err: %v, model: %+v, rid: %s", err, model, cts.Kit.Rid)
			return nil, fmt.Errorf("create biz quota failed, err: %v", err)
		}

		return quotaID, nil
	})

	if err != nil {
		return nil, err
	}

	id, ok := quotaID.(string)
	if !ok {
		return nil, fmt.Errorf("create biz quota but return id type %s is not string", reflect.TypeOf(quotaID).String())
	}

	return &core.CreateResult{ID: id}, nil
}

// BatchUpdateBizQuota batch update biz quota.
func (svc *service) BatchUpdateBizQuota(cts *rest.Contexts) (interface{}, error) {

	req := new(dsquota.UpdateBizQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if len(req.Dimensions) != 0 {
			// TODO: 补充判断云配额是否够分配和更新云配额已使用大小逻辑。
		}

		model := &tablequota.BizQuotaTable{
			CloudQuotaID: req.CloudQuotaID,
			AccountID:    req.AccountID,
			Vendor:       req.Vendor,
			Dimension:    req.Dimensions,
			Memo:         req.Memo,
			Reviser:      cts.Kit.User,
		}

		err := svc.dao.BizQuota().Update(cts.Kit, txn, req.ID, model)
		if err != nil {
			logs.Errorf("update biz quota failed, err: %v, model: %+v, rid: %s", err, model, cts.Kit.Rid)
			return nil, fmt.Errorf("update biz quota failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ListBizQuota list biz quota.
func (svc *service) ListBizQuota(cts *rest.Contexts) (interface{}, error) {

	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page:   req.Page,
		Fields: req.Fields,
	}
	result, err := svc.dao.BizQuota().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list biz quota failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list biz quota failed, err: %v", err)
	}
	if req.Page.Count {
		return &dsquota.ListBizQuotaResult{Count: result.Count}, nil
	}

	details := make([]corequota.BizQuota, 0, len(result.Details))
	for _, one := range result.Details {
		details = append(details, convertBizQuota(one))
	}

	return &dsquota.ListBizQuotaResult{Details: details}, nil
}

// DeleteBizQuota delete biz quota.
func (svc *service) DeleteBizQuota(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Filter: req.Filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}
	listResp, err := svc.dao.BizQuota().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list biz quota failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list biz quota failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delIDs[index] = one.ID
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		// TODO: 补充判断云配额是否够分配和更新云配额已使用大小逻辑。

		delFilter := tools.ContainersExpression("id", delIDs)
		if err = svc.dao.BizQuota().BatchDeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			logs.Errorf("batch delete biz quota failed, err: %v, ids: %v, rid: %s", err, delIDs, cts.Kit.Rid)
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete biz quota failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func convertBizQuota(model tablequota.BizQuotaTable) corequota.BizQuota {
	return corequota.BizQuota{
		ID:           model.ID,
		CloudQuotaID: model.CloudQuotaID,
		Vendor:       model.Vendor,
		ResType:      model.ResType,
		AccountID:    model.AccountID,
		BkBizID:      model.BkBizID,
		Region:       model.Region,
		Zone:         model.Zone,
		Levels:       model.Levels,
		Dimensions:   model.Dimension,
		Memo:         model.Memo,
		Revision: &core.Revision{
			Creator:   model.Creator,
			Reviser:   model.Reviser,
			CreatedAt: model.CreatedAt.String(),
			UpdatedAt: model.UpdatedAt.String(),
		},
	}
}
