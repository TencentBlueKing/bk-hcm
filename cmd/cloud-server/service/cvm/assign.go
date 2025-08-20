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
	"strings"

	"hcm/cmd/cloud-server/logics/cvm"
	proto "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// AssignCvmToBiz assign cvm to biz.
func (svc *cvmSvc) AssignCvmToBiz(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignCvmToBizReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	cvmIDs := make([]string, 0, len(req.Cvms))
	cvmIDBizIDMap := make(map[string]int64, len(req.Cvms))
	infos := make([]cvm.AssignedCvmInfo, 0, len(req.Cvms))
	for _, cvmInfo := range req.Cvms {
		cvmIDs = append(cvmIDs, cvmInfo.CvmID)
		cvmIDBizIDMap[cvmInfo.CvmID] = cvmInfo.BkBizID
		infos = append(infos, cvm.AssignedCvmInfo{
			CvmID:     cvmInfo.CvmID,
			BkBizID:   cvmInfo.BkBizID,
			BkCloudID: converter.PtrToVal(cvmInfo.BkCloudID),
		})
	}
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          cvmIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}

	authRes := make([]meta.ResourceAttribute, 0, len(basicInfoMap))
	for _, info := range basicInfoMap {
		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm,
			Action: meta.Assign, ResourceID: info.AccountID}, BizID: cvmIDBizIDMap[info.ID]})
	}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
	if err != nil {
		return nil, err
	}

	return nil, cvm.Assign(cts.Kit, svc.client.DataService(), infos)
}

// AssignCvmToBizPreview assign cvm to biz preview.
func (svc *cvmSvc) AssignCvmToBizPreview(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AssignCvmToBizPreviewReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          req.CvmIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		return nil, err
	}
	err = handler.ResOperateAuth(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer,
		ResType: meta.Cvm, Action: meta.Find, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	previewMap, err := cvm.AssignPreview(cts.Kit, svc.cmdbCli, svc.client, req.CvmIDs)
	if err != nil {
		logs.Errorf("cvm assign preview failed, err: %v, cvm ids: %v, rid: %s", err, req.CvmIDs, cts.Kit.Rid)
		return nil, err
	}

	details := make([]proto.AssignCvmToBizPreviewDetail, 0, len(req.CvmIDs))
	for _, cvmID := range req.CvmIDs {
		previewInfos, ok := previewMap[cvmID]
		if !ok || len(previewInfos) == 0 {
			details = append(details, proto.AssignCvmToBizPreviewDetail{CvmID: cvmID, MatchType: enumor.NoMatchCvm})
			continue
		}

		if len(previewInfos) == 1 {
			details = append(details, proto.AssignCvmToBizPreviewDetail{CvmID: cvmID, MatchType: enumor.AutoMatchCvm,
				BkCloudID: converter.ValToPtr(previewInfos[0].BkCloudID), BizID: previewInfos[0].BkBizID})
			continue
		}

		details = append(details, proto.AssignCvmToBizPreviewDetail{CvmID: cvmID, MatchType: enumor.ManualMatchCvm})
	}

	return proto.AssignCvmToBizPreviewData{Details: details}, nil
}

// ListAssignedCvmMatchHost list assigned cvm match host.
func (svc *cvmSvc) ListAssignedCvmMatchHost(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.ListAssignedCvmMatchHostReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 权限校验
	auth := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Cvm, Action: meta.Find, ResourceID: req.AccountID}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, auth); err != nil {
		return nil, err
	}

	cvmInfos := make([]cvm.PreviewAssignedCvmInfo, 0, len(req.PrivateIPv4Addresses))
	for _, innerIPv4 := range req.PrivateIPv4Addresses {
		cvmInfos = append(cvmInfos, cvm.PreviewAssignedCvmInfo{InnerIPv4: innerIPv4})
	}
	fields := []string{"bk_host_id", "bk_host_innerip", "bk_host_outerip", "bk_cloud_id", "bk_cloud_region",
		"bk_host_name", "bk_os_name", "create_time"}
	ccHosts, ccBizHostIDsMap, err := cvm.GetAssignedHostInfoFromCC(cts.Kit, svc.cmdbCli, cvmInfos, fields)
	if err != nil {
		logs.Errorf("get assign host from cc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	accountReq := &dataproto.AccountListReq{
		Filter: tools.EqualExpression("id", req.AccountID),
		Page:   &core.BasePage{Start: 0, Limit: 1},
	}
	resp, err := svc.client.DataService().Global.Account.List(cts.Kit.Ctx, cts.Kit.Header(), accountReq)
	if err != nil {
		logs.Errorf("list account failed, err: %v, account id: %s, rid: %s", err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}
	if resp == nil || len(resp.Details) == 0 {
		return nil, fmt.Errorf("not found account by id(%s)", req.AccountID)
	}

	details := make([]proto.ListAssignedCvmMatchHostDetail, 0)
	for _, bizID := range resp.Details[0].UsageBizIDs {
		hostIDs, ok := ccBizHostIDsMap[bizID]
		if !ok || len(hostIDs) == 0 {
			continue
		}

		for _, hostID := range hostIDs {
			ccHost, ok := ccHosts[hostID]
			if !ok {
				continue
			}

			detail := proto.ListAssignedCvmMatchHostDetail{
				BkHostID:             ccHost.BkHostID,
				PrivateIPv4Addresses: strings.Split(ccHost.BkHostInnerIP, ","),
				PublicIPv4Addresses:  strings.Split(ccHost.BkHostOuterIP, ","),
				BkCloudID:            ccHost.BkCloudID,
				BkBizID:              bizID,
				Region:               ccHost.BkCloudRegion,
				BkHostName:           ccHost.BkHostName,
				BkOsName:             ccHost.BkOSName,
				CreateTime:           ccHost.CreateTime,
			}
			details = append(details, detail)
		}
	}

	return &proto.ListAssignedCvmMatchHostData{Details: details}, nil
}
