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
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

func (svc *cvmSvc) initHuaWeiCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchStartHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/batch/start", svc.BatchStartHuaWeiCvm)
	h.Add("BatchStopHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/batch/stop", svc.BatchStopHuaWeiCvm)
	h.Add("BatchRebootHuaWeiCvm", http.MethodPost, "/vendors/huawei/cvms/batch/reboot", svc.BatchRebootHuaWeiCvm)
	h.Add("BatchDeleteHuaWeiCvm", http.MethodDelete, "/vendors/huawei/cvms/batch", svc.BatchDeleteHuaWeiCvm)
	h.Add("BatchResetHuaWeiCvmPwd", http.MethodPost, "/vendors/huawei/cvms/batch/reset/pwd", svc.BatchResetHuaWeiCvmPwd)

	h.Load(cap.WebService)
}

// BatchResetHuaWeiCvmPwd ...
func (svc *cvmSvc) BatchResetHuaWeiCvmPwd(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchResetPwdReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.DefaultBasePage,
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiResetPwdOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Password: req.Password,
	}
	if err = client.ResetCvmPwd(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reset huawei cvm pwd failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 操作完主机后需调用主机同步接口更新该操作相关数据。

	return nil, nil
}

// BatchStartHuaWeiCvm ...
func (svc *cvmSvc) BatchStartHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchStartReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.DefaultBasePage,
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiStartOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 操作完主机后需调用主机同步接口更新该操作相关数据。

	return nil, nil
}

// BatchStopHuaWeiCvm ...
func (svc *cvmSvc) BatchStopHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchStopReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.DefaultBasePage,
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiStopOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Force:    req.Force,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 操作完主机后需调用主机同步接口更新该操作相关数据。

	return nil, nil
}

// BatchRebootHuaWeiCvm ...
func (svc *cvmSvc) BatchRebootHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchRebootReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.DefaultBasePage,
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiRebootOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		Force:    req.Force,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 操作完主机后需调用主机同步接口更新该操作相关数据。

	return nil, nil
}

// BatchDeleteHuaWeiCvm ...
func (svc *cvmSvc) BatchDeleteHuaWeiCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.HuaWeiBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &dataproto.CvmListReq{
		Field:  []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.DefaultBasePage,
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list huawei cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	delCloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		delCloudIDs = append(delCloudIDs, one.CloudID)
	}

	client, err := svc.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.HuaWeiDeleteOption{
		Region:         req.Region,
		CloudIDs:       delCloudIDs,
		DeletePublicIP: req.DeletePublicIP,
		DeleteVolume:   req.DeleteDisk,
	}
	if err = client.DeleteCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete huawei cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("id", req.IDs),
	}
	if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete huawei cvm failed, err: %v, ids: %v, rid: %s", err,
			req.IDs, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 添加删除弹性IP 和 卷的逻辑

	return nil, nil
}
