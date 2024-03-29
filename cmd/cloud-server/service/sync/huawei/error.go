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

package huawei

import "strings"

// Error huawei 有部分地域无法new出客户端，需要将这部分错误信息过滤掉
func Error(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "failed to get project id, No project id found") {
		return nil
	}

	if strings.Contains(err.Error(), "The IAM user is forbidden in the currently selected region") {
		return nil
	}

	if strings.Contains(err.Error(), "unexpected regionId") {
		return nil
	}

	if strings.Contains(err.Error(), "failed to get project id") {
		return nil
	}

	return err
}
