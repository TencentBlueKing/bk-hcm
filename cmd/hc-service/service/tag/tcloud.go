/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package tag

import (
	cloudadaptor "hcm/cmd/hc-service/logics/cloud-adaptor"
	ressync "hcm/cmd/hc-service/logics/res-sync"
	"hcm/cmd/hc-service/service/capability"
	typestag "hcm/pkg/adaptor/types/tag"
	apitag "hcm/pkg/api/hc-service/tag"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// InitTagService initial the tag service
func InitTagService(cap *capability.Capability) {
	v := &tag{
		ad:      cap.CloudAdaptor,
		cs:      cap.ClientSet,
		syncCli: ressync.NewClient(cap.CloudAdaptor, cap.ClientSet.DataService()),
	}

	h := rest.NewHandler()

	h.Add("TCloudBatchTagRes", "POST", "/vendors/tcloud/tags/tag_resources/batch", v.TCloudBatchTagRes)

	h.Load(cap.WebService)
}

type tag struct {
	ad      *cloudadaptor.CloudAdaptorClient
	cs      *client.ClientSet
	syncCli ressync.Interface
}

// TCloudBatchTagRes 给账号下多个资源打多个标签。 注：该接口需绑定标签的资源不存在也不会报错
func (t *tag) TCloudBatchTagRes(cts *rest.Contexts) (interface{}, error) {
	req := new(apitag.TCloudBatchTagResRequest)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	account, err := t.cs.DataService().TCloud.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), req.AccountID)
	if err != nil {
		logs.Errorf("fail to get account info: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if account.Vendor != enumor.TCloud {
		return nil, errf.Newf(errf.InvalidParameter, "account %s is not tcloud account", req.AccountID)
	}
	resourceList := make([]string, len(req.Resources))
	for i := range req.Resources {
		resourceList[i] = req.Resources[i].Convert(account.Extension.CloudMainAccountID)
	}
	tcloudCLi, err := t.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("fail to get tcloud adaptor: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &typestag.TCloudTagResOpt{
		ResourceList: resourceList,
		Tags:         req.Tags,
	}
	resp, err := tcloudCLi.TagResources(cts.Kit, opt)
	if err != nil {
		return nil, err
	}
	return resp, nil

}
