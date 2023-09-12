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

// Package region ...
package region

import (
	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitRegionService initialize the region service.
func InitRegionService(cap *capability.Capability) {
	svc := &regionSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()
	h.Add("BatchCreateRegion", "POST", "/vendors/{vendor}/regions/batch/create", svc.BatchCreateRegion)
	h.Add("BatchUpdateRegion", "PATCH", "/vendors/{vendor}/regions/batch", svc.BatchUpdateRegion)
	h.Add("ListRegion", "POST", "/vendors/{vendor}/regions/list", svc.ListRegion)
	h.Add("BatchDeleteRegion", "DELETE", "/vendors/{vendor}/regions/batch", svc.BatchDeleteRegion)

	h.Load(cap.WebService)
}

type regionSvc struct {
	dao dao.Set
}

// BatchCreateRegion batch create region.
func (svc *regionSvc) BatchCreateRegion(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.BatchCreateTCloudRegion(cts)
	case enumor.Aws:
		return svc.BatchCreateAwsRegion(cts)
	case enumor.Gcp:
		return svc.BatchCreateGcpRegion(cts)
	}

	return nil, nil
}

// BatchUpdateRegion batch update region.
func (svc *regionSvc) BatchUpdateRegion(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var err error
	switch vendor {
	case enumor.TCloud:
		err = svc.BatchUpdateTCloudRegion(cts)
	case enumor.Aws:
		err = svc.BatchUpdateAwsRegion(cts)
	case enumor.Gcp:
		err = svc.BatchUpdateGcpRegion(cts)
	}

	return nil, err
}

// ListRegion list region.
func (svc *regionSvc) ListRegion(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	switch vendor {
	case enumor.TCloud:
		return svc.ListTCloudRegion(cts)
	case enumor.Aws:
		return svc.ListAwsRegion(cts)
	case enumor.Gcp:
		return svc.ListGcpRegion(cts)
	}

	return nil, nil
}

// BatchDeleteRegion batch delete regions.
func (svc *regionSvc) BatchDeleteRegion(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var err error
	switch vendor {
	case enumor.TCloud:
		err = svc.BatchDeleteTCloudRegion(cts)
	case enumor.Aws:
		err = svc.BatchDeleteAwsRegion(cts)
	case enumor.Gcp:
		err = svc.BatchDeleteGcpRegion(cts)
	}

	return nil, err
}
