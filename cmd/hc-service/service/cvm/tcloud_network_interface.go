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

package cvm

import (
	adcore "hcm/pkg/adaptor/types/core"
	networkinterface "hcm/pkg/adaptor/types/network-interface"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// ListTCloudCvmNetworkInterface 返回一个map，key为cvmID，value为cvm的网卡信息 ListCvmNetworkInterfaceResp
func (svc *cvmSvc) ListTCloudCvmNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.ListCvmNetworkInterfaceReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmList, err := svc.getCvms(cts.Kit, enumor.TCloud, req.Region, req.CvmIDs)
	if err != nil {
		logs.Errorf("get cvms failed, err: %v, cvmIDs: %v, rid: %s", err, req.CvmIDs, cts.Kit.Rid)
		return nil, err
	}
	cloudIDToIDMap := make(map[string]string)
	for _, baseCvm := range cvmList {
		cloudIDToIDMap[baseCvm.CloudID] = baseCvm.ID
	}

	result, err := svc.listTCloudNetworkInterfaceFromCloud(cts.Kit, req.Region, req.AccountID, cloudIDToIDMap)
	if err != nil {
		logs.Errorf("list tcloud network interface from cloud failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// listTCloudNetworkInterfaceFromCloud 从云端获取网络接口信息
func (svc *cvmSvc) listTCloudNetworkInterfaceFromCloud(kt *kit.Kit, region, accountID string,
	cloudIDToIDMap map[string]string) (map[string]*protocvm.ListCvmNetworkInterfaceRespItem, error) {

	cli, err := svc.ad.TCloud(kt, accountID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*protocvm.ListCvmNetworkInterfaceRespItem)
	var offset uint64 = 0
	for {
		opt := &networkinterface.TCloudNetworkInterfaceListOption{
			Region: region,
			Page: &adcore.TCloudPage{
				Offset: offset,
				Limit:  adcore.TCloudQueryLimit,
			},
			Filters: []*vpc.Filter{
				{
					Name:   common.StringPtr("attachment.instance-id"),
					Values: common.StringPtrs(cvt.MapKeyToSlice(cloudIDToIDMap)),
				},
			},
		}

		resp, err := cli.DescribeNetworkInterfaces(kt, opt)
		if err != nil {
			logs.Errorf("describe network interfaces failed, err: %v, cloudIDs: %v, rid: %s",
				err, cvt.MapKeyToSlice(cloudIDToIDMap), kt.Rid)
			return nil, err
		}
		for _, detail := range resp.Details {
			cloudID := cvt.PtrToVal(detail.Attachment.InstanceId)
			id := cloudIDToIDMap[cloudID]
			if _, ok := result[id]; !ok {
				result[id] = &protocvm.ListCvmNetworkInterfaceRespItem{
					MacAddressToPrivateIpAddresses: make(map[string][]string),
				}
			}

			privateIPs := make([]string, 0)
			for _, set := range detail.PrivateIpAddressSet {
				privateIPs = append(privateIPs, cvt.PtrToVal(set.PrivateIpAddress))
			}
			result[id].MacAddressToPrivateIpAddresses[cvt.PtrToVal(detail.MacAddress)] = privateIPs

		}
		if len(resp.Details) < adcore.TCloudQueryLimit {
			break
		}
		offset += adcore.TCloudQueryLimit
	}
	return result, nil
}
