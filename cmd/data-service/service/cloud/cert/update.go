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

package cert

import (
	"fmt"

	corecert "hcm/pkg/api/core/cloud/cert"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablecert "hcm/pkg/dal/table/cloud/cert"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateCert batch update cert
func (svc *certSvc) BatchUpdateCert(cts *rest.Contexts) (interface{}, error) {
	req := new(dataproto.CertBatchUpdateExprReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	updateData := &tablecert.SslCertTable{
		BkBizID: req.BkBizID,
		Reviser: cts.Kit.User,
	}

	if len(req.Domain) > 0 && !req.Domain.IsEmpty() {
		updateData.Domain = req.Domain
	}

	if len(req.CertType) > 0 {
		updateData.CertType = req.CertType
	}

	if len(req.EncryptAlgorithm) > 0 {
		updateData.EncryptAlgorithm = req.EncryptAlgorithm
	}

	if len(req.CertStatus) > 0 {
		updateData.CertStatus = req.CertStatus
	}

	if len(req.CloudExpiredTime) > 0 {
		updateData.CloudExpiredTime = req.CloudExpiredTime
	}

	if err := svc.dao.Cert().Update(cts.Kit, tools.ContainersExpression("id", req.IDs), updateData); err != nil {
		return nil, err
	}

	return nil, nil
}

// BatchUpdateCertExt batch update cert ext
func (svc *certSvc) BatchUpdateCertExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchUpdateCertExt[corecert.TCloudCertExtension](cts, svc)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func batchUpdateCertExt[T corecert.Extension](cts *rest.Contexts, svc *certSvc) (interface{}, error) {
	req := new(dataproto.CertExtBatchUpdateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, item := range *req {
			updateData := &tablecert.SslCertTable{
				Name:             item.Name,
				BkBizID:          int64(item.BkBizID),
				Domain:           item.Domain,
				CertType:         item.CertType,
				CertStatus:       item.CertStatus,
				EncryptAlgorithm: item.EncryptAlgorithm,
				CloudExpiredTime: item.CloudExpiredTime,
				Tags:             types.StringMap(item.Tags),
				Reviser:          cts.Kit.User,
			}

			if err := svc.dao.Cert().UpdateByIDWithTx(cts.Kit, txn, item.ID, updateData); err != nil {
				return nil, fmt.Errorf("update cert db failed, err: %v", err)
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update cert ext db failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
