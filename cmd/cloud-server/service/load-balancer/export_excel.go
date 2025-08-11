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

package loadbalancer

import (
	"fmt"
	"path/filepath"

	lblogic "hcm/cmd/cloud-server/logics/load-balancer"
	cslb "hcm/pkg/api/cloud-server/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// PreCheckExportBizListener 导出业务下监听器及其下面的资源预检
func (svc *lbSvc) PreCheckExportBizListener(cts *rest.Contexts) (interface{}, error) {
	req := new(cslb.ExportListenerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authExportBizListener(cts, req); err != nil {
		return nil, err
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	exporter, err := lblogic.NewListenerExporter(svc.client, vendor, req)
	if err != nil {
		return nil, err
	}
	if err = exporter.PreCheck(cts.Kit); err != nil {
		return &cslb.ExportListenerResp{Pass: false, Reason: err.Error()}, nil
	}

	return &cslb.ExportListenerResp{Pass: true}, nil
}

func (svc *lbSvc) authExportBizListener(cts *rest.Contexts, req *cslb.ExportListenerReq) error {
	lbIDs := make([]string, 0, len(req.Listeners))
	for _, l := range req.Listeners {
		lbIDs = append(lbIDs, l.LbID)
	}
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.LoadBalancerCloudResType,
		IDs:          lbIDs,
	}
	lbInfo, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return err
	}
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer,
		ResType: meta.LoadBalancer, Action: meta.Update, BasicInfos: lbInfo})
	if err != nil {
		return err
	}

	return nil
}

// ExportBizListener 导出业务下监听器及其下面的资源
func (svc *lbSvc) ExportBizListener(cts *rest.Contexts) (interface{}, error) {
	req := new(cslb.ExportListenerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if err := svc.authExportBizListener(cts, req); err != nil {
		return nil, err
	}

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	exporter, err := lblogic.NewListenerExporter(svc.client, vendor, req)
	if err != nil {
		return nil, err
	}
	if err = exporter.PreCheck(cts.Kit); err != nil {
		return nil, err
	}
	filePath, err := exporter.Export(cts.Kit)
	if err != nil {
		return nil, err
	}

	return &rest.FileResp{
		ContentTypeStr:        "application/octet-stream",
		ContentDispositionStr: fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(filePath)),
		FilePath:              filePath,
	}, nil
}
