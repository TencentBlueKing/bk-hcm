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

	"hcm/pkg/api/core"
	corecert "hcm/pkg/api/core/cloud/cert"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud/eip"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	tablecert "hcm/pkg/dal/table/cloud/cert"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
)

// ListCert list cert.
func (svc *certSvc) ListCert(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.CertListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Field,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.Cert().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list cert failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list cert failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.CertListResult{Count: result.Count}, nil
	}

	details := make([]corecert.BaseCert, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, tErr := convTableToBaseCert(&one)
		if tErr != nil {
			logs.Errorf("list loop cert detail failed, err: %v, rid: %s", err, cts.Kit.Rid)
			continue
		}

		details = append(details, *tmpOne)
	}

	return &protocloud.CertListResult{Details: details}, nil
}

func convTableToBaseCert(one *tablecert.SslCertTable) (*corecert.BaseCert, error) {
	domain := new([]*string)
	err := json.UnmarshalFromString(string(one.Domain), domain)
	if err != nil {
		return nil, fmt.Errorf("UnmarshalFromString db domain failed, err: %v", err)
	}

	base := &corecert.BaseCert{
		ID:               one.ID,
		CloudID:          one.CloudID,
		Name:             one.Name,
		Vendor:           one.Vendor,
		BkBizID:          one.BkBizID,
		AccountID:        one.AccountID,
		Domain:           converter.PtrToVal(domain),
		CertType:         one.CertType,
		CertStatus:       one.CertStatus,
		EncryptAlgorithm: one.EncryptAlgorithm,
		CloudCreatedTime: one.CloudCreatedTime,
		CloudExpiredTime: one.CloudExpiredTime,
		Memo:             one.Memo,
		Tags:             core.TagMap(one.Tags),
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}

	return base, nil
}

// ListCertExt list cert ext.
func (svc *certSvc) ListCertExt(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	req := new(dataproto.EipListReq)
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

	data, err := svc.dao.Cert().List(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return convCertListResult[corecert.TCloudCertExtension](cts.Kit, data.Details)
	default:
		return nil, errf.Newf(errf.InvalidParameter, "unsupported vendor: %s", vendor)
	}
}

func convCertListResult[T corecert.Extension](kt *kit.Kit, tables []tablecert.SslCertTable) (
	*protocloud.CertExtListResult[T], error) {

	details := make([]corecert.Cert[T], 0, len(tables))
	for _, one := range tables {
		tmpCert, err := convTableToBaseCert(&one)
		if err != nil {
			logs.Errorf("list loop cert detail failed, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		extension := new(T)
		details = append(details, corecert.Cert[T]{
			BaseCert:  *tmpCert,
			Extension: extension,
		})
	}

	return &protocloud.CertExtListResult[T]{
		Details: details,
	}, nil
}
