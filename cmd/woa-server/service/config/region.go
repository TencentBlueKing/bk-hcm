/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package config region config
package config

import (
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetQcloudRegion gets qcloud region config list
func (s *service) GetQcloudRegion(cts *rest.Contexts) (interface{}, error) {
	rst, err := s.logics.Region().GetRegion(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get region list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}

// GetIdcRegion gets idc region config list
func (s *service) GetIdcRegion(cts *rest.Contexts) (interface{}, error) {
	rst, err := s.logics.Region().GetIdcRegion(cts.Kit)
	if err != nil {
		logs.Errorf("failed to get idc region list, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
