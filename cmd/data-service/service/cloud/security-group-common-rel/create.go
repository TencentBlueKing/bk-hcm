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

package sgcomrel

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchCreateSgCommonRels rels.
func (svc *sgComRelSvc) BatchCreateSgCommonRels(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SGCommonRelBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablecloud.SecurityGroupCommonRelTable, 0, len(req.Rels))
		for _, one := range req.Rels {
			models = append(models, tablecloud.SecurityGroupCommonRelTable{
				ResVendor:       one.ResVendor,
				ResID:           one.ResID,
				ResType:         one.ResType,
				SecurityGroupID: one.SecurityGroupID,
				Priority:        one.Priority,
				Creator:         cts.Kit.User,
			})
		}

		// check relation resource is existed
		if err := svc.checkRelationResourceExist(cts.Kit, req.Rels); err != nil {
			logs.Errorf("check relation resource exist failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, nil
		}
		if err := svc.dao.SGCommonRel().BatchCreateWithTx(cts.Kit, txn, models); err != nil {
			return nil, fmt.Errorf("batch create sg common rels failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch create sg common rels failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchUpsertSgCommonRels rels.
func (svc *sgComRelSvc) BatchUpsertSgCommonRels(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.SGCommonRelBatchUpsertReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	delIDs := make([]uint64, 0)
	if req.DeleteReq != nil && req.DeleteReq.Filter != nil {
		opt := &types.ListOption{
			Fields: []string{"id"},
			Filter: req.DeleteReq.Filter,
			Page:   core.NewDefaultBasePage(),
		}
		listResp, err := svc.dao.SGCommonRel().List(cts.Kit, opt)
		if err != nil {
			logs.Errorf("list security group common rels failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("list security group common rels failed, err: %v", err)
		}

		if len(listResp.Details) == 0 && len(req.Rels) == 0 {
			return nil, nil
		}

		for _, one := range listResp.Details {
			delIDs = append(delIDs, one.ID)
		}
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if len(delIDs) > 0 {
			if err := svc.dao.SGCommonRel().DeleteWithTx(
				cts.Kit, txn, tools.ContainersExpression("id", delIDs)); err != nil {
				return nil, err
			}
		}

		models := make([]tablecloud.SecurityGroupCommonRelTable, 0, len(req.Rels))
		for _, one := range req.Rels {
			models = append(models, tablecloud.SecurityGroupCommonRelTable{
				ResVendor:       one.ResVendor,
				ResID:           one.ResID,
				ResType:         one.ResType,
				SecurityGroupID: one.SecurityGroupID,
				Priority:        one.Priority,
				Creator:         cts.Kit.User,
			})
		}
		// check relation resource is existed
		if err := svc.checkRelationResourceExist(cts.Kit, req.Rels); err != nil {
			logs.Errorf("check relation resource exist failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, nil
		}

		if err := svc.dao.SGCommonRel().BatchCreateWithTx(cts.Kit, txn, models); err != nil {
			return nil, fmt.Errorf("batch create sg common rels failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch upsert sg common rels failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *sgComRelSvc) checkRelationResourceExist(kt *kit.Kit, rels []protocloud.SGCommonRelCreate) error {
	// check relation resource is existed
	// 校验关联资源是否存在
	sgIDs := make([]string, 0)
	resTypeToResIDsMap := make(map[enumor.CloudResourceType][]string)
	for _, rel := range rels {
		sgIDs = append(sgIDs, rel.SecurityGroupID)
		resTypeToResIDsMap[rel.ResType] = append(resTypeToResIDsMap[rel.ResType], rel.ResID)
	}

	sgMap := make(map[string]tablecloud.SecurityGroupTable)
	for _, ids := range slice.Split(sgIDs, int(core.DefaultMaxPageLimit)) {
		listOpt := &types.ListOption{
			Filter: tools.ContainersExpression("id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := svc.dao.SecurityGroup().List(kt, listOpt)
		if err != nil {
			logs.Errorf("list security group failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
			return err
		}
		for _, detail := range resp.Details {
			sgMap[detail.ID] = detail
		}
	}

	if len(sgMap) != len(converter.StringSliceToMap(sgIDs)) {
		logs.Errorf("get security group count not right, ids: %v, count: %d, rid: %s", sgIDs, len(sgMap), kt.Rid)
		return fmt.Errorf("get security group count not right")
	}

	for resType, resIDs := range resTypeToResIDsMap {
		dbResp, err := svc.dao.Cloud().ListResourceIDs(kt, resType, tools.ContainersExpression("id", resIDs))
		if err != nil {
			logs.Errorf("list resource ids failed, err: %v, resType: %s, resIDs: %v, rid: %s",
				err, resType, resIDs, kt.Rid)
			return err
		}
		if len(dbResp) != len(converter.StringSliceToMap(resIDs)) {
			logs.Errorf("get resource count not right, err: %v, resType: %s, resIDs: %v, rid: %s",
				err, resType, resIDs, kt.Rid)
			return fmt.Errorf("get resource count not right")
		}
	}

	return nil
}
