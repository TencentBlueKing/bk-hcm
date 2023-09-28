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

package enumor

import "fmt"

// ActionName is action name.
type ActionName string

// Validate ActionName.
func (v ActionName) Validate() error {
	switch v {
	case TestCreateSG:
	case TestCreateSubnet:
	case TestCreateVpc:
	case TestCreateCvm:
	case VirRoot:
	default:
		return fmt.Errorf("unsupported action name type: %s", v)
	}

	return nil
}

const (
	// TestCreateSG test CreateSG
	TestCreateSG ActionName = "test_CreateSG"
	// TestCreateSubnet test CreateSubnet
	TestCreateSubnet ActionName = "test_CreateSubnet"
	// TestCreateVpc test CreateVpc
	TestCreateVpc ActionName = "test_CreateVpc"
	// TestCreateCvm test CreateCvm
	TestCreateCvm ActionName = "test_CreateCvm"

	// VirRoot vir root
	VirRoot ActionName = "root"

	// TestPrintTask test print task
	TestPrintTask ActionName = "test_PrintTask"
)
