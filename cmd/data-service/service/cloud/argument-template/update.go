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

package argstpl

import (
	"fmt"

	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tableargstpl "hcm/pkg/dal/table/cloud/argument-template"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateArgsTpl batch update argument template
func (svc *argsTplSvc) BatchUpdateArgsTpl(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.ArgsTplBatchUpdateExprReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateData := &tableargstpl.ArgumentTemplateTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}

	if len(req.Name) > 0 {
		updateData.Name = req.Name
	}

	if len(req.Templates) > 0 && !req.Templates.IsEmpty() {
		updateData.Templates = req.Templates
	}

	if len(req.GroupTemplates) > 0 && !req.GroupTemplates.IsEmpty() {
		updateData.GroupTemplates = req.GroupTemplates
	}

	if err := svc.dao.ArgsTpl().Update(cts.Kit, tools.ContainersExpression("id", req.IDs), updateData); err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchUpdateArgsTplExt batch update argument template ext
func (svc *argsTplSvc) BatchUpdateArgsTplExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateArgsTplExt[coreargstpl.TCloudArgsTplExtension](cts, svc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchUpdateArgsTplExt[T coreargstpl.Extension](cts *rest.Contexts, svc *argsTplSvc) (interface{}, error) {
	req := new(dataproto.ArgsTplExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, item := range *req {
			updateData := &tableargstpl.ArgumentTemplateTable{
				Name:           item.Name,
				BkBizID:        int64(item.BkBizID),
				Type:           item.Type,
				Templates:      item.Templates,
				GroupTemplates: item.GroupTemplates,
				Reviser:        cts.Kit.User,
			}

			if err := svc.dao.ArgsTpl().UpdateByIDWithTx(cts.Kit, txn, item.ID, updateData); err != nil {
				return nil, fmt.Errorf("update argument template db failed, err: %v", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update argument template ext db failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
