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

	proto "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
)

// ListCvm list cvm.
func (svc *cvmSvc) ListCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listCvm(cts, handler.ListResourceAuthRes)
}

// ListBizCvm list biz cvm.
func (svc *cvmSvc) ListBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.listCvm(cts, handler.ListBizAuthRes)
}

func (svc *cvmSvc) listCvm(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{}, error) {
	req := new(proto.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Cvm, Action: meta.Find, Filter: req.Filter})
	if err != nil {
		return nil, err
	}

	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}
	req.Filter = expr

	listReq := &dataproto.CvmListReq{
		Filter: req.Filter,
		Page:   req.Page,
	}
	return svc.client.DataService().Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
}

// GetCvm get cvm.
func (svc *cvmSvc) GetCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.getCvm(cts, handler.ResValidWithAuth)
}

// GetBizCvm get biz cvm.
func (svc *cvmSvc) GetBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.getCvm(cts, handler.BizValidWithAuth)
}

func (svc *cvmSvc) getCvm(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.CvmCloudResType, id)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	switch basicInfo.Vendor {
	case enumor.TCloud:
		return svc.client.DataService().TCloud.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Aws:
		return svc.client.DataService().Aws.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Gcp:
		return svc.client.DataService().Gcp.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.HuaWei:
		return svc.client.DataService().HuaWei.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	case enumor.Azure:
		return svc.client.DataService().Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)

	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", id, basicInfo.Vendor)
	}
}

// CheckCvmsInBiz check if cvms are in the specified biz.
func CheckCvmsInBiz(kt *kit.Kit, client *client.ClientSet, rule filter.RuleFactory, bizID int64) error {
	req := &dataproto.CvmListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "bk_biz_id", Op: filter.NotEqual.Factory(), Value: bizID}, rule,
			},
		},
		Page: &core.BasePage{
			Count: true,
		},
	}
	result, err := client.DataService().Global.Cvm.ListCvm(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("count cvms that are not in biz failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
		return err
	}

	if result.Count != 0 {
		return fmt.Errorf("%d cvms are already assigned", result.Count)
	}

	return nil
}
