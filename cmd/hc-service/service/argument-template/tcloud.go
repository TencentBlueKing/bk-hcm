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
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor/tcloud"
	typeargstpl "hcm/pkg/adaptor/types/argument-template"
	"hcm/pkg/api/core"
	argstpl "hcm/pkg/api/core/cloud/argument-template"
	dataproto "hcm/pkg/api/data-service/cloud"
	protoargstpl "hcm/pkg/api/hc-service/argument-template"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// initTCloudArgsTplService initializes the tcloud argument template service.
func (svc *argsTplSvc) initTCloudArgsTplService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("CreateTCloudAddress", http.MethodPost, "/vendors/tcloud/argument_templates/create",
		svc.CreateTCloudArgsTpl)
	h.Add("UpdateTCloudArgsTpl", http.MethodPut, "/vendors/tcloud/argument_templates/{id}", svc.UpdateTCloudArgsTpl)
	h.Add("DeleteTCloudArgsTpl", http.MethodDelete, "/vendors/tcloud/argument_templates", svc.DeleteTCloudArgsTpl)
	h.Add("ListTCloudArgsTpl", http.MethodPost, "/vendors/tcloud/argument_templates/list", svc.ListTCloudArgsTpl)

	h.Load(cap.WebService)
}

// CreateTCloudArgsTpl ...
func (svc *argsTplSvc) CreateTCloudArgsTpl(cts *rest.Contexts) (interface{}, error) {
	req := new(protoargstpl.TCloudCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	var templatesJson types.JsonField
	var groupTemplatesJson types.JsonField
	if len(req.Templates) > 0 {
		templatesJson, err = types.NewJsonField(req.Templates)
		if err != nil {
			return "", fmt.Errorf("json marshal templates failed, err: %v", err)
		}
	}

	if len(req.GroupTemplates) > 0 {
		groupTemplatesJson, err = types.NewJsonField(req.GroupTemplates)
		if err != nil {
			return "", fmt.Errorf("json marshal group templates failed, err: %v", err)
		}
	}

	cloudID, err := svc.createTCloudCloud(cts.Kit, client, req)
	if err != nil {
		return nil, err
	}

	var createReq = new(dataproto.ArgsTplBatchCreateReq[argstpl.TCloudArgsTplExtension])
	createReq.ArgumentTemplates = []dataproto.ArgsTplBatchCreate[argstpl.TCloudArgsTplExtension]{
		{
			CloudID:        cloudID,
			Name:           req.Name,
			Vendor:         string(enumor.TCloud),
			AccountID:      req.AccountID,
			BkBizID:        req.BkBizID,
			Type:           req.Type,
			Templates:      templatesJson,
			GroupTemplates: groupTemplatesJson,
		},
	}

	newData, err := svc.dataCli.TCloud.BatchCreateArgsTpl(cts.Kit, createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud argument template failed, createReq: %+v, "+
			"err: %v, rid: %s", enumor.TCloud, createReq, err, cts.Kit.Rid)
		return nil, err
	}

	if len(newData.IDs) == 0 {
		return nil, errf.Newf(errf.Aborted, "create tcloud argument template failed")
	}

	return &argstpl.ArgsTplCreateResult{ID: newData.IDs[0]}, nil
}

// createTCloudCloud creates the tcloud argument template in the cloud.
func (svc *argsTplSvc) createTCloudCloud(kt *kit.Kit, client tcloud.TCloud, req *protoargstpl.TCloudCreateReq) (
	string, error) {

	switch req.Type {
	case enumor.AddressType:
		return svc.createTCloudAddressTpl(kt, client, req)
	case enumor.AddressGroupType:
		return svc.createTCloudAddressGroupTpl(kt, client, req)
	case enumor.ServiceType:
		return svc.createTCloudServiceTpl(kt, client, req)
	case enumor.ServiceGroupType:
		return svc.createTCloudServiceGroupTpl(kt, client, req)
	default:
		return "", fmt.Errorf("unsupported template type: %s", req.Type)
	}
}

// createTCloudAddressTpl creates the tcloud argument template address in the cloud.
func (svc *argsTplSvc) createTCloudAddressTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudCreateReq) (string, error) {

	opt := &typeargstpl.TCloudCreateAddressOption{
		TemplateName: req.Name,
	}
	for _, addr := range req.Templates {
		opt.AddressesExtra = append(opt.AddressesExtra, &vpc.AddressInfo{
			Address:     addr.Address,
			Description: addr.Description,
		})
	}

	resp, err := client.CreateArgsTplAddress(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to create tcloud argument template address failed, opt: %v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return "", err
	}

	return converter.PtrToVal(resp.AddressTemplateId), nil
}

// createTCloudAddressGroupTpl creates the tcloud argument template address group in the cloud.
func (svc *argsTplSvc) createTCloudAddressGroupTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudCreateReq) (string, error) {

	opt := &typeargstpl.TCloudCreateAddressGroupOption{
		TemplateGroupName: req.Name,
		TemplateIDs:       req.GroupTemplates,
	}
	resp, err := client.CreateArgsTplAddressGroup(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to create tcloud argument template address group failed, opt: %v, "+
			"err: %v, rid: %s", opt, err, kt.Rid)
		return "", err
	}
	return converter.PtrToVal(resp.AddressTemplateGroupId), nil
}

// createTCloudServiceTpl creates the tcloud argument template service in the cloud.
func (svc *argsTplSvc) createTCloudServiceTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudCreateReq) (string, error) {

	opt := &typeargstpl.TCloudCreateServiceOption{
		TemplateName: req.Name,
	}
	for _, addr := range req.Templates {
		opt.ServicesExtra = append(opt.ServicesExtra, &vpc.ServicesInfo{
			Service:     addr.Address,
			Description: addr.Description,
		})
	}
	resp, cErr := client.CreateArgsTplService(kt, opt)
	if cErr != nil {
		logs.Errorf("request adaptor to create tcloud argument template service failed, opt: %v, err: %v, rid: %s",
			opt, cErr, kt.Rid)
		return "", cErr
	}
	return converter.PtrToVal(resp.ServiceTemplateId), nil
}

// createTCloudServiceGroupTpl creates the tcloud argument template service group in the cloud.
func (svc *argsTplSvc) createTCloudServiceGroupTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudCreateReq) (string, error) {

	opt := &typeargstpl.TCloudCreateServiceGroupOption{
		TemplateGroupName: req.Name,
		TemplateIDs:       req.GroupTemplates,
	}
	resp, cErr := client.CreateArgsTplServiceGroup(kt, opt)
	if cErr != nil {
		logs.Errorf("request adaptor to create tcloud argument template service group failed, opt: %v, "+
			"err: %v, rid: %s", opt, cErr, kt.Rid)
		return "", cErr
	}
	return converter.PtrToVal(resp.ServiceTemplateGroupId), nil
}

// UpdateTCloudArgsTpl ...
func (svc *argsTplSvc) UpdateTCloudArgsTpl(cts *rest.Contexts) (interface{}, error) {
	req := new(protoargstpl.TCloudUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	id := cts.PathParameter("id").String()
	listOpt := &core.ListReq{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	listRes, err := svc.dataCli.Global.ArgsTpl.ListArgsTpl(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	if len(listRes.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "tcloud argument template [%s] not found", id)
	}

	accountID := listRes.Details[0].AccountID
	cloudTemplateID := listRes.Details[0].CloudID
	templateType := listRes.Details[0].Type

	client, err := svc.ad.TCloud(cts.Kit, accountID)
	if err != nil {
		return nil, err
	}

	var templatesJson types.JsonField
	var groupTemplatesJson types.JsonField
	if len(req.Templates) > 0 {
		templatesJson, err = types.NewJsonField(req.Templates)
		if err != nil {
			return "", fmt.Errorf("json marshal templates failed, err: %v", err)
		}
	}

	if len(req.GroupTemplates) > 0 {
		groupTemplatesJson, err = types.NewJsonField(req.GroupTemplates)
		if err != nil {
			return "", fmt.Errorf("json marshal group templates failed, err: %v", err)
		}
	}

	err = svc.updateTCloudCloud(cts.Kit, client, req, templateType, cloudTemplateID)
	if err != nil {
		return nil, err
	}

	updateReq := &dataproto.ArgsTplBatchUpdateExprReq{
		IDs:            []string{id},
		BkBizID:        req.BkBizID,
		Name:           req.Name,
		Templates:      templatesJson,
		GroupTemplates: groupTemplatesJson,
	}
	if _, err = svc.dataCli.Global.ArgsTpl.BatchUpdateArgsTpl(cts.Kit, updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to update tcloud argument template failed, updateReq: %+v, "+
			"err: %v, rid: %s", enumor.TCloud, updateReq, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// updateTCloudCloud updates the tcloud argument template in the cloud.
func (svc *argsTplSvc) updateTCloudCloud(kt *kit.Kit, client tcloud.TCloud, req *protoargstpl.TCloudUpdateReq,
	templateType enumor.TemplateType, cloudTemplateID string) error {

	switch templateType {
	case enumor.AddressType:
		return svc.updateTCloudAddressTpl(kt, client, req, cloudTemplateID)
	case enumor.AddressGroupType:
		return svc.updateTCloudAddressGroupTpl(kt, client, req, cloudTemplateID)
	case enumor.ServiceType:
		return svc.updateTCloudServiceTpl(kt, client, req, cloudTemplateID)
	case enumor.ServiceGroupType:
		return svc.updateTCloudServiceGroupTpl(kt, client, req, cloudTemplateID)
	default:
		return fmt.Errorf("unsupported template type: %s", templateType)
	}
}

func (svc *argsTplSvc) updateTCloudAddressGroupTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudUpdateReq, cloudTemplateID string) error {

	opt := &typeargstpl.TCloudUpdateAddressGroupOption{
		TemplateGroupID:   cloudTemplateID,
		TemplateGroupName: req.Name,
		TemplateIDs:       req.GroupTemplates,
	}
	resp, err := client.UpdateArgsTplAddressGroup(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to update tcloud argument template address group failed, opt: %v, "+
			"err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	if len(resp.FailedCloudIDs) > 0 {
		return errf.Newf(errf.Aborted, "update tcloud argument template address group failed, "+
			"failedCloudIDs: %v", resp.FailedCloudIDs)
	}
	return nil
}

// updateTCloudAddressTpl updates the tcloud argument template address in the cloud.
func (svc *argsTplSvc) updateTCloudAddressTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudUpdateReq, cloudTemplateID string) error {

	opt := &typeargstpl.TCloudUpdateAddressOption{
		TemplateID:   cloudTemplateID,
		TemplateName: req.Name,
	}
	for _, addr := range req.Templates {
		opt.AddressesExtra = append(opt.AddressesExtra, &vpc.AddressInfo{
			Address:     addr.Address,
			Description: addr.Description,
		})
	}

	resp, err := client.UpdateArgsTplAddress(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to update tcloud argument template address failed, opt: %v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return err
	}

	if len(resp.FailedCloudIDs) > 0 {
		return errf.Newf(errf.Aborted, "update tcloud argument template address failed, failedCloudIDs: %v",
			resp.FailedCloudIDs)
	}
	return nil
}

// updateTCloudServiceTpl updates the tcloud argument template service in the cloud.
func (svc *argsTplSvc) updateTCloudServiceTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudUpdateReq, cloudTemplateID string) error {

	opt := &typeargstpl.TCloudUpdateServiceOption{
		TemplateID:   cloudTemplateID,
		TemplateName: req.Name,
	}
	for _, addr := range req.Templates {
		opt.ServicesExtra = append(opt.ServicesExtra, &vpc.ServicesInfo{
			Service:     addr.Address,
			Description: addr.Description,
		})
	}
	resp, err := client.UpdateArgsTplService(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to create tcloud argument template service failed, opt: %v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return err
	}

	if len(resp.FailedCloudIDs) > 0 {
		return errf.Newf(errf.Aborted, "update tcloud argument template service failed, failedCloudIDs: %v",
			resp.FailedCloudIDs)
	}

	return nil
}

// updateTCloudServiceGroupTpl updates the tcloud argument template service group in the cloud.
func (svc *argsTplSvc) updateTCloudServiceGroupTpl(kt *kit.Kit, client tcloud.TCloud,
	req *protoargstpl.TCloudUpdateReq, cloudTemplateID string) error {

	opt := &typeargstpl.TCloudUpdateServiceGroupOption{
		TemplateGroupID:   cloudTemplateID,
		TemplateGroupName: req.Name,
		TemplateIDs:       req.GroupTemplates,
	}
	resp, err := client.UpdateArgsTplServiceGroup(kt, opt)
	if err != nil {
		logs.Errorf("request adaptor to update tcloud argument template service group failed, opt: %v, "+
			"err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	if len(resp.FailedCloudIDs) > 0 {
		return errf.Newf(errf.Aborted, "update tcloud argument template service group failed, "+
			"failedCloudIDs: %v", resp.FailedCloudIDs)
	}
	return nil
}

// DeleteTCloudArgsTpl ...
func (svc *argsTplSvc) DeleteTCloudArgsTpl(cts *rest.Contexts) (interface{}, error) {
	req := new(protoargstpl.TCloudDeleteReq)
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
	listResp, err := svc.dataCli.Global.ArgsTpl.ListArgsTpl(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud argument template address failed, id: %s, err: %v, rid: %s",
			req.ID, err, cts.Kit.Rid)
		return nil, err
	}

	if len(listResp.Details) == 0 {
		logs.Errorf("request dataservice list tcloud argument template empty, id: %s, rid: %s", req.ID, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("get adaptor to tcloud client failed, accID: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typeargstpl.TCloudDeleteOption{
		CloudID: listResp.Details[0].CloudID,
	}

	switch listResp.Details[0].Type {
	case enumor.AddressType:
		if err = client.DeleteArgsTplAddress(cts.Kit, opt); err != nil {
			logs.Errorf("request adaptor to delete tcloud argument template address failed, opt: %v, err: %v, rid: %s",
				opt, err, cts.Kit.Rid)
			return nil, err
		}
	case enumor.AddressGroupType:
		if err = client.DeleteArgsTplAddressGroup(cts.Kit, opt); err != nil {
			logs.Errorf("request adaptor to delete tcloud argument template address group failed, opt: %v, "+
				"err: %v, rid: %s", opt, err, cts.Kit.Rid)
			return nil, err
		}
	case enumor.ServiceType:
		if err = client.DeleteArgsTplService(cts.Kit, opt); err != nil {
			logs.Errorf("request adaptor to delete tcloud argument template service failed, opt: %v, err: %v, rid: %s",
				opt, err, cts.Kit.Rid)
			return nil, err
		}
	case enumor.ServiceGroupType:
		if err = client.DeleteArgsTplServiceGroup(cts.Kit, opt); err != nil {
			logs.Errorf("request adaptor to delete tcloud argument template service group failed, opt: %v, "+
				"err: %v, rid: %s", opt, err, cts.Kit.Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported template type: %s", listResp.Details[0].Type)
	}

	delReq := &dataproto.ArgsTplBatchDeleteReq{
		Filter: tools.EqualExpression("id", req.ID),
	}
	if err = svc.dataCli.Global.ArgsTpl.BatchDeleteArgsTpl(cts.Kit, delReq); err != nil {
		logs.Errorf("request dataservice delete tcloud argument template failed, id: %s, err: %v, rid: %s",
			req.ID, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListTCloudArgsTpl list tcloud argument template
func (svc *argsTplSvc) ListTCloudArgsTpl(cts *rest.Contexts) (interface{}, error) {
	req := new(protoargstpl.ArgsTplListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typeargstpl.TCloudListOption{}
	if len(req.Filters) > 0 {
		opt.Filters = req.Filters
	}

	if req.Page != nil {
		opt.Page = req.Page
	}

	switch req.Type {
	case enumor.AddressType:
		resp, _, lErr := client.ListArgsTplAddress(cts.Kit, opt)
		if lErr != nil {
			logs.Errorf("request adaptor to list tcloud argument template address failed, opt: %v, err: %v, rid: %s",
				opt, lErr, cts.Kit.Rid)
			return nil, lErr
		}

		return resp, nil
	case enumor.AddressGroupType:
		resp, _, lErr := client.ListArgsTplAddressGroup(cts.Kit, opt)
		if lErr != nil {
			logs.Errorf("request adaptor to list tcloud argument template address group failed, opt: %v, "+
				"err: %v, rid: %s", opt, lErr, cts.Kit.Rid)
			return nil, lErr
		}

		return resp, nil
	case enumor.ServiceType:
		resp, _, lErr := client.ListArgsTplService(cts.Kit, opt)
		if lErr != nil {
			logs.Errorf("request adaptor to list tcloud argument template service failed, opt: %v, err: %v, rid: %s",
				opt, lErr, cts.Kit.Rid)
			return nil, lErr
		}

		return resp, nil
	case enumor.ServiceGroupType:
		resp, _, lErr := client.ListArgsTplServiceGroup(cts.Kit, opt)
		if lErr != nil {
			logs.Errorf("request adaptor to list tcloud argument template service group failed, opt: %v, "+
				"err: %v, rid: %s", opt, lErr, cts.Kit.Rid)
			return nil, lErr
		}

		return resp, nil
	default:
		return nil, fmt.Errorf("unsupported template type: %s", req.Type)
	}
}
