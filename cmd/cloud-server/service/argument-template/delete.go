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

// Package argstpl ...
package argstpl

import (
	"fmt"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	protoargstpl "hcm/pkg/api/hc-service/argument-template"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// DeleteBizArgsTpl delete biz argument template.
func (svc *argsTplSvc) DeleteBizArgsTpl(cts *rest.Contexts) (interface{}, error) {
	return svc.deleteArgsTplSvc(cts, handler.BizOperateAuth)
}

func (svc *argsTplSvc) deleteArgsTplSvc(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.ArgumentTemplateResType,
		IDs:          req.IDs,
		Fields:       types.CommonBasicInfoFields,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("list argument template basic info failed, req: %+v, err: %v, rid: %s",
			basicInfoReq, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.ArgumentTemplate,
		Action: meta.Delete, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	if err = svc.audit.ResDeleteAudit(cts.Kit, enumor.ArgumentTemplateAuditResType, basicInfoReq.IDs); err != nil {
		logs.Errorf("create operation audit argument template failed, ids: %v, err: %v, rid: %s",
			basicInfoReq.IDs, err, cts.Kit.Rid)
		return nil, err
	}

	// delete tcloud cloud argument template
	for _, id := range req.IDs {
		basicInfo, exists := basicInfoMap[id]
		if !exists {
			logs.Errorf("argument template record is not found, id: %s, rid: %s", id, cts.Kit.Rid)
			return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("id %s record is not found", id))
		}

		err = svc.client.HCService().TCloud.ArgsTpl.DeleteArgsTpl(cts.Kit, &protoargstpl.TCloudDeleteReq{
			AccountID: basicInfo.AccountID,
			ID:        id,
		})
		if err != nil {
			logs.Errorf("[%s] request hcservice to delete argument template failed, id: %s, err: %v, rid: %s",
				enumor.TCloud, id, err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}
