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

package other

import (
	"fmt"
	"strconv"
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/hooks"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

const cloudIDPrefix = "_cmdb_"

// BuildCloudIDFromHostID ...
func BuildCloudIDFromHostID(hostID int64) string {
	return fmt.Sprintf("%s%d", cloudIDPrefix, hostID)
}

// GetHostIDFromCloudID ...
func GetHostIDFromCloudID(cloudID string) (int64, error) {
	hostIDStr := strings.TrimPrefix(cloudID, cloudIDPrefix)
	hostID, err := strconv.ParseInt(hostIDStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return hostID, nil
}

// RemoveHostByCCInfo 根据cc的主机信息，删除本地多余的主机
func (cli *client) RemoveHostByCCInfo(kt *kit.Kit, params *DelHostParams) error {
	if params == nil {
		logs.Errorf("params is nil, rid: %s", kt.Rid)
		return fmt.Errorf("params is nil")
	}

	if err := params.Validate(); err != nil {
		logs.Errorf("param is invalid, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(params.DelHostIDs) != 0 {
		return cli.deleteHostByHostID(kt, params.DelHostIDs)
	}

	return cli.removeBizHost(kt, params.BizID, params.CCBizExistHostIDs)
}

func (cli *client) deleteHostByHostID(kt *kit.Kit, hostIDs []int64) error {
	if len(hostIDs) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		deleteReq := &cloud.CvmBatchDeleteReq{Filter: tools.ExpressionAnd(tools.RuleIn("bk_host_id", batch),
			tools.RuleEqual("vendor", enumor.Other))}

		if err := cli.dbCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), deleteReq); err != nil {
			logs.Errorf("[%s] request dataservice to batch delete host failed, err: %v, req: %+v, rid: %s",
				enumor.Other, err, deleteReq, kt.Rid)
			return err
		}
	}

	return nil
}

func (cli *client) removeBizHost(kt *kit.Kit, bizID int64, ccBizExistHostIDs map[int64]struct{}) error {
	dbHosts, err := cli.listHostFromDBByBizID(kt, bizID, []string{"id", "bk_host_id"})
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, bizID: %d, rid: %s", err, bizID, kt.Rid)
		return err
	}

	delHostIDs := make([]int64, 0)
	for _, host := range dbHosts {
		// 需要忽略可能存在的存量没有bk_host_id的数据
		if host.BkHostID == constant.UnBindBkHostID {
			continue
		}
		if _, ok := ccBizExistHostIDs[host.BkHostID]; ok {
			continue
		}
		if host.Vendor != enumor.Other {
			if bizID == constant.UnassignedBiz {
				continue
			}
			// todo 如果在hcm分配到业务下的公有云机器不在cc，那么需要告警
			logs.Errorf("can not find host from cc, hostID: %d, vendor: %v, bizID: %d, rid: %s", host.BkHostID,
				host.Vendor, bizID, kt.Rid)
			continue
		}
		// todo 本期实现的other云厂商数据是从cc同步过来的，这里会和cc保持一致，后续以hcm为准后，调整该实现
		delHostIDs = append(delHostIDs, host.BkHostID)
	}

	if err = cli.deleteHostByHostID(kt, delHostIDs); err != nil {
		logs.Errorf("delete host by host id failed, err: %v, ids: %v, rid: %s", err, delHostIDs, kt.Rid)
		return err
	}

	return nil
}

func (cli *client) listHostFromDBByBizID(kt *kit.Kit, bizID int64, fields []string) ([]cvm.BaseCvm, error) {
	rules := []*filter.AtomRule{tools.RuleEqual("bk_biz_id", bizID)}
	newFilter, err := hooks.GetOtherSyncerListDBHostFilter(kt, rules)
	if err != nil {
		logs.Errorf("adjust other syncer list db host filter failed, err: %v, filter: %v, rid: %s", err, newFilter,
			kt.Rid)
		return nil, err
	}

	return cli.listHostFromDB(kt, fields, newFilter)
}

// listHostFromDB 从db中查询主机
func (cli *client) listHostFromDB(kt *kit.Kit, fields []string, filter *filter.Expression) ([]cvm.BaseCvm, error) {
	req := &core.ListReq{
		Fields: fields,
		Filter: filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "id",
		},
	}
	hosts := make([]cvm.BaseCvm, 0)
	for {
		result, err := cli.dbCli.Global.Cvm.ListCvm(kt, req)
		if err != nil {
			logs.ErrorJson("request dataservice to list cvm failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		hosts = append(hosts, result.Details...)

		if len(result.Details) < int(req.Page.Limit) {
			break
		}
		req.Page.Start += uint32(req.Page.Limit)
	}

	return hosts, nil
}

// Host 对比从cc获取的主机，与本地的主机的差异，进行本地主机的新增、更新、删除操作
func (cli *client) Host(kt *kit.Kit, params *SyncHostParams) error {
	if params == nil {
		logs.Errorf("params is nil, rid: %s", kt.Rid)
		return fmt.Errorf("params is nil")
	}

	if err := params.Validate(); err != nil {
		logs.Errorf("param is invalid, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	hostCache := params.HostCache
	if hostCache == nil {
		hostCache = make(map[int64]cmdb.HostWithCloudID)
	}
	ccHosts := make([]cmdb.HostWithCloudID, 0, len(params.HostIDs))
	needFindHostIDs := make([]int64, 0)
	for _, hostID := range params.HostIDs {
		if host, exist := hostCache[hostID]; exist {
			ccHosts = append(ccHosts, host)
			continue
		}
		needFindHostIDs = append(needFindHostIDs, hostID)
	}
	if len(needFindHostIDs) != 0 {
		// 需要不带业务id去查询主机，防止在前面耗时过程中，主机已被转移到其他业务，这里查不到主机导致把db里的数据误删的问题
		hosts, err := cli.getHostFromCCByHostIDs(kt, needFindHostIDs, cmdb.HostFields)
		if err != nil {
			logs.Errorf("get host from cc by host id failed, err: %v, ids: %v, rid: %s", err, params.HostIDs, kt.Rid)
			return err
		}
		ccHosts = append(ccHosts, hosts...)
	}

	allFields := make([]string, 0) // 当数组为空时，返回所有字段
	dbHosts, err := cli.listHostFromDBByCCHost(kt, ccHosts, allFields)
	if err != nil {
		logs.Errorf("list host from db failed, err: %v, hostIDs: %v, rid: %s", err, params.HostIDs, kt.Rid)
		return err
	}

	if err = cli.doHostDiff(kt, ccHosts, dbHosts); err != nil {
		logs.Errorf("do host diff failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (cli *client) getHostFromCCByHostIDs(kt *kit.Kit, hostIDs []int64, fields []string) ([]cmdb.HostWithCloudID,
	error) {

	hostBizIDMap, err := cli.getHostBizID(kt, hostIDs)
	if err != nil {
		logs.Errorf("get host biz id failed, err: %v, host ids: %v, rid: %s", err, hostIDs, kt.Rid)
		return nil, err
	}

	res := make([]cmdb.HostWithCloudID, 0)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		params := &cmdb.ListHostWithoutBizParams{
			Fields: fields,
			Page:   &cmdb.BasePage{Start: 0, Limit: int64(core.DefaultMaxPageLimit)},
			HostPropertyFilter: &cmdb.QueryFilter{
				Rule: &cmdb.CombinedRule{
					Condition: "AND",
					Rules: []cmdb.Rule{
						&cmdb.AtomRule{Field: "bk_host_id", Operator: "in", Value: batch},
					},
				},
			},
		}
		resp, err := cmdb.CmdbClient().ListHostWithoutBiz(kt, params)
		if err != nil {
			logs.Errorf("get host from cc failed, err: %v, params: %+v, rid: %s", err, params, kt.Rid)
			return nil, err
		}

		for _, host := range resp.Info {
			bizID, ok := hostBizIDMap[host.BkHostID]
			if !ok {
				logs.Errorf("can not find host(%d) biz id, rid: %s", host.BkHostID, kt.Rid)
				continue
			}

			res = append(res, cmdb.HostWithCloudID{
				BizID:   bizID,
				Host:    host,
				CloudID: BuildCloudIDFromHostID(host.BkHostID),
			},
			)
		}
	}

	return res, nil
}

type vendorCloudID struct {
	Vendor  enumor.Vendor `json:"vendor"`
	CloudID string        `json:"cloud_id"`
}

// listHostFromDBByCCHost 先根据host id查询db里的主机，如果查不到，则通过vendor+cloud id查询db里的主机
func (cli *client) listHostFromDBByCCHost(kt *kit.Kit, ccHosts []cmdb.HostWithCloudID, fields []string) ([]cvm.BaseCvm,
	error) {

	hostIDs := make([]int64, 0, len(ccHosts))
	for _, host := range ccHosts {
		hostIDs = append(hostIDs, host.BkHostID)
	}

	res := make([]cvm.BaseCvm, 0)
	existHostIDMap := make(map[int64]struct{})
	for _, batch := range slice.Split(hostIDs, constant.BatchOperationMaxLimit) {
		cond := tools.ExpressionAnd(tools.RuleIn("bk_host_id", batch))
		hosts, err := cli.listHostFromDB(kt, fields, cond)
		if err != nil {
			logs.Errorf("list host from db failed, err: %v, filter: %+v, rid: %s", err, cond, kt.Rid)
			return nil, err
		}
		for _, host := range hosts {
			existHostIDMap[host.BkHostID] = struct{}{}
		}
		res = append(res, hosts...)
	}

	vendorCloudIDs := make(map[enumor.Vendor][]string)
	for _, host := range ccHosts {
		if _, ok := existHostIDMap[host.BkHostID]; ok {
			continue
		}

		if host.BkCloudVendor == "" || host.BkCloudInstID == "" {
			continue
		}

		vendor := cmdb.CmdbHcmVendorMap[host.BkCloudVendor]
		if _, ok := vendorCloudIDs[vendor]; !ok {
			vendorCloudIDs[vendor] = make([]string, 0)
		}
		vendorCloudIDs[vendor] = append(vendorCloudIDs[vendor], host.BkCloudInstID)
	}

	for vendor, cloudIDs := range vendorCloudIDs {
		for _, batch := range slice.Split(cloudIDs, constant.BatchOperationMaxLimit) {
			cond := tools.ExpressionAnd(tools.RuleIn("cloud_id", batch), tools.RuleEqual("vendor", vendor))
			hosts, err := cli.listHostFromDB(kt, fields, cond)
			if err != nil {
				logs.Errorf("list host from db failed, err: %v, filter: %+v, rid: %s", err, cond, kt.Rid)
				return nil, err
			}
			res = append(res, hosts...)
		}
	}

	return res, nil
}

func (cli *client) doHostDiff(kt *kit.Kit, ccHosts []cmdb.HostWithCloudID, dbHosts []cvm.BaseCvm) error {
	if len(ccHosts) == 0 && len(dbHosts) == 0 {
		return nil
	}

	dbVendorCloudIDHostMap := make(map[vendorCloudID]cvm.BaseCvm)
	dbHostIDHostMap := make(map[int64]cvm.BaseCvm)
	for _, host := range dbHosts {
		dbVendorCloudIDHostMap[vendorCloudID{Vendor: host.Vendor, CloudID: host.CloudID}] = host
		dbHostIDHostMap[host.BkHostID] = host
	}

	addHosts := make([]cloud.CvmBatchCreate[cvm.OtherCvmExtension], 0)
	updateHosts := make([]cloud.CvmCommonInfoBatchUpdateData, 0)

	for _, ccHost := range ccHosts {
		if dbHost, ok := dbHostIDHostMap[ccHost.BkHostID]; ok {
			if dbHost.Vendor != enumor.Other {
				continue
			}

			update, isChange := isHostChange(ccHost, dbHost)
			if !isChange {
				continue
			}
			updateHosts = append(updateHosts, update)
			continue
		}

		host, ok := dbVendorCloudIDHostMap[vendorCloudID{Vendor: cmdb.CmdbHcmVendorMap[ccHost.BkCloudVendor],
			CloudID: ccHost.BkCloudInstID}]
		if !ok {
			addHosts = append(addHosts, convToCreate(ccHost, cli.accountID))
			continue
		}

		if host.Vendor != enumor.Other && host.BkHostID == constant.UnBindBkHostID {
			updateHosts = append(updateHosts,
				cloud.CvmCommonInfoBatchUpdateData{ID: host.ID, BkHostID: converter.ValToPtr(ccHost.BkHostID)})
		}
	}

	if len(addHosts) != 0 {
		if err := cli.createHost(kt, addHosts); err != nil {
			logs.Errorf("add host failed, err: %v, req: %+v, rid: %s", err, addHosts, kt.Rid)
			return err
		}
	}

	if len(updateHosts) != 0 {
		if err := cli.updateHost(kt, updateHosts); err != nil {
			logs.Errorf("update host id failed, err: %v, data: %+v, rid: %s", err, updateHosts, kt.Rid)
			return err
		}
	}

	return nil
}

func splitIP(ip string) []string {
	return strings.Split(ip, ",")
}

func isHostChange(ccHost cmdb.HostWithCloudID, dbHost cvm.BaseCvm) (cloud.CvmCommonInfoBatchUpdateData, bool) {
	innerIpv4 := make([]string, 0)
	if len(ccHost.BkHostInnerIP) != 0 {
		innerIpv4 = splitIP(ccHost.BkHostInnerIP)
	}
	innerIpv6 := make([]string, 0)
	if len(ccHost.BkHostInnerIPv6) != 0 {
		innerIpv6 = splitIP(ccHost.BkHostInnerIPv6)
	}
	outerIpv4 := make([]string, 0)
	if len(ccHost.BkHostOuterIP) != 0 {
		outerIpv4 = splitIP(ccHost.BkHostOuterIP)
	}
	outerIpv6 := make([]string, 0)
	if len(ccHost.BkHostOuterIPv6) != 0 {
		outerIpv6 = splitIP(ccHost.BkHostOuterIPv6)
	}

	update := cloud.CvmCommonInfoBatchUpdateData{ID: dbHost.ID}
	isChange := false

	if !assert.IsStringSliceEqual(innerIpv4, dbHost.PrivateIPv4Addresses) {
		update.PrivateIPv4Addresses = converter.ValToPtr(innerIpv4)
		isChange = true
	}

	if !assert.IsStringSliceEqual(innerIpv6, dbHost.PrivateIPv6Addresses) {
		update.PrivateIPv6Addresses = converter.ValToPtr(innerIpv6)
		isChange = true
	}

	if !assert.IsStringSliceEqual(outerIpv4, dbHost.PublicIPv4Addresses) {
		update.PublicIPv4Addresses = converter.ValToPtr(outerIpv4)
		isChange = true
	}

	if !assert.IsStringSliceEqual(outerIpv6, dbHost.PublicIPv6Addresses) {
		update.PublicIPv6Addresses = converter.ValToPtr(outerIpv6)
		isChange = true
	}

	if ccHost.BkHostName != dbHost.Name {
		update.Name = converter.ValToPtr(ccHost.BkHostName)
		isChange = true
	}

	if ccHost.BkHostID != dbHost.BkHostID {
		update.BkHostID = converter.ValToPtr(ccHost.BkHostID)
		isChange = true
	}

	if ccHost.BizID != dbHost.BkBizID {
		update.BkBizID = converter.ValToPtr(ccHost.BizID)
		isChange = true
	}

	if ccHost.BkCloudID != dbHost.BkCloudID {
		update.BkCloudID = converter.ValToPtr(ccHost.BkCloudID)
		isChange = true
	}

	return update, isChange
}

func convToCreate(ccHost cmdb.HostWithCloudID, accountID string) cloud.CvmBatchCreate[cvm.OtherCvmExtension] {
	innerIpv4 := make([]string, 0)
	if len(ccHost.BkHostInnerIP) != 0 {
		innerIpv4 = splitIP(ccHost.BkHostInnerIP)
	}
	innerIpv6 := make([]string, 0)
	if len(ccHost.BkHostInnerIPv6) != 0 {
		innerIpv6 = splitIP(ccHost.BkHostInnerIPv6)
	}
	outerIpv4 := make([]string, 0)
	if len(ccHost.BkHostOuterIP) != 0 {
		outerIpv4 = splitIP(ccHost.BkHostOuterIP)
	}
	outerIpv6 := make([]string, 0)
	if len(ccHost.BkHostOuterIPv6) != 0 {
		outerIpv6 = splitIP(ccHost.BkHostOuterIPv6)
	}

	return cloud.CvmBatchCreate[cvm.OtherCvmExtension]{
		CloudID:              BuildCloudIDFromHostID(ccHost.BkHostID),
		Name:                 ccHost.BkHostName,
		BkBizID:              ccHost.BizID,
		BkHostID:             ccHost.BkHostID,
		BkCloudID:            ccHost.BkCloudID,
		AccountID:            accountID,
		Region:               ccHost.BkCloudRegion,
		PrivateIPv4Addresses: innerIpv4,
		PrivateIPv6Addresses: innerIpv6,
		PublicIPv4Addresses:  outerIpv4,
		PublicIPv6Addresses:  outerIpv6,
		Extension:            &cvm.OtherCvmExtension{},
	}
}

func (cli *client) getHostBizID(kt *kit.Kit, hostIDs []int64) (map[int64]int64, error) {
	if len(hostIDs) == 0 {
		return make(map[int64]int64), nil
	}

	hostBizIDMap := make(map[int64]int64)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		req := &cmdb.HostModuleRelationParams{HostID: batch}
		relationRes, err := cmdb.CmdbClient().FindHostBizRelations(kt, req)
		if err != nil {
			logs.Errorf("fail to find cmdb topo relation, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		for _, relation := range converter.PtrToVal(relationRes) {
			bizID := relation.BizID
			if bizID == cli.ccHostPoolBiz {
				bizID = constant.HostPoolBiz
			}
			hostBizIDMap[relation.HostID] = bizID
		}
	}

	return hostBizIDMap, nil
}

func (cli *client) createHost(kt *kit.Kit, hosts []cloud.CvmBatchCreate[cvm.OtherCvmExtension]) error {
	if len(hosts) == 0 {
		return nil
	}

	for _, batch := range slice.Split(hosts, constant.BatchOperationMaxLimit) {
		createReq := &cloud.CvmBatchCreateReq[cvm.OtherCvmExtension]{Cvms: batch}
		if _, err := cli.dbCli.Other.Cvm.BatchCreateCvm(kt, createReq); err != nil {
			logs.Errorf("create host failed, err: %v, req: %+v, rid: %s", err, createReq, kt.Rid)
			return err
		}
	}

	return nil
}

func (cli *client) updateHost(kt *kit.Kit, updateData []cloud.CvmCommonInfoBatchUpdateData) error {
	if len(updateData) == 0 {
		return nil
	}

	for _, batch := range slice.Split(updateData, constant.BatchOperationMaxLimit) {
		update := &cloud.CvmCommonInfoBatchUpdateReq{Cvms: batch}
		if err := cli.dbCli.Global.Cvm.BatchUpdateCvmCommonInfo(kt, update); err != nil {
			logs.Errorf("update host common info failed, err: %v, req: %+v, rid: %s", err, update, kt.Rid)
			return err
		}
	}

	return nil
}
