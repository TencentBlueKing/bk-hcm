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

package azure

import (
	"errors"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/tidwall/gjson"

	"hcm/pkg/logs"
)

// errorf azure的error格式不可读，需要进行一层转换，取错误原因做为error返回。
/*
e.g:
PUT https://management.azure.com/subscriptions/xx/resourceGroups/xx/providers/Microsoft.Compute/virtualMachines/xx
--------------------------------------------------------------------------------
RESPONSE 409: 409 Conflict
ERROR CODE: OperationNotAllowed
--------------------------------------------------------------------------------
{
  "error": {
    "code": "OperationNotAllowed",
    "message": "Managed disk storage account type change through Virtual Machine 'xx' is not allowed.",
    "target": "osDisk.managedDisk.storageAccountType"
  }
}
--------------------------------------------------------------------------------
*/
func errorf(azureError error) error {
	var azureErr *azcore.ResponseError
	ok := errors.As(azureError, &azureErr)
	if ok {
		all, err := io.ReadAll(azureErr.RawResponse.Body)
		if err != nil {
			logs.ErrorDepthf(1, "read azure error body failed, skip read error message, err: %v")
			return azureError
		}

		message := gjson.GetBytes(all, "error.message").String()
		if len(message) == 0 {
			return azureErr
		}

		return errors.New(message)
	} else {
		return azureErr
	}
}
