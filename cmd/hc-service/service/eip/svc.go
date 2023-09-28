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

package eip

import (
	"fmt"

	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/cmd/hc-service/service/eip/aws"
	"hcm/cmd/hc-service/service/eip/azure"
	"hcm/cmd/hc-service/service/eip/gcp"
	"hcm/cmd/hc-service/service/eip/huawei"
	"hcm/cmd/hc-service/service/eip/tcloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

type eipAdaptor struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}

// EipService ...
type EipService interface {
	DeleteEip(cts *rest.Contexts) (interface{}, error)
	AssociateEip(cts *rest.Contexts) (interface{}, error)
	DisassociateEip(cts *rest.Contexts) (interface{}, error)
	CreateEip(cts *rest.Contexts) (interface{}, error)
	CountEip(cts *rest.Contexts) (interface{}, error)
}

// DeleteEip ...
func (da *eipAdaptor) CountEip(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var svc EipService
	switch vendor {
	case enumor.TCloud:
		svc = &tcloud.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.HuaWei:
		svc = &huawei.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Aws:
		svc = &aws.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Azure:
		svc = &azure.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Gcp:
		svc = &gcp.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	default:
		return nil, fmt.Errorf("%s does not support the count of cloud eips", vendor)
	}

	return svc.CountEip(cts)
}

// DeleteEip ...
func (da *eipAdaptor) DeleteEip(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var svc EipService
	switch vendor {
	case enumor.TCloud:
		svc = &tcloud.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.HuaWei:
		svc = &huawei.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Aws:
		svc = &aws.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Azure:
		svc = &azure.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Gcp:
		svc = &gcp.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	default:
		return nil, fmt.Errorf("%s does not support the delete of cloud eips", vendor)
	}
	return svc.DeleteEip(cts)
}

// AssociateEip ...
func (da *eipAdaptor) AssociateEip(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var svc EipService
	switch vendor {
	case enumor.TCloud:
		svc = &tcloud.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.HuaWei:
		svc = &huawei.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Aws:
		svc = &aws.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Azure:
		svc = &azure.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Gcp:
		svc = &gcp.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	default:
		return nil, fmt.Errorf("%s does not support the delete of cloud eips", vendor)
	}
	return svc.AssociateEip(cts)
}

// DisassociateEip ...
func (da *eipAdaptor) DisassociateEip(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var svc EipService
	switch vendor {
	case enumor.TCloud:
		svc = &tcloud.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.HuaWei:
		svc = &huawei.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Aws:
		svc = &aws.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Azure:
		svc = &azure.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Gcp:
		svc = &gcp.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	default:
		return nil, fmt.Errorf("%s does not support the detach of cloud disks", vendor)
	}
	return svc.DisassociateEip(cts)
}

// CreateEip ...
func (da *eipAdaptor) CreateEip(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.Request.PathParameter("vendor"))
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	var svc EipService
	switch vendor {
	case enumor.TCloud:
		svc = &tcloud.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.HuaWei:
		svc = &huawei.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Aws:
		svc = &aws.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Azure:
		svc = &azure.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	case enumor.Gcp:
		svc = &gcp.EipSvc{Adaptor: da.adaptor, DataCli: da.dataCli}
	default:
		return nil, fmt.Errorf("%s does not support the detach of cloud disks", vendor)
	}
	return svc.CreateEip(cts)
}
