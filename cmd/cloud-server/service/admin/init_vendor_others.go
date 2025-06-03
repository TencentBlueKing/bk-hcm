/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package admin

import (
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// OtherAccountInit 查找是否存在vendor为other的用户，若有则返回，没有则创建
func (s *adminService) OtherAccountInit(cts *rest.Contexts) (any, error) {

	// 这里不鉴权，在web-server中屏蔽请求，只允许系统内部调用
	resp, err := s.adminLogics.InitVendorOtherAccount(cts.Kit)
	if err != nil {
		logs.Errorf("init vendor other account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return resp, nil
}
