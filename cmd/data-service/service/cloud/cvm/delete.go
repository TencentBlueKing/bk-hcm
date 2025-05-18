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

package cvm

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// BatchDeleteCvm cvm.
func (svc *cvmSvc) BatchDeleteCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CvmBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "vendor", "cloud_id", "bk_biz_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.Cvm().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cvm failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	delCvmIDs := make([]string, len(listResp.Details))
	for index, one := range listResp.Details {
		delCvmIDs[index] = one.ID
	}

	niIDs, err := svc.listCvmAssNetworkInterface(cts.Kit, delCvmIDs)
	if err != nil {
		return nil, err
	}

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if len(niIDs) != 0 {
			delFilter := tools.ContainersExpression("id", niIDs)
			if err := svc.dao.NetworkInterface().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
				return nil, err
			}
		}

		// 删除安全组关联关系
		sgRelFilter := tools.ExpressionAnd(
			tools.RuleIn("res_id", delCvmIDs),
			tools.RuleEqual("res_type", enumor.CvmCloudResType),
		)
		err := svc.dao.SGCommonRel().DeleteWithTx(cts.Kit, txn, sgRelFilter)
		if err != nil {
			logs.Errorf("delete cvm sg rel failed , err: %v, cvm_ids: %v, rid: %s", err, delCvmIDs, cts.Kit.Rid)
			return nil, err
		}

		delFilter := tools.ContainersExpression("id", delCvmIDs)
		if err := svc.dao.Cvm().DeleteWithTx(cts.Kit, txn, delFilter); err != nil {
			return nil, err
		}

		// delete cmdb cloud hosts
		if err = deleteCmdbHosts(svc, cts.Kit, listResp.Details); err != nil {
			logs.Errorf("delete cmdb hosts failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, nil
		}

		return nil, nil
	})
	if err != nil {
		logs.Errorf("delete cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *cvmSvc) listCvmAssNetworkInterface(kt *kit.Kit, cvmIDs []string) ([]string, error) {

	ids := make([]string, 0)
	split := slice.Split(cvmIDs, int(filter.DefaultMaxInLimit))
	for _, partID := range split {
		opt := &types.ListOption{
			Filter: tools.ContainersExpression("cvm_id", partID),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := svc.dao.NiCvmRel().List(kt, opt)
		if err != nil {
			logs.Errorf("list ni_cvm_rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			ids = append(ids, one.NetworkInterfaceID)
		}
	}

	return ids, nil
}
