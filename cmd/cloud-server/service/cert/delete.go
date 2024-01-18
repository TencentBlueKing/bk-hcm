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
	dataproto "hcm/pkg/api/data-service/cloud"
	protocert "hcm/pkg/api/hc-service/cert"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// DeleteCert delete resource cert.
func (svc *certSvc) DeleteCert(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteCertSvc(cts, handler.ResOperateAuth)
}

// DeleteBizCert delete biz cert.
func (svc *certSvc) DeleteBizCert(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteCertSvc(cts, handler.BizOperateAuth)
}

func (svc *certSvc) deleteCertSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CertCloudResType,
		IDs:          []string{id},
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list cert basic info failed, req: %+v, err: %v, rid: %s", basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cert,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		logs.Errorf("delete cert auth failed, id: %s, err: %v, rid: %s", id, err, cts.Kit.Rid)
		return nil, err
	}

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.SslCertAuditResType, basicInfoReq.IDs); err != nil {
		logs.Errorf("create operation audit cert failed, ids: %v, err: %v, rid: %s", basicInfoReq.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	// delete tcloud cloud cert
	certInfo, ok := basicInfoMap[id]
	if !ok {
		logs.Errorf("cert record is not found, id: %s, rid: %s", id, cts.Kit.Rid)
		return nil, errf.Newf(errf.Aborted, "cert %s record is not found", id)
	}

	err = svc.client.HCService().TCloud.Cert.DeleteCert(cts.Kit, &protocert.TCloudDeleteReq{
		AccountID: certInfo.AccountID,
		ID:        id,
	})
	if err != nil {
		logs.Errorf("[%s] request hcservice to delete cert failed, id: %s, err: %v, rid: %s",
			enumor.TCloud, id, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
