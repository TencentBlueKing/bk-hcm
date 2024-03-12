/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http!=//opensource.org/licenses/MIT
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

package tcloud

import (
	"fmt"
	"strconv"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/adaptor/tcloud"
	typecore "hcm/pkg/adaptor/types/core"
	typesclb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	coreclb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// LoadBalancer 同步指定负载均衡
func (cli *client) LoadBalancer(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) (*SyncResult, error) {
	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lbFromCloud, err := cli.listLBFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	lbFromDB, err := cli.listLBFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(lbFromCloud) == 0 && len(lbFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesclb.TCloudClb, coreclb.TCloudLoadBalancer](
		lbFromCloud, lbFromDB, isLBChange)

	// 删除云上已经删除的负载均衡实例
	if err = cli.deleteLoadBalancer(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
		return nil, err
	}

	// 创建云上新增负载均衡实例
	_, err = cli.createLoadBalancer(kt, params.AccountID, params.Region, addSlice)
	if err != nil {
		return nil, err
	}
	// 更新变更负载均衡
	if err = cli.updateLoadBalancer(kt, params.AccountID, params.Region, updateMap); err != nil {
		return nil, err
	}

	// 同步安全组规则
	lbFromDB, err = cli.listLBFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	lbIDs := make([]string, 0, len(lbFromDB))
	for _, one := range lbFromDB {
		lbIDs = append(lbIDs, one.ID)
	}

	return new(SyncResult), nil
}

// RemoveLoadBalancerDeleteFromCloud 删除存在本地但是在云上被删除的数据
func (cli *client) RemoveLoadBalancerDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
	req := &core.ListReq{
		Filter: tools.EqualWithOpExpression(filter.And, map[string]interface{}{
			"account_id": accountID,
			"region":     region,
		}),
		Page: &core.BasePage{
			Start: 0,
			Limit: constant.BatchOperationMaxLimit,
		},
	}
	for {
		lbFromDB, err := cli.dbCli.Global.LoadBalancer.ListLoadBalancer(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list lb failed, err: %v, req: %v, rid: %s",
				enumor.TCloud, err, req, kt.Rid)
			return err
		}

		cloudIDs := slice.Map(lbFromDB.Details, func(lb coreclb.BaseLoadBalancer) string { return lb.CloudID })

		if len(cloudIDs) == 0 {
			break
		}

		var delCloudIDs []string

		params := &SyncBaseParams{AccountID: accountID, Region: region, CloudIDs: cloudIDs}
		delCloudIDs, err = cli.listRemoveLBID(kt, params)
		if err != nil {
			return err
		}

		if len(delCloudIDs) != 0 {
			if err = cli.deleteLoadBalancer(kt, accountID, region, delCloudIDs); err != nil {
				return err
			}
		}

		if len(lbFromDB.Details) < constant.BatchOperationMaxLimit {
			break
		}

		req.Page.Start += constant.BatchOperationMaxLimit
	}

	return nil
}

// listRemoveLBID check lb exists, return its id if one can not be found
func (cli *client) listRemoveLBID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	batchParam := &SyncBaseParams{
		AccountID: params.AccountID,
		Region:    params.Region,
	}
	lbMap := cvt.StringSliceToMap(params.CloudIDs)

	for _, batchCloudID := range slice.Split(params.CloudIDs, DescribeMax) {
		batchParam.CloudIDs = batchCloudID
		found, err := cli.listLBFromCloud(kt, batchParam)
		if err != nil {
			return nil, err
		}
		for _, lb := range found {
			delete(lbMap, lb.GetCloudID())
		}
	}

	return cvt.MapKeyToSlice(lbMap), nil
}

// call data service to create lb
func (cli *client) createLoadBalancer(kt *kit.Kit, accountID string, region string,
	addSlice []typesclb.TCloudClb) (interface{}, error) {

	if len(addSlice) <= 0 {
		return nil, nil
	}

	cloudVpcIds := slice.Map(addSlice, func(lb typesclb.TCloudClb) string { return cvt.PtrToVal(lb.VpcId) })
	cloudSubnetIDs := slice.Map(addSlice, func(lb typesclb.TCloudClb) string { return cvt.PtrToVal(lb.SubnetId) })

	vpcMap, subnetMap, err := cli.getLoadBalancerRelatedRes(kt, accountID, region, cloudVpcIds, cloudSubnetIDs)
	if err != nil {
		return nil, err
	}

	var lbCreateReq protocloud.TCloudCLBCreateReq

	for _, cloud := range addSlice {
		lbCreateReq.Lbs = append(lbCreateReq.Lbs, convCloudToDBCreate(cloud, accountID, region, vpcMap, subnetMap))
	}

	if _, err := cli.dbCli.TCloud.LoadBalancer.BatchCreateTCloudClb(kt, &lbCreateReq); err != nil {
		logs.Errorf("[%s] call data service to create tcloud load balancer failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return nil, err
	}

	logs.Infof("[%s] sync load balancer to create lb success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(addSlice), kt.Rid)

	return nil, nil
}

// return vpc map and subnet map of given cloud id
func (cli *client) getLoadBalancerRelatedRes(kt *kit.Kit, accountID string, region string, cloudVpcIds []string,
	cloudSubnetIDs []string) (vpcMap map[string]*common.VpcDB, subnetMap map[string]string, err error) {
	vpcMap, err = cli.getVpcMap(kt, accountID, region, cloudVpcIds)
	if err != nil {
		logs.Errorf("fail to get vpc of load balancer during syncing, err: %v, account: %s, vpcIds: %v, rid:%s",
			err, accountID, cloudSubnetIDs, kt.Rid)
		return nil, nil, err
	}

	subnetMap, err = cli.getSubnetMap(kt, accountID, region, cloudSubnetIDs)
	if err != nil {
		logs.Errorf("fail to get subnet of load balancer during syncing, err: %v, account: %s, subnetIDs: %v, rid:%s",
			err, accountID, cloudSubnetIDs, kt.Rid)
		return nil, nil, err
	}
	return vpcMap, subnetMap, nil
}

// call data service to update lb
func (cli *client) updateLoadBalancer(kt *kit.Kit, accountID string, region string,
	updateMap map[string]typesclb.TCloudClb) error {

	if len(updateMap) == 0 {
		return nil
	}

	var updateReq protocloud.TCloudClbBatchUpdateReq
	updateReq.Lbs = cvt.MapToSlice(updateMap, convCloudToDBUpdate)

	if err := cli.dbCli.TCloud.LoadBalancer.BatchUpdate(kt, &updateReq); err != nil {
		logs.Errorf("[%s] call data service to update tcloud load balancer failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return err
	}
	return nil
}

// deleteLoadBalancer call data service to delete lb
func (cli *client) deleteLoadBalancer(kt *kit.Kit, accountID string, region string, delCloudIDs []string) error {

	if len(delCloudIDs) <= 0 {
		return nil
	}

	checkParams := &SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  delCloudIDs,
	}
	delLBFromCloud, err := cli.listLBFromCloud(kt, checkParams)
	if err != nil {
		return err
	}

	if len(delLBFromCloud) > 0 {
		logs.Errorf("[%s] lb not exist before sync deletion, opt: %v, failed_count: %d, rid: %s",
			enumor.TCloud, checkParams, len(delLBFromCloud), kt.Rid)
		return fmt.Errorf("lb not exist before sync deletion")
	}

	deleteReq := &protocloud.ClbBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.LoadBalancer.BatchDelete(kt, deleteReq); err != nil {
		logs.Errorf("[%s] call data service to batch delete lb failed, err: %v, rid: %s",
			enumor.TCloud, err, kt.Rid)
		return err
	}

	logs.Infof("[%s] sync to delete lb success, accountID: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, len(delCloudIDs), kt.Rid)

	return nil
}

// listLBFromCloud list load balancer from cloud vendor
func (cli *client) listLBFromCloud(kt *kit.Kit, params *SyncBaseParams) ([]typesclb.TCloudClb, error) {
	opt := &typesclb.TCloudListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &typecore.TCloudPage{
			Offset: 0,
			Limit:  typecore.TCloudQueryLimit,
		},
	}
	result, err := cli.cloudCli.ListLoadBalancer(kt, opt)
	if err != nil {
		if tcloud.IsNotFoundErr(err) {
			return nil, nil
		}

		logs.Errorf("[%s] list lb from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
			enumor.TCloud, err, params.AccountID, opt, kt.Rid)
		return nil, err
	}

	return result, nil

}

// list load balancer from database
func (cli *client) listLBFromDB(kt *kit.Kit, params *SyncBaseParams) ([]coreclb.TCloudLoadBalancer, error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", params.AccountID),
			tools.RuleEqual("region", params.Region),
			tools.RuleIn("cloud_id", params.CloudIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.TCloud.LoadBalancer.ListLoadBalancer(kt, req)
	if err != nil {
		logs.Errorf("[%s] list lb from db failed, err: %v, account: %s, req: %v, rid: %s",
			enumor.TCloud, err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func convCloudToDBCreate(cloud typesclb.TCloudClb, accountID string, region string, vpcMap map[string]*common.VpcDB,
	subnetMap map[string]string) protocloud.ClbBatchCreate[coreclb.TCloudClbExtension] {

	cloudVpcID := cvt.PtrToVal(cloud.VpcId)
	subnetID := cvt.PtrToVal(cloud.SubnetId)
	lb := protocloud.ClbBatchCreate[coreclb.TCloudClbExtension]{
		CloudID:          cloud.GetCloudID(),
		Name:             cvt.PtrToVal(cloud.LoadBalancerName),
		Vendor:           enumor.TCloud,
		AccountID:        accountID,
		BkBizID:          constant.UnassignedBiz,
		LoadBalancerType: cvt.PtrToVal(cloud.LoadBalancerType),
		Region:           region,
		VpcID:            vpcMap[cloudVpcID].VpcID,
		CloudVpcID:       cloudVpcID,
		SubnetID:         subnetMap[subnetID],
		CloudSubnetID:    subnetID,
		Domain:           cvt.PtrToVal(cloud.LoadBalancerDomain),
		Status:           strconv.FormatUint(cvt.PtrToVal(cloud.Status), 10),
		CloudCreatedTime: cvt.PtrToVal(cloud.CreateTime),
		CloudStatusTime:  cvt.PtrToVal(cloud.StatusTime),
		CloudExpiredTime: cvt.PtrToVal(cloud.ExpireTime),
		// 备注字段云上没有
		Memo: nil,

		Extension: &coreclb.TCloudClbExtension{
			SlaType:                  cvt.PtrToVal(cloud.SlaType),
			VipIsp:                   cvt.PtrToVal(cloud.VipIsp),
			AddressIpVersion:         cvt.PtrToVal(cloud.AddressIPVersion),
			LoadBalancerPassToTarget: cvt.PtrToVal(cloud.LoadBalancerPassToTarget),
			IPv6Mode:                 cvt.PtrToVal(cloud.IPv6Mode),
			Snat:                     cvt.PtrToVal(cloud.Snat),
			SnatPro:                  cvt.PtrToVal(cloud.SnatPro),
			// 该接口无法获取下列字段
			BandwidthPackageId: nil,
		},
	}
	if cloud.NetworkAttributes != nil {
		lb.Extension.InternetMaxBandwidthOut = cvt.PtrToVal(cloud.NetworkAttributes.InternetMaxBandwidthOut)
		lb.Extension.InternetChargeType = cvt.PtrToVal(cloud.NetworkAttributes.InternetChargeType)
		lb.Extension.BandwidthpkgSubType = cvt.PtrToVal(cloud.NetworkAttributes.BandwidthpkgSubType)
	}
	if cloud.SnatIps != nil {
		ipList := make([]coreclb.SnatIp, 0, len(cloud.SnatIps))
		for _, snatIP := range cloud.SnatIps {
			if snatIP == nil {
				continue
			}
			ipList = append(ipList, coreclb.SnatIp{SubnetId: snatIP.SubnetId, Ip: snatIP.Ip})
		}
		lb.Extension.SnatIps = ipList
	}
	// IP地址判断
	if len(cloud.LoadBalancerVips) != 0 {
		switch typesclb.TCloudLoadBalancerType(cvt.PtrToVal(cloud.LoadBalancerType)) {
		case typesclb.InternalLoadBalancerType:
			lb.PrivateIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		case typesclb.OpenLoadBalancerType:
			lb.PublicIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		}
	}
	if ipv6 := cvt.PtrToVal(cloud.AddressIPv6); len(ipv6) > 0 {
		lb.PublicIPv6Addresses = []string{ipv6}
	}

	//  可用区判断
	if typesclb.TCloudLoadBalancerType(lb.LoadBalancerType) == typesclb.OpenLoadBalancerType && cloud.MasterZone != nil {
		lb.Zones = []string{cvt.PtrToVal(cloud.MasterZone.Zone)}
	}

	// 没有碰到的则默认是false
	for _, flag := range cloud.AttributeFlags {
		switch cvt.PtrToVal(flag) {
		case DeleteProtectAttrFlag:
			lb.Extension.DeleteProtect = true
		}
	}

	return lb
}

func convCloudToDBUpdate(id string,
	cloud typesclb.TCloudClb) *protocloud.LoadBalancerExtUpdateReq[coreclb.TCloudClbExtension] {

	lb := protocloud.LoadBalancerExtUpdateReq[coreclb.TCloudClbExtension]{
		ID:               id,
		Name:             cvt.PtrToVal(cloud.LoadBalancerName),
		BkBizID:          constant.UnassignedBiz,
		Domain:           cvt.PtrToVal(cloud.LoadBalancerDomain),
		Status:           strconv.FormatUint(cvt.PtrToVal(cloud.Status), 10),
		CloudCreatedTime: cvt.PtrToVal(cloud.CreateTime),
		CloudStatusTime:  cvt.PtrToVal(cloud.StatusTime),
		CloudExpiredTime: cvt.PtrToVal(cloud.ExpireTime),

		Extension: &coreclb.TCloudClbExtension{
			SlaType:                  cvt.PtrToVal(cloud.SlaType),
			VipIsp:                   cvt.PtrToVal(cloud.VipIsp),
			AddressIpVersion:         cvt.PtrToVal(cloud.AddressIPVersion),
			LoadBalancerPassToTarget: cvt.PtrToVal(cloud.LoadBalancerPassToTarget),

			IPv6Mode: cvt.PtrToVal(cloud.IPv6Mode),
			Snat:     cvt.PtrToVal(cloud.Snat),
			SnatPro:  cvt.PtrToVal(cloud.SnatPro),
		},
	}
	if cloud.NetworkAttributes != nil {
		lb.Extension.InternetMaxBandwidthOut = cvt.PtrToVal(cloud.NetworkAttributes.InternetMaxBandwidthOut)
		lb.Extension.InternetChargeType = cvt.PtrToVal(cloud.NetworkAttributes.InternetChargeType)
		lb.Extension.BandwidthpkgSubType = cvt.PtrToVal(cloud.NetworkAttributes.BandwidthpkgSubType)
	}
	if cloud.SnatIps != nil {
		ipList := make([]coreclb.SnatIp, 0, len(cloud.SnatIps))
		for _, snatIP := range cloud.SnatIps {
			if snatIP == nil {
				continue
			}
			ipList = append(ipList, coreclb.SnatIp{SubnetId: snatIP.SubnetId, Ip: snatIP.Ip})
		}
		lb.Extension.SnatIps = ipList
	}

	if len(cloud.LoadBalancerVips) != 0 {
		switch typesclb.TCloudLoadBalancerType(cvt.PtrToVal(cloud.LoadBalancerType)) {
		case typesclb.InternalLoadBalancerType:
			lb.PrivateIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		case typesclb.OpenLoadBalancerType:
			lb.PublicIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		}
	}
	if ipv6 := cvt.PtrToVal(cloud.AddressIPv6); len(ipv6) > 0 {
		lb.PublicIPv6Addresses = []string{ipv6}
	}
	// AttributeFlags
	// 没有碰到的则默认是false
	for _, flag := range cloud.AttributeFlags {
		switch cvt.PtrToVal(flag) {
		case DeleteProtectAttrFlag:
			lb.Extension.DeleteProtect = true
		}
	}

	if cloud.Egress != nil {
		lb.Extension.Egress = cvt.PtrToVal(cloud.Egress)
	}
	return &lb
}

func isLBChange(cloud typesclb.TCloudClb, db coreclb.TCloudLoadBalancer) bool {

	if db.Name != cvt.PtrToVal(cloud.LoadBalancerName) {
		return true
	}

	if db.Domain != cvt.PtrToVal(cloud.LoadBalancerDomain) {
		return true
	}
	if db.Status != strconv.FormatUint(cvt.PtrToVal(cloud.Status), 10) {
		return true
	}
	if db.CloudCreatedTime != cvt.PtrToVal(cloud.CreateTime) {
		return true
	}
	if db.CloudStatusTime != cvt.PtrToVal(cloud.StatusTime) {
		return true
	}
	if db.CloudExpiredTime != cvt.PtrToVal(cloud.ExpireTime) {
		return true
	}

	if len(cloud.LoadBalancerVips) != 0 {
		var dbIPList []string
		switch typesclb.TCloudLoadBalancerType(cvt.PtrToVal(cloud.LoadBalancerType)) {
		case typesclb.InternalLoadBalancerType:
			dbIPList = db.PrivateIPv4Addresses
		case typesclb.OpenLoadBalancerType:
			dbIPList = db.PublicIPv4Addresses
		}
		if len(dbIPList) == 0 {
			return true
		}

		tmpMap := cvt.StringSliceToMap(cvt.PtrToSlice(cloud.LoadBalancerVips))
		for _, address := range dbIPList {
			delete(tmpMap, address)
		}

		if len(tmpMap) != 0 {
			return true
		}
	}
	if ipv6 := cvt.PtrToVal(cloud.AddressIPv6); len(db.PublicIPv6Addresses) == 0 || db.PublicIPv6Addresses[0] != ipv6 {
		return true
	}

	return isLBExtensionChange(cloud, db)
}

func isLBExtensionChange(cloud typesclb.TCloudClb, db coreclb.TCloudLoadBalancer) bool {
	if db.Extension == nil {
		return true
	}

	if cloud.NetworkAttributes != nil {
		if db.Extension.InternetMaxBandwidthOut != cvt.PtrToVal(cloud.NetworkAttributes.InternetMaxBandwidthOut) {
			return true
		}
		if db.Extension.InternetChargeType != cvt.PtrToVal(cloud.NetworkAttributes.InternetChargeType) {
			return true
		}
		if db.Extension.BandwidthpkgSubType != cvt.PtrToVal(cloud.NetworkAttributes.BandwidthpkgSubType) {
			return true
		}
	}

	if db.Extension.SlaType != cvt.PtrToVal(cloud.SlaType) {
		return true
	}
	if db.Extension.VipIsp != cvt.PtrToVal(cloud.VipIsp) {
		return true
	}
	if db.Extension.AddressIpVersion != cvt.PtrToVal(cloud.AddressIPVersion) {
		return true
	}
	if db.Extension.LoadBalancerPassToTarget != cvt.PtrToVal(cloud.LoadBalancerPassToTarget) {
		return true
	}
	if db.Extension.IPv6Mode != cvt.PtrToVal(cloud.IPv6Mode) {
		return true
	}
	if db.Extension.Snat != cvt.PtrToVal(cloud.Snat) {
		return true
	}
	if db.Extension.SnatPro != cvt.PtrToVal(cloud.SnatPro) {
		return true
	}

	// SnatIP列表对比
	if len(db.Extension.SnatIps) != len(cloud.SnatIps) {
		return true
	} else if len(cloud.SnatIps) != 0 {
		// 将云上的SnatIP转化为map，key为 {SubnetId},{Ip}
		cloudSnatMap := cvt.SliceToMap(cloud.SnatIps, func(ip *tclb.SnatIp) (string, struct{}) {
			if ip == nil {
				return ",", struct{}{}
			}
			return cvt.PtrToVal(ip.SubnetId) + "," + cvt.PtrToVal(ip.Ip), struct{}{}
		})
		for _, local := range db.Extension.SnatIps {
			delete(cloudSnatMap, local.Hash())
		}
		// 数量相等的情况下，应该刚好删除干净
		if len(cloudSnatMap) != 0 {
			return true
		}
	}

	// AttributeFlags
	attrs := make(map[string]struct{}, len(cloud.AttributeFlags))
	for _, flag := range cloud.AttributeFlags {
		attrs[cvt.PtrToVal(flag)] = struct{}{}
	}

	if _, deleteProtect := attrs[DeleteProtectAttrFlag]; deleteProtect != db.Extension.DeleteProtect {
		return true
	}

	if db.Extension.Egress != cvt.PtrToVal(cloud.Egress) {
		return true
	}

	return false
}

// SyncLBOption ...
type SyncLBOption struct {
}

// Validate ...
func (o *SyncLBOption) Validate() error {
	return validator.Validate.Struct(o)
}

const DescribeMax = 20
const DeleteProtectAttrFlag = "DeleteProtect"
