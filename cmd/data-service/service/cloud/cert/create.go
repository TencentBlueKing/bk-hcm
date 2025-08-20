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
	"reflect"

	"hcm/pkg/api/core"
	corecert "hcm/pkg/api/core/cloud/cert"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablecert "hcm/pkg/dal/table/cloud/cert"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// CreateCert create cert.
func (svc *certSvc) CreateCert(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return batchCreateCert[corecert.TCloudCertExtension](cts, svc, vendor)
	default:
		return nil, fmt.Errorf("unsupport %s vendor for now", vendor)
	}
}

func batchCreateCert[T corecert.Extension](cts *rest.Contexts, svc *certSvc, vendor enumor.Vendor) (
	interface{}, error) {

	req := new(protocloud.CertBatchCreateReq[T])
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]*tablecert.SslCertTable, 0, len(req.Certs))
		for _, one := range req.Certs {
			models = append(models, &tablecert.SslCertTable{
				CloudID:          one.CloudID,
				Name:             one.Name,
				Vendor:           vendor,
				BkBizID:          one.BkBizID,
				AccountID:        one.AccountID,
				Domain:           one.Domain,
				CertType:         one.CertType,
				CertStatus:       one.CertStatus,
				EncryptAlgorithm: one.EncryptAlgorithm,
				CloudCreatedTime: one.CloudCreatedTime,
				CloudExpiredTime: one.CloudExpiredTime,
				Memo:             one.Memo,
				Tags:             types.StringMap(one.Tags),
				Creator:          cts.Kit.User,
				Reviser:          cts.Kit.User,
			})
		}

		ids, err := svc.dao.Cert().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create cert failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create cert but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
