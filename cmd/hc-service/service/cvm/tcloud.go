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
	adcore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	tcvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
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
	h.Add("QueryTCloudCvm", http.MethodPost, "/vendors/tcloud/cvms/query", svc.BatchQueryTCloudCVM)

	h.Add("ListTCloudCvmNetworkInterface", http.MethodPost, "/vendors/tcloud/cvms/network_interfaces/list",
		svc.ListTCloudCvmNetworkInterface)
	h.Add("BatchAssociateTCloudSecurityGroup", http.MethodPost, "/vendors/tcloud/cvms/security_groups/batch/associate",
		svc.BatchAssociateTCloudSecurityGroup)

	h.Load(cap.WebService)
}

// BatchQueryTCloudCVM 到云上查询cvm
func (svc *cvmSvc) BatchQueryTCloudCVM(cts *rest.Contexts) (any, error) {
	req := new(corecvm.QueryCloudCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloudCli, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	cvmWithCount, err := tcloudCli.ListCvmWithCount(cts.Kit, &typecvm.ListCvmWithCountOption{
		Region:   req.Region,
		CloudIDs: req.CvmIDs,
		SGIDs:    req.SGIDs,
		Page: &adcore.TCloudPage{
			Offset: uint64(req.Page.Start),
			Limit:  uint64(req.Page.Limit),
		},
	})
	if err != nil {
		logs.Errorf("fail to list cvm with count, err: %v, req: %+v ,rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	cmvList := slice.Map(cvmWithCount.Cvms, func(c typecvm.TCloudCvm) corecvm.Cvm[corecvm.TCloudCvmExtension] {
		return convTCloudCvm(c, req.AccountID, req.Region)
	})
	return types.ListResult[corecvm.Cvm[corecvm.TCloudCvmExtension]]{Count: uint64(cvmWithCount.TotalCount),
		Details: cmvList}, nil
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

	// create common rels in db
	err = svc.createSGCommonRels(cts.Kit, enumor.TCloud, enumor.CvmCloudResType, req.CvmID, req.SecurityGroupIDs)
	if err != nil {
		// 不抛出err, 尽最大努力交付
		logs.Errorf("create sg common rels failed, err: %v, cvmID: %s, sgIDs: %v, rid: %s",
			err, req.CvmID, req.SecurityGroupIDs, cts.Kit.Rid)
	}
	return nil, nil
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

	// sync cvm
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

	// sync cvm
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

	// delete cvm in db
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

func convTCloudCvm(c typecvm.TCloudCvm, accountID, region string) corecvm.Cvm[corecvm.TCloudCvmExtension] {
	var cloudVpcIDs, cloudSubnetIDs []string
	if c.VirtualPrivateCloud != nil {
		cloudVpcIDs = append(cloudVpcIDs, cvt.PtrToVal(c.VirtualPrivateCloud.VpcId))
		cloudSubnetIDs = append(cloudSubnetIDs, cvt.PtrToVal(c.VirtualPrivateCloud.SubnetId))
	}
	baseCvm := corecvm.BaseCvm{
		CloudID:              c.GetCloudID(),
		Name:                 cvt.PtrToVal(c.InstanceName),
		Vendor:               enumor.TCloud,
		AccountID:            accountID,
		Region:               region,
		Zone:                 cvt.PtrToVal(c.Placement.Zone),
		CloudVpcIDs:          cloudVpcIDs,
		CloudSubnetIDs:       cloudSubnetIDs,
		OsName:               cvt.PtrToVal(c.OsName),
		Status:               cvt.PtrToVal(c.InstanceState),
		PrivateIPv4Addresses: cvt.PtrToSlice(c.PrivateIpAddresses),
		PublicIPv4Addresses:  cvt.PtrToSlice(c.PublicIpAddresses),
		PrivateIPv6Addresses: cvt.PtrToSlice(c.IPv6Addresses),
		MachineType:          cvt.PtrToVal(c.InstanceType),
		CloudImageID:         cvt.PtrToVal(c.ImageId),
		CloudCreatedTime:     cvt.PtrToVal(c.CreatedTime),
	}
	var internetAccessible *corecvm.TCloudInternetAccessible
	if c.InternetAccessible != nil {
		internetAccessible = &corecvm.TCloudInternetAccessible{
			InternetChargeType:      c.InternetAccessible.InternetChargeType,
			InternetMaxBandwidthOut: c.InternetAccessible.InternetMaxBandwidthOut,
			PublicIPAssigned:        c.InternetAccessible.PublicIpAssigned,
			CloudBandwidthPackageID: c.InternetAccessible.BandwidthPackageId,
		}
	}

	return corecvm.Cvm[corecvm.TCloudCvmExtension]{
		BaseCvm: baseCvm,
		Extension: &corecvm.TCloudCvmExtension{
			CloudSecurityGroupIDs: cvt.PtrToSlice(c.SecurityGroupIds),
			Placement:             &corecvm.TCloudPlacement{CloudProjectID: cvt.PtrToVal(c.Placement).ProjectId},
			InstanceChargeType:    c.InstanceChargeType,
			Cpu:                   c.CPU,
			Memory:                c.Memory,
			CloudSystemDiskID:     cvt.PtrToVal(c.SystemDisk).DiskId,
			CloudDataDiskIDs: slice.Map(c.DataDisks,
				func(dd *tcvm.DataDisk) string { return cvt.PtrToVal(dd.DiskId) }),
			InternetAccessible: internetAccessible,

			VirtualPrivateCloud: &corecvm.TCloudVirtualPrivateCloud{
				AsVpcGateway: c.VirtualPrivateCloud.AsVpcGateway,
			},
			RenewFlag:             c.RenewFlag,
			StopChargingMode:      c.StopChargingMode,
			UUID:                  c.Uuid,
			IsolatedSource:        c.IsolatedSource,
			DisableApiTermination: c.DisableApiTermination,
		},
	}
}
