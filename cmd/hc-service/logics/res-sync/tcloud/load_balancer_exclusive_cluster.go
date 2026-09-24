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

package tcloud

import (
	"hcm/cmd/hc-service/logics/res-sync/common"
	adcore "hcm/pkg/adaptor/types/core"
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
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"
)

// exclusiveClusterNetwork、exclusiveClusterTypes 全量同步产品范围固定为公网 TGW/STGW，非云 API 限制。
var exclusiveClusterTypes = []enumor.ClusterType{enumor.TGWClusterType, enumor.STGWClusterType}

const exclusiveClusterNetwork = enumor.PublicClusterNetwork

// SyncExclusiveClusterOption ...
type SyncExclusiveClusterOption struct{}

// Validate ...
func (opt SyncExclusiveClusterOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ExclusiveCluster 同步独占集群，范围固定为公网 TGW/STGW
func (cli *client) ExclusiveCluster(kt *kit.Kit, params *SyncBaseParams, opt *SyncExclusiveClusterOption) (
	*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	clusterFromCloud, err := cli.listExclusiveClusterFromCloud(kt, params)
	if err != nil {
		return nil, err
	}

	clusterFromDB, err := cli.listExclusiveClusterFromDB(kt, params)
	if err != nil {
		return nil, err
	}

	if len(clusterFromCloud) == 0 && len(clusterFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typeslb.TCloudExclusiveCluster,
		corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]](clusterFromCloud, clusterFromDB,
		isExclusiveClusterChange)

	if err = cli.deleteExclusiveCluster(kt, params.AccountID, params.Region, delCloudIDs); err != nil {
		return nil, err
	}

	if err = cli.createExclusiveCluster(kt, params.AccountID, params.Region, addSlice); err != nil {
		return nil, err
	}

	if err = cli.updateExclusiveCluster(kt, params.AccountID, params.Region, updateMap); err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

// listExclusiveClusterFromCloud 按cloud_id拉取账号+地域下公网TGW/STGW独占集群
func (cli *client) listExclusiveClusterFromCloud(kt *kit.Kit, params *SyncBaseParams) (
	[]typeslb.TCloudExclusiveCluster, error) {

	if err := params.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &typeslb.TCloudExclusiveClusterListOption{
		Region:       params.Region,
		Network:      exclusiveClusterNetwork,
		ClusterTypes: exclusiveClusterTypes,
		Page: &adcore.TCloudPage{
			Offset: 0,
			Limit:  constant.TCLBDescribeMax,
		},
	}
	result := make([]typeslb.TCloudExclusiveCluster, 0, len(params.CloudIDs))

	for _, cloudIDs := range slice.Split(params.CloudIDs, constant.TCLBDescribeMax) {
		opt.CloudIDs = cloudIDs
		batch, err := cli.cloudCli.ListExclusiveClusters(kt, opt)
		if err != nil {
			logs.Errorf("[%s] list exclusive cluster from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
				enumor.TCloud, err, params.AccountID, opt, kt.Rid)
			return nil, err
		}
		result = append(result, batch...)
	}

	return result, nil
}

// listExclusiveClusterFromDB 按account_id+region+vendor+cloud_id列出DB独占集群，Raw extension反序列化为强类型
func (cli *client) listExclusiveClusterFromDB(kt *kit.Kit, params *SyncBaseParams) (
	[]corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension], error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", params.AccountID),
			tools.RuleEqual("region", params.Region),
			tools.RuleEqual("vendor", enumor.TCloud),
			tools.RuleIn("cloud_id", params.CloudIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Global.ListExclusiveCluster(kt, req)
	if err != nil {
		logs.Errorf("[%s] list exclusive cluster from db failed, err: %v, account: %s, req: %v, rid: %s",
			enumor.TCloud, err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	list := make([]corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension], 0, len(result.Details))
	for _, raw := range result.Details {
		ext := new(corelb.TCloudExclusiveClusterExtension)
		if len(raw.Extension) != 0 {
			if err := json.Unmarshal(raw.Extension, ext); err != nil {
				logs.Errorf("[%s] unmarshal exclusive cluster extension failed, err: %v, id: %s, rid: %s",
					enumor.TCloud, err, raw.ID, kt.Rid)
				return nil, err
			}
		}
		list = append(list, corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]{
			BaseExclusiveCluster: raw.BaseExclusiveCluster,
			Extension:            ext,
		})
	}

	return list, nil
}

// createExclusiveCluster 调用data-service新增独占集群，请求体不含bk_biz_id，由data-service写-1
func (cli *client) createExclusiveCluster(kt *kit.Kit, accountID, region string,
	addSlice []typeslb.TCloudExclusiveCluster) error {

	if len(addSlice) == 0 {
		return nil
	}

	createReq := new(protocloud.TCloudExclusiveClusterBatchCreateReq)
	for _, cloud := range addSlice {
		createReq.Clusters = append(createReq.Clusters,
			protocloud.ExclusiveClusterCreate[corelb.TCloudExclusiveClusterExtension]{
				CloudID:          cloud.GetCloudID(),
				Name:             cvt.PtrToVal(cloud.ClusterName),
				AccountID:        accountID,
				Region:           region,
				Zone:             cvt.PtrToVal(cloud.Zone),
				ClusterType:      enumor.ClusterType(cvt.PtrToVal(cloud.ClusterType)),
				ClusterTag:       cvt.PtrToVal(cloud.ClusterTag),
				Network:          enumor.ClusterNetwork(cvt.PtrToVal(cloud.Network)),
				Isp:              enumor.ClusterIsp(cvt.PtrToVal(cloud.Isp)),
				Egress:           cvt.PtrToVal(cloud.Egress),
				IPVersion:        cvt.PtrToVal(cloud.IPVersion),
				MaxConn:          cloud.MaxConn,
				ClbResourceCount: deriveClbResourceCount(cloud),
				Extension:        convertTCloudExclusiveClusterExtension(cloud),
			})
	}

	if _, err := cli.dbCli.TCloud.BatchCreateExclusiveCluster(kt, createReq); err != nil {
		logs.Errorf("[%s] request dataservice to create tcloud exclusive cluster failed, err: %v, "+
			"account: %s, region: %s, count: %d, rid: %s", enumor.TCloud, err, accountID, region,
			len(addSlice), kt.Rid)
		return err
	}

	logs.Infof("[%s] sync exclusive cluster to create success, account: %s, region: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, region, len(addSlice), kt.Rid)

	return nil
}

// updateExclusiveCluster 调用data-service更新独占集群云属性，请求体不含bk_biz_id
func (cli *client) updateExclusiveCluster(kt *kit.Kit, accountID, region string,
	updateMap map[string]typeslb.TCloudExclusiveCluster) error {

	if len(updateMap) == 0 {
		return nil
	}

	updateReq := new(protocloud.TCloudExclusiveClusterBatchUpdateReq)
	for id, cloud := range updateMap {
		updateReq.Clusters = append(updateReq.Clusters,
			protocloud.ExclusiveClusterUpdate[corelb.TCloudExclusiveClusterExtension]{
				ID:               id,
				Name:             cvt.PtrToVal(cloud.ClusterName),
				Zone:             cvt.PtrToVal(cloud.Zone),
				ClusterType:      enumor.ClusterType(cvt.PtrToVal(cloud.ClusterType)),
				ClusterTag:       cvt.PtrToVal(cloud.ClusterTag),
				Network:          enumor.ClusterNetwork(cvt.PtrToVal(cloud.Network)),
				Isp:              enumor.ClusterIsp(cvt.PtrToVal(cloud.Isp)),
				Egress:           cvt.PtrToVal(cloud.Egress),
				IPVersion:        cvt.PtrToVal(cloud.IPVersion),
				MaxConn:          cloud.MaxConn,
				ClbResourceCount: deriveClbResourceCount(cloud),
				Extension:        convertTCloudExclusiveClusterExtension(cloud),
			})
	}

	if err := cli.dbCli.TCloud.BatchUpdateExclusiveCluster(kt, updateReq); err != nil {
		logs.Errorf("[%s] request dataservice to update tcloud exclusive cluster failed, err: %v, "+
			"account: %s, region: %s, count: %d, rid: %s", enumor.TCloud, err, accountID, region,
			len(updateMap), kt.Rid)
		return err
	}

	logs.Infof("[%s] sync exclusive cluster to update success, account: %s, region: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, region, len(updateMap), kt.Rid)

	return nil
}

// deleteExclusiveCluster 调用data-service删除云上已不存在的独占集群
func (cli *client) deleteExclusiveCluster(kt *kit.Kit, accountID, region string, delCloudIDs []string) error {
	if len(delCloudIDs) == 0 {
		return nil
	}

	deleteReq := &protocloud.ExclusiveClusterBatchDeleteReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("cloud_id", delCloudIDs),
			tools.RuleEqual("account_id", accountID),
			tools.RuleEqual("region", region),
			tools.RuleEqual("vendor", enumor.TCloud),
		),
	}
	if err := cli.dbCli.Global.BatchDeleteExclusiveCluster(kt, deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete exclusive cluster failed, err: %v, "+
			"account: %s, region: %s, count: %d, rid: %s", enumor.TCloud, err, accountID, region,
			len(delCloudIDs), kt.Rid)
		return err
	}

	logs.Infof("[%s] sync to delete exclusive cluster success, account: %s, region: %s, count: %d, rid: %s",
		enumor.TCloud, accountID, region, len(delCloudIDs), kt.Rid)

	return nil
}

// RemoveExclusiveClusterDeleteFromCloud 清理云上已删除的独占集群，DB过滤范围与全量拉取保持同集合（公网TGW/STGW），
// 避免误删将来可能存在的Private/VPCGW集群
func (cli *client) RemoveExclusiveClusterDeleteFromCloud(kt *kit.Kit, params *SyncRemovedParams,
	allCloudIDMap map[string]struct{}) error {

	if err := params.Validate(); err != nil {
		return err
	}

	rules := []*filter.AtomRule{
		tools.RuleEqual("account_id", params.AccountID),
		tools.RuleEqual("region", params.Region),
		tools.RuleEqual("vendor", enumor.TCloud),
		tools.RuleEqual("network", exclusiveClusterNetwork),
		tools.RuleIn("cluster_type", exclusiveClusterTypes),
	}
	if len(params.CloudIDs) > 0 {
		rules = append(rules, tools.RuleIn("cloud_id", params.CloudIDs))
	}
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(rules...),
		Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
	}

	var delCloudIDs []string
	for {
		resultFromDB, err := cli.dbCli.Global.ListExclusiveCluster(kt, req)
		if err != nil {
			logs.Errorf("[%s] request dataservice to list exclusive cluster failed, req: %v, err: %v, rid: %s",
				enumor.TCloud, req, err, kt.Rid)
			return err
		}

		for _, detail := range resultFromDB.Details {
			if _, ok := allCloudIDMap[detail.CloudID]; !ok {
				delCloudIDs = append(delCloudIDs, detail.CloudID)
			}
		}

		if uint(len(resultFromDB.Details)) < core.DefaultMaxPageLimit {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	if len(delCloudIDs) == 0 {
		return nil
	}

	logs.Infof("[%s] will remove %d deleted exclusive cluster from cloud, account: %s, region: %s, rid: %s",
		enumor.TCloud, len(delCloudIDs), params.AccountID, params.Region, kt.Rid)

	return cli.deleteExclusiveCluster(kt, params.AccountID, params.Region, delCloudIDs)
}

// deriveClbResourceCount 计算集群内已用CLB资源数：resourceCount - idleResourceCount，负值钳为0
func deriveClbResourceCount(cloud typeslb.TCloudExclusiveCluster) int64 {
	count := cvt.PtrToVal(cloud.ResourceCount) - cvt.PtrToVal(cloud.IdleResourceCount)
	if count < 0 {
		return 0
	}
	return count
}

// convertTCloudExclusiveClusterExtension 转换独占集群云属性到扩展字段，Tag不落库
func convertTCloudExclusiveClusterExtension(
	cloud typeslb.TCloudExclusiveCluster) *corelb.TCloudExclusiveClusterExtension {

	ext := &corelb.TCloudExclusiveClusterExtension{
		MaxInFlow:                cvt.PtrToVal(cloud.MaxInFlow),
		MaxOutFlow:               cvt.PtrToVal(cloud.MaxOutFlow),
		MaxInPkg:                 cvt.PtrToVal(cloud.MaxInPkg),
		MaxOutPkg:                cvt.PtrToVal(cloud.MaxOutPkg),
		MaxNewConn:               cvt.PtrToVal(cloud.MaxNewConn),
		HTTPMaxNewConn:           cvt.PtrToVal(cloud.HTTPMaxNewConn),
		HTTPSMaxNewConn:          cvt.PtrToVal(cloud.HTTPSMaxNewConn),
		HTTPQps:                  cvt.PtrToVal(cloud.HTTPQps),
		HTTPSQps:                 cvt.PtrToVal(cloud.HTTPSQps),
		LoadBalanceDirectorCount: cvt.PtrToVal(cloud.LoadBalanceDirectorCount),
		ClustersVersion:          cvt.PtrToVal(cloud.ClustersVersion),
		DisasterRecoveryType:     cvt.PtrToVal(cloud.DisasterRecoveryType),
	}
	if cloud.ClustersZone != nil {
		ext.ClustersZone = corelb.TCloudExclusiveClusterZone{
			MasterZone: cvt.PtrToSlice(cloud.ClustersZone.MasterZone),
			SlaveZone:  cvt.PtrToSlice(cloud.ClustersZone.SlaveZone),
		}
	}
	return ext
}

func isExclusiveClusterChange(cloud typeslb.TCloudExclusiveCluster,
	db corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]) bool {

	if db.Name != cvt.PtrToVal(cloud.ClusterName) {
		return true
	}
	if db.Zone != cvt.PtrToVal(cloud.Zone) {
		return true
	}
	if db.ClusterType != enumor.ClusterType(cvt.PtrToVal(cloud.ClusterType)) {
		return true
	}
	if db.ClusterTag != cvt.PtrToVal(cloud.ClusterTag) {
		return true
	}
	if db.Network != enumor.ClusterNetwork(cvt.PtrToVal(cloud.Network)) {
		return true
	}
	if db.Isp != enumor.ClusterIsp(cvt.PtrToVal(cloud.Isp)) {
		return true
	}
	if db.Egress != cvt.PtrToVal(cloud.Egress) {
		return true
	}
	if db.IPVersion != cvt.PtrToVal(cloud.IPVersion) {
		return true
	}
	if !assert.IsPtrInt64Equal(db.MaxConn, cloud.MaxConn) {
		return true
	}
	if db.ClbResourceCount != deriveClbResourceCount(cloud) {
		return true
	}

	return isExclusiveClusterExtensionChange(cloud, db)
}

func isExclusiveClusterExtensionChange(cloud typeslb.TCloudExclusiveCluster,
	db corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]) bool {

	if db.Extension == nil {
		return true
	}

	if db.Extension.MaxInFlow != cvt.PtrToVal(cloud.MaxInFlow) {
		return true
	}
	if db.Extension.MaxOutFlow != cvt.PtrToVal(cloud.MaxOutFlow) {
		return true
	}
	if db.Extension.MaxInPkg != cvt.PtrToVal(cloud.MaxInPkg) {
		return true
	}
	if db.Extension.MaxOutPkg != cvt.PtrToVal(cloud.MaxOutPkg) {
		return true
	}
	if db.Extension.MaxNewConn != cvt.PtrToVal(cloud.MaxNewConn) {
		return true
	}
	if db.Extension.HTTPMaxNewConn != cvt.PtrToVal(cloud.HTTPMaxNewConn) {
		return true
	}
	if db.Extension.HTTPSMaxNewConn != cvt.PtrToVal(cloud.HTTPSMaxNewConn) {
		return true
	}
	if db.Extension.HTTPQps != cvt.PtrToVal(cloud.HTTPQps) {
		return true
	}
	if db.Extension.HTTPSQps != cvt.PtrToVal(cloud.HTTPSQps) {
		return true
	}
	if db.Extension.LoadBalanceDirectorCount != cvt.PtrToVal(cloud.LoadBalanceDirectorCount) {
		return true
	}
	if db.Extension.ClustersVersion != cvt.PtrToVal(cloud.ClustersVersion) {
		return true
	}
	if db.Extension.DisasterRecoveryType != cvt.PtrToVal(cloud.DisasterRecoveryType) {
		return true
	}

	var cloudMasterZone, cloudSlaveZone []*string
	if cloud.ClustersZone != nil {
		cloudMasterZone = cloud.ClustersZone.MasterZone
		cloudSlaveZone = cloud.ClustersZone.SlaveZone
	}
	if !assert.IsPtrStringSliceEqual(cvt.SliceToPtr(db.Extension.ClustersZone.MasterZone), cloudMasterZone) {
		return true
	}
	if !assert.IsPtrStringSliceEqual(cvt.SliceToPtr(db.Extension.ClustersZone.SlaveZone), cloudSlaveZone) {
		return true
	}

	return false
}
