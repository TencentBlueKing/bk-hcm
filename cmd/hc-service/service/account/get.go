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

package account

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"hcm/pkg/adaptor/types"
	typeaccount "hcm/pkg/adaptor/types/account"
	"hcm/pkg/api/core/cloud"
	hsaccount "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
)

// TCloudGetInfoBySecret 根据秘钥信息去云上获取账号信息
func (svc *service) TCloudGetInfoBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 参数解析与校验
	req := new(cloud.TCloudSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().TCloud(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		})
	if err != nil {
		return nil, err
	}
	// 2. 云上信息获取
	return client.GetAccountInfoBySecret(cts.Kit)

}

// AwsGetInfoBySecret 根据秘钥信息去云上获取账号信息
func (svc *service) AwsGetInfoBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 参数解析与校验
	req := new(cloud.AwsSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// cloudAccountID 通过接口获取
	client, err := svc.ad.Adaptor().Aws(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		}, "", req.Site)
	if err != nil {
		return nil, err
	}
	// 2. 云上信息获取
	return client.GetAccountInfoBySecret(cts.Kit)

}

// HuaWeiGetInfoBySecret 根据秘钥信息去云上获取账号信息
func (svc *service) HuaWeiGetInfoBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 参数解析与校验
	req := new(cloud.HuaWeiSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().HuaWei(&types.BaseSecret{
		CloudSecretID:  req.CloudSecretID,
		CloudSecretKey: req.CloudSecretKey,
	})
	if err != nil {
		return nil, err
	}
	// 2. 云上信息获取
	return client.GetAccountInfoBySecret(cts.Kit, req.CloudSecretID)

}

// GcpGetInfoBySecret 根据秘钥信息去云上获取账号信息
func (svc *service) GcpGetInfoBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 参数解析与校验
	req := new(cloud.GcpSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// project id以接口查询为准，这里用不上
	cred := &types.GcpCredential{CloudProjectID: "xxx", Json: []byte(req.CloudServiceSecretKey)}
	client, err := svc.ad.Adaptor().Gcp(cred)
	if err != nil {
		return nil, err
	}
	// 2. 云上信息获取
	return client.GetAccountInfoBySecret(cts.Kit, req.CloudServiceSecretKey)
}

// AzureGetInfoBySecret 根据秘钥信息去云上获取账号信息
func (svc *service) AzureGetInfoBySecret(cts *rest.Contexts) (interface{}, error) {
	// 1. 参数解析与校验
	req := new(cloud.AzureSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().Azure(&types.AzureCredential{
		// 订阅id以接口查询获取为准，这里用不上
		CloudSubscriptionID:  "xxx",
		CloudTenantID:        req.CloudTenantID,
		CloudApplicationID:   req.CloudApplicationID,
		CloudClientSecretKey: req.CloudClientSecretKey,
	})
	if err != nil {
		return nil, err
	}
	// 2. 云上信息获取
	return client.GetAccountInfoBySecret(cts.Kit)
}

// GetGcpResCountBySecret 根据秘钥信息获取资源数量
func (svc *service) GetGcpResCountBySecret(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.GcpCredential)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().Gcp(&types.GcpCredential{
		CloudProjectID: req.CloudProjectID,
		Json:           []byte(req.CloudServiceSecretKey),
	})
	if err != nil {
		logs.Errorf("new gcp client failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	kt := cts.Kit
	cvmCount, niCount, err := client.CountCvmAndNI(kt)
	if err != nil {
		logs.Errorf("count cvm and network interface failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	diskCount, err := client.CountDisk(kt)
	if err != nil {
		logs.Errorf("count disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	firewallCount, err := client.CountFirewall(kt)
	if err != nil {
		logs.Errorf("count firewall failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	vpcCount, err := client.CountVpc(kt)
	if err != nil {
		logs.Errorf("count vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	subnetCount, err := client.CountSubnet(kt)
	if err != nil {
		logs.Errorf("count subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	eipCount, err := client.CountEip(kt)
	if err != nil {
		logs.Errorf("count eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	routeCount, err := client.CountRoute(kt)
	if err != nil {
		logs.Errorf("count route failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	accountCount, err := client.CountAccount(kt)
	if err != nil {
		logs.Errorf("count account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := &hsaccount.ResCount{
		Items: []*hsaccount.ResCountItem{
			{Type: enumor.CvmCloudResType, Count: cvmCount},
			{Type: enumor.DiskCloudResType, Count: diskCount},
			{Type: enumor.GcpFirewallRuleCloudResType, Count: firewallCount},
			{Type: enumor.VpcCloudResType, Count: vpcCount},
			{Type: enumor.SubnetCloudResType, Count: subnetCount},
			{Type: enumor.EipCloudResType, Count: eipCount},
			{Type: enumor.NetworkInterfaceCloudResType, Count: niCount},
			{Type: enumor.RouteTableCloudResType, Count: routeCount},
			{Type: enumor.SubAccountCloudResType, Count: accountCount},
		},
	}

	return result, nil
}

// GetAzureResCountBySecret 根据秘钥信息获取资源数量
func (svc *service) GetAzureResCountBySecret(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.AzureAuthSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().Azure(&types.AzureCredential{
		CloudTenantID:        req.CloudTenantID,
		CloudSubscriptionID:  req.CloudSubscriptionID,
		CloudApplicationID:   req.CloudApplicationID,
		CloudClientSecretKey: req.CloudClientSecretKey,
	})
	if err != nil {
		logs.Errorf("new gcp client failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	kt := cts.Kit
	cvmCount, err := client.CountCvm(kt)
	if err != nil {
		logs.Errorf("count cvm interface failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	diskCount, err := client.CountDisk(kt)
	if err != nil {
		logs.Errorf("count disk failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	sgCount, err := client.CountSecurityGroup(kt)
	if err != nil {
		logs.Errorf("count firewall failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	vpcCount, subnetCount, err := client.CountVpcAndSubnet(kt)
	if err != nil {
		logs.Errorf("count vpc and subnet failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	niCount, err := client.CountNI(kt)
	if err != nil {
		logs.Errorf("count network interface failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	eipCount, err := client.CountEip(kt)
	if err != nil {
		logs.Errorf("count eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	routeTableCount, err := client.CountRouteTable(kt)
	if err != nil {
		logs.Errorf("count route table failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	accountCount, err := client.CountAccount(kt)
	if err != nil {
		logs.Errorf("count account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := &hsaccount.ResCount{Items: []*hsaccount.ResCountItem{
		{Type: enumor.CvmCloudResType, Count: cvmCount},
		{Type: enumor.DiskCloudResType, Count: diskCount},
		{Type: enumor.SecurityGroupCloudResType, Count: sgCount},
		{Type: enumor.VpcCloudResType, Count: vpcCount},
		{Type: enumor.SubnetCloudResType, Count: subnetCount},
		{Type: enumor.EipCloudResType, Count: eipCount},
		{Type: enumor.NetworkInterfaceCloudResType, Count: niCount},
		{Type: enumor.RouteTableCloudResType, Count: routeTableCount},
		{Type: enumor.SubAccountCloudResType, Count: accountCount},
	}}

	return result, nil
}

// HuaWeiGetResCountBySecret 根据秘钥信息获取资源数量
func (svc *service) HuaWeiGetResCountBySecret(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.HuaWeiSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().HuaWei(&types.BaseSecret{CloudSecretID: req.CloudSecretID,
		CloudSecretKey: req.CloudSecretKey})
	if err != nil {
		return nil, err
	}
	ret := new(hsaccount.ResCount)
	// 获取cvm数量
	retCvm, err := client.CountAllResources(cts.Kit, enumor.HuaWeiCvmProviderType)
	if err != nil {
		logs.Errorf("get cvm count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items,
		&hsaccount.ResCountItem{Type: enumor.CvmCloudResType, Count: converter.PtrToVal(retCvm.TotalCount)})
	// 获取disk数量
	retDisk, err := client.CountAllResources(cts.Kit, enumor.HuaWeiDiskProviderType)
	if err != nil {
		logs.Errorf("get disk count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items,
		&hsaccount.ResCountItem{Type: enumor.DiskCloudResType, Count: converter.PtrToVal(retDisk.TotalCount)})
	// 获取vpc数量
	retVpc, err := client.CountAllResources(cts.Kit, enumor.HuaWeiVpcProviderType)
	if err != nil {
		logs.Errorf("get vpc count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{Type: enumor.VpcCloudResType,
		Count: converter.PtrToVal(retVpc.TotalCount)})
	// 获取eip数量
	retEip, err := client.CountAllResources(cts.Kit, enumor.HuaWeiEipProviderType)
	if err != nil {
		logs.Errorf("get eip count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items,
		&hsaccount.ResCountItem{Type: enumor.EipCloudResType, Count: converter.PtrToVal(retEip.TotalCount)})
	// 获取安全组数量
	retSG, err := client.CountAllResources(cts.Kit, enumor.HuaWeiSGProviderType)
	if err != nil {
		logs.Errorf("get sg count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items,
		&hsaccount.ResCountItem{Type: enumor.SecurityGroupCloudResType, Count: converter.PtrToVal(retSG.TotalCount)})
	// 获取子账号数量
	saCount, err := client.CountSubAccountResources(cts.Kit)
	if err != nil {
		logs.Errorf("get sub account count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{Type: enumor.SubAccountCloudResType, Count: saCount})
	// 获取子网和路由表数量
	sCount, rCount, err := client.CountSubnetRouteTableRes(cts.Kit)
	if err != nil {
		logs.Errorf("get subnet or routetable count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{Type: enumor.SubnetCloudResType, Count: sCount})
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{Type: enumor.RouteTableCloudResType, Count: rCount})
	// 获取网络接口数量
	niCount, err := client.CountNIResources(cts.Kit)
	if err != nil {
		logs.Errorf("get ni count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{Type: enumor.NetworkInterfaceCloudResType, Count: niCount})
	return ret, nil
}

type counterFunc func(kt *kit.Kit, region string) (int32, error)

const ResGetMaxConcurrency = 5

// TCloudGetResCountBySecret 根据秘钥获取云上资源数量
func (svc *service) TCloudGetResCountBySecret(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.TCloudSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	tcloudClient, err := svc.ad.Adaptor().
		TCloud(&types.BaseSecret{CloudSecretID: req.CloudSecretID, CloudSecretKey: req.CloudSecretKey})
	if err != nil {
		return nil, err
	}
	regionListResult, err := tcloudClient.ListRegion(cts.Kit)
	if err != nil {
		return nil, err
	}

	var globalErr error
	wg := sync.WaitGroup{}
	newKit, cancelCtx := shallowCopyKitWithCancel(cts.Kit)
	defer cancelCtx()
	counterMap := map[enumor.CloudResourceType]counterFunc{
		enumor.CvmCloudResType:           tcloudClient.CountCvm,
		enumor.DiskCloudResType:          tcloudClient.CountDisk,
		enumor.VpcCloudResType:           tcloudClient.CountVpc,
		enumor.SubnetCloudResType:        tcloudClient.CountSubnet,
		enumor.RouteTableCloudResType:    tcloudClient.CountRouteTable,
		enumor.EipCloudResType:           tcloudClient.CountEip,
		enumor.SecurityGroupCloudResType: tcloudClient.CountSecurityGroup,
	}
	resultMap := make(map[enumor.CloudResourceType]*int32, len(counterMap))
	for resourceType := range counterMap {
		resultMap[resourceType] = new(int32)
	}
	// 以每种资源的每个地域为粒度并发
	limiter := make(chan struct{}, ResGetMaxConcurrency)
	countByRegion := func(resType enumor.CloudResourceType, counter counterFunc) {
		for _, region := range regionListResult.Details {
			wg.Add(1)
			limiter <- struct{}{}
			go func(region string) {
				defer func() {
					wg.Done()
					<-limiter
				}()
				count, countErr := counter(newKit, region)
				if countErr != nil {
					// 过滤因其他goroutine失败导致的错误
					if !errf.IsContextCanceled(countErr) {
						globalErr = countErr
						cancelCtx()
					}
					return
				}
				atomic.AddInt32(resultMap[resType], count)
			}(region.RegionID)
		}
	}
	for resType, counter := range counterMap {
		countByRegion(resType, counter)
	}
	// 单独处理子账号
	accountCount, err := tcloudClient.CountAccount(newKit)
	if err != nil && !errf.IsContextCanceled(err) {
		globalErr = err
		cancelCtx()
	}
	wg.Wait()
	if globalErr != nil {
		return nil, globalErr
	}
	result := &hsaccount.ResCount{Items: make([]*hsaccount.ResCountItem, 0, len(counterMap)+1)}
	result.Items = append(result.Items,
		&hsaccount.ResCountItem{Type: enumor.SubAccountCloudResType, Count: accountCount})
	for resourceType, count := range resultMap {
		result.Items = append(result.Items, &hsaccount.ResCountItem{Type: resourceType, Count: *count})
	}
	return result, nil
}

// AwsGetResCountBySecret 根据秘钥获取云上资源数量
func (svc *service) AwsGetResCountBySecret(cts *rest.Contexts) (interface{}, error) {
	req := new(cloud.AwsSecret)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Site == "" {
		req.Site = enumor.InternationalSite
	}
	secret := &types.BaseSecret{CloudSecretID: req.CloudSecretID, CloudSecretKey: req.CloudSecretKey}
	awsClient, err := svc.ad.Adaptor().Aws(secret, "", req.Site)
	if err != nil {
		return nil, err
	}

	regionListResult, err := awsClient.ListRegion(cts.Kit)
	if err != nil {
		return nil, err
	}

	var globalErr error
	wg := sync.WaitGroup{}
	// aws没有账号权限，不处理账号部分
	counterMap := map[enumor.CloudResourceType]counterFunc{
		enumor.CvmCloudResType:           awsClient.CountCvm,
		enumor.DiskCloudResType:          awsClient.CountDisk,
		enumor.VpcCloudResType:           awsClient.CountVpc,
		enumor.SubnetCloudResType:        awsClient.CountSubnet,
		enumor.RouteTableCloudResType:    awsClient.CountRouteTable,
		enumor.EipCloudResType:           awsClient.CountEip,
		enumor.SecurityGroupCloudResType: awsClient.CountSecurityGroup,
	}
	resultMap := make(map[enumor.CloudResourceType]*int32, len(counterMap))
	for resourceType := range counterMap {
		resultMap[resourceType] = new(int32)
	}
	// 保证这个context cancel 不会影响其他context
	newKit, cancelCtx := shallowCopyKitWithCancel(cts.Kit)
	defer cancelCtx()

	limiter := make(chan struct{}, ResGetMaxConcurrency)
	countByRegion := func(resType enumor.CloudResourceType, counter counterFunc) {
		for _, region := range regionListResult.Details {
			wg.Add(1)
			limiter <- struct{}{}
			go func(region string) {
				defer func() {
					wg.Done()
					<-limiter
				}()
				count, countErr := counter(newKit, region)
				if countErr != nil {
					// 过滤因其他goroutine失败导致的错误
					var aErr awserr.Error
					if !(errors.As(countErr, &aErr) && aErr.Code() == request.CanceledErrorCode) {
						globalErr = countErr
						cancelCtx()
					}
					return
				}
				atomic.AddInt32(resultMap[resType], count)
			}(region.RegionID)
		}
	}

	for resType, counter := range counterMap {
		countByRegion(resType, counter)
	}
	wg.Wait()

	if globalErr != nil {
		return nil, globalErr
	}
	result := &hsaccount.ResCount{Items: make([]*hsaccount.ResCountItem, 0, len(counterMap)+1)}
	for resourceType, count := range resultMap {
		result.Items = append(result.Items, &hsaccount.ResCountItem{Type: resourceType, Count: *count})
	}
	return result, nil
}

func shallowCopyKitWithCancel(kt *kit.Kit) (*kit.Kit, func()) {
	newKit := converter.ValToPtr(*kt)
	ctxNew, cancel := context.WithCancel(kt.Ctx)
	newKit.Ctx = ctxNew
	return newKit, cancel
}

// ListTCloudAuthPolicies 查询账号授权策略
func (svc *service) ListTCloudAuthPolicies(cts *rest.Contexts) (interface{}, error) {
	req := new(hsaccount.ListTCloudAuthPolicyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.Adaptor().TCloud(
		&types.BaseSecret{
			CloudSecretID:  req.CloudSecretID,
			CloudSecretKey: req.CloudSecretKey,
		})
	if err != nil {
		logs.Errorf("build tcloud client failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typeaccount.TCloudListPolicyOption{
		Uin:         req.Uin,
		ServiceType: req.ServiceType,
	}
	result, err := client.ListPoliciesGrantingServiceAccess(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list policies granting service access failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
