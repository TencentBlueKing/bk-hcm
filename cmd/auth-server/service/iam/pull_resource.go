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

package iam

import (
	"fmt"

	"hcm/cmd/auth-server/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// PullResource callback function for iam to pull auth resource.
func (i *IAM) PullResource(cts *rest.Contexts) (interface{}, error) {
	req := new(types.PullResourceReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	// if auth is disabled, returns error if iam calls pull resource callback function
	if i.disableAuth {
		logs.Errorf("authorize function is disabled, can not pull auth resource, rid: %s", cts.Kit.Rid)
		return nil, errf.New(errf.Aborted, "authorize function is disabled, can not pull auth resource.")
	}

	// get response data for each iam req method, if callback method is not set, returns empty data
	switch req.Method {
	case types.ListInstanceMethod, types.SearchInstanceMethod:
		filter, ok := req.Filter.(types.ListInstanceFilter)
		if !ok {
			logs.Errorf("filter %v is not the right type for list_instance method, rid: %s", filter, cts.Kit.Rid)
			return nil, errf.New(errf.InvalidParameter, "filter type not right")
		}

		instance, err := i.ListInstances(cts.Kit, req.Type, &filter, req.Page)
		if err != nil {
			logs.Errorf("list instance failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return instance, nil

	case types.FetchInstanceInfoMethod:
		filter, ok := req.Filter.(types.FetchInstanceInfoFilter)
		if !ok {
			logs.Errorf("filter %v is not the right type for fetch_instance_info method, rid: %s", filter, cts.Kit.Rid)
			return nil, errf.New(errf.InvalidParameter, "filter type not right")
		}

		info, err := i.FetchInstanceInfo(cts.Kit, req.Type, &filter)
		if err != nil {
			logs.Errorf("fetch instance info failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return info, nil

	case types.ListAttrMethod:
		// attribute authentication is not needed for the time being,
		// so the interface does not need to be implemented
		logs.Errorf("pull resource method list_attr not support, rid: %s", cts.Kit.Rid)
		return nil, errf.New(errf.InvalidParameter, "list_attr not support")

	case types.ListAttrValueMethod:
		// attribute authentication is not needed for the time being,
		// so the interface does not need to be implemented
		logs.Errorf("pull resource method list_attr_value not support, rid: %s", cts.Kit.Rid)
		return nil, errf.New(errf.InvalidParameter, "list_attr_value not support")

	case types.ListInstanceByPolicyMethod:
		// sdk authentication is used, and there is no need to support this interface.
		logs.Errorf("pull resource method list_instance_by_policy not support, rid: %s", cts.Kit.Rid)
		return nil, errf.New(errf.InvalidParameter, "list_instance_by_policy not support")

	default:
		logs.Errorf("pull resource method %s not support, rid: %s", req.Method, cts.Kit.Rid)
		return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("%s not support", req.Method))
	}
}
