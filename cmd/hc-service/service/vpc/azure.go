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

// Package vpc defines vpc service.
package vpc

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	dataservice "hcm/pkg/api/data-service"
	"hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/uuid"
)

// AzureVpcUpdate update azure vpc.
func (v vpc) AzureVpcUpdate(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	req := new(hcservice.VpcUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getRes, err := v.cs.DataService().Azure.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	updateOpt := new(types.AzureVpcUpdateOption)
	err = cli.UpdateVpc(cts.Kit, updateOpt)
	if err != nil {
		return nil, err
	}

	updateReq := &cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
		Vpcs: []cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{{
			ID: id,
			VpcUpdateBaseInfo: cloud.VpcUpdateBaseInfo{
				Memo: req.Memo,
			},
		}},
	}
	err = v.cs.DataService().Azure.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AzureVpcDelete delete azure vpc.
func (v vpc) AzureVpcDelete(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()

	getRes, err := v.cs.DataService().Azure.Vpc.Get(cts.Kit.Ctx, cts.Kit.Header(), id)
	if err != nil {
		return nil, err
	}

	cli, err := v.ad.Azure(cts.Kit, getRes.AccountID)
	if err != nil {
		return nil, err
	}

	delOpt := &adcore.AzureDeleteOption{
		BaseDeleteOption:  adcore.BaseDeleteOption{ResourceID: getRes.Name},
		ResourceGroupName: getRes.Extension.ResourceGroup,
	}
	err = cli.DeleteVpc(cts.Kit, delOpt)
	if err != nil {
		return nil, err
	}

	deleteReq := &dataservice.BatchDeleteReq{
		Filter: tools.EqualExpression("id", id),
	}
	err = v.cs.DataService().Global.Vpc.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// AzureVpcSync sync azure cloud vpc.
func (v vpc) AzureVpcSync(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.ResourceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if len(req.ResourceGroupName) == 0 {
		return nil, errf.New(errf.InvalidParameter, "resource_group_name is required")
	}

	// batch get vpc list from cloudapi.
	list, err := v.BatchGetAzureVpcList(cts, req)
	if err != nil {
		logs.Errorf("[%s-vpc] request cloudapi response failed. accountID:%s, region:%s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch get vpc map from db.
	resourceDBMap, err := v.BatchGetVpcMapFromDB(cts, req, enumor.Azure)
	if err != nil {
		logs.Errorf("[%s-vpc] batch get vpcdblist failed. accountID:%s, region:%s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	// batch sync vendor vpc list.
	err = v.BatchSyncAzureVpcList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-vpc] compare api and dblist failed. accountID:%s, region:%s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}

	return hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetAzureVpcList batch get vpc list from cloudapi.
func (v vpc) BatchGetAzureVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.AzureVpcListResult, error) {
	cli, err := v.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &adcore.AzureListOption{
		ResourceGroupName: req.ResourceGroupName,
	}
	list, err := cli.ListVpc(cts.Kit, opt)
	if err != nil {
		logs.Errorf("[%s-vpc]batch get cloud api failed. accountID:%s, region:%s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}
	return list, nil
}

// BatchSyncAzureVpcList batch sync vendor vpc list.
func (v vpc) BatchSyncAzureVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.AzureVpcListResult, resourceDBMap map[string]cloudcore.BaseVpc) error {
	createResources, updateResources, existIDMap, err := v.filterAzureVpcList(cts, req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.VpcBatchUpdateReq[cloud.AzureVpcUpdateExt]{
			Vpcs: updateResources,
		}
		if err = v.cs.DataService().Azure.Vpc.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("[%s-vpc]batch compare db update failed. accountID:%s, region:%s, err: %v",
				enumor.Azure, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		createReq := &cloud.VpcBatchCreateReq[cloud.AzureVpcCreateExt]{
			Vpcs: createResources,
		}
		if _, err = v.cs.DataService().Azure.Vpc.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			logs.Errorf("[%s-vpc]batch compare db create failed. accountID:%s, region:%s, err: %v",
				enumor.Azure, req.AccountID, req.Region, err)
			return err
		}
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	for _, resItem := range resourceDBMap {
		if _, ok := existIDMap[resItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resItem.ID)
		}
	}
	if len(deleteIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}
		if err = v.cs.DataService().Global.Vpc.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
			logs.Errorf("[%s-vpc]batch compare db delete failed. accountID:%s, region:%s, delIDs:%v, err: %v",
				enumor.Azure, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}
	return nil
}

// filterAzureVpcList filter azure vpc list
func (v vpc) filterAzureVpcList(cts *rest.Contexts, req *hcservice.ResourceSyncReq, list *types.AzureVpcListResult,
	resourceDBMap map[string]cloudcore.BaseVpc) (createResources []cloud.VpcCreateReq[cloud.AzureVpcCreateExt],
	updateResources []cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpclist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update vpc data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.VpcUpdateReq[cloud.AzureVpcUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AzureVpcUpdateExt{
					DNSServers: item.Extension.DNSServers,
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Memo = item.Memo

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AzureCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}
			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true

			// sync azure cloud subnet.
			req.VpcName = item.Name
			_, err = v.AzureSubnetSync(cts, req, item.CloudID)
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			// need add vpc data
			tmpRes := cloud.VpcCreateReq[cloud.AzureVpcCreateExt]{
				AccountID: req.AccountID,
				CloudID:   item.CloudID,
				Name:      converter.ValToPtr(item.Name),
				Category:  enumor.BizVpcCategory,
				Memo:      item.Memo,
				Extension: &cloud.AzureVpcCreateExt{
					ResourceGroup: item.Extension.ResourceGroup,
					Region:        item.Extension.Region,
					DNSServers:    item.Extension.DNSServers,
				},
			}

			if item.Extension.Cidr != nil {
				tmpCidrs := []cloud.AzureCidr{}
				for _, cidrItem := range item.Extension.Cidr {
					tmpCidrs = append(tmpCidrs, cloud.AzureCidr{
						Type: cidrItem.Type,
						Cidr: cidrItem.Cidr,
					})
				}
				tmpRes.Extension.Cidr = tmpCidrs
			}
			createResources = append(createResources, tmpRes)

			// sync azure cloud subnet.
			req.VpcName = item.Name
			_, err = v.AzureSubnetSync(cts, req, item.CloudID)
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}
	return createResources, updateResources, existIDMap, nil
}

// AzureSubnetSync sync azure cloud subnet.
func (v vpc) AzureSubnetSync(cts *rest.Contexts, req *hcservice.ResourceSyncReq, cloudVpcID string) (
	interface{}, error) {
	if len(req.VpcName) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vpc_name is required")
	}

	// batch get subnet list from cloudapi.
	list, err := v.BatchGetAzureSubnetList(cts, req)
	if err != nil {
		logs.Errorf("[%s-vpc-subnet]request cloudapi response failed. accountID:%s, cloudVpcID:%s, err: %v",
			enumor.Azure, req.AccountID, cloudVpcID, err)
		return nil, err
	}
	if len(list.Details) == 0 {
		return nil, nil
	}

	// batch get subnet map from db.
	resourceDBMap, err := v.BatchGetSubnetMapFromDB(cts, req, enumor.Azure, cloudVpcID)
	if err != nil {
		logs.Errorf("[%s-vpc-subnet]batch get subnetdblist failed. accountID:%s, cloudVpcID:%s, err: %v",
			enumor.Azure, req.AccountID, cloudVpcID, err)
		return nil, err
	}

	// batch sync vendor subnet list.
	err = v.BatchSyncAzureSubnetList(cts, req, list, resourceDBMap)
	if err != nil {
		logs.Errorf("[%s-vpc-subnet]compare api and dblist failed. accountID:%s, cloudVpcID:%s, err: %v",
			enumor.Azure, req.AccountID, cloudVpcID, err)
		return nil, err
	}

	return hcservice.ResourceSyncResult{
		TaskID: uuid.UUID(),
	}, nil
}

// BatchGetAzureSubnetList batch get subnet list from cloudapi.
func (v vpc) BatchGetAzureSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq) (
	*types.AzureSubnetListResult, error) {
	cli, err := v.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &types.AzureSubnetListOption{
		VpcID: req.VpcName,
	}
	opt.ResourceGroupName = req.ResourceGroupName
	list, err := cli.ListSubnet(cts.Kit, opt)
	if err != nil {
		logs.Errorf("[%s-vpc-subnet]batch get cloud api failed. accountID:%s, region:%s, err: %v",
			enumor.Azure, req.AccountID, req.Region, err)
		return nil, err
	}
	return list, nil
}

// BatchGetSubnetMapFromDB batch get subnet map from db.
func (v vpc) BatchGetSubnetMapFromDB(cts *rest.Contexts, req *hcservice.ResourceSyncReq, vendor enumor.Vendor,
	cloudVpcID string) (map[string]cloudcore.BaseSubnet, error) {
	var (
		page        uint32
		count       = core.DefaultMaxPageLimit
		resourceMap = make(map[string]cloudcore.BaseSubnet, 0)
	)

	for {
		offset := page * uint32(count)
		expr := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
				&filter.AtomRule{
					Field: "account_id",
					Op:    filter.Equal.Factory(),
					Value: req.AccountID,
				},
				&filter.AtomRule{
					Field: "cloud_vpc_id",
					Op:    filter.Equal.Factory(),
					Value: cloudVpcID,
				},
			},
		}
		dbQueryReq := &core.ListReq{
			Filter: expr,
			Page:   &core.BasePage{Count: false, Start: offset, Limit: count},
		}
		dbList, err := v.cs.DataService().Global.Subnet.List(cts.Kit.Ctx, cts.Kit.Header(), dbQueryReq)
		if err != nil {
			logs.Errorf("[%s-vpc-subnet]batch list db error. accountID:%s, region:%s, offset:%d, limit:%d, "+
				"cloudVpcID:%s, err: %v", vendor, req.AccountID, req.Region, offset, count, cloudVpcID, err)
			return nil, err
		}
		if len(dbList.Details) == 0 {
			return resourceMap, nil
		}

		for _, item := range dbList.Details {
			resourceMap[item.CloudID] = item
		}
		if len(dbList.Details) < int(count) {
			break
		}
		page++
	}
	return resourceMap, nil
}

// BatchSyncAzureSubnetList batch sync vendor subnet list.
func (v vpc) BatchSyncAzureSubnetList(cts *rest.Contexts, req *hcservice.ResourceSyncReq,
	list *types.AzureSubnetListResult, resourceDBMap map[string]cloudcore.BaseSubnet) error {
	createResources, updateResources, existIDMap, err := v.filterAzureSubnetList(req, list, resourceDBMap)
	if err != nil {
		return err
	}

	// update resource data
	if len(updateResources) > 0 {
		updateReq := &cloud.SubnetBatchUpdateReq[cloud.AzureSubnetUpdateExt]{
			Subnets: updateResources,
		}
		if err = v.cs.DataService().Azure.Subnet.BatchUpdate(cts.Kit.Ctx, cts.Kit.Header(), updateReq); err != nil {
			logs.Errorf("[%s-vpc-subnet]batch compare db update failed. accountID:%s, region:%s, err: %v",
				enumor.Azure, req.AccountID, req.Region, err)
			return err
		}
	}

	// add resource data
	if len(createResources) > 0 {
		err = v.batchCreateAzureSubnet(cts, createResources)
		if err != nil {
			logs.Errorf("[%s-vpc-subnet]batch compare db create failed. accountID:%s, region:%s, err: %v",
				enumor.Azure, req.AccountID, req.Region, err)
			return err
		}
	}

	// delete resource data
	deleteIDs := make([]string, 0)
	for _, resItem := range resourceDBMap {
		if _, ok := existIDMap[resItem.ID]; !ok {
			deleteIDs = append(deleteIDs, resItem.ID)
		}
	}
	if len(deleteIDs) > 0 {
		deleteReq := &dataservice.BatchDeleteReq{
			Filter: tools.ContainersExpression("id", deleteIDs),
		}
		if err = v.cs.DataService().Global.Subnet.BatchDelete(cts.Kit.Ctx, cts.Kit.Header(), deleteReq); err != nil {
			logs.Errorf("[%s-vpc-subnet]batch compare db delete failed. accountID:%s, region:%s, delIDs:%v, "+
				"err: %v", enumor.Azure, req.AccountID, req.Region, deleteIDs, err)
			return err
		}
	}
	return nil
}

func (v vpc) filterAzureSubnetList(req *hcservice.ResourceSyncReq, list *types.AzureSubnetListResult,
	resourceDBMap map[string]cloudcore.BaseSubnet) (createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt],
	updateResources []cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt], existIDMap map[string]bool, err error) {
	if list == nil || len(list.Details) == 0 {
		return nil, nil, nil,
			fmt.Errorf("cloudapi vpcsubnetlist is empty, accountID:%s, region:%s", req.AccountID, req.Region)
	}

	existIDMap = make(map[string]bool, 0)
	for _, item := range list.Details {
		// need compare and update subnet data
		if resourceInfo, ok := resourceDBMap[item.CloudID]; ok {
			tmpRes := cloud.SubnetUpdateReq[cloud.AzureSubnetUpdateExt]{
				ID: resourceInfo.ID,
				Extension: &cloud.AzureSubnetUpdateExt{
					NatGateway:           converter.ValToPtr(item.Extension.NatGateway),
					NetworkSecurityGroup: converter.ValToPtr(item.Extension.NetworkSecurityGroup),
				},
			}
			tmpRes.Name = converter.ValToPtr(item.Name)
			tmpRes.Ipv4Cidr = item.Ipv4Cidr
			if len(item.Ipv6Cidr) > 0 {
				tmpRes.Ipv6Cidr = item.Ipv6Cidr
			} else {
				tmpRes.Ipv6Cidr = []string{""}
			}
			tmpRes.Memo = item.Memo

			updateResources = append(updateResources, tmpRes)
			existIDMap[resourceInfo.ID] = true
		} else {
			// need add subnet data
			tmpRes := cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]{
				AccountID:  req.AccountID,
				CloudVpcID: item.CloudVpcID,
				CloudID:    item.CloudID,
				Name:       converter.ValToPtr(item.Name),
				Ipv4Cidr:   item.Ipv4Cidr,
				Memo:       item.Memo,
				Extension: &cloud.AzureSubnetCreateExt{
					ResourceGroup:        item.Extension.ResourceGroup,
					NatGateway:           item.Extension.NatGateway,
					NetworkSecurityGroup: item.Extension.NetworkSecurityGroup,
				},
			}
			if len(item.Ipv6Cidr) > 0 {
				tmpRes.Ipv6Cidr = item.Ipv6Cidr
			} else {
				tmpRes.Ipv6Cidr = []string{""}
			}

			createResources = append(createResources, tmpRes)
		}
	}
	return createResources, updateResources, existIDMap, nil
}

func (v vpc) batchCreateAzureSubnet(cts *rest.Contexts,
	createResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]) error {
	querySize := int(filter.DefaultMaxInLimit)
	times := len(createResources) / querySize
	if len(createResources)%querySize != 0 {
		times++
	}
	for i := 0; i < times; i++ {
		var newResources []cloud.SubnetCreateReq[cloud.AzureSubnetCreateExt]
		if i == times-1 {
			newResources = append(newResources, createResources[i*querySize:]...)
		} else {
			newResources = append(newResources, createResources[i*querySize:(i+1)*querySize]...)
		}

		createReq := &cloud.SubnetBatchCreateReq[cloud.AzureSubnetCreateExt]{
			Subnets: newResources,
		}
		if _, err := v.cs.DataService().Azure.Subnet.BatchCreate(cts.Kit.Ctx, cts.Kit.Header(), createReq); err != nil {
			return err
		}
	}
	return nil
}
