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

// Package tcloud ...
package tcloud

import (
	"fmt"

	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

func defaultPrepare(cts *rest.Contexts, cli ressync.Interface) (*sync.TCloudSyncReq, tcloud.Interface, error) {
	req := new(sync.TCloudSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := cli.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, nil, err
	}

	return req, syncCli, nil
}

// baseHandler ...
type baseHandler struct {
	resType enumor.CloudResourceType
	request *sync.TCloudSyncReq
	cli     ressync.Interface

	syncCli tcloud.Interface
}

// Describe load_balancer
func (hd *baseHandler) Describe() string {
	if hd.request == nil {
		return fmt.Sprintf("tcloud %s(-)", hd.Resource())
	}
	return fmt.Sprintf("tcloud %s(region=%s,account=%s)", hd.Resource(), hd.request.Region, hd.request.AccountID)
}

// SyncConcurrent use request specified or 1
func (hd *baseHandler) SyncConcurrent() uint {
	// TODO read from config
	if hd.request != nil && hd.request.Concurrent != 0 {
		return hd.request.Concurrent
	}
	return 1
}

// Resource return resource type of handler
func (hd *baseHandler) Resource() enumor.CloudResourceType {
	return hd.resType
}

// Prepare ...
func (hd *baseHandler) Prepare(cts *rest.Contexts) error {
	request, syncCli, err := defaultPrepare(cts, hd.cli)
	if err != nil {
		return err
	}

	hd.request = request
	hd.syncCli = syncCli

	return nil
}
