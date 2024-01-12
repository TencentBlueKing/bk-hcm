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
	"reflect"

	"hcm/pkg/api/core"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/table/cloud/argument-template"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// CreateArgsTpl create argument template.
func (svc *argsTplSvc) CreateArgsTpl(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateArgsTpl[coreargstpl.TCloudArgsTplExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func batchCreateArgsTpl[T coreargstpl.Extension](cts *rest.Contexts, svc *argsTplSvc, vendor enumor.Vendor) (
	interface{}, error) {

	req := new(protocloud.ArgsTplBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]*tableargstpl.ArgumentTemplateTable, 0, len(req.ArgumentTemplates))
		for _, one := range req.ArgumentTemplates {
			models = append(models, &tableargstpl.ArgumentTemplateTable{
				CloudID:        one.CloudID,
				Name:           one.Name,
				Vendor:         vendor,
				BkBizID:        one.BkBizID,
				AccountID:      one.AccountID,
				Type:           one.Type,
				Templates:      one.Templates,
				GroupTemplates: one.GroupTemplates,
				Memo:           one.Memo,
				Creator:        cts.Kit.User,
				Reviser:        cts.Kit.User,
			})
		}

		ids, err := svc.dao.ArgsTpl().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create argument template failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create argument template but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}
	return &core.BatchCreateResult{IDs: ids}, nil
}
