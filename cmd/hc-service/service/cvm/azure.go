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
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

func (svc *cvmSvc) initAzureCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("StartAzureCvm", http.MethodPost, "/vendors/azure/cvms/{id}/start", svc.StartAzureCvm)
	h.Add("StopAzureCvm", http.MethodPost, "/vendors/azure/cvms/{id}/stop", svc.StopAzureCvm)
	h.Add("RebootAzureCvm", http.MethodPost, "/vendors/azure/cvms/{id}/reboot", svc.RebootAzureCvm)
	h.Add("DeleteAzureCvm", http.MethodDelete, "/vendors/azure/cvms/{id}", svc.DeleteAzureCvm)

	h.Load(cap.WebService)
}

// StartAzureCvm ...
func (svc *cvmSvc) StartAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvm, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvm.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureStartOption{
		ResourceGroupName: cvm.Extension.ResourceGroupName,
		Name:              cvm.Name,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 操作完主机后需调用主机同步接口更新该操作相关数据。

	return nil, nil
}

// StopAzureCvm ...
func (svc *cvmSvc) StopAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(protocvm.AzureStopReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvm, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvm.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureStopOption{
		ResourceGroupName: cvm.Extension.ResourceGroupName,
		Name:              cvm.Name,
		SkipShutdown:      req.SkipShutdown,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 操作完主机后需调用主机同步接口更新该操作相关数据。

	return nil, nil
}

// RebootAzureCvm ...
func (svc *cvmSvc) RebootAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	cvm, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvm.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureRebootOption{
		ResourceGroupName: cvm.Extension.ResourceGroupName,
		Name:              cvm.Name,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	// TODO: 操作完主机后需调用主机同步接口更新该操作相关数据。

	return nil, nil
}

// DeleteAzureCvm ...
func (svc *cvmSvc) DeleteAzureCvm(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(protocvm.AzureDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvm, err := svc.dataCli.Azure.Cvm.GetCvm(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		logs.Errorf("request dataservice get tcloud security group failed, err: %v, id: %s, rid: %s", err, id,
			cts.Kit.Rid)
		return nil, err
	}

	client, err := svc.ad.Azure(cts.Kit, cvm.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.AzureDeleteOption{
		ResourceGroupName: cvm.Extension.ResourceGroupName,
		Name:              cvm.Name,
		Force:             req.Force,
	}
	if err = client.DeleteCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete azure cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.CvmBatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	if err = svc.dataCli.Global.Cvm.BatchDeleteCvm(cts.Kit.Ctx, cts.Kit.Header(), delReq); err != nil {
		logs.Errorf("request dataservice delete azure cvm failed, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
