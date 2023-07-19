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

package subnet

import (
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	hcsubnet "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/maps"
)

// CountSubnetAvailableIPs count subnet available ips. **NOTICE** only for ui.
func (svc *subnetSvc) CountSubnetAvailableIPs(cts *rest.Contexts) (interface{}, error) {
	return svc.countSubnetAvailableIPs(cts, handler.ResValidWithAuth)
}

// CountBizSubnetAvailIPs count biz subnet available ips. **NOTICE** only for ui.
func (svc *subnetSvc) CountBizSubnetAvailIPs(cts *rest.Contexts) (interface{}, error) {
	return svc.countSubnetAvailableIPs(cts, handler.BizValidWithAuth)
}

func (svc *subnetSvc) countSubnetAvailableIPs(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SubnetCloudResType, id, "region", "vendor", "account_id", "bk_biz_id")
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Subnet,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	// get subnet detail info
	switch basicInfo.Vendor {
	case enumor.TCloud:
		req := &hcsubnet.ListCountIPReq{
			Region:    basicInfo.Region,
			AccountID: basicInfo.AccountID,
			IDs:       []string{id},
		}
		idIPMap, err := svc.client.HCService().TCloud.Subnet.ListCountIP(cts.Kit.Ctx, cts.Kit.Header(), req)
		if err != nil {
			return nil, err
		}
		return getCountIPResultFromList(id, idIPMap)
	case enumor.Aws:
		req := &hcsubnet.ListCountIPReq{
			Region:    basicInfo.Region,
			AccountID: basicInfo.AccountID,
			IDs:       []string{id},
		}
		idIPMap, err := svc.client.HCService().Aws.Subnet.ListCountIP(cts.Kit.Ctx, cts.Kit.Header(), req)
		if err != nil {
			return nil, err
		}
		return getCountIPResultFromList(id, idIPMap)
	case enumor.Gcp:
		req := &hcsubnet.ListCountIPReq{
			Region:    basicInfo.Region,
			AccountID: basicInfo.AccountID,
			IDs:       []string{id},
		}
		idIPMap, err := svc.client.HCService().Gcp.Subnet.ListCountIP(cts.Kit.Ctx, cts.Kit.Header(), req)
		if err != nil {
			return nil, err
		}
		return getCountIPResultFromList(id, idIPMap)
	case enumor.HuaWei:
		ipInfo, err := svc.client.HCService().HuaWei.Subnet.CountIP(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}
		return ipInfo, err
	case enumor.Azure:
		subnet, err := svc.client.DataService().Azure.Subnet.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
		if err != nil {
			return nil, err
		}

		req := &hcsubnet.ListAzureCountIPReq{
			ResourceGroupName: subnet.Extension.ResourceGroupName,
			VpcID:             subnet.VpcID,
			AccountID:         basicInfo.AccountID,
			IDs:               []string{id},
		}
		idIPMap, err := svc.client.HCService().Azure.Subnet.ListCountIP(cts.Kit.Ctx, cts.Kit.Header(), req)
		if err != nil {
			return nil, err
		}
		return getCountIPResultFromList(id, idIPMap)
	default:
		return nil, errf.New(errf.InvalidParameter, "vendor is invalid")
	}
}

func getCountIPResultFromList(id string, m map[string]hcsubnet.AvailIPResult) (
	*cloudserver.SubnetCountIPResult, error) {

	count, exist := m[id]
	if !exist {
		return nil, fmt.Errorf("subnet: %s not count ip result", id)
	}

	return &cloudserver.SubnetCountIPResult{
		AvailableIPCount: count.AvailableIPCount,
		TotalIPCount:     count.TotalIPCount,
		UsedIPCount:      count.UsedIPCount,
	}, nil
}

// ListCountResSubnetAvailIPs list count resource subnet avail ips.
func (svc *subnetSvc) ListCountResSubnetAvailIPs(cts *rest.Contexts) (interface{}, error) {
	return svc.listCountSubnetAvailIPs(cts, handler.ListResourceAuthRes)
}

// ListCountBizSubnetAvailIPs list count biz subnet avail ips.
func (svc *subnetSvc) ListCountBizSubnetAvailIPs(cts *rest.Contexts) (interface{}, error) {
	return svc.listCountSubnetAvailIPs(cts, handler.ListBizAuthRes)
}

// listCountSubnetAvailIPs list count subnet avail ips.
func (svc *subnetSvc) listCountSubnetAvailIPs(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (
	interface{}, error) {

	req := new(cloudserver.ListSubnetCountIPReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	flt := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: req.IDs,
			},
		},
	}

	// list authorized instances
	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{Authorizer: svc.authorizer,
		ResType: meta.Subnet, Action: meta.Find, Filter: flt})
	if err != nil {
		return nil, err
	}

	result := make(map[string]cloudserver.SubnetCountIPResult)
	if noPermFlag {
		return result, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"vendor", "region", "account_id", "id"},
	}
	resp, err := svc.client.DataService().Global.Subnet.List(cts.Kit.Ctx, cts.Kit.Header(), listReq)
	if err != nil {
		return nil, err
	}

	subnetMap := make(map[enumor.Vendor][]cloud.BaseSubnet)
	for _, one := range resp.Details {
		if _, exist := subnetMap[one.Vendor]; !exist {
			subnetMap[one.Vendor] = make([]cloud.BaseSubnet, 0)
		}

		subnetMap[one.Vendor] = append(subnetMap[one.Vendor], one)
	}

	for vendor, subnets := range subnetMap {
		tmp := make(map[string]cloudserver.SubnetCountIPResult)
		switch vendor {
		case enumor.TCloud:
			tmp, err = svc.listTCloudAvailIP(cts.Kit, subnets)
		case enumor.Aws:
			tmp, err = svc.listAwsAvailIP(cts.Kit, subnets)
		case enumor.Azure:
			tmp, err = svc.listAzureAvailIP(cts.Kit, subnets)
		case enumor.HuaWei:
			tmp, err = svc.listHuaWeiAvailIP(cts.Kit, subnets)
		case enumor.Gcp:
			tmp, err = svc.listGcpAvailIP(cts.Kit, subnets)
		default:
			// 如果这个云没有获取可用IP的能力的话，则返回接口不包括这个云的数据，避免造成误解。
			continue
		}
		if err != nil {
			logs.Errorf("list %s avail ip failed, err: %v, ids: %v, rid: %s", vendor, err, req.IDs, cts.Kit.Rid)
			return nil, err
		}

		result = maps.MapAppend(result, tmp)
	}

	return result, nil
}

func (svc *subnetSvc) listHuaWeiAvailIP(kt *kit.Kit, subnets []cloud.BaseSubnet) (
	map[string]cloudserver.SubnetCountIPResult, error) {

	result := make(map[string]cloudserver.SubnetCountIPResult)
	for _, one := range subnets {
		respResult, err := svc.client.HCService().HuaWei.Subnet.CountIP(kt.Ctx, kt.Header(), one.ID)
		if err != nil {
			logs.Errorf("get huawei subnet avail ip result failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		result[one.ID] = cloudserver.SubnetCountIPResult{
			AvailableIPCount: respResult.AvailableIPCount,
			TotalIPCount:     respResult.TotalIPCount,
			UsedIPCount:      respResult.UsedIPCount,
		}
	}

	return result, nil
}

func (svc *subnetSvc) listTCloudAvailIP(kt *kit.Kit, subnets []cloud.BaseSubnet) (
	map[string]cloudserver.SubnetCountIPResult, error) {

	classSubnets := classSubnet(subnets)

	result := make(map[string]cloudserver.SubnetCountIPResult)
	for accountID, regionMap := range classSubnets {
		for region, ids := range regionMap {
			req := &hcsubnet.ListCountIPReq{
				Region:    region,
				AccountID: accountID,
				IDs:       ids,
			}
			respData, err := svc.client.HCService().TCloud.Subnet.ListCountIP(kt.Ctx, kt.Header(), req)
			if err != nil {
				logs.Errorf("list tcloud count ip failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
				return nil, err
			}

			for id, ipResult := range respData {
				result[id] = cloudserver.SubnetCountIPResult{
					AvailableIPCount: ipResult.AvailableIPCount,
					TotalIPCount:     ipResult.TotalIPCount,
					UsedIPCount:      ipResult.UsedIPCount,
				}
			}
		}
	}

	return result, nil
}

func (svc *subnetSvc) listGcpAvailIP(kt *kit.Kit, subnets []cloud.BaseSubnet) (
	map[string]cloudserver.SubnetCountIPResult, error) {

	classSubnets := classSubnet(subnets)

	result := make(map[string]cloudserver.SubnetCountIPResult)
	for accountID, regionMap := range classSubnets {
		for region, ids := range regionMap {
			req := &hcsubnet.ListCountIPReq{
				Region:    region,
				AccountID: accountID,
				IDs:       ids,
			}
			respData, err := svc.client.HCService().Gcp.Subnet.ListCountIP(kt.Ctx, kt.Header(), req)
			if err != nil {
				logs.Errorf("list gcp count ip failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
				return nil, err
			}

			for id, ipResult := range respData {
				result[id] = cloudserver.SubnetCountIPResult{
					AvailableIPCount: ipResult.AvailableIPCount,
					TotalIPCount:     ipResult.TotalIPCount,
					UsedIPCount:      ipResult.UsedIPCount,
				}
			}
		}
	}

	return result, nil
}

func (svc *subnetSvc) listAwsAvailIP(kt *kit.Kit, subnets []cloud.BaseSubnet) (
	map[string]cloudserver.SubnetCountIPResult, error) {

	classSubnets := classSubnet(subnets)

	result := make(map[string]cloudserver.SubnetCountIPResult)
	for accountID, regionMap := range classSubnets {
		for region, ids := range regionMap {
			req := &hcsubnet.ListCountIPReq{
				Region:    region,
				AccountID: accountID,
				IDs:       ids,
			}
			respData, err := svc.client.HCService().Aws.Subnet.ListCountIP(kt.Ctx, kt.Header(), req)
			if err != nil {
				logs.Errorf("list aws count ip failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
				return nil, err
			}

			for id, ipResult := range respData {
				result[id] = cloudserver.SubnetCountIPResult{
					AvailableIPCount: ipResult.AvailableIPCount,
					TotalIPCount:     ipResult.TotalIPCount,
					UsedIPCount:      ipResult.UsedIPCount,
				}
			}
		}
	}

	return result, nil
}

func (svc *subnetSvc) listAzureAvailIP(kt *kit.Kit, subnets []cloud.BaseSubnet) (
	map[string]cloudserver.SubnetCountIPResult, error) {

	ids := make([]string, 0, len(subnets))
	for _, one := range subnets {
		ids = append(ids, one.ID)
	}

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	listResult, err := svc.client.DataService().Azure.Subnet.ListSubnetExt(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list azure subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	classSubnets := make(map[string]map[string]map[string][]string)
	for _, one := range listResult.Details {
		if _, exist := classSubnets[one.AccountID]; !exist {
			classSubnets[one.AccountID] = make(map[string]map[string][]string)
		}

		resGroupName := one.Extension.ResourceGroupName
		vpcID := one.VpcID
		if _, exist := classSubnets[one.AccountID][resGroupName]; !exist {
			classSubnets[one.AccountID][resGroupName] = make(map[string][]string, 0)
		}

		if _, exist := classSubnets[one.AccountID][resGroupName][vpcID]; !exist {
			classSubnets[one.AccountID][resGroupName][vpcID] = make([]string, 0)
		}

		classSubnets[one.AccountID][resGroupName][vpcID] = append(classSubnets[one.AccountID][resGroupName][vpcID],
			one.ID)
	}

	result := make(map[string]cloudserver.SubnetCountIPResult)
	for accountID, resGroupMap := range classSubnets {
		for resGroupName, vpcMap := range resGroupMap {
			for vpcID, subnetIDs := range vpcMap {
				req := &hcsubnet.ListAzureCountIPReq{
					ResourceGroupName: resGroupName,
					VpcID:             vpcID,
					AccountID:         accountID,
					IDs:               subnetIDs,
				}
				respData, err := svc.client.HCService().Azure.Subnet.ListCountIP(kt.Ctx, kt.Header(), req)
				if err != nil {
					logs.Errorf("list azure count ip failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
					return nil, err
				}

				for id, ipResult := range respData {
					result[id] = cloudserver.SubnetCountIPResult{
						AvailableIPCount: ipResult.AvailableIPCount,
						TotalIPCount:     ipResult.TotalIPCount,
						UsedIPCount:      ipResult.UsedIPCount,
					}
				}
			}
		}
	}

	return result, nil
}

func classSubnet(subnets []cloud.BaseSubnet) map[string]map[string][]string {
	classSubnets := make(map[string]map[string][]string)
	for _, one := range subnets {
		if _, exist := classSubnets[one.AccountID]; !exist {
			classSubnets[one.AccountID] = make(map[string][]string)
		}

		if _, exist := classSubnets[one.AccountID][one.Region]; !exist {
			classSubnets[one.AccountID][one.Region] = make([]string, 0)
		}

		classSubnets[one.AccountID][one.Region] = append(classSubnets[one.AccountID][one.Region], one.ID)
	}
	return classSubnets
}
