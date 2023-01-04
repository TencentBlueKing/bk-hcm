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

package cloud

import (
	"fmt"
	"net/http"

	"hcm/cmd/data-service/service/capability"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitCloudService initial the cloud service
func InitCloudService(cap *capability.Capability) {
	svc := &cloudSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("GetResourceVendor", http.MethodGet, "/cloud/resources/vendors/{type}/id/{id}", svc.GetResourceVendor)
	h.Add("ListResourceVendor", http.MethodPost, "/cloud/resources/vendors/list", svc.ListResourceVendor)

	h.Load(cap.WebService)
}

type cloudSvc struct {
	dao dao.Set
}

// GetResourceVendor get resource vendor.
func (svc cloudSvc) GetResourceVendor(cts *rest.Contexts) (interface{}, error) {
	resourceType := cts.PathParameter("type").String()
	if len(resourceType) == 0 {
		return nil, errf.New(errf.InvalidParameter, "resource type is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "resource id is required")
	}

	list, err := svc.dao.Cloud().ListResourceVendor(cts.Kit, enumor.CloudResourceType(resourceType), []string{id})
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "%s not found resource: %s", resourceType, id)
	}

	if len(list) != 1 {
		logs.Errorf("list resource vendor return count not right, count: %s, resource type: %s, id: %s, rid: %s",
			len(list), resourceType, id, cts.Kit.Rid)
		return nil, fmt.Errorf("list resource vendor return count not right")
	}

	return list[0].Vendor, nil
}

// ListResourceVendor list resource vendor.
func (svc cloudSvc) ListResourceVendor(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.ListResourceVendorReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	list, err := svc.dao.Cloud().ListResourceVendor(cts.Kit, req.ResourceType, req.IDs)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "%s not found resource: %v", req.ResourceType, req.IDs)
	}

	result := make(map[string]enumor.Vendor, len(list))
	for _, vendor := range list {
		result[vendor.ID] = vendor.Vendor
	}

	return result, nil
}
