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

import (
	"strings"
	"sync"
	"sync/atomic"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/cvm"
	networkinterface "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
)

const (
	countWorkerNum = 5
)

// CountAllResources count resources for cvm disk vpc sg eip.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-rms/rms_04_0107.html
func (h *HuaWei) CountAllResources(kt *kit.Kit,
	typ enumor.HuaWeiProviderType) (*model.CountAllResourcesResponse, error) {

	client, err := h.clientSet.newRmsClient()
	if err != nil {
		logs.Errorf("[%s] count new rms client failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return nil, err
	}

	request := &model.CountAllResourcesRequest{}
	var listType = []string{
		string(typ),
	}
	request.Type = &listType
	response, err := client.CountAllResources(request)
	if err != nil {
		logs.Errorf("[%s] count all resources failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return nil, err
	}

	return response, nil
}

// CountSubAccountResources count subaccount.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-iam/iam_08_0001.html
func (h *HuaWei) CountSubAccountResources(kt *kit.Kit) (int32, error) {
	accounts, err := h.ListAccount(kt)
	if err != nil {
		logs.Errorf("[%s] count list account failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return 0, err
	}

	return int32(len(accounts)), nil
}

// CountSubnetRouteTableRes count subnet and routeTable.
// reference: https://support.huaweicloud.com/intl/zh-cn/api-vpc/vpc_apiv3_0003.html
func (h *HuaWei) CountSubnetRouteTableRes(kt *kit.Kit) (int32, int32, error) {
	regions, err := h.getAvailableRegions(kt, Vpc)
	if err != nil {
		logs.Errorf("[%s] count get region failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return 0, 0, err
	}

	var subnetCount int32
	var routeTableCount int32
	wg := &sync.WaitGroup{}
	pipeline := make(chan bool, countWorkerNum)
	for _, region := range regions {
		pipeline <- true
		wg.Add(1)

		go func(region string) {
			defer func() {
				wg.Done()
				<-pipeline
			}()

			opt := &types.HuaWeiVpcListOption{
				HuaWeiListOption: core.HuaWeiListOption{
					Region: region,
					Page: &core.HuaWeiPage{
						Limit: converter.ValToPtr(int32(constant.CloudResourceSyncMaxLimit)),
					},
				},
			}
			for {
				vpcs, err := h.ListVpcRaw(kt, opt)
				if countError(err) != nil || vpcs == nil {
					logs.Errorf("[%s] count list vpc failed, err: %v, rid: %s", enumor.HuaWei,
						err, kt.Rid)
					return
				}

				for _, vpc := range converter.PtrToVal(vpcs.Vpcs) {
					for _, one := range vpc.CloudResources {
						if one.ResourceType == "virsubnet" {
							atomic.AddInt32(&subnetCount, one.ResourceCount)
						}

						if one.ResourceType == "routetable" {
							atomic.AddInt32(&routeTableCount, one.ResourceCount)
						}
					}
				}

				if len(converter.PtrToVal(vpcs.Vpcs)) < constant.CloudResourceSyncMaxLimit {
					break
				}

				opt.HuaWeiListOption.Page.Marker = vpcs.PageInfo.NextMarker
			}

		}(region)
	}
	wg.Wait()

	return atomic.LoadInt32(&subnetCount), atomic.LoadInt32(&routeTableCount), nil
}

// CountNIResources count network interface.
// reference: https://support.huaweicloud.com/api-ecs/zh-cn_topic_0094148850.html
// reference: https://support.huaweicloud.com/intl/zh-cn/api-ecs/ecs_02_0505.html
func (h *HuaWei) CountNIResources(kt *kit.Kit) (int32, error) {
	regions, err := h.getAvailableRegions(kt, Ecs)
	if err != nil {
		logs.Errorf("[%s] count get region failed, err: %v, rid: %s", enumor.HuaWei,
			err, kt.Rid)
		return 0, err
	}

	niCount := int32(0)
	wg := &sync.WaitGroup{}
	pipeline := make(chan bool, countWorkerNum)
	for _, region := range regions {
		pipeline <- true
		wg.Add(1)
		go func(region string) {
			defer func() {
				wg.Done()
				<-pipeline
			}()
			opt := &cvm.HuaWeiListOption{
				Region: region,
				Page: &core.HuaWeiCvmOffsetPage{
					Offset: 1,
					Limit:  int32(constant.CloudResourceSyncMaxLimit),
				},
			}
			for {
				cvms, err := h.ListCvm(kt, opt)
				if countError(err) != nil {
					logs.Errorf("[%s] count list cvm failed, err: %v, rid: %s", enumor.HuaWei,
						err, kt.Rid)
					return
				}
				for _, cvm := range cvms {
					opt := &networkinterface.HuaWeiNIListOption{
						Region:   region,
						ServerID: cvm.Id,
					}
					nis, err := h.ListNetworkInterface(kt, opt)
					if countError(err) != nil {
						logs.Errorf("[%s] count list ni failed, err: %v, rid: %s", enumor.HuaWei,
							err, kt.Rid)
						return
					}

					atomic.AddInt32(&niCount, int32(len(nis.Details)))
				}

				if len(cvms) < constant.CloudResourceSyncMaxLimit {
					break
				}

				opt.Page.Offset += 1
			}
		}(region)
	}
	wg.Wait()

	return atomic.LoadInt32(&niCount), nil
}

func (h *HuaWei) getAvailableRegions(kt *kit.Kit, typ string) ([]string, error) {
	ret := make([]string, 0)

	regions, err := h.ListRegion(kt)
	if err != nil {
		return ret, err
	}

	for _, region := range regions {
		if region.Service == typ {
			ret = append(ret, region.RegionID)
		}
	}

	return ret, nil
}

func countError(err error) error {
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
