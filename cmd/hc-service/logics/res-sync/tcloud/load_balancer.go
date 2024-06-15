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
	typecore "hcm/pkg/adaptor/types/core"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// LoadBalancerWithListener 同步指定负载均衡及下属监听器、规则
// 1. 同步该负载均衡自身属性，同步关联安全组信息
// 2. 同步该负载均衡下的监听器
// 3. 同步监听器下的规则
func (cli *client) LoadBalancerWithListener(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) (*SyncResult,
	error) {

	_, err := cli.LoadBalancer(kt, params, opt)
	if err != nil {
		logs.Errorf("fail to sync load balancer with rel, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// 同步下属监听器
	requiredLBCloudIds := params.CloudIDs
	lbList, err := cli.listLBFromDB(kt, params)
	if err != nil {
		logs.Errorf("fail to get lb from db after lb layer sync, before listener sync, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 同步对应安全组关联关系
	err = cli.lbSgRel(kt, params, lbList)
	if err != nil {
		logs.Errorf("fail to sync load balancer sg rel, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(lbList) == 0 {
		return new(SyncResult), nil
	}

	lblParams := &SyncListenerBatchOption{
		AccountID: params.AccountID,
		Region:    params.Region,
		LbInfos:   lbList,
	}

	if _, err = cli.listenerByLbBatch(kt, lblParams); err != nil {
		logs.Errorf("fail to sync listener of lbs, err: %v, ids: %v, rid: %s", err, requiredLBCloudIds, kt.Rid)
		return nil, err
	}

	return new(SyncResult), nil
}

// LoadBalancer 同步指定负载均衡自身属性，不同步关联资源
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

	addSlice, updateMap, delCloudIDs := common.Diff[typeslb.TCloudClb, corelb.TCloudLoadBalancer](
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
	return new(SyncResult), nil
}

// RemoveLoadBalancerDeleteFromCloud 删除存在本地但是在云上被删除的数据
func (cli *client) RemoveLoadBalancerDeleteFromCloud(kt *kit.Kit, accountID string, region string) error {
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("region", region),
		),
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

		cloudIDs := slice.Map(lbFromDB.Details, func(lb corelb.BaseLoadBalancer) string { return lb.CloudID })

		if len(cloudIDs) == 0 {
			break
		}

		var delCloudIDs []string

		params := &SyncBaseParams{AccountID: accountID, Region: region, CloudIDs: cloudIDs}
		delCloudIDs, err = cli.listRemovedLBID(kt, params)
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

// listRemovedLBID check lb exists, return its id if one can not be found
func (cli *client) listRemovedLBID(kt *kit.Kit, params *SyncBaseParams) ([]string, error) {
	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	batchParam := &SyncBaseParams{
		AccountID: params.AccountID,
		Region:    params.Region,
	}
	lbMap := cvt.StringSliceToMap(params.CloudIDs)

	for _, batchCloudID := range slice.Split(params.CloudIDs, constant.TCLBDescribeMax) {
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

// createLoadBalancer call data service to create lb
func (cli *client) createLoadBalancer(kt *kit.Kit, accountID string, region string,
	addSlice []typeslb.TCloudClb) (interface{}, error) {

	if len(addSlice) <= 0 {
		return nil, nil
	}

	cloudVpcIds := slice.Map(addSlice, func(lb typeslb.TCloudClb) string { return cvt.PtrToVal(lb.VpcId) })
	cloudSubnetIDs := slice.Map(addSlice, func(lb typeslb.TCloudClb) string { return cvt.PtrToVal(lb.SubnetId) })

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

// getLoadBalancerRelatedRes return vpc map and subnet map of given cloud id
func (cli *client) getLoadBalancerRelatedRes(kt *kit.Kit, accountID string, region string, cloudVpcIds []string,
	cloudSubnetIDs []string) (vpcMap map[string]*common.VpcDB, subnetMap map[string]string, err error) {

	vpcMap, err = cli.getVpcMap(kt, accountID, region, cloudVpcIds)
	if err != nil {
		logs.Errorf("fail to get vpc of load balancer during syncing, err: %v, account: %s, vpcIds: %v, rid:%s",
			err, accountID, cloudVpcIds, kt.Rid)
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

// updateLoadBalancer call data service to update lb
func (cli *client) updateLoadBalancer(kt *kit.Kit, accountID string, region string,
	updateMap map[string]typeslb.TCloudClb) error {

	if len(updateMap) == 0 {
		return nil
	}
	var cloudSubnetIDs, cloudVpcIds []string
	for _, clb := range updateMap {
		cloudVpcIds = append(cloudVpcIds, cvt.PtrToVal(clb.VpcId))
		cloudSubnetIDs = append(cloudSubnetIDs, cvt.PtrToVal(clb.SubnetId))
	}

	vpcMap, subnetMap, err := cli.getLoadBalancerRelatedRes(kt, accountID, region, cloudVpcIds, cloudSubnetIDs)
	if err != nil {
		logs.Errorf("fail to get load balancer related res for update db, err: %v, account: %s, "+
			"vpcIds: %v, cloud subnet ids: %v, rid: %s", err, accountID, cloudVpcIds, cloudSubnetIDs, kt.Rid)
		return err
	}

	var updateReq protocloud.TCloudClbBatchUpdateReq

	for id, clb := range updateMap {
		updateReq.Lbs = append(updateReq.Lbs, convCloudToDBUpdate(id, clb, vpcMap, subnetMap, region))
	}
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

	deleteReq := &protocloud.LoadBalancerBatchDeleteReq{
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
func (cli *client) listLBFromCloud(kt *kit.Kit, params *SyncBaseParams) (result []typeslb.TCloudClb, err error) {

	opt := &typeslb.TCloudListOption{
		Region:   params.Region,
		CloudIDs: params.CloudIDs,
		Page: &typecore.TCloudPage{
			Offset: 0,
			Limit:  constant.TCLBDescribeMax,
		},
	}
	// 指定id时一次只能查询20个
	for _, cloudIds := range slice.Split(params.CloudIDs, constant.TCLBDescribeMax) {
		opt.CloudIDs = cloudIds
		batch, err := cli.cloudCli.ListLoadBalancer(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list lb from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
				enumor.TCloud, err, params.AccountID, opt, kt.Rid)
			return nil, err
		}
		result = append(result, batch...)
	}

	return result, nil

}

// listLBFromDB list load balancer from database
func (cli *client) listLBFromDB(kt *kit.Kit, params *SyncBaseParams) ([]corelb.TCloudLoadBalancer, error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", params.AccountID),
			tools.RuleEqual("region", params.Region),
		),
		Page: core.NewDefaultBasePage(),
	}
	// 支持查询所有，或者指定cloud_id
	if len(params.CloudIDs) > 0 {
		req.Filter.Rules = append(req.Filter.Rules, tools.RuleIn("cloud_id", params.CloudIDs))
	}
	result, err := cli.dbCli.TCloud.LoadBalancer.ListLoadBalancer(kt, req)
	if err != nil {
		logs.Errorf("[%s] list lb from db failed, err: %v, account: %s, req: %v, rid: %s",
			enumor.TCloud, err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

func convCloudToDBCreate(cloud typeslb.TCloudClb, accountID string, region string, vpcMap map[string]*common.VpcDB,
	subnetMap map[string]string) protocloud.LbBatchCreate[corelb.TCloudClbExtension] {

	cloudVpcID := cvt.PtrToVal(cloud.VpcId)
	cloudSubnetID := cvt.PtrToVal(cloud.SubnetId)
	lb := protocloud.LbBatchCreate[corelb.TCloudClbExtension]{
		CloudID:          cloud.GetCloudID(),
		Name:             cvt.PtrToVal(cloud.LoadBalancerName),
		Vendor:           enumor.TCloud,
		AccountID:        accountID,
		BkBizID:          constant.UnassignedBiz,
		LoadBalancerType: cvt.PtrToVal(cloud.LoadBalancerType),
		IPVersion:        cloud.GetIPVersion(),
		Region:           region,
		VpcID:            cvt.PtrToVal(vpcMap[cloudVpcID]).VpcID,
		CloudVpcID:       cloudVpcID,
		SubnetID:         subnetMap[cloudSubnetID],
		CloudSubnetID:    cloudSubnetID,
		Domain:           cvt.PtrToVal(cloud.LoadBalancerDomain),
		Status:           strconv.FormatUint(cvt.PtrToVal(cloud.Status), 10),
		CloudCreatedTime: cvt.PtrToVal(cloud.CreateTime),
		CloudStatusTime:  cvt.PtrToVal(cloud.StatusTime),
		CloudExpiredTime: cvt.PtrToVal(cloud.ExpireTime),
		// 备注字段云上没有
		Memo: nil,
	}

	// IP地址判断
	if len(cloud.LoadBalancerVips) != 0 {
		switch typeslb.TCloudLoadBalancerType(cvt.PtrToVal(cloud.LoadBalancerType)) {
		case typeslb.InternalLoadBalancerType:
			lb.PrivateIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		case typeslb.OpenLoadBalancerType:
			lb.PublicIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		}
	}
	if ipv6 := cvt.PtrToVal(cloud.AddressIPv6); len(ipv6) > 0 {
		lb.PublicIPv6Addresses = []string{ipv6}
	}

	//  可用区判断
	if typeslb.TCloudLoadBalancerType(lb.LoadBalancerType) == typeslb.OpenLoadBalancerType && cloud.MasterZone != nil {
		lb.Zones = []string{cvt.PtrToVal(cloud.MasterZone.Zone)}
	}

	lb.Extension = convertTCloudExtension(cloud, region)

	return lb
}

func convertTCloudExtension(cloud typeslb.TCloudClb, region string) *corelb.TCloudClbExtension {
	ext := &corelb.TCloudClbExtension{
		SlaType:                  cloud.SlaType,
		VipIsp:                   cloud.VipIsp,
		LoadBalancerPassToTarget: cloud.LoadBalancerPassToTarget,
		IPv6Mode:                 cloud.IPv6Mode,
		Snat:                     cloud.Snat,
		SnatPro:                  cloud.SnatPro,
		MixIpTarget:              cloud.MixIpTarget,
		ChargeType:               cloud.ChargeType,
		// 该接口无法获取下列字段
		BandwidthPackageId: nil,
	}
	if cloud.NetworkAttributes != nil {
		ext.InternetMaxBandwidthOut = cloud.NetworkAttributes.InternetMaxBandwidthOut
		ext.InternetChargeType = cloud.NetworkAttributes.InternetChargeType
		ext.BandwidthpkgSubType = cloud.NetworkAttributes.BandwidthpkgSubType
	}
	if cloud.SnatIps != nil {
		ipList := make([]corelb.SnatIp, 0, len(cloud.SnatIps))
		for _, snatIP := range cloud.SnatIps {
			if snatIP == nil {
				continue
			}
			ipList = append(ipList, corelb.SnatIp{SubnetId: snatIP.SubnetId, Ip: snatIP.Ip})
		}
		ext.SnatIps = cvt.ValToPtr(ipList)
	}

	flagMap := make(map[string]bool)
	// 没有碰到的则默认是false
	for _, flag := range cloud.AttributeFlags {
		flagMap[cvt.PtrToVal(flag)] = true
	}
	// 逐个赋值flag
	ext.DeleteProtect = cvt.ValToPtr(flagMap[constant.TCLBDeleteProtect])

	// 跨域1.0 信息
	if cvt.PtrToVal(cloud.TargetRegionInfo.Region) != region ||
		!assert.IsPtrStringEqual(cloud.TargetRegionInfo.VpcId, cloud.VpcId) {
		ext.TargetRegion = cloud.TargetRegionInfo.Region
		ext.TargetCloudVpcID = cloud.TargetRegionInfo.VpcId
	}

	return ext
}

func convCloudToDBUpdate(id string, cloud typeslb.TCloudClb, vpcMap map[string]*common.VpcDB,
	subnetMap map[string]string, region string) *protocloud.LoadBalancerExtUpdateReq[corelb.TCloudClbExtension] {

	cloudVpcID := cvt.PtrToVal(cloud.VpcId)
	cloudSubnetID := cvt.PtrToVal(cloud.SubnetId)
	lb := protocloud.LoadBalancerExtUpdateReq[corelb.TCloudClbExtension]{
		ID:               id,
		Name:             cvt.PtrToVal(cloud.LoadBalancerName),
		Domain:           cvt.PtrToVal(cloud.LoadBalancerDomain),
		IPVersion:        cloud.GetIPVersion(),
		Status:           strconv.FormatUint(cvt.PtrToVal(cloud.Status), 10),
		CloudCreatedTime: cvt.PtrToVal(cloud.CreateTime),
		CloudStatusTime:  cvt.PtrToVal(cloud.StatusTime),
		CloudExpiredTime: cvt.PtrToVal(cloud.ExpireTime),
		VpcID:            cvt.PtrToVal(vpcMap[cloudVpcID]).VpcID,
		CloudVpcID:       cloudVpcID,
		SubnetID:         subnetMap[cloudSubnetID],
		CloudSubnetID:    cloudSubnetID,
		Extension: &corelb.TCloudClbExtension{
			SlaType:                  cloud.SlaType,
			VipIsp:                   cloud.VipIsp,
			LoadBalancerPassToTarget: cloud.LoadBalancerPassToTarget,
			ChargeType:               cloud.ChargeType,

			IPv6Mode: cloud.IPv6Mode,
			Snat:     cloud.Snat,
			SnatPro:  cloud.SnatPro,
		},
	}
	if cloud.NetworkAttributes != nil {
		lb.Extension.InternetMaxBandwidthOut = cloud.NetworkAttributes.InternetMaxBandwidthOut
		lb.Extension.InternetChargeType = cloud.NetworkAttributes.InternetChargeType
		lb.Extension.BandwidthpkgSubType = cloud.NetworkAttributes.BandwidthpkgSubType
	}
	if cloud.SnatIps != nil {
		ipList := make([]corelb.SnatIp, 0, len(cloud.SnatIps))
		for _, snatIP := range cloud.SnatIps {
			if snatIP == nil {
				continue
			}
			ipList = append(ipList, corelb.SnatIp{SubnetId: snatIP.SubnetId, Ip: snatIP.Ip})
		}
		lb.Extension.SnatIps = cvt.ValToPtr(ipList)
	}

	if len(cloud.LoadBalancerVips) != 0 {
		switch typeslb.TCloudLoadBalancerType(cvt.PtrToVal(cloud.LoadBalancerType)) {
		case typeslb.InternalLoadBalancerType:
			lb.PrivateIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		case typeslb.OpenLoadBalancerType:
			lb.PublicIPv4Addresses = cvt.PtrToSlice(cloud.LoadBalancerVips)
		}
	}
	if ipv6 := cvt.PtrToVal(cloud.AddressIPv6); len(ipv6) > 0 {
		lb.PublicIPv6Addresses = []string{ipv6}
	}

	// AttributeFlags
	flagMap := make(map[string]bool)
	// 没有碰到的则默认是false
	for _, flag := range cloud.AttributeFlags {
		flagMap[cvt.PtrToVal(flag)] = true
	}
	// 逐个赋值flag
	lb.Extension.DeleteProtect = cvt.ValToPtr(flagMap[constant.TCLBDeleteProtect])

	if cloud.Egress != nil {
		lb.Extension.Egress = cloud.Egress
	}

	// 跨域1.0 信息
	if cvt.PtrToVal(cloud.TargetRegionInfo.Region) != region ||
		!assert.IsPtrStringEqual(cloud.TargetRegionInfo.VpcId, cloud.VpcId) {
		lb.Extension.TargetRegion = cloud.TargetRegionInfo.Region
		lb.Extension.TargetCloudVpcID = cloud.TargetRegionInfo.VpcId
	}
	return &lb
}

func isLBChange(cloud typeslb.TCloudClb, db corelb.TCloudLoadBalancer) bool {

	if db.Name != cvt.PtrToVal(cloud.LoadBalancerName) {
		return true
	}

	if db.IPVersion != cloud.GetIPVersion() {
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
	if db.CloudVpcID != cvt.PtrToVal(cloud.VpcId) {
		return true
	}
	if db.CloudSubnetID != cvt.PtrToVal(cloud.SubnetId) {
		return true
	}

	if len(cloud.LoadBalancerVips) != 0 {
		var dbIPList []string
		switch typeslb.TCloudLoadBalancerType(cvt.PtrToVal(cloud.LoadBalancerType)) {
		case typeslb.InternalLoadBalancerType:
			dbIPList = db.PrivateIPv4Addresses
		case typeslb.OpenLoadBalancerType:
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
	ipv6 := cvt.PtrToVal(cloud.AddressIPv6)
	if len(db.PublicIPv6Addresses) == 0 && len(ipv6) != 0 ||
		len(db.PublicIPv6Addresses) > 0 && db.PublicIPv6Addresses[0] != ipv6 {
		return true
	}

	return isLBExtensionChange(cloud, db)
}

func isLBExtensionChange(cloud typeslb.TCloudClb, db corelb.TCloudLoadBalancer) bool {
	if db.Extension == nil {
		return true
	}

	if cloud.NetworkAttributes != nil {
		if !assert.IsPtrInt64Equal(db.Extension.InternetMaxBandwidthOut,
			cloud.NetworkAttributes.InternetMaxBandwidthOut) {
			return true
		}
		if !assert.IsPtrStringEqual(db.Extension.InternetChargeType, cloud.NetworkAttributes.InternetChargeType) {
			return true
		}
		if !assert.IsPtrStringEqual(db.Extension.BandwidthpkgSubType, cloud.NetworkAttributes.BandwidthpkgSubType) {
			return true
		}
	}

	if !assert.IsPtrStringEqual(db.Extension.SlaType, cloud.SlaType) {
		return true
	}
	if !assert.IsPtrStringEqual(db.Extension.VipIsp, cloud.VipIsp) {
		return true
	}

	if !assert.IsPtrBoolEqual(db.Extension.LoadBalancerPassToTarget, cloud.LoadBalancerPassToTarget) {
		return true
	}
	if !assert.IsPtrStringEqual(db.Extension.IPv6Mode, cloud.IPv6Mode) {
		return true
	}
	if !assert.IsPtrStringEqual(db.Extension.Egress, cloud.Egress) {
		return true
	}
	if !assert.IsPtrBoolEqual(db.Extension.Snat, cloud.Snat) {
		return true
	}
	if !assert.IsPtrBoolEqual(db.Extension.SnatPro, cloud.SnatPro) {
		return true
	}
	if !assert.IsPtrStringEqual(db.Extension.ChargeType, cloud.ChargeType) {
		return true
	}
	// SnatIP列表对比
	if isSnatIPChange(cloud, db) {
		return true
	}

	// 云上AttributeFlags 转map
	attrs := make(map[string]bool, len(cloud.AttributeFlags))
	for _, flag := range cloud.AttributeFlags {
		attrs[cvt.PtrToVal(flag)] = true
	}

	// 逐个判断每种类型
	if attrs[constant.TCLBDeleteProtect] != cvt.PtrToVal(db.Extension.DeleteProtect) {
		return true
	}

	// 跨域1.0 信息
	if !assert.IsPtrStringEqual(db.Extension.TargetRegion, cloud.TargetRegionInfo.Region) {
		return true
	}
	if !assert.IsPtrStringEqual(db.Extension.TargetCloudVpcID, cloud.TargetRegionInfo.VpcId) {
		return true
	}

	return false
}

// 云上SnatIP列表与本地对比
func isSnatIPChange(cloud typeslb.TCloudClb, db corelb.TCloudLoadBalancer) bool {
	dbSnatIps := cvt.PtrToVal(db.Extension.SnatIps)
	if len(dbSnatIps) != len(cloud.SnatIps) {
		return true
	}
	if len(cloud.SnatIps) == 0 {
		// 相等，且都为零
		return false
	}
	// 转为map逐个比较
	cloudSnatMap := cloudSnatSliceToMap(cloud.SnatIps)
	for _, local := range dbSnatIps {
		delete(cloudSnatMap, local.Hash())
	}
	// 数量相等的情况下，应该刚好删除干净。因此非零就是存在不同
	return len(cloudSnatMap) != 0
}

// 将云上的SnatIP转化为map，key为 {SubnetId},{Ip}
func cloudSnatSliceToMap(cloudSlice []*tclb.SnatIp) map[string]struct{} {
	cloudSnatMap := make(map[string]struct{}, len(cloudSlice))
	for _, ip := range cloudSlice {
		cloudSnatMap[hashCloudSnatIP(ip)] = struct{}{}
	}
	return cloudSnatMap
}

// hashCloudSnatIP key为 {SubnetId},{Ip}
func hashCloudSnatIP(ip *tclb.SnatIp) string {
	if ip == nil {
		return ","
	}
	return cvt.PtrToVal(ip.SubnetId) + "," + cvt.PtrToVal(ip.Ip)
}

// SyncLBOption ...
type SyncLBOption struct {
}

// Validate ...
func (o *SyncLBOption) Validate() error {
	return validator.Validate.Struct(o)
}
