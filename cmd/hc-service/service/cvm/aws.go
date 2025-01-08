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
	"net/http"

	syncaws "hcm/cmd/hc-service/logics/res-sync/aws"
	"hcm/cmd/hc-service/service/capability"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

func (svc *cvmSvc) initAwsCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateAwsCvm", http.MethodPost, "/vendors/aws/cvms/batch/create", svc.BatchCreateAwsCvm)
	h.Add("BatchStartAwsCvm", http.MethodPost, "/vendors/aws/cvms/batch/start", svc.BatchStartAwsCvm)
	h.Add("BatchStopAwsCvm", http.MethodPost, "/vendors/aws/cvms/batch/stop", svc.BatchStopAwsCvm)
	h.Add("BatchRebootAwsCvm", http.MethodPost, "/vendors/aws/cvms/batch/reboot", svc.BatchRebootAwsCvm)
	h.Add("BatchDeleteAwsCvm", http.MethodDelete, "/vendors/aws/cvms/batch", svc.BatchDeleteAwsCvm)

	h.Add("BatchAssociateAwsSecurityGroup", http.MethodPost, "/vendors/aws/cvms/security_groups/batch/associate",
		svc.BatchAssociateAwsSecurityGroup)

	h.Load(cap.WebService)
}

// BatchAssociateAwsSecurityGroup batch associate aws security group.
func (svc *cvmSvc) BatchAssociateAwsSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req := new(protocvm.AwsCvmBatchAssociateSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	awsCli, err := svc.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	cvmList, err := svc.listCvms(cts.Kit, req.CvmID)
	if err != nil {
		logs.Errorf("get cvms failed, err: %v, cvmID: %s, rid: %s", err, req.CvmID, cts.Kit.Rid)
		return nil, err
	}
	if len(cvmList) == 0 {
		logs.Errorf("cvm not found, cvmID: %s, rid: %s", req.CvmID, cts.Kit.Rid)
		return nil, fmt.Errorf("cvm (%s) not found", req.CvmID)
	}
	cvmCloudID := cvmList[0].CloudID

	sgMap, err := svc.listSecurityGroupMap(cts.Kit, req.SecurityGroupIDs...)
	if err != nil {
		logs.Errorf("list security groups failed, err: %v, sgIDs: %v, rid: %s", err, req.SecurityGroupIDs, cts.Kit.Rid)
		return nil, err
	}
	sgCloudIDs := make([]string, 0, len(req.SecurityGroupIDs))
	for _, id := range req.SecurityGroupIDs {
		sg, ok := sgMap[id]
		if !ok {
			logs.Errorf("security group not found, sgID: %s, rid: %s", id, cts.Kit.Rid)
			return nil, fmt.Errorf("security group (%s) not found", id)
		}
		sgCloudIDs = append(sgCloudIDs, sg.CloudID)
	}

	opt := &typecvm.AwsAssociateSecurityGroupsOption{
		Region:                req.Region,
		CloudSecurityGroupIDs: sgCloudIDs,
		CloudCvmID:            cvmCloudID,
	}
	err = awsCli.BatchAssociateSecurityGroup(cts.Kit, opt)
	if err != nil {
		logs.Errorf("batch associate security group failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	createReq := &protocloud.SGCvmRelBatchCreateReq{}
	for _, sgID := range req.SecurityGroupIDs {
		createReq.Rels = append(createReq.Rels, protocloud.SGCvmRelCreate{
			SecurityGroupID: sgID,
			CvmID:           req.CvmID,
		})
	}
	if err = svc.dataCli.Global.SGCvmRel.BatchCreateSgCvmRels(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
		logs.Errorf("request dataservice create security group cvm rels failed, err: %v, req: %+v, rid: %s",
			err, createReq, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// BatchCreateAwsCvm ...
func (svc *cvmSvc) BatchCreateAwsCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.AwsBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	awsCli, err := svc.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	createOpt := &typecvm.AwsCreateOption{
		DryRun:                req.DryRun,
		Region:                req.Region,
		Name:                  req.Name,
		Zone:                  req.Zone,
		InstanceType:          req.InstanceType,
		CloudImageID:          req.CloudImageID,
		Password:              req.Password,
		RequiredCount:         req.RequiredCount,
		CloudSecurityGroupIDs: req.CloudSecurityGroupIDs,
		ClientToken:           req.ClientToken,
		CloudSubnetID:         req.CloudSubnetID,
		BlockDeviceMapping:    req.BlockDeviceMapping,
		PublicIPAssigned:      req.PublicIPAssigned,
	}
	result, err := awsCli.CreateCvm(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("create aws cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	respData := &protocvm.BatchCreateResult{
		UnknownCloudIDs: result.UnknownCloudIDs,
		SuccessCloudIDs: result.SuccessCloudIDs,
		FailedCloudIDs:  result.FailedCloudIDs,
		FailedMessage:   result.FailedMessage,
	}

	if len(result.SuccessCloudIDs) == 0 {
		return respData, nil
	}

	syncClient := syncaws.NewClient(svc.dataCli, awsCli)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  result.SuccessCloudIDs,
	}

	_, err = syncClient.CvmWithRelRes(cts.Kit, params, &syncaws.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync aws cvm with res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return respData, nil
}

// BatchStartAwsCvm ...
func (svc *cvmSvc) BatchStartAwsCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.AwsBatchStartReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list aws cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AwsStartOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start aws cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncaws.NewClient(svc.dataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncaws.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync aws cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchStopAwsCvm ...
func (svc *cvmSvc) BatchStopAwsCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.AwsBatchStopReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list aws cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AwsStopOption{
		Region:    req.Region,
		CloudIDs:  cloudIDs,
		Force:     req.Force,
		Hibernate: req.Hibernate,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop aws cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncaws.NewClient(svc.dataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncaws.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync aws cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchRebootAwsCvm ...
func (svc *cvmSvc) BatchRebootAwsCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.AwsBatchRebootReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list aws cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AwsRebootOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot aws cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := syncaws.NewClient(svc.dataCli, client)

	params := &syncaws.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &syncaws.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync aws cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteAwsCvm ...
func (svc *cvmSvc) BatchDeleteAwsCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.AwsBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list aws cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	delCloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		delCloudIDs = append(delCloudIDs, one.CloudID)
	}

	client, err := svc.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AwsDeleteOption{
		Region:   req.Region,
		CloudIDs: delCloudIDs,
	}
	if err = client.DeleteCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete aws cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("id", req.IDs),
	}
	if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete aws cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
