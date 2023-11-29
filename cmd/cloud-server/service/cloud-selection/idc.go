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

package csselection

import (
	"errors"

	"hcm/pkg/api/core"
	coreselection "hcm/pkg/api/core/cloud-selection"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// ListIdc ...
func (svc *service) ListIdc(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	res := meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   meta.CloudSelectionIdc,
			Action: meta.Find,
		},
	}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, res); err != nil {
		logs.Errorf("list idc auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	result, err := svc.client.DataService().Global.CloudSelection.ListIdc(cts.Kit, req)
	if err != nil {
		logs.Errorf("call dataservice to list idc failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	// 添加idc价格，临时方案
	withPrice := slice.Map(result.Details, func(i coreselection.Idc) coreselection.IdcWithPrice {
		return coreselection.IdcWithPrice{Idc: i, Price: svc.cfg.DefaultIdcPrice[i.Vendor]}
	})
	return withPrice, nil
}

func (svc *service) getIdcVendorByIDs(kt *kit.Kit, ids []string) (types.StringArray, error) {

	req := &core.ListReq{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"vendor"},
	}
	result, err := svc.client.DataService().Global.CloudSelection.ListIdc(kt, req)
	if err != nil {
		logs.Errorf("call dataservice to list idc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(result.Details) != len(ids) {
		logs.Errorf("some idc not found, ids: %v, result count: %d, rid: %s", ids, len(result.Details), kt.Rid)
		return nil, errors.New("some idc not found")
	}

	m := make(map[string]struct{})
	for _, one := range result.Details {
		m[string(one.Vendor)] = struct{}{}
	}

	return converter.MapKeyToSlice(m), nil
}
