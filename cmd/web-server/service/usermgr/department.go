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

package usermgr

import (
	"fmt"

	"hcm/cmd/web-server/service/capability"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/usermgr"
)

// InitUsermgrService initial the usermgrSvc service
func InitUsermgrService(c *capability.Capability) {
	svr := &usermgrSvc{
		esbClient: c.EsbClient,
	}

	h := rest.NewHandler()
	h.Add("GetDepartment", "GET", "/departments/{department_id}", svr.GetDepartment)

	h.Load(c.WebService)
}

type usermgrSvc struct {
	esbClient esb.Client
}

// GetDepartment retrieve department from usermgr
func (u *usermgrSvc) GetDepartment(cts *rest.Contexts) (interface{}, error) {
	departmentID, err := cts.PathParameter("department_id").Int64()
	if err != nil {
		return nil, fmt.Errorf("department id invalid, err: %v", err)
	}
	params := &usermgr.RetrieveDepartmentReq{
		ID:     departmentID,
		Fields: []string{},
	}
	data, err := u.esbClient.Usermgr().RetrieveDepartment(cts.Kit.Ctx, params)
	if err != nil {
		return nil, fmt.Errorf("call usermgr retrieve department api failed, err: %v", err)
	}

	return data, nil
}
