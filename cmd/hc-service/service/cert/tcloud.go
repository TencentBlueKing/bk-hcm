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

// Package cert ...
package cert

import (
	"net/http"

	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/capability"
	typecert "hcm/pkg/adaptor/types/cert"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cert"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocert "hcm/pkg/api/hc-service/cert"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

func (svc *certSvc) initTCloudCertService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("CreateTCloudCert", http.MethodPost, "/vendors/tcloud/certs/create", svc.CreateTCloudCert)
	h.Add("DeleteTCloudCert", http.MethodDelete, "/vendors/tcloud/certs", svc.DeleteTCloudCert)
	h.Add("ListTCloudCert", http.MethodPost, "/vendors/tcloud/certs/list", svc.ListTCloudCert)

	h.Load(cap.WebService)
}

// CreateTCloudCert ...
func (svc *certSvc) CreateTCloudCert(cts *rest.Contexts) (interface{}, error) {
	req := new(protocert.TCloudCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	createOpt := &typecert.TCloudCreateOption{
		Name:       req.Name,
		CertType:   string(req.CertType),
		PublicKey:  req.PublicKey,
		PrivateKey: req.PrivateKey,
		Repeatable: true,
	}
	result, err := tcloud.CreateCert(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("request adaptor tcloud upload cert error, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	respData := &cert.CertCreateResult{}
	if len(result.SuccessCloudIDs) == 0 {
		logs.Errorf("request adaptor tcloud upload cert failed, req: %+v, rid: %s", req, cts.Kit.Rid)
		return nil, errf.Newf(errf.Aborted, "upload certificate failed")
	}

	cloudIDs := result.SuccessCloudIDs
	syncClient := synctcloud.NewClient(svc.dataCli, tcloud)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    "region",
		CloudIDs:  cloudIDs,
	}
	_, err = syncClient.Cert(cts.Kit, params, &synctcloud.SyncCertOption{BkBizID: req.BkBizID})
	if err != nil {
		logs.Errorf("sync tcloud cert failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	// 查询证书云ID对应的DB记录
	resp, err := svc.dataCli.Global.ListCert(
		cts.Kit,
		&core.ListReq{Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "cloud_id",
					Op:    filter.In.Factory(),
					Value: cloudIDs,
				}, &filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: req.Vendor,
				},
			},
		}, Page: &core.BasePage{Limit: uint(len(cloudIDs))}, Fields: []string{"id"}},
	)
	if err != nil {
		logs.Errorf("request dataservice cert list failed, cloudIDs: %v, err: %v, rid: %s", cloudIDs, err, cts.Kit.Rid)
		return nil, err
	}

	if len(resp.Details) == 0 {
		return respData, nil
	}

	return &cert.CertCreateResult{ID: resp.Details[0].ID}, nil
}

// DeleteTCloudCert ...
func (svc *certSvc) DeleteTCloudCert(cts *rest.Contexts) (interface{}, error) {
	req := new(protocert.TCloudDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.EqualExpression("id", req.ID),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.ListCert(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cert failed, id: %s, err: %v, rid: %s", req.ID, err, cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		logs.Errorf("request dataservice list tcloud cert empty, id: %s, rid: %s", err, req.ID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("get adaptor to tcloud client failed, accID: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typecert.TCloudDeleteOption{
		CloudID: listResp.Details[0].CloudID,
	}
	if err = client.DeleteCert(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete tcloud cert failed, err: %v, opt: %+v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CertBatchDeleteReq{
		Filter: tools.EqualExpression("id", req.ID),
	}
	if err = svc.dataCli.Global.BatchDeleteCert(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete tcloud cert failed, err: %v, id: %s, rid: %s", err, req.ID, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListTCloudCert list tcloud cert
func (svc *certSvc) ListTCloudCert(cts *rest.Contexts) (interface{}, error) {
	req := new(protocert.TCloudListOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloud, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecert.TCloudListOption{
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  adcore.TCloudQueryLimit,
		},
	}
	result, err := tcloud.ListCert(cts.Kit, opt)
	if err != nil {
		logs.Errorf("[%s] list cert failed, req: %+v, err: %v, rid: %s", enumor.TCloud, req, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
