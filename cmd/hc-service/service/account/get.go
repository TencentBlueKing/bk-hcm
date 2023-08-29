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
	"hcm/pkg/adaptor/types"
	"hcm/pkg/api/core/cloud"
	hsaccount "hcm/pkg/api/hc-service/account"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
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
		}, "")
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

// HuaWeiGetResCountBySecret 根据秘钥信息获取资源数量
func (svc *service) HuaWeiGetResCountBySecret(cts *rest.Contexts) (interface{}, error) {
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

	ret := new(hsaccount.ResCount)

	// 获取cvm数量
	retCvm, err := client.CountAllResources(cts.Kit, enumor.HuaWeiCvmProviderType)
	if err != nil {
		logs.Errorf("get cvm count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.CvmCloudResType,
		Count: converter.PtrToVal(retCvm.TotalCount),
	})

	// 获取disk数量
	retDisk, err := client.CountAllResources(cts.Kit, enumor.HuaWeiDiskProviderType)
	if err != nil {
		logs.Errorf("get disk count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.DiskCloudResType,
		Count: converter.PtrToVal(retDisk.TotalCount),
	})

	// 获取vpc数量
	retVpc, err := client.CountAllResources(cts.Kit, enumor.HuaWeiVpcProviderType)
	if err != nil {
		logs.Errorf("get vpc count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.VpcCloudResType,
		Count: converter.PtrToVal(retVpc.TotalCount),
	})

	// 获取eip数量
	retEip, err := client.CountAllResources(cts.Kit, enumor.HuaWeiEipProviderType)
	if err != nil {
		logs.Errorf("get eip count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.EipCloudResType,
		Count: converter.PtrToVal(retEip.TotalCount),
	})

	// 获取安全组数量
	retSG, err := client.CountAllResources(cts.Kit, enumor.HuaWeiSGProviderType)
	if err != nil {
		logs.Errorf("get sg count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.SecurityGroupCloudResType,
		Count: converter.PtrToVal(retSG.TotalCount),
	})

	// 获取子账号数量
	saCount, err := client.CountSubAccountResources(cts.Kit)
	if err != nil {
		logs.Errorf("get sub account count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.SubAccountCloudResType,
		Count: saCount,
	})

	// 获取子网和路由表数量
	sCount, rCount, err := client.CountSubnetRouteTableRes(cts.Kit)
	if err != nil {
		logs.Errorf("get subnet or routetable count failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.SubnetCloudResType,
		Count: sCount,
	})
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.RouteTableCloudResType,
		Count: rCount,
	})

	// 获取网络接口数量
	niCount, err := client.CountNIResources(cts.Kit)
	if err != nil {
		logs.Errorf("get ni count failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	ret.Items = append(ret.Items, &hsaccount.ResCountItem{
		Type:  enumor.NetworkInterfaceCloudResType,
		Count: niCount,
	})

	return ret, nil
}
