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

	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor/tcloud"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/cvm"
	typecvm "hcm/pkg/adaptor/types/cvm"
	networkinterface "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

func (svc *cvmSvc) initTCloudCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/batch/create", svc.BatchCreateTCloudCvm)
	h.Add("InquiryPriceTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/prices/inquiry", svc.InquiryPriceTCloudCvm)
	h.Add("BatchStartTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/batch/start", svc.BatchStartTCloudCvm)
	h.Add("BatchStopTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/batch/stop", svc.BatchStopTCloudCvm)
	h.Add("BatchRebootTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/batch/reboot", svc.BatchRebootTCloudCvm)
	h.Add("BatchDeleteTCloudCvm", http.MethodDelete, "/vendors/tcloud/cvms/batch", svc.BatchDeleteTCloudCvm)
	h.Add("BatchResetTCloudCvmPwd", http.MethodPost, "/vendors/tcloud/cvms/batch/reset/pwd", svc.BatchResetTCloudCvmPwd)
	h.Add("BatchResetTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/reset", svc.BatchResetTCloudCvm)

	h.Add("ListTCloudCvmNetworkInterface", http.MethodPost, "/vendors/tcloud/cvms/network_interfaces/list",
		svc.ListTCloudCvmNetworkInterface)
	h.Add("BatchAssociateTCloudSecurityGroup", http.MethodPost, "/vendors/tcloud/cvms/security_groups/batch/associate",
		svc.BatchAssociateTCloudSecurityGroup)

	h.Load(cap.WebService)
}

// BatchAssociateTCloudSecurityGroup batch associate tcloud security group.
func (svc *cvmSvc) BatchAssociateTCloudSecurityGroup(cts *rest.Contexts) (interface{}, error) {

	req := new(protocvm.TCloudCvmBatchAssociateSecurityGroupReq)
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

	defer func() {
		err = svc.syncTCloudCvmWithRelRes(cts.Kit, tcloud, req.AccountID, req.Region, []string{cvmCloudID})
		if err != nil {
			logs.Errorf("sync tcloud cvm with rel res failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return
		}
	}()

	sgMap, err := svc.listSecurityGroupMap(cts.Kit, req.SecurityGroupIDs...)
	if err != nil {
		logs.Errorf("list security groups failed, err: %v, sgIDs: %v, rid: %s",
			err, req.SecurityGroupIDs, cts.Kit.Rid)
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

	opt := &typecvm.TCloudAssociateSecurityGroupsOption{
		Region:                req.Region,
		CloudSecurityGroupIDs: sgCloudIDs,
		CloudCvmID:            cvmCloudID,
	}

	err = tcloud.BatchCvmAssociateSecurityGroups(cts.Kit, opt)
	if err != nil {
		logs.Errorf("batch associate security group failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	err = svc.createSGCommonRels(cts.Kit, enumor.TCloud, enumor.CvmCloudResType, req.CvmID, req.SecurityGroupIDs)
	if err != nil {
		// 不抛出err, 尽最大努力交付
		logs.Errorf("create sg common rels failed, err: %v, cvmID: %s, sgIDs: %v, rid: %s",
			err, req.CvmID, req.SecurityGroupIDs, cts.Kit.Rid)
	}
	return nil, nil
}

func (svc *cvmSvc) syncTCloudCvmWithRelRes(kt *kit.Kit, tcloud tcloud.TCloud, accountID, region string,
	cloudIDs []string) error {

	syncClient := synctcloud.NewClient(svc.dataCli, tcloud)
	params := &synctcloud.SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  cloudIDs,
	}

	_, err := syncClient.CvmWithRelRes(kt, params, &synctcloud.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm with res failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// InquiryPriceTCloudCvm inquiry price tcloud cvm.
func (svc *cvmSvc) InquiryPriceTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchCreateReq)
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

	createOpt := &typecvm.TCloudCreateOption{
		DryRun:                  req.DryRun,
		Region:                  req.Region,
		Name:                    req.Name,
		Zone:                    req.Zone,
		InstanceType:            req.InstanceType,
		CloudImageID:            req.CloudImageID,
		Password:                req.Password,
		RequiredCount:           req.RequiredCount,
		CloudSecurityGroupIDs:   req.CloudSecurityGroupIDs,
		ClientToken:             req.ClientToken,
		CloudVpcID:              req.CloudVpcID,
		CloudSubnetID:           req.CloudSubnetID,
		InstanceChargeType:      req.InstanceChargeType,
		InstanceChargePrepaid:   req.InstanceChargePrepaid,
		SystemDisk:              req.SystemDisk,
		DataDisk:                req.DataDisk,
		PublicIPAssigned:        req.PublicIPAssigned,
		InternetMaxBandwidthOut: req.InternetMaxBandwidthOut,
		InternetChargeType:      req.InternetChargeType,
		BandwidthPackageID:      req.BandwidthPackageID,
	}
	result, err := tcloud.InquiryPriceCvm(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("inquiry cvm price failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// BatchCreateTCloudCvm ...
func (svc *cvmSvc) BatchCreateTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchCreateReq)
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

	createOpt := &typecvm.TCloudCreateOption{
		DryRun:                  req.DryRun,
		Region:                  req.Region,
		Name:                    req.Name,
		Zone:                    req.Zone,
		InstanceType:            req.InstanceType,
		CloudImageID:            req.CloudImageID,
		Password:                req.Password,
		RequiredCount:           req.RequiredCount,
		CloudSecurityGroupIDs:   req.CloudSecurityGroupIDs,
		ClientToken:             req.ClientToken,
		CloudVpcID:              req.CloudVpcID,
		CloudSubnetID:           req.CloudSubnetID,
		InstanceChargeType:      req.InstanceChargeType,
		InstanceChargePrepaid:   req.InstanceChargePrepaid,
		SystemDisk:              req.SystemDisk,
		DataDisk:                req.DataDisk,
		PublicIPAssigned:        req.PublicIPAssigned,
		InternetMaxBandwidthOut: req.InternetMaxBandwidthOut,
		InternetChargeType:      req.InternetChargeType,
		BandwidthPackageID:      req.BandwidthPackageID,
	}
	result, err := tcloud.CreateCvm(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("create cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
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

	err = svc.syncTCloudCvmWithRelRes(cts.Kit, tcloud, req.AccountID, req.Region, result.SuccessCloudIDs)
	if err != nil {
		logs.Errorf("sync tcloud cvm with rel res failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return respData, nil
}

// BatchResetTCloudCvmPwd ...
func (svc *cvmSvc) BatchResetTCloudCvmPwd(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchResetPwdReq)
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
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudResetPwdOption{
		Region:    req.Region,
		CloudIDs:  cloudIDs,
		UserName:  req.UserName,
		Password:  req.Password,
		ForceStop: req.ForceStop,
	}
	if err = client.ResetCvmPwd(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reset tcloud cvm pwd failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchStartTCloudCvm ...
func (svc *cvmSvc) BatchStartTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchStartReq)
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
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudStartOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchStopTCloudCvm ...
func (svc *cvmSvc) BatchStopTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchStopReq)
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
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudStopOption{
		Region:      req.Region,
		CloudIDs:    cloudIDs,
		StopType:    req.StopType,
		StoppedMode: req.StoppedMode,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchRebootTCloudCvm ...
func (svc *cvmSvc) BatchRebootTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchRebootReq)
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
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudRebootOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		StopType: req.StopType,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchDeleteTCloudCvm ...
func (svc *cvmSvc) BatchDeleteTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchDeleteReq)
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
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	delCloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		delCloudIDs = append(delCloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudDeleteOption{
		Region:   req.Region,
		CloudIDs: delCloudIDs,
	}
	if err = client.DeleteCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("id", req.IDs),
	}
	if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchResetTCloudCvm 重装系统
func (svc *cvmSvc) BatchResetTCloudCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchResetReq)
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

	for _, cloudID := range req.CloudIDs {
		opt := &cvm.ResetInstanceOption{
			Region:   req.Region,
			CloudID:  cloudID,
			ImageID:  req.ImageID,
			Password: req.Password,
		}
		if _, err = client.ResetCvmInstance(cts.Kit, opt); err != nil {
			logs.Errorf("request adaptor to tcloud reset cvm instance failed, err: %v, opt: %+v, cloudID: %s, "+
				"rid: %s", err, cvt.PtrToVal(req), cloudID, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}

// ListTCloudCvmNetworkInterface 返回一个map，key为cvmID，value为cvm的网卡信息 ListCvmNetworkInterfaceResp
func (svc *cvmSvc) ListTCloudCvmNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.ListCvmNetworkInterfaceReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmList, err := svc.getCvms(cts.Kit, enumor.TCloud, req.Region, req.CvmIDs)
	if err != nil {
		logs.Errorf("get cvms failed, err: %v, cvmIDs: %v, rid: %s", err, req.CvmIDs, cts.Kit.Rid)
		return nil, err
	}
	cloudIDToIDMap := make(map[string]string)
	for _, baseCvm := range cvmList {
		cloudIDToIDMap[baseCvm.CloudID] = baseCvm.ID
	}

	result, err := svc.listTCloudNetworkInterfaceFromCloud(cts.Kit, req.Region, req.AccountID, cloudIDToIDMap)
	if err != nil {
		logs.Errorf("list tcloud network interface from cloud failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func (svc *cvmSvc) listTCloudNetworkInterfaceFromCloud(kt *kit.Kit, region, accountID string,
	cloudIDToIDMap map[string]string) (map[string]*protocvm.ListCvmNetworkInterfaceRespItem, error) {

	cli, err := svc.ad.TCloud(kt, accountID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*protocvm.ListCvmNetworkInterfaceRespItem)
	var offset uint64 = 0
	for {
		opt := &networkinterface.TCloudNetworkInterfaceListOption{
			Region: region,
			Page: &adcore.TCloudPage{
				Offset: offset,
				Limit:  adcore.TCloudQueryLimit,
			},
			Filters: []*vpc.Filter{
				{
					Name:   common.StringPtr("attachment.instance-id"),
					Values: common.StringPtrs(cvt.MapKeyToSlice(cloudIDToIDMap)),
				},
			},
		}

		resp, err := cli.DescribeNetworkInterfaces(kt, opt)
		if err != nil {
			logs.Errorf("describe network interfaces failed, err: %v, cloudIDs: %v, rid: %s",
				err, cvt.MapKeyToSlice(cloudIDToIDMap), kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			cloudID := cvt.PtrToVal(detail.Attachment.InstanceId)
			id := cloudIDToIDMap[cloudID]
			if _, ok := result[id]; !ok {
				result[id] = &protocvm.ListCvmNetworkInterfaceRespItem{
					MacAddressToPrivateIpAddresses: make(map[string][]string),
				}
			}

			privateIPs := make([]string, 0)
			for _, set := range detail.PrivateIpAddressSet {
				privateIPs = append(privateIPs, cvt.PtrToVal(set.PrivateIpAddress))
			}
			result[id].MacAddressToPrivateIpAddresses[cvt.PtrToVal(detail.MacAddress)] = privateIPs

		}
		if len(resp.Details) < adcore.TCloudQueryLimit {
			break
		}
		offset += adcore.TCloudQueryLimit
	}
	return result, nil
}
