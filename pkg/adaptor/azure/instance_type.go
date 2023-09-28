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
	"sync"

	typesinstancetype "hcm/pkg/adaptor/types/instance-type"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// ArmInstanceMap azure instance type is arm64
var ArmInstanceMap map[string]struct{} = map[string]struct{}{
	"Standard_D2ps_v5":    {},
	"Standard_D4ps_v5":    {},
	"Standard_D8ps_v5":    {},
	"Standard_D16ps_v5":   {},
	"Standard_D32ps_v5":   {},
	"Standard_D48ps_v5":   {},
	"Standard_D64ps_v5":   {},
	"Standard_D2pds_v5":   {},
	"Standard_D4pds_v5":   {},
	"Standard_D8pds_v5":   {},
	"Standard_D16pds_v5":  {},
	"Standard_D32pds_v5":  {},
	"Standard_D48pds_v5":  {},
	"Standard_D64pds_v5":  {},
	"Standard_D2pls_v5":   {},
	"Standard_D4pls_v5":   {},
	"Standard_D8pls_v5":   {},
	"Standard_D16pls_v5":  {},
	"Standard_D32pls_v5":  {},
	"Standard_D48pls_v5":  {},
	"Standard_D64pls_v5":  {},
	"Standard_D2plds_v5":  {},
	"Standard_D4plds_v5":  {},
	"Standard_D8plds_v5":  {},
	"Standard_D16plds_v5": {},
	"Standard_D32plds_v5": {},
	"Standard_D48plds_v5": {},
	"Standard_D64plds_v5": {},
	"Standard_E2ps_v5":    {},
	"Standard_E4ps_v5":    {},
	"Standard_E8ps_v5":    {},
	"Standard_E16ps_v5":   {},
	"Standard_E20ps_v5":   {},
	"Standard_E32ps_v5":   {},
	"Standard_E2pds_v5":   {},
	"Standard_E4pds_v5":   {},
	"Standard_E8pds_v5":   {},
	"Standard_E16pds_v5":  {},
	"Standard_E20pds_v5":  {},
	"Standard_E32pds_v5":  {},
}

// ListInstanceType ...
// reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machine-sizes/list?tabs=HTTP
func (az *Azure) ListInstanceType(kt *kit.Kit, opt *typesinstancetype.AzureInstanceTypeListOption) (
	its []*typesinstancetype.AzureInstanceType, err error) {

	var typeFamilyMap map[string]string

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		typeFamilyMap, err = az.getInstanceTypeFamilyMap(kt)
		wg.Done()
	}()

	go func() {
		its, err = az.getInstanceTypeList(kt, opt.Region)
		wg.Done()
	}()

	wg.Wait()
	if err != nil {
		return nil, err
	}

	for _, one := range its {
		one.InstanceFamily = typeFamilyMap[one.InstanceType]
	}

	return its, nil
}

func (az *Azure) getInstanceTypeList(kt *kit.Kit, region string) ([]*typesinstancetype.AzureInstanceType, error) {

	client, err := az.clientSet.virtualMachineSizeClient()
	if err != nil {
		return nil, err
	}

	its := make([]*typesinstancetype.AzureInstanceType, 0)
	pager := client.NewListPager(region, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			logs.Errorf("failed to list instance type, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, v := range nextResult.Value {
			if v != nil {
				name := converter.PtrToVal(v.Name)
				its = append(its, &typesinstancetype.AzureInstanceType{
					Architecture: changeToAzureInstanceType(name),
					InstanceType: name,
					CPU:          int64(converter.PtrToVal(v.NumberOfCores)),
					Memory:       int64(converter.PtrToVal(v.MemoryInMB)),
				})
			}
		}
	}

	return its, nil
}

func (az *Azure) getInstanceTypeFamilyMap(kt *kit.Kit) (map[string]string, error) {
	cli, err := az.clientSet.clientFactory()
	if err != nil {
		logs.Errorf("new client factory failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	pager := cli.NewResourceSKUsClient().NewListPager(nil)

	m := make(map[string]string, 0)
	for pager.More() {
		page, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, err
		}
		for _, v := range page.Value {
			if v.Family == nil || v.Name == nil {
				continue
			}

			m[converter.PtrToVal(v.Name)] = converter.PtrToVal(v.Family)
		}
	}

	return m, nil
}

func changeToAzureInstanceType(name string) string {
	if len(name) == 0 {
		return constant.X86
	}

	if _, ok := ArmInstanceMap[name]; ok {
		return constant.Arm64
	}

	return constant.X86
}
